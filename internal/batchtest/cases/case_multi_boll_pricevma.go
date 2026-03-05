package cases

import "fmt"

// addMultiBollPriceVmaSections 双因子 Boll+PriceVsMA 的参数与权重网格探索。
//
// 网格：Boll(period×multiplier) × PriceVsMA(period) × 权重(4组)
func (b *caseBuilder) addMultiBollPriceVmaSections() {
	weights := [][2]float64{{1, 1}, {2, 1}}
	for _, bp := range []int{12, 13} {
		for _, bm := range []float64{2.0, 2.2} {
			for _, pv := range []int{8} {
				for _, w := range weights {
					b.add(fmt.Sprintf("Bo%dM%.1f_w%.0f+Pv%d_w%.0f", bp, bm, w[0], pv, w[1]), TestCase{
						UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: w[0],
						UsePriceVsMA: true, PriceVsMAPeriod: pv, PriceVsMAWeight: w[1],
					})
				}
			}
		}
	}
}
