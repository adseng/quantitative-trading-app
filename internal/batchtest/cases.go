package batchtest

import "fmt"

// TestCase defines a single backtest parameter combination.
type TestCase struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	UseMA    bool    `json:"useMA"`
	MaShort  int     `json:"maShort"`
	MaLong   int     `json:"maLong"`
	MaWeight float64 `json:"maWeight"`

	UseTrend    bool    `json:"useTrend"`
	TrendN      int     `json:"trendN"`
	TrendWeight float64 `json:"trendWeight"`

	UseRSI        bool    `json:"useRSI"`
	RSIPeriod     int     `json:"rsiPeriod"`
	RSIOverbought float64 `json:"rsiOverbought"`
	RSIOversold   float64 `json:"rsiOversold"`
	RSIWeight     float64 `json:"rsiWeight"`

	UseMACD    bool    `json:"useMACD"`
	MACDFast   int     `json:"macdFast"`
	MACDSlow   int     `json:"macdSlow"`
	MACDSignal int     `json:"macdSignal"`
	MACDWeight float64 `json:"macdWeight"`

	UseBoll        bool    `json:"useBoll"`
	BollPeriod     int     `json:"bollPeriod"`
	BollMultiplier float64 `json:"bollMultiplier"`
	BollWeight     float64 `json:"bollWeight"`

	UseBreakout    bool    `json:"useBreakout"`
	BreakoutPeriod int     `json:"breakoutPeriod"`
	BreakoutWeight float64 `json:"breakoutWeight"`

	UsePriceVsMA    bool    `json:"usePriceVsMA"`
	PriceVsMAPeriod int     `json:"priceVsMAPeriod"`
	PriceVsMAWeight float64 `json:"priceVsMAWeight"`

	UseATR    bool    `json:"useATR"`
	ATRPeriod int     `json:"atrPeriod"`
	ATRWeight float64 `json:"atrWeight"`

	UseVolume    bool    `json:"useVolume"`
	VolumePeriod int     `json:"volumePeriod"`
	VolumeWeight float64 `json:"volumeWeight"`

	UseSession    bool    `json:"useSession"`
	SessionWeight float64 `json:"sessionWeight"`

	UseMACross     bool    `json:"useMACross"`
	MACrossShort   int     `json:"macrossShort"`
	MACrossLong    int     `json:"macrossLong"`
	MACrossWeight  float64 `json:"macrossWeight"`
	MACrossWindow  int     `json:"macrossWindow"`
	MACrossPreempt float64 `json:"macrossPreempt"`
}

// TestResult holds one test case result.
type TestResult struct {
	TestCase       TestCase `json:"testCase"`
	Accuracy       float64  `json:"accuracy"`
	Correct        int      `json:"correct"`
	Total          int      `json:"total"`
	SignalCount    int      `json:"signalCount"`
	SignalAccuracy float64  `json:"signalAccuracy"`
	AvgScore       float64  `json:"avgScore"`
	AvgAbsScore    float64  `json:"avgAbsScore"`
}

// V4 test plan — focused on the golden zone discovered in V3:
//
// V3 best performers:
//   Boll15_2.0(+1): 58.1% @ 10.9%   Boll+Break-: 57.2% @ 18.5%
//   Break5(-1): 56.7% @ 21.4%        Boll1.5+Break: 55.9% @ 30.3%
//   RSI7+Boll15+Break: 55.5% @ 32.9% RSI7+Boll1.5: 55.3% @ 37.5%
//
// Strategy: fine-tune Boll/Break/RSI parameters around the sweet spot.

