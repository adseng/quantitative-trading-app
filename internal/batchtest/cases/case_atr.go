package cases

import "fmt"

// addAtrSections ATR 单因子（FactorATR 已输出正向信号：放大+跌→看涨）
func (b *caseBuilder) addAtrSections() {
	for _, p := range []int{5, 7, 10, 12, 14, 18, 20, 25} {
		b.add(fmt.Sprintf("ATR_P%d", p), b.atr(p, 1))
	}
}
