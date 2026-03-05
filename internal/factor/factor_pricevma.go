package factor

// FactorPriceVsMA 价格相对均线位置因子：价格在 SMA(N) 之上 → 看涨；之下 → 看跌。
// period: 均线周期（常用20），weight: 权重。
func (e *SignalContext) FactorPriceVsMA(period int, weight float64) *SignalContext {
	if e.KLine == nil {
		return e
	}
	prices := e.KLine.ClosePrices()
	if len(prices) < period {
		return e
	}

	sma := avg(prices[:period])
	currentPrice := prices[0]

	if currentPrice > sma {
		e.AddBull(weight)
	} else if currentPrice < sma {
		e.AddBear(weight)
	}
	return e
}
