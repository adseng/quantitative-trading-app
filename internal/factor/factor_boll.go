package factor

import "math"

// FactorBoll 布林带因子：价格 < 下轨 → 看涨(信号1)；价格 > 上轨 → 看跌(信号-1)。
// period: SMA周期（常用20），multiplier: 标准差倍数（常用2.0），weight: 权重。
func (e *SignalContext) FactorBoll(period int, multiplier, weight float64) *SignalContext {
	if e.KLine == nil || len(e.KLine.History) < period {
		return e
	}
	prices := e.KLine.ClosePrices()
	if len(prices) < period {
		return e
	}

	slice := prices[:period]
	middle := avg(slice)

	var variance float64
	for _, p := range slice {
		d := p - middle
		variance += d * d
	}
	stddev := math.Sqrt(variance / float64(period))

	upper := middle + multiplier*stddev
	lower := middle - multiplier*stddev

	currentPrice := prices[0] // newest price

	if currentPrice < lower {
		e.AddBull(weight) // oversold, expect bounce
	} else if currentPrice > upper {
		e.AddBear(weight) // overbought, expect drop
	}
	return e
}
