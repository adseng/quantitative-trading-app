package factor

// FactorMA 均线因子：短期均线 vs 长期均线（多空排列）。
// 短期 > 长期 → 信号 1 → 影响力 +weight；短期 < 长期 → 信号 -1 → 影响力 -weight；相等 → 0。
//
// 参数：
//   - shortPeriod, longPeriod: 短期、长期均线周期（如 5, 20）
//   - weight: 该因子权重
func (e *SignalContext) FactorMA(shortPeriod, longPeriod int, weight float64) *SignalContext {
	if e.KLine == nil || len(e.KLine.History) < longPeriod {
		return e
	}
	prices := e.KLine.ClosePrices()
	if len(prices) < longPeriod {
		return e
	}
	// prices[0] 为最新（最近一根）收盘价
	shortMA := avg(prices[:shortPeriod])
	longMA := avg(prices[:longPeriod])

	// 多头排列：短期在长期之上
	if shortMA > longMA {
		e.AddBull(weight)
	} else if shortMA < longMA {
		e.AddBear(weight)
	}
	return e
}

// avg 计算 float64 切片的算术平均值。
func avg(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	var s float64
	for _, v := range x {
		s += v
	}
	return s / float64(len(x))
}

// MA 计算 prices 最后 period 根的简单移动平均。
func MA(prices []float64, period int) float64 {
	if len(prices) < period || period <= 0 {
		return 0
	}
	return avg(prices[len(prices)-period:])
}
