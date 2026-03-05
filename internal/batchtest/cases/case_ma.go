package cases

import "fmt"

// addMaSections MA 单因子：短/长期均线组合 + 权重方向
func (b *caseBuilder) addMaSections() {
	// 20 轮迭代最优：MA_1_7 准确率 52.84%（FactorMA 已输出正向信号，权重仅用于多因子）
	for _, pair := range [][2]int{
		{1, 5}, {1, 6}, {1, 7}, {1, 8}, {1, 9}, {1, 10}, {1, 11}, {1, 12},
		{2, 5}, {2, 7}, {2, 9}, {3, 6}, {3, 7}, {3, 9}, {4, 8}, {5, 10},
	} {
		b.add(fmt.Sprintf("MA_%d_%d", pair[0], pair[1]), b.ma(pair[0], pair[1], 1))
	}
}
