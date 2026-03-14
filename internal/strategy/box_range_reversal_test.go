package strategy

import "testing"

func TestBoxRangeTakeProfitFactorBounds(t *testing.T) {
	params := BoxRangeReversalParams{TakeProfitFactor: -1}.Normalize()
	if params.TakeProfitFactor != 0 {
		t.Fatalf("expected lower bound 0, got %.2f", params.TakeProfitFactor)
	}

	params = BoxRangeReversalParams{TakeProfitFactor: 3}.Normalize()
	if params.TakeProfitFactor != 1 {
		t.Fatalf("expected upper bound 1, got %.2f", params.TakeProfitFactor)
	}
}

func TestInterpolateTakeProfitFactor(t *testing.T) {
	if got := interpolateTakeProfit(DirectionLong, 90, 100, 110, 0); got != 100 {
		t.Fatalf("expected long factor 0 to hit midline 100, got %.2f", got)
	}
	if got := interpolateTakeProfit(DirectionLong, 90, 100, 110, 1); got != 110 {
		t.Fatalf("expected long factor 1 to hit opposite edge 110, got %.2f", got)
	}
	if got := interpolateTakeProfit(DirectionShort, 90, 100, 110, 0); got != 100 {
		t.Fatalf("expected short factor 0 to hit midline 100, got %.2f", got)
	}
	if got := interpolateTakeProfit(DirectionShort, 90, 100, 110, 1); got != 90 {
		t.Fatalf("expected short factor 1 to hit opposite edge 90, got %.2f", got)
	}
}
