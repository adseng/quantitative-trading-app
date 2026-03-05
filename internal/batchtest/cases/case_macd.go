package cases

import "fmt"

// addMacdSections MACD 单因子（FactorMACD 已输出正向信号）
func (b *caseBuilder) addMacdSections() {
	for _, p := range [][3]int{
		{5, 15, 5}, {6, 13, 5}, {5, 20, 7}, {12, 26, 5}, {8, 17, 6}, {10, 20, 8},
		{5, 15, 4}, {5, 15, 6}, {4, 12, 5}, {6, 15, 5},
	} {
		b.add(fmt.Sprintf("MACD_%d_%d_%d", p[0], p[1], p[2]), b.macd(p[0], p[1], p[2], 1))
	}
}
