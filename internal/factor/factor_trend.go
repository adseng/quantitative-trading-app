package factor

// FactorTrend 趋势因子：统计前 n 根 K 线的涨跌根数。
// 净差 > 0 → 信号 1 → 影响力 +weight；净差 < 0 → 信号 -1 → 影响力 -weight；净差 = 0 → 0。
//
// 参数：
//   - n: 统计的 K 线根数（需至少 n+1 根历史数据）
//   - weight: 该因子权重
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
	if diff > 0 {
		e.AddBull(weight) // 信号 1，影响力 = weight × 1
	} else if diff < 0 {
		e.AddBear(weight) // 信号 -1，影响力 = weight × (-1)
	}
	return e
}
