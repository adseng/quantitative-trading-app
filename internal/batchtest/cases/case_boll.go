package cases

import "fmt"

// addBollSections 布林带单因子（已输出正向信号：下轨超卖看涨、上轨超买看跌）
func (b *caseBuilder) addBollSections() {
	for _, p := range []int{9, 10, 11, 12, 13, 15} {
		for _, m := range []float64{2.0, 2.2, 2.4} {
			b.add(fmt.Sprintf("Boll_P%d_M%.1f", p, m), TestCase{UseBoll: true, BollPeriod: p, BollMultiplier: m, BollWeight: 1})
		}
	}
}
