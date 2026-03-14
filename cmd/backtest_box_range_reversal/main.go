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

	defaults := backtest.DefaultBoxRangeRunRequest()

	dataPath := flag.String("data", datafile.DefaultKlinePath("15m", 10000), "path to the local kline data file")
	resultPath := flag.String("out", "", "path to write backtest result file")
	initialBalance := flag.Float64("capital", defaults.InitialBalance, "initial balance in USDT")
	positionSize := flag.Float64("position", defaults.PositionSizeUSDT, "order value per trade in USDT")
	impulseLookback := flag.Int("impulse-lookback", defaults.Params.ImpulseLookback, "lookback bars for impulse detection")
	consolidationLookback := flag.Int("consolidation-lookback", defaults.Params.ConsolidationLookback, "lookback bars for consolidation box")
	atrPeriod := flag.Int("atr-period", defaults.Params.ATRPeriod, "ATR period")
	minImpulsePercent := flag.Float64("min-impulse", defaults.Params.MinImpulsePercent, "minimum impulse move percent")
	minImpulseATRRatio := flag.Float64("min-impulse-atr-ratio", defaults.Params.MinImpulseATRRatio, "minimum impulse strength ratio against consolidation ATR")
	minBoxWidthPercent := flag.Float64("min-box-width", defaults.Params.MinBoxWidthPercent, "minimum box width percent")
	maxBoxWidthPercent := flag.Float64("max-box-width", defaults.Params.MaxBoxWidthPercent, "maximum box width percent")
	consolidationVolumeRatio := flag.Float64("consolidation-volume-ratio", defaults.Params.ConsolidationVolumeRatio, "maximum consolidation volume ratio")
	consolidationATRRatio := flag.Float64("consolidation-atr-ratio", defaults.Params.ConsolidationATRRatio, "maximum consolidation ATR ratio")
	minBoundaryTouches := flag.Int("min-boundary-touches", defaults.Params.MinBoundaryTouches, "minimum touches on both box edges")
	edgeTolerancePercent := flag.Float64("edge-tolerance", defaults.Params.EdgeTolerancePercent, "edge touch tolerance percent")
	minRejectWickBodyRatio := flag.Float64("reject-wick-body-ratio", defaults.Params.MinRejectWickBodyRatio, "minimum wick body ratio for rejection bar")
	stopATRMultiplier := flag.Float64("stop-atr-multiplier", defaults.Params.StopATRMultiplier, "ATR multiplier for stop loss")
	cooldownBars := flag.Int("cooldown", defaults.Params.CooldownBars, "cooldown bars after each signal")
	takeProfitFactor := flag.Float64("take-profit-factor", defaults.Params.TakeProfitFactor, "target factor between 0(midline) and 1(opposite edge)")
	flag.Parse()

	req := backtest.RunBoxRangeRequest{
		DataPath:         *dataPath,
		StrategyName:     strategy.BoxRangeReversalName,
		InitialBalance:   *initialBalance,
		PositionSizeUSDT: *positionSize,
		ResultPath:       *resultPath,
		Params: strategy.BoxRangeReversalParams{
			ImpulseLookback:          *impulseLookback,
			ConsolidationLookback:    *consolidationLookback,
			ATRPeriod:                *atrPeriod,
			MinImpulsePercent:        *minImpulsePercent,
			MinImpulseATRRatio:       *minImpulseATRRatio,
			MinBoxWidthPercent:       *minBoxWidthPercent,
			MaxBoxWidthPercent:       *maxBoxWidthPercent,
			ConsolidationVolumeRatio: *consolidationVolumeRatio,
			ConsolidationATRRatio:    *consolidationATRRatio,
			MinBoundaryTouches:       *minBoundaryTouches,
			EdgeTolerancePercent:     *edgeTolerancePercent,
			MinRejectWickBodyRatio:   *minRejectWickBodyRatio,
			StopATRMultiplier:        *stopATRMultiplier,
			CooldownBars:             *cooldownBars,
			TakeProfitFactor:         *takeProfitFactor,
		},
	}.Normalize()

	if req.ResultPath == "" {
		req.ResultPath = backtest.DefaultBoxRangeResultPath(req.StrategyName)
	}

	report, err := backtest.RunBoxRange(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "backtest error: %v\n", err)
		os.Exit(1)
	}
	if err := backtest.SaveBoxRangeReport(req.ResultPath, report); err != nil {
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
