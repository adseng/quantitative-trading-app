package backtest

import (
	"testing"

	"quantitative-trading-app/internal/market"
	"quantitative-trading-app/internal/strategy"
)

func TestRunOnKlinesProducesWinningTrade(t *testing.T) {
	klines := []market.KLine{
		{OpenTime: 1, Open: 98, High: 100, Low: 97, Close: 99},
		{OpenTime: 2, Open: 99, High: 101, Low: 98, Close: 100},
		{OpenTime: 3, Open: 100, High: 110, Low: 99, Close: 108},
		{OpenTime: 4, Open: 112, High: 113, Low: 106, Close: 111},
		{OpenTime: 5, Open: 111.2, High: 136, Low: 111, Close: 128},
		{OpenTime: 6, Open: 128, High: 140, Low: 126, Close: 130},
	}

	report, err := RunOnKlines(RunRequest{
		StrategyName:     strategy.BoxPullbackName,
		InitialBalance:   10000,
		PositionSizeUSDT: 100,
		Params: strategy.BoxPullbackParams{
			LookaheadN:              2,
			MinK1BodyPercent:        0.02,
			K1StrengthLookback:      2,
			MinK1BodyToAvgRatio:     1.2,
			TrendMAPeriod:           2,
			MinBoxRangePercent:      0.001,
			MaxBoxRangePercent:      0.2,
			TouchTolerancePercent:   0,
			MinConfirmWickBodyRatio: 1,
			CooldownBars:            0,
			RiskRewardRatio:         2,
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

func TestRunOnKlinesAllowsMultipleOpenTrades(t *testing.T) {
	klines := []market.KLine{
		{OpenTime: 1, Open: 98, High: 100, Low: 97, Close: 99},
		{OpenTime: 2, Open: 99, High: 101, Low: 98, Close: 100},
		{OpenTime: 3, Open: 100, High: 110, Low: 99, Close: 108},
		{OpenTime: 4, Open: 112, High: 113, Low: 106, Close: 111},
		{OpenTime: 5, Open: 111.2, High: 121, Low: 111, Close: 120},
		{OpenTime: 6, Open: 123, High: 124, Low: 118, Close: 122},
		{OpenTime: 7, Open: 126, High: 132, Low: 125, Close: 131},
		{OpenTime: 8, Open: 132, High: 140, Low: 124, Close: 138},
		{OpenTime: 9, Open: 138, High: 160, Low: 130, Close: 150},
	}

	report, err := RunOnKlines(RunRequest{
		StrategyName:     strategy.BoxPullbackName,
		InitialBalance:   10000,
		PositionSizeUSDT: 100,
		Params: strategy.BoxPullbackParams{
			LookaheadN:              2,
			MinK1BodyPercent:        0.02,
			K1StrengthLookback:      2,
			MinK1BodyToAvgRatio:     1.2,
			TrendMAPeriod:           2,
			MinBoxRangePercent:      0.001,
			MaxBoxRangePercent:      0.2,
			TouchTolerancePercent:   0,
			MinConfirmWickBodyRatio: 1,
			CooldownBars:            0,
			RiskRewardRatio:         2,
		},
	}, klines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Summary.ExecutedTrades != 2 {
		t.Fatalf("expected 2 trades, got %d", report.Summary.ExecutedTrades)
	}
	if report.Summary.SkippedSignals != 0 {
		t.Fatalf("expected 0 skipped signals, got %d", report.Summary.SkippedSignals)
	}
	if report.Trades[0].ExitIndex <= report.Trades[1].EntryIndex {
		t.Fatalf("expected overlapping trades, got exit %d entry %d", report.Trades[0].ExitIndex, report.Trades[1].EntryIndex)
	}
}
