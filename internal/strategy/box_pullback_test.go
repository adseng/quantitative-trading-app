package strategy

import (
	"testing"

	"quantitative-trading-app/internal/market"
)

func TestEvaluateBoxPullbackFindsLongSignal(t *testing.T) {
	klines := []market.KLine{
		{OpenTime: 1, Open: 98, High: 100, Low: 97, Close: 99},
		{OpenTime: 2, Open: 99, High: 101, Low: 98, Close: 100},
		{OpenTime: 3, Open: 100, High: 110, Low: 99, Close: 108},
		{OpenTime: 4, Open: 112, High: 113, Low: 106, Close: 111},
		{OpenTime: 5, Open: 111.2, High: 120, Low: 111, Close: 118},
		{OpenTime: 6, Open: 118, High: 121, Low: 117, Close: 119},
	}

	signals := EvaluateBoxPullback(klines, BoxPullbackParams{
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
	})

	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}

	signal := signals[0]
	if signal.Direction != DirectionLong {
		t.Fatalf("expected LONG signal, got %s", signal.Direction)
	}
	if signal.EntryIndex != 4 {
		t.Fatalf("expected entry index 4, got %d", signal.EntryIndex)
	}
	if signal.StopLoss != 99 {
		t.Fatalf("expected stop loss 99, got %v", signal.StopLoss)
	}
}
