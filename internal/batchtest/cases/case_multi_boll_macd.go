package cases

import "fmt"

// addMultiBollMacdSections 双因子 Boll+MACD 的参数与权重网格探索。
//
// 网格：Boll(period×multiplier) × MACD(fast/slow/signal) × 权重(4组)
func (b *caseBuilder) addMultiBollMacdSections() {
	weights := [][2]float64{{1, 1}}
	for _, bp := range []int{12, 13} {
		for _, bm := range []float64{2.0, 2.2} {
			for _, mcfg := range [][3]int{{5, 15, 4}} {
				for _, w := range weights {
					b.add(fmt.Sprintf("Bo%dM%.1f_w%.0f+Macd%d_%d_%d_w%.0f", bp, bm, w[0], mcfg[0], mcfg[1], mcfg[2], w[1]), TestCase{
						UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: w[0],
						UseMACD: true, MACDFast: mcfg[0], MACDSlow: mcfg[1], MACDSignal: mcfg[2], MACDWeight: w[1],
					})
				}
			}
		}
	}
}
