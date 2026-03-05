package factor

// FactorPriceVsMA 价格相对均线位置因子。
// 基于回测结论：价格在 SMA(N) 之下时下一根更易上涨，故输出正向看涨信号。
// 价格 > SMA → 看跌；价格 < SMA → 看涨。
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

	if currentPrice < sma {
		e.AddBull(weight)
	} else if currentPrice > sma {
		e.AddBear(weight)
	}
	return e
}
