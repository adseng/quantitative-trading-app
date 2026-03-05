package cases

import "fmt"

// addMultiTripleBollBreakMaSections 三因子 Boll+Breakout+MA 的参数与权重网格探索。
//
// 网格：Boll × Breakout × MA(long) × 三因子权重
func (b *caseBuilder) addMultiTripleBollBreakMaSections() {
	for _, bp := range []int{12, 13} {
		for _, bm := range []float64{2.0, 2.2} {
			for _, kp := range []int{12, 15} {
				for _, maLong := range []int{7} {
					for _, bw := range []float64{1} {
						for _, kw := range []float64{1} {
							for _, mw := range []float64{0.5, 1} {
								b.add(fmt.Sprintf("Bo%dM%.1f_w%.0f+Br%d_w%.0f+M1_%d_w%.1f", bp, bm, bw, kp, kw, maLong, mw), TestCase{
									UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: bw,
									UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: kw,
									UseMA: true, MaShort: 1, MaLong: maLong, MaWeight: mw,
								})
							}
						}
					}
				}
			}
		}
	}
}
