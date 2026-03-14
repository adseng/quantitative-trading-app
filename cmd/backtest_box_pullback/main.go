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

	defaults := backtest.DefaultRunRequest()

	dataPath := flag.String("data", datafile.DefaultKlinePath("15m", 10000), "path to the local kline data file")
	resultPath := flag.String("out", "", "path to write backtest result file")
	initialBalance := flag.Float64("capital", defaults.InitialBalance, "initial balance in USDT")
	positionSize := flag.Float64("position", defaults.PositionSizeUSDT, "order value per trade in USDT")
	lookaheadN := flag.Int("lookahead", defaults.Params.LookaheadN, "max number of bars to wait for pullback confirmation")
	minK1BodyPercent := flag.Float64("min-k1-body", defaults.Params.MinK1BodyPercent, "minimum K1 body percent, e.g. 0.003 = 0.3%")
	k1StrengthLookback := flag.Int("k1-strength-lookback", defaults.Params.K1StrengthLookback, "lookback bars for K1 relative body strength")
	minK1BodyToAvgRatio := flag.Float64("k1-body-avg-ratio", defaults.Params.MinK1BodyToAvgRatio, "minimum ratio between K1 body and recent average body")
	trendMAPeriod := flag.Int("trend-ma", defaults.Params.TrendMAPeriod, "trend moving average period")
	minBoxRangePercent := flag.Float64("min-box-range", defaults.Params.MinBoxRangePercent, "minimum box range percent")
	maxBoxRangePercent := flag.Float64("max-box-range", defaults.Params.MaxBoxRangePercent, "maximum box range percent")
	touchTolerancePercent := flag.Float64("touch-tolerance", defaults.Params.TouchTolerancePercent, "touch tolerance percent")
	minConfirmWickBodyRatio := flag.Float64("confirm-wick-ratio", defaults.Params.MinConfirmWickBodyRatio, "minimum confirmation wick/body ratio")
	cooldownBars := flag.Int("cooldown", defaults.Params.CooldownBars, "cooldown bars after each signal")
	riskRewardRatio := flag.Float64("rr", defaults.Params.RiskRewardRatio, "risk reward ratio")
	flag.Parse()

	req := backtest.RunRequest{
		DataPath:         *dataPath,
		StrategyName:     strategy.BoxPullbackName,
		InitialBalance:   *initialBalance,
		PositionSizeUSDT: *positionSize,
		ResultPath:       *resultPath,
		Params: strategy.BoxPullbackParams{
			LookaheadN:              *lookaheadN,
			MinK1BodyPercent:        *minK1BodyPercent,
			K1StrengthLookback:      *k1StrengthLookback,
			MinK1BodyToAvgRatio:     *minK1BodyToAvgRatio,
			TrendMAPeriod:           *trendMAPeriod,
			MinBoxRangePercent:      *minBoxRangePercent,
			MaxBoxRangePercent:      *maxBoxRangePercent,
			TouchTolerancePercent:   *touchTolerancePercent,
			MinConfirmWickBodyRatio: *minConfirmWickBodyRatio,
			CooldownBars:            *cooldownBars,
			RiskRewardRatio:         *riskRewardRatio,
		},
	}.Normalize()

	if req.ResultPath == "" {
		req.ResultPath = backtest.DefaultResultPath(req.StrategyName)
	}

	report, err := backtest.Run(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "backtest error: %v\n", err)
		os.Exit(1)
	}
	if err := backtest.SaveReport(req.ResultPath, report); err != nil {
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
