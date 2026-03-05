package cases

import "fmt"

// addTrendSections Trend 单因子：统计根数 N（FactorTrend 已输出正向信号）
func (b *caseBuilder) addTrendSections() {
	for _, n := range []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 15, 20} {
		b.add(fmt.Sprintf("Trend_N%d", n), b.trend(n, 1))
	}
}
