package cases

import "fmt"

// addMultiBollBreakSections 双因子 Boll+Breakout 的参数与权重网格探索。
//
// 网格：Boll(period×multiplier) × Breakout(period) × 权重(4组)
// 用于找到该组合下的最优参数和权重配比。
func (b *caseBuilder) addMultiBollBreakSections() {
	weights := [][2]float64{{1, 1}, {2, 1}} // 精简：2组权重
	for _, bp := range []int{12, 13} {
		for _, bm := range []float64{2.0, 2.2} {
			for _, kp := range []int{12, 15} {
				for _, w := range weights {
					b.add(fmt.Sprintf("Bo%dM%.1f_w%.0f+Br%d_w%.0f", bp, bm, w[0], kp, w[1]), TestCase{
						UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: w[0],
						UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: w[1],
					})
				}
			}
		}
	}
}
