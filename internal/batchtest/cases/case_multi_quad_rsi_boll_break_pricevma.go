package cases

import "fmt"

// addMultiQuadRsiBollBreakPriceVmaSections 四因子 RSI+Boll+Breakout+PriceVsMA 的参数与权重网格探索。
//
// 基于最优三因子叠加 PriceVsMA 价格动量，网格约 100 组。
func (b *caseBuilder) addMultiQuadRsiBollBreakPriceVmaSections() {
	for _, rp := range []int{5, 7} {
		for _, bp := range []int{12, 13} {
			for _, bm := range []float64{2.0, 2.2} {
				for _, kp := range []int{12, 15} {
					for _, pv := range []int{5, 8, 12} {
						for _, pvw := range []float64{0.25, 0.5, 0.75, 1} {
							b.add(fmt.Sprintf("R%d+Bo%dM%.1f+Br%d+Pv%d_w%.2f", rp, bp, bm, kp, pv, pvw), TestCase{
								UseRSI: true, RSIPeriod: rp, RSIOverbought: 80, RSIOversold: 20, RSIWeight: 1,
								UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: 1,
								UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: 1,
								UsePriceVsMA: true, PriceVsMAPeriod: pv, PriceVsMAWeight: pvw,
							})
						}
					}
				}
			}
		}
	}
}
