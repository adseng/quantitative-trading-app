package main

import (
	"flag"
	"fmt"
	"os"

	"quantitative-trading-app/internal/backtest"
	"quantitative-trading-app/internal/config"
	"quantitative-trading-app/internal/datafile"
	"quantitative-trading-app/internal/strategy"
)

func main() {
	_ = config.Load()

	defaults := backtest.DefaultEMARunRequest()

	dataPath := flag.String("data", datafile.DefaultKlinePath("15m", 10000), "path to the local kline data file")
	resultPath := flag.String("out", "", "path to write backtest result file")
	initialBalance := flag.Float64("capital", defaults.InitialBalance, "initial balance in USDT")
	positionSize := flag.Float64("position", defaults.PositionSizeUSDT, "order value per trade in USDT")
	fastPeriod := flag.Int("fast", defaults.Params.FastPeriod, "fast EMA period")
	slowPeriod := flag.Int("slow", defaults.Params.SlowPeriod, "slow EMA period")
	breakoutLookback := flag.Int("breakout-lookback", defaults.Params.BreakoutLookback, "breakout lookback bars")
	pullbackLookahead := flag.Int("pullback-lookahead", defaults.Params.PullbackLookahead, "pullback confirmation lookahead bars")
	pullbackTolerancePercent := flag.Float64("pullback-tolerance", defaults.Params.PullbackTolerancePercent, "pullback tolerance percent")
	atrPeriod := flag.Int("atr-period", defaults.Params.ATRPeriod, "ATR period")
	stopATRMultiplier := flag.Float64("stop-atr-multiplier", defaults.Params.StopATRMultiplier, "ATR multiplier for stop loss")
	cooldownBars := flag.Int("cooldown", defaults.Params.CooldownBars, "cooldown bars after each signal")
	riskRewardRatio := flag.Float64("rr", defaults.Params.RiskRewardRatio, "risk reward ratio")
	flag.Parse()

	req := backtest.RunEMARequest{
		DataPath:         *dataPath,
		StrategyName:     strategy.EMATrendPullbackName,
		InitialBalance:   *initialBalance,
		PositionSizeUSDT: *positionSize,
		ResultPath:       *resultPath,
		Params: strategy.EMATrendPullbackParams{
			FastPeriod:               *fastPeriod,
			SlowPeriod:               *slowPeriod,
			BreakoutLookback:         *breakoutLookback,
			PullbackLookahead:        *pullbackLookahead,
			PullbackTolerancePercent: *pullbackTolerancePercent,
			ATRPeriod:                *atrPeriod,
			StopATRMultiplier:        *stopATRMultiplier,
			CooldownBars:             *cooldownBars,
			RiskRewardRatio:          *riskRewardRatio,
		},
	}.Normalize()

	if req.ResultPath == "" {
		req.ResultPath = backtest.DefaultEMAResultPath(req.StrategyName)
	}

	report, err := backtest.RunEMA(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "backtest error: %v\n", err)
		os.Exit(1)
	}
	if err := backtest.SaveEMAReport(req.ResultPath, report); err != nil {
		fmt.Fprintf(os.Stderr, "save result error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Strategy: %s\n", report.StrategyName)
	fmt.Printf("Signals: %d\n", report.Summary.TotalSignals)
	fmt.Printf("Trades: %d\n", report.Summary.ExecutedTrades)
	fmt.Printf("Wins/Losses: %d/%d\n", report.Summary.Wins, report.Summary.Losses)
	fmt.Printf("Win rate: %.2f%%\n", report.Summary.WinRate*100)
	fmt.Printf("PnL: %.2f USDT\n", report.Summary.TotalPnL)
	fmt.Printf("Final balance: %.2f USDT\n", report.Summary.FinalBalance)
	fmt.Printf("Saved to: %s\n", req.ResultPath)
}