func GenerateTestCases() []TestCase {
	cases := make([]TestCase, 0, 200)
	id := 0
	add := func(name string, c TestCase) {
		id++
		c.ID = id
		c.Name = name
		cases = append(cases, c)
	}

	boll := func(period int, mult float64) TestCase {
		return TestCase{UseBoll: true, BollPeriod: period, BollMultiplier: mult, BollWeight: 1}
	}
	brk := func(period int) TestCase {
		return TestCase{UseBreakout: true, BreakoutPeriod: period, BreakoutWeight: -1}
	}
	rsi := func(period int, ob, os float64) TestCase {
		return TestCase{UseRSI: true, RSIPeriod: period, RSIOverbought: ob, RSIOversold: os, RSIWeight: 1}
	}

	// ============================================================
	// SECTION 1: Boll 单因子参数精扫 (30 cases)
	// Period: 8,10,12,15,18,20  ×  Multiplier: 1.0,1.2,1.5,1.8,2.0
	// ============================================================
	for _, p := range []int{8, 10, 12, 15, 18, 20} {
		for _, m := range []float64{1.0, 1.2, 1.5, 1.8, 2.0} {
			name := fmt.Sprintf("Boll_P%d_M%.1f", p, m)
			add(name, TestCase{UseBoll: true, BollPeriod: p, BollMultiplier: m, BollWeight: 1})
		}
	}

	// ============================================================
	// SECTION 2: Breakout 单因子周期精扫 (12 cases)
	// Period: 3,5,7,8,10,12,15,18,20,25,30,40
	// ============================================================
	for _, p := range []int{3, 5, 7, 8, 10, 12, 15, 18, 20, 25, 30, 40} {
		name := fmt.Sprintf("Break_P%d", p)
		add(name, TestCase{UseBreakout: true, BreakoutPeriod: p, BreakoutWeight: -1})
	}

	// ============================================================
	// SECTION 3: Boll+Break 二因子参数网格 (36 cases)
	// Boll: P(10,15,20) × M(1.2,1.5,1.8,2.0)
	// Break: P(5,10,15)
	// ============================================================
	for _, bp := range []int{10, 15, 20} {
		for _, bm := range []float64{1.2, 1.5, 1.8, 2.0} {
			for _, kp := range []int{5, 10, 15} {
				name := fmt.Sprintf("Bo%dM%.0f+Br%d", bp, bm*10, kp)
				b := boll(bp, bm)
				k := brk(kp)
				add(name, TestCase{
					UseBoll: true, BollPeriod: b.BollPeriod, BollMultiplier: b.BollMultiplier, BollWeight: 1,
					UseBreakout: true, BreakoutPeriod: k.BreakoutPeriod, BreakoutWeight: -1,
				})
			}
		}
	}

	// ============================================================
	// SECTION 4: Boll+Break 权重比例 (6 cases，精简)
	// ============================================================
	for _, bw := range []float64{1, 2} {
		for _, kw := range []float64{-1, -2, -3} {
			name := fmt.Sprintf("Bo15M2w%d+Br10w%d", int(bw), int(kw))
			add(name, TestCase{
				UseBoll: true, BollPeriod: 15, BollMultiplier: 2.0, BollWeight: bw,
				UseBreakout: true, BreakoutPeriod: 10, BreakoutWeight: kw,
			})
		}
	}

	// ============================================================
	// SECTION 5: RSI 单因子参数精扫 (12 cases)
	// Period: 5,7,9,14  ×  Thresholds: 75/25, 70/30, 65/35
	// ============================================================
	for _, p := range []int{5, 7, 9, 14} {
		for _, th := range [][2]float64{{75, 25}, {70, 30}, {65, 35}} {
			name := fmt.Sprintf("RSI_P%d_%d_%d", p, int(th[0]), int(th[1]))
			add(name, TestCase{UseRSI: true, RSIPeriod: p, RSIOverbought: th[0], RSIOversold: th[1], RSIWeight: 1})
		}
	}

	// ============================================================
	// SECTION 6: RSI+Boll 二因子网格 (18 cases)
	// RSI: P(7,14) × Th(75/25, 70/30, 65/35)
	// Boll: M(1.2, 1.5, 1.8)
	// ============================================================
	for _, rp := range []int{7, 14} {
		for _, th := range [][2]float64{{75, 25}, {70, 30}, {65, 35}} {
			for _, bm := range []float64{1.2, 1.5, 1.8} {
				name := fmt.Sprintf("R%d_%d+Bo20M%.0f", rp, int(th[0]), bm*10)
				add(name, TestCase{
					UseRSI: true, RSIPeriod: rp, RSIOverbought: th[0], RSIOversold: th[1], RSIWeight: 1,
					UseBoll: true, BollPeriod: 20, BollMultiplier: bm, BollWeight: 1,
				})
			}
		}
	}

	// ============================================================
	// SECTION 7: RSI+Boll+Break 三因子黄金组合网格 (30 cases)
	// RSI: (7,75/25), (7,70/30), (14,70/30), (14,65/35), (9,70/30)
	// Boll: P(10,15,20) × M(1.5,2.0)
	// Break: P10
	// ============================================================
	type rsiCfg struct {
		p  int
		ob float64
		os float64
	}
	rsiList := []rsiCfg{
		{7, 75, 25}, {7, 70, 30}, {14, 70, 30}, {9, 70, 30},
	}
	for _, rc := range rsiList {
		for _, bp := range []int{10, 15, 20} {
			for _, bm := range []float64{1.5, 2.0} {
				name := fmt.Sprintf("R%d_%d+Bo%dM%.0f+Br10", rc.p, int(rc.ob), bp, bm*10)
				add(name, TestCase{
					UseRSI: true, RSIPeriod: rc.p, RSIOverbought: rc.ob, RSIOversold: rc.os, RSIWeight: 1,
					UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: 1,
					UseBreakout: true, BreakoutPeriod: 10, BreakoutWeight: -1,
				})
			}
		}
	}

	// ============================================================
	// SECTION 8: 三因子 Break 周期变体 (8 cases)
	// ============================================================
	for _, kp := range []int{5, 7, 10, 12, 15, 18, 20, 25} {
		name := fmt.Sprintf("R7+Bo15M20+Br%d", kp)
		add(name, TestCase{
			UseRSI: true, RSIPeriod: 7, RSIOverbought: 75, RSIOversold: 25, RSIWeight: 1,
			UseBoll: true, BollPeriod: 15, BollMultiplier: 2.0, BollWeight: 1,
			UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: -1,
		})
	}

	// ============================================================
	// SECTION 9: 三因子权重比例调优 (12 cases)
	// 基于 RSI7+Boll15M2+Break10
	// ============================================================
	type w3 struct {
		rw, bw, kw float64
		tag        string
	}
	w3List := []w3{
		{1, 2, -1, "r1b2k1"}, {2, 1, -1, "r2b1k1"}, {2, 2, -1, "r2b2k1"},
		{1, 1, -2, "r1b1k2"}, {2, 2, -2, "r2b2k2"}, {1, 1, -3, "r1b1k3"},
		{3, 2, -1, "r3b2k1"}, {1, 3, -2, "r1b3k2"}, {3, 3, -3, "r3b3k3"},
	}
	for _, w := range w3List {
		name := fmt.Sprintf("R7Bo15Br10_%s", w.tag)
		add(name, TestCase{
			UseRSI: true, RSIPeriod: 7, RSIOverbought: 75, RSIOversold: 25, RSIWeight: w.rw,
			UseBoll: true, BollPeriod: 15, BollMultiplier: 2.0, BollWeight: w.bw,
			UseBreakout: true, BreakoutPeriod: 10, BreakoutWeight: w.kw,
		})
	}

	// ============================================================
	// SECTION 10: 四因子 - Top3+ATR/Vol (16 cases)
	// ============================================================
	_ = rsi
	_ = brk

	// RSI7+Boll+Break+ATR
	for _, atrP := range []int{7, 14} {
		for _, bp := range []int{10, 15} {
			for _, bm := range []float64{1.5, 2.0} {
				name := fmt.Sprintf("R7+Bo%dM%.0f+Br10+A%d", bp, bm*10, atrP)
				add(name, TestCase{
					UseRSI: true, RSIPeriod: 7, RSIOverbought: 75, RSIOversold: 25, RSIWeight: 1,
					UseBoll: true, BollPeriod: bp, BollMultiplier: bm, BollWeight: 1,
					UseBreakout: true, BreakoutPeriod: 10, BreakoutWeight: -1,
					UseATR: true, ATRPeriod: atrP, ATRWeight: -1,
				})
			}
		}
	}

	// RSI7+Boll+Break+Vol
	for _, vp := range []int{10, 20} {
		for _, bp := range []int{10, 15} {
			name := fmt.Sprintf("R7+Bo%dM20+Br10+V%d", bp, vp)
			add(name, TestCase{
				UseRSI: true, RSIPeriod: 7, RSIOverbought: 75, RSIOversold: 25, RSIWeight: 1,
				UseBoll: true, BollPeriod: bp, BollMultiplier: 2.0, BollWeight: 1,
				UseBreakout: true, BreakoutPeriod: 10, BreakoutWeight: -1,
				UseVolume: true, VolumePeriod: vp, VolumeWeight: -1,
			})
		}
	}

	// Boll+Break+ATR (无RSI)
	for _, atrP := range []int{7, 14} {
		name := fmt.Sprintf("Bo15M20+Br10+A%d", atrP)
		add(name, TestCase{
			UseBoll: true, BollPeriod: 15, BollMultiplier: 2.0, BollWeight: 1,
			UseBreakout: true, BreakoutPeriod: 10, BreakoutWeight: -1,
			UseATR: true, ATRPeriod: atrP, ATRWeight: -1,
		})
	}

	// ============================================================
	// SECTION 11: Boll 窄倍数 (7 cases)
	// ============================================================
	for _, p := range []int{10, 15, 20} {
		for _, m := range []float64{1.0, 1.1} {
			name := fmt.Sprintf("Boll_P%d_M%.1f", p, m)
			add(name, TestCase{UseBoll: true, BollPeriod: p, BollMultiplier: m, BollWeight: 1})
		}
	}
	add("Boll_P20_M1.3", TestCase{UseBoll: true, BollPeriod: 20, BollMultiplier: 1.3, BollWeight: 1})

	// ============================================================
	// SECTION 12: RSI+Break 无Boll (8 cases)
	// ============================================================
	for _, rp := range []int{5, 7, 9, 14} {
		for _, kp := range []int{5, 10} {
			name := fmt.Sprintf("R%d_70+Br%d", rp, kp)
			add(name, TestCase{
				UseRSI: true, RSIPeriod: rp, RSIOverbought: 70, RSIOversold: 30, RSIWeight: 1,
				UseBreakout: true, BreakoutPeriod: kp, BreakoutWeight: -1,
			})
		}
	}

	// ============================================================
	// SECTION 12b: 金叉/死叉（事件型+时间容错+预判）
	// window=容错根数(左邻有效)，preempt=预判阈值(0=关)
	// ============================================================
	for _, params := range [][3]int{
		{5, 10, 1}, {5, 20, 1}, {5, 30, 1},
		{10, 20, 1}, {10, 30, 1}, {15, 30, 1}, {20, 30, 1},
		{20, 60, 1}, {30, 60, 1}, {50, 120, 1}, {50, 200, 1},
	} {
		name := fmt.Sprintf("MACr%d_%d_w2", params[0], params[1])
		add(name, TestCase{
			UseMACross: true, MACrossShort: params[0], MACrossLong: params[1], MACrossWeight: float64(params[2]),
			MACrossWindow: 2, MACrossPreempt: 0,
		})
	}
	for _, params := range [][3]int{
		{5, 20, -1}, {10, 30, -1}, {20, 60, -1},
	} {
		name := fmt.Sprintf("MACr%d_%d_neg", params[0], params[1])
		add(name, TestCase{
			UseMACross: true, MACrossShort: params[0], MACrossLong: params[1], MACrossWeight: float64(params[2]),
			MACrossWindow: 2, MACrossPreempt: 0,
		})
	}
	// 容错+预判组合
	add("MACr5_20_w2p", TestCase{UseMACross: true, MACrossShort: 5, MACrossLong: 20, MACrossWeight: 1, MACrossWindow: 2, MACrossPreempt: 0.002})
	add("MACr10_30_w3p", TestCase{UseMACross: true, MACrossShort: 10, MACrossLong: 30, MACrossWeight: 1, MACrossWindow: 3, MACrossPreempt: 0.002})
	add("MACr5_20_neg_w2p", TestCase{UseMACross: true, MACrossShort: 5, MACrossLong: 20, MACrossWeight: -1, MACrossWindow: 2, MACrossPreempt: 0.002})

	// ============================================================
	// SECTION 13: 五因子+ 最优方向组合 (6 cases)
	// ============================================================
	add("R7+Bo15M20+Br10+A7+V10", TestCase{
		UseRSI: true, RSIPeriod: 7, RSIOverbought: 75, RSIOversold: 25, RSIWeight: 1,
		UseBoll: true, BollPeriod: 15, BollMultiplier: 2.0, BollWeight: 1,
		UseBreakout: true, BreakoutPeriod: 10, BreakoutWeight: -1,
		UseATR: true, ATRPeriod: 7, ATRWeight: -1,
		UseVolume: true, VolumePeriod: 10, VolumeWeight: -1,
	})
	add("R7+Bo15M20+Br10+A7+Sess", TestCase{
		UseRSI: true, RSIPeriod: 7, RSIOverbought: 75, RSIOversold: 25, RSIWeight: 1,
		UseBoll: true, BollPeriod: 15, BollMultiplier: 2.0, BollWeight: 1,
		UseBreakout: true, BreakoutPeriod: 10, BreakoutWeight: -1,
		UseATR: true, ATRPeriod: 7, ATRWeight: -1,
		UseSession: true, SessionWeight: -1,
	})
	add("Bo15M15+Br5+A7+V10", TestCase{
		UseBoll: true, BollPeriod: 15, BollMultiplier: 1.5, BollWeight: 1,
		UseBreakout: true, BreakoutPeriod: 5, BreakoutWeight: -1,
		UseATR: true, ATRPeriod: 7, ATRWeight: -1,
		UseVolume: true, VolumePeriod: 10, VolumeWeight: -1,
	})
	add("R7+Bo10M15+Br5+A7", TestCase{
		UseRSI: true, RSIPeriod: 7, RSIOverbought: 75, RSIOversold: 25, RSIWeight: 1,
		UseBoll: true, BollPeriod: 10, BollMultiplier: 1.5, BollWeight: 1,
		UseBreakout: true, BreakoutPeriod: 5, BreakoutWeight: -1,
		UseATR: true, ATRPeriod: 7, ATRWeight: -1,
	})
	add("ALL_opt_direction", TestCase{
		UseMA: true, MaShort: 5, MaLong: 20, MaWeight: -0.5,
		UseTrend: true, TrendN: 5, TrendWeight: -0.5,
		UseRSI: true, RSIPeriod: 7, RSIOverbought: 75, RSIOversold: 25, RSIWeight: 2,
		UseMACD: true, MACDFast: 12, MACDSlow: 26, MACDSignal: 9, MACDWeight: -0.5,
		UseBoll: true, BollPeriod: 15, BollMultiplier: 2.0, BollWeight: 3,
		UseBreakout: true, BreakoutPeriod: 10, BreakoutWeight: -2,
		UsePriceVsMA: true, PriceVsMAPeriod: 10, PriceVsMAWeight: -0.5,
		UseATR: true, ATRPeriod: 7, ATRWeight: -1,
		UseVolume: true, VolumePeriod: 10, VolumeWeight: -1,
		UseSession: true, SessionWeight: -0.5,
	})
	add("Top3_strong_weight", TestCase{
		UseRSI: true, RSIPeriod: 7, RSIOverbought: 75, RSIOversold: 25, RSIWeight: 3,
		UseBoll: true, BollPeriod: 15, BollMultiplier: 2.0, BollWeight: 5,
		UseBreakout: true, BreakoutPeriod: 10, BreakoutWeight: -4,
	})

	if len(cases) > 200 {
		cases = cases[:200]
	}
	return cases
}

// ensure fmt import is used
var _ = fmt.Sprintf
