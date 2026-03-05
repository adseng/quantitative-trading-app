package cases

import "fmt"

// addMultiBollRsiSections 双因子 Boll+RSI 的参数与权重网格探索。
//
// 网格：Boll(period×multiplier) × RSI(period×阈值) × 权重(4组)
func (b *caseBuilder) addMultiBollRsiSections() {
	weights := [][2]float64{{1, 1}, {2, 1}}
	for _, bp := range []int{12, 13} {
		for _, bm := range []float64{2.0, 2.2} {
			for _, rp := range []int{5, 7} {
				for _, th := range [][2]float64{{80, 20}} {
					for _, w := range weights {
						b.add(fmt.Sprintf("Bo%dM%.1f_w%.0f+R%d_%d_w%.0f", bp, bm, w[0], rp, int(th[0]), w[1]), TestCase{
							UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: w[0],
							UseRSI: true, RSIPeriod: rp, RSIOverbought: th[0], RSIOversold: th[1], RSIWeight: w[1],
						})
					}
				}
			}
		}
	}
}
