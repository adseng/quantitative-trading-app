package cases

import "fmt"

// addMultiTripleRsiBreakAtrSections 三因子 RSI+Breakout+ATR 的参数与权重网格探索。
//
// 网格：RSI(period×阈值) × Breakout(period) × ATR(period) × 三因子权重
func (b *caseBuilder) addMultiTripleRsiBreakAtrSections() {
	for _, rp := range []int{5, 7} {
		for _, th := range [][2]float64{{80, 20}} {
			for _, kp := range []int{12, 15} {
				for _, ap := range []int{14} {
					for _, rw := range []float64{1} {
						for _, kw := range []float64{1} {
							for _, aw := range []float64{0.5, 1} {
								b.add(fmt.Sprintf("R%d_%d_w%.0f+Br%d_w%.0f+A%d_w%.1f", rp, int(th[0]), rw, kp, kw, ap, aw), TestCase{
									UseRSI: true, RSIPeriod: rp, RSIOverbought: th[0], RSIOversold: th[1], RSIWeight: rw,
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
