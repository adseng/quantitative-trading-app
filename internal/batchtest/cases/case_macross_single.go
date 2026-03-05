package cases

import "fmt"

// addMacrossSingleSections MACross 单因子（FactorMACross 已输出正向信号：死叉→看涨）
func (b *caseBuilder) addMacrossSingleSections() {
	for _, params := range [][2]int{
		{5, 10}, {5, 20}, {5, 30}, {10, 20}, {10, 30}, {15, 30}, {20, 30},
		{20, 60}, {30, 60}, {50, 120}, {50, 200}, {8, 20}, {8, 30}, {12, 30},
	} {
		b.add(fmt.Sprintf("MACrS_%d_%d", params[0], params[1]), b.macross(params[0], params[1], 1, 2, 0))
	}
	b.add("MACrS_5_20_p", b.macross(5, 20, 1, 2, 0.002))
	b.add("MACrS_10_30_p", b.macross(10, 30, 1, 3, 0.002))
}
