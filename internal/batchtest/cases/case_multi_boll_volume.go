package cases

import "fmt"

// addMultiBollVolumeSections 双因子 Boll+Volume 的参数与权重网格探索。
//
// 网格：Boll(period×multiplier) × Volume(period) × 权重(4组)
func (b *caseBuilder) addMultiBollVolumeSections() {
	weights := [][2]float64{{1, 1}, {2, 1}}
	for _, bp := range []int{12, 13} {
		for _, bm := range []float64{2.0, 2.2} {
			for _, vp := range []int{5, 10} {
				for _, w := range weights {
					b.add(fmt.Sprintf("Bo%dM%.1f_w%.0f+V%d_w%.0f", bp, bm, w[0], vp, w[1]), TestCase{
						UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: w[0],
						UseVolume: true, VolumePeriod: vp, VolumeWeight: w[1],
					})
				}
			}
		}
	}
}
