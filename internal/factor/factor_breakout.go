package factor

// FactorBreakout 高低点突破因子。
// 回测结论：跌破最低价时下一根更易上涨，故输出正向看涨信号。
// 突破最高价 → 看跌；跌破最低价 → 看涨。
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

	if currentClose < lowest {
		e.AddBull(weight)
	} else if currentClose > highest {
		e.AddBear(weight)
	}
	return e
}
