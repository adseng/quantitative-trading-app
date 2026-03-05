package factor

// FactorVolume 量价配合因子：成交量高于 N 周期均量时，结合价格方向给出信号。
// 回测结论：放量下跌时下一根更易上涨，故输出正向看涨信号。放量上涨 → 看跌；放量下跌 → 看涨。
// period: 均量周期（常用20），weight: 权重。
func (e *SignalContext) FactorVolume(period int, weight float64) *SignalContext {
	if e.KLine == nil || len(e.KLine.History) < period+1 {
		return e
	}

	vols := e.KLine.Volumes()
	if len(vols) < period+1 {
		return e
	}

	currentVol := vols[0]
	var avgVol float64
	for i := 1; i <= period; i++ {
		avgVol += vols[i]
	}
	avgVol /= float64(period)

	if currentVol <= avgVol {
		return e
	}

	priceChange := e.KLine.History[0].Close - e.KLine.History[1].Close
	if priceChange < 0 {
		e.AddBull(weight)
	} else if priceChange > 0 {
		e.AddBear(weight)
	}
	return e
}
