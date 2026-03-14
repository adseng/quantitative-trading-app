package strategy

import (
	"testing"

	"quantitative-trading-app/internal/market"
)

func TestEvaluateEMATrendPullbackFindsLongSignal(t *testing.T) {
	klines := []market.KLine{
		{OpenTime: 1, Open: 100, High: 101, Low: 99.5, Close: 100.8},
		{OpenTime: 2, Open: 100.8, High: 102, Low: 100.5, Close: 101.6},
		{OpenTime: 3, Open: 101.6, High: 103, Low: 101.2, Close: 102.5},
		{OpenTime: 4, Open: 102.5, High: 104, Low: 102.1, Close: 103.4},
		{OpenTime: 5, Open: 103.4, High: 105, Low: 103, Close: 104.3},
		{OpenTime: 6, Open: 104.3, High: 106, Low: 104, Close: 105.5},
		{OpenTime: 7, Open: 105.5, High: 109, Low: 105.2, Close: 108.2},
		{OpenTime: 8, Open: 108.2, High: 109.2, Low: 107.6, Close: 108.8},
		{OpenTime: 9, Open: 108.9, High: 112, Low: 108.7, Close: 111.4},
		{OpenTime: 10, Open: 111.4, High: 114, Low: 111, Close: 113.8},
	}

	signals := EvaluateEMATrendPullback(klines, EMATrendPullbackParams{
		FastPeriod:               2,
		SlowPeriod:               4,
		BreakoutLookback:         3,
		PullbackLookahead:        2,
		PullbackTolerancePercent: 0.02,
		ATRPeriod:                2,
		StopATRMultiplier:        0.5,
		CooldownBars:             0,
		RiskRewardRatio:          1.5,
	})

	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}

	signal := signals[0]
	if signal.Direction != DirectionLong {
		t.Fatalf("expected LONG signal, got %s", signal.Direction)
	}
	if signal.EntryIndex != 8 {
		t.Fatalf("expected entry index 8, got %d", signal.EntryIndex)
	}
	if signal.StopLoss >= signal.EntryPrice {
		t.Fatalf("expected stop loss below entry price, got stop %.4f entry %.4f", signal.StopLoss, signal.EntryPrice)
	}
}
