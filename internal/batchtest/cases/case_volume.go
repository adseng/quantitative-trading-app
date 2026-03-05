package cases

import "fmt"

// addVolumeSections Volume 单因子（FactorVolume 已输出正向信号）
func (b *caseBuilder) addVolumeSections() {
	for _, p := range []int{5, 8, 10, 15, 20, 25, 30, 40} {
		b.add(fmt.Sprintf("Vol_P%d", p), b.volume(p, 1))
	}
}
