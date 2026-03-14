package backtest

import (
	"testing"

	"quantitative-trading-app/internal/market"
	"quantitative-trading-app/internal/strategy"
)

func TestRunEMAOnKlinesProducesWinningTrade(t *testing.T) {
	klines := []market.KLine{
		{OpenTime: 1, Open: 100, High: 101, Low: 99.5, Close: 100.8},
		{OpenTime: 2, Open: 100.8, High: 102, Low: 100.5, Close: 101.6},
		{OpenTime: 3, Open: 101.6, High: 103, Low: 101.2, Close: 102.5},
		{OpenTime: 4, Open: 102.5, High: 104, Low: 102.1, Close: 103.4},
		{OpenTime: 5, Open: 103.4, High: 105, Low: 103, Close: 104.3},
		{OpenTime: 6, Open: 104.3, High: 106, Low: 104, Close: 105.5},
		{OpenTime: 7, Open: 105.5, High: 109, Low: 105.2, Close: 108.2},
		{OpenTime: 8, Open: 108.2, High: 109.2, Low: 107.6, Close: 108.8},
		{OpenTime: 9, Open: 108.9, High: 114, Low: 108.7, Close: 113.2},
		{OpenTime: 10, Open: 113.2, High: 116, Low: 112.9, Close: 115.4},
	}

	report, err := RunEMAOnKlines(RunEMARequest{
		StrategyName:     strategy.EMATrendPullbackName,
		InitialBalance:   10000,
		PositionSizeUSDT: 100,
		Params: strategy.EMATrendPullbackParams{
			FastPeriod:               2,
			SlowPeriod:               4,
			BreakoutLookback:         3,
			PullbackLookahead:        2,
			PullbackTolerancePercent: 0.02,
			ATRPeriod:                2,
			StopATRMultiplier:        0.5,
			CooldownBars:             0,
			RiskRewardRatio:          1.2,
		},
	}, klines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Summary.ExecutedTrades != 1 {
		t.Fatalf("expected 1 trade, got %d", report.Summary.ExecutedTrades)
	}
	if report.Trades[0].ExitReason != "take_profit" {
		t.Fatalf("expected take profit exit, got %s", report.Trades[0].ExitReason)
	}
	if report.Trades[0].PnL <= 0 {
		t.Fatalf("expected positive pnl, got %.2f", report.Trades[0].PnL)
	}
}
