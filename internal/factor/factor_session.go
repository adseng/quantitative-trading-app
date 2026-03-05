package factor

import "time"

// FactorSession 时段因子：根据 K 线所在交易时段给出信号。
// 亚盘(00:00-08:00 UTC) 通常波动较小 → 0；
// 欧盘(08:00-13:00 UTC) 开盘常延续趋势 → 跟随价格方向；
// 美盘(13:00-21:00 UTC) 波动最大 → 跟随价格方向且置信更高。
// weight: 权重。
func (e *SignalContext) FactorSession(weight float64) *SignalContext {
	if e.KLine == nil || len(e.KLine.History) < 2 {
		return e
	}

	t := time.UnixMilli(e.KLine.History[0].OpenTime).UTC()
	hour := t.Hour()

	// 亚盘：波动低，不产出信号
	if hour >= 0 && hour < 8 {
		return e
	}

	priceChange := e.KLine.History[0].Close - e.KLine.History[1].Close
	// 回测结论：欧美盘下跌时下一根更易上涨
	if priceChange < 0 {
		e.AddBull(weight)
	} else if priceChange > 0 {
		e.AddBear(weight)
	}
	return e
}
