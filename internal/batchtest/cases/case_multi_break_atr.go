package cases

import "fmt"

// addMultiBreakAtrSections 双因子 Breakout+ATR 的参数与权重网格探索。
//
// 网格：Breakout(period) × ATR(period) × 权重(4组)
func (b *caseBuilder) addMultiBreakAtrSections() {
	weights := [][2]float64{{1, 1}, {2, 1}}
	for _, kp := range []int{12, 15} {
		for _, ap := range []int{14, 20} {
			for _, w := range weights {
				b.add(fmt.Sprintf("Br%d_w%.0f+A%d_w%.0f", kp, w[0], ap, w[1]), TestCase{
					UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: w[0],
					UseATR: true, ATRPeriod: ap, ATRWeight: w[1],
				})
			}
		}
	}
}
