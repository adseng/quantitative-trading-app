package cases

import "fmt"

// addMultiTripleRsiBollBreakSections 三因子 RSI+Boll+Breakout 的参数与权重网格探索。
//
// 网格：RSI(period) × Boll(period×multiplier) × Breakout(period) × 三因子权重(2×2×2)
// 探索该三因子组合的最优参数和权重配比。
func (b *caseBuilder) addMultiTripleRsiBollBreakSections() {
	for _, rp := range []int{5, 7} {
		for _, bp := range []int{12, 13} {
			for _, bm := range []float64{2.0, 2.2} {
				for _, kp := range []int{12, 15} {
					for _, rw := range []float64{1, 2} {
						for _, bw := range []float64{1} {
							for _, kw := range []float64{1} {
								b.add(fmt.Sprintf("R%d_w%.0f+Bo%dM%.1f_w%.0f+Br%d_w%.0f", rp, rw, bp, bm, bw, kp, kw), TestCase{
									UseRSI: true, RSIPeriod: rp, RSIOverbought: 80, RSIOversold: 20, RSIWeight: rw,
									UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: bw,
									UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: kw,
								})
							}
						}
					}
				}
			}
		}
	}
}
