package cases

import "fmt"

// addRsiSingleSections RSI 单因子（原生方向正确：超卖看涨、超买看跌）
func (b *caseBuilder) addRsiSingleSections() {
	for _, p := range []int{5, 6, 7, 8, 9, 10, 14} {
		for _, th := range [][2]float64{{75, 25}, {78, 22}, {80, 20}, {70, 30}, {65, 35}} {
			b.add(fmt.Sprintf("RSI_P%d_%d_%d", p, int(th[0]), int(th[1])), TestCase{UseRSI: true, RSIPeriod: p, RSIOverbought: th[0], RSIOversold: th[1], RSIWeight: 1})
		}
	}
}

// addRsiSections RSI 单因子、RSI+Boll、RSI+Break、RSI+Boll+Break
func (b *caseBuilder) addRsiSections() {
	for _, p := range []int{5, 7, 9, 14} {
		for _, th := range [][2]float64{{75, 25}, {70, 30}, {65, 35}} {
			b.add(fmt.Sprintf("RSI_P%d_%d_%d", p, int(th[0]), int(th[1])), TestCase{UseRSI: true, RSIPeriod: p, RSIOverbought: th[0], RSIOversold: th[1], RSIWeight: 1})
		}
	}

	// SECTION 6: RSI+Boll 二因子网格 (18 cases)
	for _, rp := range []int{7, 14} {
		for _, th := range [][2]float64{{75, 25}, {70, 30}, {65, 35}} {
			for _, bm := range []float64{1.2, 1.5, 1.8} {
				b.add(fmt.Sprintf("R%d_%d+Bo20M%.0f", rp, int(th[0]), bm*10), TestCase{
					UseRSI: true, RSIPeriod: rp, RSIOverbought: th[0], RSIOversold: th[1], RSIWeight: 1,
					UseBoll: true, BollPeriod: 20, BollMultiplier: bm, BollWeight: 1,
				})
			}
		}
	}

	// SECTION 7: RSI+Boll+Break 三因子黄金组合网格 (30 cases)
	type rsiCfg struct{ p int; ob, os float64 }
	for _, rc := range []rsiCfg{{7, 75, 25}, {7, 70, 30}, {14, 70, 30}, {9, 70, 30}} {
		for _, bp := range []int{10, 15, 20} {
			for _, bm := range []float64{1.5, 2.0} {
				b.add(fmt.Sprintf("R%d_%d+Bo%dM%.0f+Br10", rc.p, int(rc.ob), bp, bm*10), TestCase{
					UseRSI: true, RSIPeriod: rc.p, RSIOverbought: rc.ob, RSIOversold: rc.os, RSIWeight: 1,
					UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: 1,
					UseBreakout: true, BreakoutPeriod: 10, BreakoutWeight: -1,
				})
			}
		}
	}

	// SECTION 8: 三因子 Break 周期变体 (8 cases)
	for _, kp := range []int{5, 7, 10, 12, 15, 18, 20, 25} {
		b.add(fmt.Sprintf("R7+Bo15M20+Br%d", kp), TestCase{
			UseRSI: true, RSIPeriod: 7, RSIOverbought: 75, RSIOversold: 25, RSIWeight: 1,
			UseBoll: true, BollPeriod: 15, BollMultiplier: 2.0, BollWeight: 1,
			UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: -1,
		})
	}

	// SECTION 9: 三因子权重比例调优 (12 cases)
	type w3 struct {
		rw, bw, kw float64
		tag        string
	}
	for _, w := range []w3{
		{1, 2, -1, "r1b2k1"}, {2, 1, -1, "r2b1k1"}, {2, 2, -1, "r2b2k1"},
		{1, 1, -2, "r1b1k2"}, {2, 2, -2, "r2b2k2"}, {1, 1, -3, "r1b1k3"},
		{3, 2, -1, "r3b2k1"}, {1, 3, -2, "r1b3k2"}, {3, 3, -3, "r3b3k3"},
	} {
		b.add(fmt.Sprintf("R7Bo15Br10_%s", w.tag), TestCase{
			UseRSI: true, RSIPeriod: 7, RSIOverbought: 75, RSIOversold: 25, RSIWeight: w.rw,
			UseBoll: true, BollPeriod: 15, BollMultiplier: 2.0, BollWeight: w.bw,
			UseBreakout: true, BreakoutPeriod: 10, BreakoutWeight: w.kw,
		})
	}

	// SECTION 12: RSI+Break 无Boll (8 cases)
	for _, rp := range []int{5, 7, 9, 14} {
		for _, kp := range []int{5, 10} {
			b.add(fmt.Sprintf("R%d_70+Br%d", rp, kp), TestCase{
				UseRSI: true, RSIPeriod: rp, RSIOverbought: 70, RSIOversold: 30, RSIWeight: 1,
				UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: -1,
			})
		}
	}
}
