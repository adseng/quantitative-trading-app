package cases

import "fmt"

// addBreakoutSections 突破单因子（FactorBreakout 已输出正向信号：跌破最低→看涨）
func (b *caseBuilder) addBreakoutSections() {
	for _, p := range []int{5, 8, 10, 12, 15, 18, 20, 25, 30} {
		b.add(fmt.Sprintf("Break_P%d", p), TestCase{UseBreakout: true, BreakoutPeriod: p, BreakoutWeight: 1})
	}
}
