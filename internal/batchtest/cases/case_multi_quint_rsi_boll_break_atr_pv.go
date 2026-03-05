package cases

import "fmt"

// addMultiQuintRsiBollBreakAtrPvSections 五因子 RSI+Boll+Break+ATR+PriceVsMA 参数网格探索。
//
// 基于最优四因子 Bo+R+Br+A，叠加 PriceVsMA 价格相对均线。
func (b *caseBuilder) addMultiQuintRsiBollBreakAtrPvSections() {
	for _, rp := range []int{5, 7, 9} {
		for _, bp := range []int{12, 13} {
			for _, bm := range []float64{2.0, 2.2} {
				for _, kp := range []int{12, 15} {
					for _, ap := range []int{14, 20} {
						for _, pv := range []int{5, 8} {
							for _, pvw := range []float64{0.25, 0.5} {
								b.add(fmt.Sprintf("R%d+Bo%dM%.1f+Br%d+A%d+Pv%d_w%.2f", rp, bp, bm, kp, ap, pv, pvw), TestCase{
									UseRSI: true, RSIPeriod: rp, RSIOverbought: 80, RSIOversold: 20, RSIWeight: 1,
									UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: 1,
									UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: 1,
									UseATR: true, ATRPeriod: ap, ATRWeight: 1,
									UsePriceVsMA: true, PriceVsMAPeriod: pv, PriceVsMAWeight: pvw,
								})
							}
						}
					}
				}
			}
		}
	}
}
