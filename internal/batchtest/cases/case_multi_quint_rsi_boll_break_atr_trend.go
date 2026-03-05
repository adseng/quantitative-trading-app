package cases

import "fmt"

// addMultiQuintRsiBollBreakAtrTrendSections 五因子 RSI+Boll+Break+ATR+Trend 参数网格探索。
//
// 基于最优四因子 Bo+R+Br+A，叠加 Trend 趋势确认。
func (b *caseBuilder) addMultiQuintRsiBollBreakAtrTrendSections() {
	for _, rp := range []int{5, 7, 9} {
		for _, bp := range []int{12, 13} {
			for _, bm := range []float64{2.0, 2.2} {
				for _, kp := range []int{12, 15} {
					for _, ap := range []int{14, 20} {
						for _, tn := range []int{6, 8} {
							for _, tw := range []float64{0.25, 0.5} {
								b.add(fmt.Sprintf("R%d+Bo%dM%.1f+Br%d+A%d+T%d_w%.2f", rp, bp, bm, kp, ap, tn, tw), TestCase{
									UseRSI: true, RSIPeriod: rp, RSIOverbought: 80, RSIOversold: 20, RSIWeight: 1,
									UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: 1,
									UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: 1,
									UseATR: true, ATRPeriod: ap, ATRWeight: 1,
									UseTrend: true, TrendN: tn, TrendWeight: tw,
								})
							}
						}
					}
				}
			}
		}
	}
}
