package cases

import "fmt"

// addMultiQuadRsiBollBreakTrendSections 四因子 RSI+Boll+Breakout+Trend 的参数与权重网格探索。
//
// 基于最优三因子叠加 Trend 趋势确认，网格约 100 组。
func (b *caseBuilder) addMultiQuadRsiBollBreakTrendSections() {
	for _, rp := range []int{5, 7} {
		for _, bp := range []int{12, 13} {
			for _, bm := range []float64{2.0, 2.2} {
				for _, kp := range []int{12, 15} {
					for _, tn := range []int{5, 6, 8, 12} {
						for _, tw := range []float64{0.25, 0.5, 0.75} {
							b.add(fmt.Sprintf("R%d+Bo%dM%.1f+Br%d+T%d_w%.2f", rp, bp, bm, kp, tn, tw), TestCase{
								UseRSI: true, RSIPeriod: rp, RSIOverbought: 80, RSIOversold: 20, RSIWeight: 1,
								UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: 1,
								UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: 1,
								UseTrend: true, TrendN: tn, TrendWeight: tw,
							})
						}
					}
				}
			}
		}
	}
}
