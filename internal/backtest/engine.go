package backtest

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"quantitative-trading-app/internal/datafile"
	"quantitative-trading-app/internal/market"
	"quantitative-trading-app/internal/strategy"
)

func Run(req RunRequest) (*Report, error) {
	req = req.Normalize()
	klines, err := datafile.LoadKlines(req.DataPath)
	if err != nil {
		return nil, err
	}
	if len(klines) == 0 {
		return nil, fmt.Errorf("no klines loaded from %s", req.DataPath)
	}
	return RunOnKlines(req, marketPointersToValues(klines))
}

func RunOnKlines(req RunRequest, klines []market.KLine) (*Report, error) {
	req = req.Normalize()
	if len(klines) < 3 {
		return nil, fmt.Errorf("not enough klines for backtest")
	}

	var signals []strategy.Signal
	switch req.StrategyName {
	case "", strategy.BoxPullbackName:
		req.StrategyName = strategy.BoxPullbackName
		signals = strategy.EvaluateBoxPullback(klines, req.Params)
	default:
		return nil, fmt.Errorf("unsupported strategy: %s", req.StrategyName)
	}

	report := NewReport(req, klines, signals)
	report.Trades, report.Summary = simulateTrades(klines, signals, req.InitialBalance, req.PositionSizeUSDT)
	return &report, nil
}

func DefaultResultPath(strategyName string) string {
	if strategyName == "" {
		strategyName = strategy.BoxPullbackName
	}
	return filepath.ToSlash(filepath.Join(datafile.DefaultDir, fmt.Sprintf("test-%s.txt", strategyName)))
}

