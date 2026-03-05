package cases

import "fmt"

// addPriceVsMASections 价格 vs SMA 单因子（FactorPriceVsMA 已输出正向信号）
func (b *caseBuilder) addPriceVsMASections() {
	for _, p := range []int{5, 6, 7, 8, 10, 12, 15, 20, 25} {
		b.add(fmt.Sprintf("PriceVsMA_P%d", p), b.priceVsMA(p, 1))
	}
}
