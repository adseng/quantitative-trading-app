package cases

import "fmt"

// addBollBreakSections Boll+Break 二因子
func (b *caseBuilder) addBollBreakSections() {
	// SECTION 3: Boll+Break 二因子参数网格 (36 cases)
	for _, bp := range []int{10, 15, 20} {
		for _, bm := range []float64{1.2, 1.5, 1.8, 2.0} {
			for _, kp := range []int{5, 10, 15} {
				name := fmt.Sprintf("Bo%dM%.0f+Br%d", bp, bm*10, kp)
				b.add(name, TestCase{
					UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: 1,
					UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: 1,
				})
			}
		}
	}

	// SECTION 4: Boll+Break 权重比例 (6 cases)
	for _, bw := range []float64{1, 2} {
		for _, kw := range []float64{1, 2, 3} {
			b.add(fmt.Sprintf("Bo15M2w%d+Br10w%d", int(bw), int(kw)), TestCase{
				UseBoll: true, BollPeriod: 15, BollMultiplier: 2.0, BollWeight: bw,
				UseBreakout: true, BreakoutPeriod: 10, BreakoutWeight: kw,
			})
		}
	}
}
