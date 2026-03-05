package cases

import "fmt"

// addMultiMaTrendSections 双因子 MA+Trend 的参数与权重网格探索。
//
// 网格：MA(long周期) × Trend(N) × 权重(4组)
func (b *caseBuilder) addMultiMaTrendSections() {
	weights := [][2]float64{{1, 1}}
	for _, maLong := range []int{7, 10} {
		for _, tn := range []int{6, 8} {
			for _, w := range weights {
				b.add(fmt.Sprintf("M1_%d_w%.0f+T%d_w%.0f", maLong, w[0], tn, w[1]), TestCase{
					UseMA: true, MaShort: 1, MaLong: maLong, MaWeight: w[0],
					UseTrend: true, TrendN: tn, TrendWeight: w[1],
				})
			}
		}
	}
}
