package cases

import "fmt"

// addMultiQuadRsiBollBreakSessionSections 四因子 RSI+Boll+Breakout+Session 的参数与权重网格探索。
//
// 基于最优三因子叠加 Session 时段过滤（避开低流动性），网格约 100 组。
func (b *caseBuilder) addMultiQuadRsiBollBreakSessionSections() {
	for _, rp := range []int{5, 7} {
		for _, bp := range []int{12, 13} {
			for _, bm := range []float64{2.0, 2.2} {
				for _, kp := range []int{12, 15} {
					for _, rw := range []float64{1, 2} {
						for _, bw := range []float64{1, 2} {
							for _, kw := range []float64{1} {
								for _, sw := range []float64{0.25, 0.5, 0.75} {
									b.add(fmt.Sprintf("R%d_w%.0f+Bo%dM%.1f_w%.0f+Br%d_w%.0f+Sess_w%.2f", rp, rw, bp, bm, bw, kp, kw, sw), TestCase{
										UseRSI: true, RSIPeriod: rp, RSIOverbought: 80, RSIOversold: 20, RSIWeight: rw,
										UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: bw,
										UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: kw,
										UseSession: true, SessionWeight: sw,
									})
								}
							}
						}
					}
				}
			}
		}
	}
}