func SaveReport(path string, report *Report) error {
	if path == "" {
		path = report.ResultPath
	}
	if path == "" {
		path = DefaultResultPath(report.StrategyName)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}

func simulateTrades(klines []market.KLine, signals []strategy.Signal, initialBalance, positionSize float64) ([]Trade, Summary) {
	trades := make([]Trade, 0)
	balance := initialBalance
	peakBalance := initialBalance
	maxDrawdown := 0.0
	signalCursor := 0
	skippedSignals := 0
	nextTradeID := 1

	type openTrade struct {
		trade  Trade
		signal strategy.Signal
	}

	activeTrades := make([]*openTrade, 0)

	for barIndex := 0; barIndex < len(klines); barIndex++ {
		bar := klines[barIndex]

		if len(activeTrades) > 0 {
			nextActive := activeTrades[:0]
			for _, active := range activeTrades {
				exitPrice, exitReason, shouldClose := shouldCloseTrade(active.signal.Direction, bar, active.trade.StopLoss, active.trade.TakeProfit)
				if !shouldClose {
					nextActive = append(nextActive, active)
					continue
				}
				active.trade.ExitIndex = barIndex
				active.trade.ExitTime = bar.OpenTime
				active.trade.ExitPrice = exitPrice
				active.trade.ExitReason = exitReason
				active.trade.HoldBars = barIndex - active.trade.EntryIndex
				active.trade.PnL = calculatePnL(active.signal.Direction, active.trade.Quantity, active.trade.EntryPrice, exitPrice)
				if active.trade.OrderValueUSDT > 0 {
					active.trade.PnLPercent = (active.trade.PnL / active.trade.OrderValueUSDT) * 100
				}
				balance += active.trade.PnL
				active.trade.BalanceAfter = balance
				if balance > peakBalance {
					peakBalance = balance
				}
				drawdown := 0.0
				if peakBalance > 0 {
					drawdown = (peakBalance - balance) / peakBalance
				}
				if drawdown > maxDrawdown {
					maxDrawdown = drawdown
				}
				trades = append(trades, active.trade)
			}
			activeTrades = nextActive
		}

		for signalCursor < len(signals) && signals[signalCursor].EntryIndex < barIndex {
			skippedSignals++
			signalCursor++
		}

		for signalCursor < len(signals) && signals[signalCursor].EntryIndex == barIndex {
			signal := signals[signalCursor]
			entryPrice := bar.Open
			if entryPrice <= 0 || positionSize <= 0 {
				skippedSignals++
				signalCursor++
				continue
			}

			qty := positionSize / entryPrice
			activeTrades = append(activeTrades, &openTrade{
				signal: signal,
				trade: Trade{
					ID:             nextTradeID,
					StrategyName:   signal.StrategyName,
					Direction:      signal.Direction,
					SignalIndex:    signal.TriggerIndex,
					EntryIndex:     signal.EntryIndex,
					SignalTime:     signal.TriggerTime,
					EntryTime:      signal.EntryTime,
					EntryPrice:     entryPrice,
					StopLoss:       signal.StopLoss,
					TakeProfit:     signal.TakeProfit,
					Quantity:       qty,
					OrderValueUSDT: positionSize,
				},
			})
			nextTradeID++
			signalCursor++
		}
	}

	if len(activeTrades) > 0 {
		lastBar := klines[len(klines)-1]
		for _, active := range activeTrades {
			active.trade.ExitIndex = len(klines) - 1
			active.trade.ExitTime = lastBar.OpenTime
			active.trade.ExitPrice = lastBar.Close
			active.trade.ExitReason = "end_of_data"
			active.trade.HoldBars = active.trade.ExitIndex - active.trade.EntryIndex
			active.trade.PnL = calculatePnL(active.signal.Direction, active.trade.Quantity, active.trade.EntryPrice, active.trade.ExitPrice)
			if active.trade.OrderValueUSDT > 0 {
				active.trade.PnLPercent = (active.trade.PnL / active.trade.OrderValueUSDT) * 100
			}
			balance += active.trade.PnL
			active.trade.BalanceAfter = balance
			if balance > peakBalance {
				peakBalance = balance
			}
			drawdown := 0.0
			if peakBalance > 0 {
				drawdown = (peakBalance - balance) / peakBalance
			}
			if drawdown > maxDrawdown {
				maxDrawdown = drawdown
			}
			trades = append(trades, active.trade)
		}
	}

	wins := 0
	losses := 0
	totalPnL := 0.0
	for _, trade := range trades {
		totalPnL += trade.PnL
		if trade.PnL >= 0 {
			wins++
		} else {
			losses++
		}
	}

	winRate := 0.0
	if len(trades) > 0 {
		winRate = float64(wins) / float64(len(trades))
	}

	roi := 0.0
	if initialBalance > 0 {
		roi = totalPnL / initialBalance
	}

	return trades, Summary{
		TotalSignals:   len(signals),
		ExecutedTrades: len(trades),
		Wins:           wins,
		Losses:         losses,
		SkippedSignals: skippedSignals,
		WinRate:        winRate,
		FinalBalance:   balance,
		TotalPnL:       totalPnL,
		ROI:            roi,
		MaxDrawdown:    maxDrawdown,
	}
}

func shouldCloseTrade(direction strategy.Direction, bar market.KLine, stopLoss, takeProfit float64) (float64, string, bool) {
	switch direction {
	case strategy.DirectionLong:
		hitStop := bar.Low <= stopLoss
		hitTP := bar.High >= takeProfit
		switch {
		case hitStop && hitTP:
			return stopLoss, "stop_loss_priority_same_bar", true
		case hitStop:
			return stopLoss, "stop_loss", true
		case hitTP:
			return takeProfit, "take_profit", true
		}
	case strategy.DirectionShort:
		hitStop := bar.High >= stopLoss
		hitTP := bar.Low <= takeProfit
		switch {
		case hitStop && hitTP:
			return stopLoss, "stop_loss_priority_same_bar", true
		case hitStop:
			return stopLoss, "stop_loss", true
		case hitTP:
			return takeProfit, "take_profit", true
		}
	}
	return 0, "", false
}

func calculatePnL(direction strategy.Direction, quantity, entryPrice, exitPrice float64) float64 {
	switch direction {
	case strategy.DirectionLong:
		return quantity * (exitPrice - entryPrice)
	case strategy.DirectionShort:
		return quantity * (entryPrice - exitPrice)
	default:
		return 0
	}
}

func marketPointersToValues(items []*market.KLine) []market.KLine {
	out := make([]market.KLine, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, *item)
	}
	return out
}

func round(value float64, digits int) float64 {
	factor := math.Pow(10, float64(digits))
	return math.Round(value*factor) / factor
}
