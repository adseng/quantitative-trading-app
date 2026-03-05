package factor

// FactorBreakout 高低点突破因子：价格突破近 N 根 K 线最高价 → 看涨；跌破最低价 → 看跌。
// period: 回看周期（常用20），weight: 权重。
func (e *SignalContext) FactorBreakout(period int, weight float64) *SignalContext {
	if e.KLine == nil || len(e.KLine.History) < period+1 {
		return e
	}

	currentClose := e.KLine.History[0].Close
	var highest, lowest float64
	for i := 1; i <= period; i++ {
		h := e.KLine.History[i].High
		l := e.KLine.History[i].Low
		if i == 1 || h > highest {
			highest = h
		}
		if i == 1 || l < lowest {
			lowest = l
		}
	}

	if currentClose > highest {
		e.AddBull(weight)
	} else if currentClose < lowest {
		e.AddBear(weight)
	}
	return e
}
