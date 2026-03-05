package factor

import "math"

// FactorATR 波动率因子：当前 ATR > 历史 ATR 均值（波动放大），结合价格方向给出信号。
// ATR 放大 + 价格上涨 → 看涨（趋势加速）；ATR 放大 + 价格下跌 → 看跌。
// period: ATR 周期（常用14），weight: 权重。
func (e *SignalContext) FactorATR(period int, weight float64) *SignalContext {
	if e.KLine == nil || len(e.KLine.History) < period+1 {
		return e
	}

	trs := make([]float64, period)
	for i := 0; i < period; i++ {
		cur := e.KLine.History[i]
		prev := e.KLine.History[i+1]
		hl := cur.High - cur.Low
		hc := math.Abs(cur.High - prev.Close)
		lc := math.Abs(cur.Low - prev.Close)
		tr := hl
		if hc > tr {
			tr = hc
		}
		if lc > tr {
			tr = lc
		}
		trs[i] = tr
	}

	currentATR := trs[0]
	avgATR := 0.0
	for _, v := range trs {
		avgATR += v
	}
	avgATR /= float64(period)

	if currentATR <= avgATR {
		return e
	}

	priceChange := e.KLine.History[0].Close - e.KLine.History[1].Close
	// 回测结论：ATR放大+价格下跌时下一根更易上涨（超卖反弹）
	if priceChange < 0 {
		e.AddBull(weight)
	} else if priceChange > 0 {
		e.AddBear(weight)
	}
	return e
}
