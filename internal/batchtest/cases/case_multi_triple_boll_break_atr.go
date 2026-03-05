package cases

import "fmt"

// addMultiTripleBollBreakAtrSections 三因子 Boll+Breakout+ATR 的参数与权重网格探索。
//
// 网格：Boll × Breakout × ATR(period) × 三因子权重
// 探索该三因子组合的最优参数和权重配比。
func (b *caseBuilder) addMultiTripleBollBreakAtrSections() {
	for _, bp := range []int{12, 13} {
		for _, bm := range []float64{2.0, 2.2} {
			for _, kp := range []int{12, 15} {
				for _, ap := range []int{14, 20} {
					for _, bw := range []float64{1} {
						for _, kw := range []float64{1} {
							for _, aw := range []float64{0.5, 1} {
								b.add(fmt.Sprintf("Bo%dM%.1f_w%.0f+Br%d_w%.0f+A%d_w%.1f", bp, bm, bw, kp, kw, ap, aw), TestCase{
									UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: bw,
									UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: kw,
									UseATR: true, ATRPeriod: ap, ATRWeight: aw,
								})
							}
						}
					}
				}
			}
		}
	}
}
