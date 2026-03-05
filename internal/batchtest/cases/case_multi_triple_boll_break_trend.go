package cases

import "fmt"

// addMultiTripleBollBreakTrendSections 三因子 Boll+Breakout+Trend 的参数与权重网格探索。
//
// 网格：Boll × Breakout × Trend(N) × 三因子权重
func (b *caseBuilder) addMultiTripleBollBreakTrendSections() {
	for _, bp := range []int{12, 13} {
		for _, bm := range []float64{2.0, 2.2} {
			for _, kp := range []int{12, 15} {
				for _, tn := range []int{6, 8} {
					for _, bw := range []float64{1} {
						for _, kw := range []float64{1} {
							for _, tw := range []float64{0.5} {
								b.add(fmt.Sprintf("Bo%dM%.1f_w%.0f+Br%d_w%.0f+T%d_w%.1f", bp, bm, bw, kp, kw, tn, tw), TestCase{
									UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: bw,
									UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: kw,
									UseTrend: true, TrendN: tn, TrendWeight: tw,
								})
							}
						}
					}
				}
			}
		}
	}
}
