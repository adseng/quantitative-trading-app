package cases

import "fmt"

// addMultiBollTrendSections 双因子 Boll+Trend 的参数与权重网格探索。
//
// 网格：Boll(period×multiplier) × Trend(N) × 权重(4组)
func (b *caseBuilder) addMultiBollTrendSections() {
	weights := [][2]float64{{1, 1}, {2, 1}}
	for _, bp := range []int{12, 13} {
		for _, bm := range []float64{2.0, 2.2} {
			for _, tn := range []int{6, 8} {
				for _, w := range weights {
					b.add(fmt.Sprintf("Bo%dM%.1f_w%.0f+T%d_w%.0f", bp, bm, w[0], tn, w[1]), TestCase{
						UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: w[0],
						UseTrend: true, TrendN: tn, TrendWeight: w[1],
					})
				}
			}
		}
	}
}
