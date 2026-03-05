package cases

import "fmt"

// addMultiRsiAtrSections 双因子 RSI+ATR 的参数与权重网格探索。
//
// 网格：RSI(period×阈值) × ATR(period) × 权重(4组)
func (b *caseBuilder) addMultiRsiAtrSections() {
	weights := [][2]float64{{1, 1}}
	for _, rp := range []int{5, 7} {
		for _, th := range [][2]float64{{80, 20}} {
			for _, ap := range []int{14, 20} {
				for _, w := range weights {
					b.add(fmt.Sprintf("R%d_%d_w%.0f+A%d_w%.0f", rp, int(th[0]), w[0], ap, w[1]), TestCase{
						UseRSI: true, RSIPeriod: rp, RSIOverbought: th[0], RSIOversold: th[1], RSIWeight: w[0],
						UseATR: true, ATRPeriod: ap, ATRWeight: w[1],
					})
				}
			}
		}
	}
}
