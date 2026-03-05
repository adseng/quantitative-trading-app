package cases

import "fmt"

// addMultiRsiBreakSections 双因子 RSI+Breakout 的参数与权重网格探索。
//
// 网格：RSI(period×阈值) × Breakout(period) × 权重(4组)
func (b *caseBuilder) addMultiRsiBreakSections() {
	weights := [][2]float64{{1, 1}, {2, 1}}
	for _, rp := range []int{5, 7} {
		for _, th := range [][2]float64{{80, 20}} {
			for _, kp := range []int{12, 15} {
				for _, w := range weights {
					b.add(fmt.Sprintf("R%d_%d_w%.0f+Br%d_w%.0f", rp, int(th[0]), w[0], kp, w[1]), TestCase{
						UseRSI: true, RSIPeriod: rp, RSIOverbought: th[0], RSIOversold: th[1], RSIWeight: w[0],
						UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: w[1],
					})
				}
			}
		}
	}
}
