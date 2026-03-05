package cases

import "fmt"

// addMultiQuadRsiBollBreakAtrSections 四因子 RSI+Boll+Breakout+ATR 的参数与权重网格探索。
//
// 基于最优三因子 Bo+R+Br 叠加 ATR 波动率过滤，网格约 100 组。
func (b *caseBuilder) addMultiQuadRsiBollBreakAtrSections() {
	for _, rp := range []int{5, 7} {
		for _, bp := range []int{12, 13} {
			for _, bm := range []float64{2.0, 2.2} {
				for _, kp := range []int{12, 15} {
					for _, ap := range []int{7, 14, 20} {
						for _, aw := range []float64{0.25, 0.5, 0.75, 1} {
							b.add(fmt.Sprintf("R%d+Bo%dM%.1f+Br%d+A%d_w%.2f", rp, bp, bm, kp, ap, aw), TestCase{
								UseRSI: true, RSIPeriod: rp, RSIOverbought: 80, RSIOversold: 20, RSIWeight: 1,
								UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: 1,
								UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: 1,
								UseATR: true, ATRPeriod: ap, ATRWeight: aw,
							})
						}
					}
				}
			}
		}
	}
}
