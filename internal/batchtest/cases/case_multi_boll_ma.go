package cases

import "fmt"

// addMultiBollMaSections 双因子 Boll+MA 的参数与权重网格探索。
//
// 网格：Boll(period×multiplier) × MA(long周期) × 权重(4组)
func (b *caseBuilder) addMultiBollMaSections() {
	weights := [][2]float64{{1, 1}, {2, 1}}
	for _, bp := range []int{12, 13} {
		for _, bm := range []float64{2.0, 2.2} {
			for _, maLong := range []int{7, 10} {
				for _, w := range weights {
					b.add(fmt.Sprintf("Bo%dM%.1f_w%.0f+M1_%d_w%.0f", bp, bm, w[0], maLong, w[1]), TestCase{
						UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: w[0],
						UseMA: true, MaShort: 1, MaLong: maLong, MaWeight: w[1],
					})
				}
			}
		}
	}
}
