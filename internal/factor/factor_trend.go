package factor

// FactorTrend 趋势因子：统计前 n 根 K 线的涨跌根数。
// 基于回测结论：净差 < 0（近期跌多涨少）时下一根更易上涨，故输出正向看涨信号。
// 净差 > 0 → 看跌；净差 < 0 → 看涨；净差 = 0 → 0。
//
// 参数：
//   - n: 统计的 K 线根数（需至少 n+1 根历史数据）
//   - weight: 该因子权重（多因子组合时用于调节贡献度）
func (e *SignalContext) FactorTrend(n int, weight float64) *SignalContext {
	if e.KLine == nil || len(e.KLine.History) < n+1 {
		return e
	}
	need := n + 1
	if len(e.KLine.History) < need {
		return e
	}

	upCount := 0
	downCount := 0
	for i := 0; i < n; i++ {
		curr := e.KLine.History[i].Close
		prev := e.KLine.History[i+1].Close
		if curr > prev {
			upCount++
		} else if curr < prev {
			downCount++
		}
	}

	diff := upCount - downCount
	if diff < 0 {
		e.AddBull(weight)
	} else if diff > 0 {
		e.AddBear(weight)
	}
	return e
}
