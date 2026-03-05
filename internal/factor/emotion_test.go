package factor

import (
	"testing"
)

func TestFactorMA(t *testing.T) {
	// 构造上涨序列：多头排列
	closes := []float64{}
	for i := 0; i < 25; i++ {
		closes = append(closes, 100.0+float64(i))
	}
	hist := &KLineHistory{
		Current: &KLine{Close: closes[24]},
		History: makeHistory(closes),
	}
	ctx := NewSignalContext(hist)
	ctx.FactorMA(5, 20, 1.0)
	if ctx.BullScore <= 0 {
		t.Errorf("expected bull score > 0 for uptrend, got bull=%.2f bear=%.2f", ctx.BullScore, ctx.BearScore)
	}
}

func TestFactorTrend(t *testing.T) {
	// 10 根里 8 涨 2 跌
	closes := []float64{100, 101, 102, 103, 99, 100, 101, 102, 103, 104, 105}
	hist := &KLineHistory{
		Current: &KLine{Close: closes[10]},
		History: makeHistory(closes),
	}
	ctx := NewSignalContext(hist)
	ctx.FactorTrend(10, 1.0)
	if ctx.BullScore <= 0 {
		t.Errorf("expected bull score > 0, got bull=%.2f bear=%.2f", ctx.BullScore, ctx.BearScore)
	}
}

func makeHistory(closes []float64) []*KLine {
	n := len(closes)
	hist := make([]*KLine, n)
	for i := 0; i < n; i++ {
		hist[i] = &KLine{Close: closes[n-1-i]}
	}
	return hist
}
