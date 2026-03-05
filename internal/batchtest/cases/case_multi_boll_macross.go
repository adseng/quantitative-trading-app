package cases

import "fmt"

// addMultiBollMacrossSections 双因子 Boll+MACross 的参数与权重网格探索。
//
// 网格：Boll(period×multiplier) × MACross(short/long) × 权重(4组)
func (b *caseBuilder) addMultiBollMacrossSections() {
	weights := [][2]float64{{1, 1}}
	for _, bp := range []int{12, 13} {
		for _, bm := range []float64{2.0, 2.2} {
			for _, mc := range [][2]int{{5, 30}} {
				for _, w := range weights {
					b.add(fmt.Sprintf("Bo%dM%.1f_w%.0f+Mc%d_%d_w%.0f", bp, bm, w[0], mc[0], mc[1], w[1]), TestCase{
						UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: w[0],
						UseMACross: true, MACrossShort: mc[0], MACrossLong: mc[1], MACrossWeight: w[1], MACrossWindow: 2, MACrossPreempt: 0,
					})
				}
			}
		}
	}
}
