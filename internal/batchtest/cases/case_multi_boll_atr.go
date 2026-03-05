package cases

import "fmt"

// addMultiBollAtrSections 双因子 Boll+ATR 的参数与权重网格探索。
//
// 网格：Boll(period×multiplier) × ATR(period) × 权重(4组)
func (b *caseBuilder) addMultiBollAtrSections() {
	weights := [][2]float64{{1, 1}, {2, 1}}
	for _, bp := range []int{12, 13} {
		for _, bm := range []float64{2.0, 2.2} {
			for _, ap := range []int{14, 20} {
				for _, w := range weights {
					b.add(fmt.Sprintf("Bo%dM%.1f_w%.0f+A%d_w%.0f", bp, bm, w[0], ap, w[1]), TestCase{
						UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: w[0],
						UseATR: true, ATRPeriod: ap, ATRWeight: w[1],
					})
				}
			}
		}
	}
}
