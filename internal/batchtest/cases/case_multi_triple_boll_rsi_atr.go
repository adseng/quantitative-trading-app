package cases

import "fmt"

// addMultiTripleBollRsiAtrSections 三因子 Boll+RSI+ATR 的参数与权重网格探索。
//
// 网格：Boll × RSI(period×阈值) × ATR(period) × 三因子权重
func (b *caseBuilder) addMultiTripleBollRsiAtrSections() {
	for _, bp := range []int{12, 13} {
		for _, bm := range []float64{2.0, 2.2} {
			for _, rp := range []int{5, 7} {
				for _, th := range [][2]float64{{80, 20}} {
					for _, ap := range []int{14} {
						for _, bw := range []float64{1} {
							for _, rw := range []float64{1} {
								for _, aw := range []float64{0.5, 1} {
									b.add(fmt.Sprintf("Bo%dM%.1f_w%.0f+R%d_w%.0f+A%d_w%.1f", bp, bm, bw, rp, rw, ap, aw), TestCase{
										UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: bw,
										UseRSI: true, RSIPeriod: rp, RSIOverbought: th[0], RSIOversold: th[1], RSIWeight: rw,
										UseATR: true, ATRPeriod: ap, ATRWeight: aw,
									})
								}
							}
						}
					}
				}
			}
		}
	}
}
