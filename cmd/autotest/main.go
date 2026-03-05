package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"quantitative-trading-app/internal/batchtest"
	"quantitative-trading-app/internal/datafile"
	"quantitative-trading-app/internal/factor"
)

const casesPerIter = 200

type iterSummary struct {
	iter      int
	bestAcc   float64
	bestName  string
	top10Avg  float64
	avgAcc    float64
	excelFile string
	elapsed   time.Duration
}

func main() {
	loops := flag.Int("n", 1, "number of evolution iterations")
	dataPath := flag.String("data", datafile.DefaultPath, "path to kline data file")
	flag.Parse()

	fmt.Println("=== Evolutionary Auto Test ===")
	fmt.Printf("总轮数: %d | 每轮用例: %d\n", *loops, casesPerIter)
	fmt.Printf("数据文件: %s\n\n", *dataPath)

	fmt.Println("[加载数据] 读取本地K线文件...")
	klines, err := datafile.LoadKlines(*dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载失败: %v\n先运行: go run ./cmd/fetchdata\n", err)
		os.Exit(1)
	}
	fmt.Printf("[加载数据] 完成，共 %d 根K线 (%.0f天)\n\n", len(klines), float64(len(klines))*15/60/24)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var summaries []iterSummary
	currentCases := batchtest.GenerateTestCases()
	var globalBestAcc float64
	var globalBestName string
	staleCount := 0

	for iter := 1; iter <= *loops; iter++ {
		fmt.Printf("┌─────────────────────────────────────────┐\n")
		fmt.Printf("│          第 %2d / %2d 轮                   │\n", iter, *loops)
		fmt.Printf("└─────────────────────────────────────────┘\n")

		if iter == 1 {
			fmt.Println("[策略] 使用 V4 手工设计的初始种群 (200组)")
			fmt.Println("  思路: Boll/Break/RSI 三大核心因子的参数精细网格搜索")
		} else {
			printEvolutionStrategy(summaries, staleCount)
		}

		s, results := runOneIteration(iter, klines, currentCases)
		summaries = append(summaries, s)

		prevBest := globalBestAcc
		if s.bestAcc > globalBestAcc {
			globalBestAcc = s.bestAcc
			globalBestName = s.bestName
			staleCount = 0
			fmt.Printf("  ★ 发现新的全局最优! %.2f%% → %.2f%% (+%.2f%%)\n",
				prevBest*100, globalBestAcc*100, (globalBestAcc-prevBest)*100)
		} else {
			staleCount++
			fmt.Printf("  - 本轮未突破全局最优 (已连续 %d 轮停滞)\n", staleCount)
		}
		fmt.Printf("  本轮最佳: %.2f%% (%s)\n", s.bestAcc*100, s.bestName)
		fmt.Printf("  全局最佳: %.2f%% (%s)\n", globalBestAcc*100, globalBestName)

		if iter < *loops {
			fmt.Println("\n[进化] 生成下一代种群...")
			currentCases = evolve(results, rng, staleCount)
		}
		fmt.Println()
	}

	printTrend(summaries, globalBestAcc, globalBestName)

	trendFile := filepath.Join(batchtest.ExcelDirPath(),
		fmt.Sprintf("trend_%s.txt", time.Now().Format("20060102150405")))
	writeTrendFile(trendFile, summaries, globalBestAcc, globalBestName)
	fmt.Printf("趋势文件: %s\n", trendFile)
}

func printEvolutionStrategy(summaries []iterSummary, staleCount int) {
	last := summaries[len(summaries)-1]
	fmt.Printf("[策略] 上轮 best=%.2f%% avg=%.2f%%\n", last.bestAcc*100, last.avgAcc*100)

	if staleCount == 0 {
		fmt.Println("  思路: 上轮有提升 → 在最优参数附近继续精细搜索")
		fmt.Println("  构成: 20精英 + 120变异(小幅微调) + 40交叉 + 20随机探索")
	} else if staleCount < 5 {
		fmt.Printf("  思路: 已停滞%d轮 → 保持精英，加大变异幅度寻找新方向\n", staleCount)
		fmt.Println("  构成: 20精英 + 120变异(扩大搜索范围) + 40交叉 + 20随机")
	} else {
		fmt.Printf("  思路: 停滞%d轮 → 可能已接近全局最优或陷入局部最优\n", staleCount)
		fmt.Println("  构成: 20精英 + 120变异(大幅探索) + 40交叉 + 20随机")
	}
}

func runOneIteration(iter int, klines []*factor.KLine, cases []batchtest.TestCase) (iterSummary, []batchtest.TestResult) {
	runner := batchtest.NewBatchRunnerWithCases(cases)
	fmt.Printf("[回测] 开始 %d 组参数回测...\n", len(cases))
	start := time.Now()

	runner.Run(klines, func(evt batchtest.ProgressEvent) {
		if evt.Phase == "testing" && evt.CaseIndex%50 == 0 {
			fmt.Printf("  进度 %s\n", evt.Message)
		}
		if evt.Phase == "done" {
			fmt.Printf("  %s\n", evt.Message)
		}
	})
	elapsed := time.Since(start)
	results := runner.Results()
	a := analyzeResults(results)

	fmt.Printf("[回测] 完成 (%s) | Excel: %s\n", elapsed.Round(time.Millisecond), runner.ExcelFile())
	fmt.Printf("[分析] %d组 | best=%.2f%% top10avg=%.2f%% avg=%.2f%% worst=%.2f%%\n",
		a.total, a.maxSigAcc*100, a.top10Avg*100, a.avgSigAcc*100, a.minSigAcc*100)

	fmt.Println("[分析] Top 5 策略:")
	for i := 0; i < 5 && i < len(a.top10); i++ {
		r := a.top10[i]
		rate := 0.0
		if r.Total > 0 {
			rate = float64(r.SignalCount) / float64(r.Total) * 100
		}
		fmt.Printf("    %d. %-38s %.2f%% (信号率%.1f%%, 强度%.2f)\n",
			i+1, r.TestCase.Name, r.SignalAccuracy*100, rate, r.AvgAbsScore)
	}

	printFactorDistribution(results)

	logFile := filepath.Join(batchtest.ExcelDirPath(),
		fmt.Sprintf("analysis_iter%d_%s.txt", iter, time.Now().Format("20060102150405")))
	writeAnalysisLog(logFile, iter, runner.ExcelFile(), a, elapsed)

	s := iterSummary{
		iter:      iter,
		bestAcc:   a.maxSigAcc,
		avgAcc:    a.avgSigAcc,
		excelFile: runner.ExcelFile(),
		elapsed:   elapsed,
	}
	if len(a.top10) > 0 {
		s.bestName = a.top10[0].TestCase.Name
		s.top10Avg = a.top10Avg
	}
	return s, results
}

func printFactorDistribution(results []batchtest.TestResult) {
	sorted := make([]batchtest.TestResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SignalAccuracy > sorted[j].SignalAccuracy
	})

	top20 := sorted
	if len(top20) > 20 {
		top20 = sorted[:20]
	}

	bollCnt, brkCnt, rsiCnt, atrCnt, volCnt, otherCnt := 0, 0, 0, 0, 0, 0
	for _, r := range top20 {
		tc := r.TestCase
		if tc.UseBoll {
			bollCnt++
		}
		if tc.UseBreakout {
			brkCnt++
		}
		if tc.UseRSI {
			rsiCnt++
		}
		if tc.UseATR {
			atrCnt++
		}
		if tc.UseVolume {
			volCnt++
		}
		if tc.UseMA || tc.UseTrend || tc.UseMACD || tc.UsePriceVsMA || tc.UseSession || tc.UseMACross {
			otherCnt++
		}
	}
	fmt.Printf("[分析] Top20因子分布: Boll=%d Break=%d RSI=%d ATR=%d Vol=%d 其他=%d\n",
		bollCnt, brkCnt, rsiCnt, atrCnt, volCnt, otherCnt)
}

// ======================== EVOLUTION ========================

func evolve(results []batchtest.TestResult, rng *rand.Rand, staleCount int) []batchtest.TestCase {
	sorted := make([]batchtest.TestResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SignalAccuracy > sorted[j].SignalAccuracy
	})

	cases := make([]batchtest.TestCase, 0, casesPerIter)
	id := 0
	add := func(prefix string, tc batchtest.TestCase) {
		id++
		tc.ID = id
		tc.Name = prefix + genName(tc)
		if len(tc.Name) > 45 {
			tc.Name = tc.Name[:45]
		}
		cases = append(cases, tc)
	}

	elite := minInt(20, len(sorted))
	for i := 0; i < elite; i++ {
		add("E:", sorted[i].TestCase)
	}

	mutParentPool := minInt(30, len(sorted))
	if staleCount >= 5 {
		mutParentPool = minInt(50, len(sorted))
	}

	for len(cases) < elite+120 {
		pi := rng.Intn(mutParentPool)
		child := mutate(sorted[pi].TestCase, rng, staleCount)
		child = ensureValid(child)
		add("M:", child)
	}

	crossPool := minInt(20, len(sorted))
	for len(cases) < elite+120+40 && len(sorted) >= 2 {
		p1 := sorted[rng.Intn(crossPool)].TestCase
		p2 := sorted[rng.Intn(crossPool)].TestCase
		child := crossover(p1, p2, rng)
		child = ensureValid(child)
		add("X:", child)
	}

	for len(cases) < casesPerIter {
		child := randomCase(rng)
		child = ensureValid(child)
		add("R:", child)
	}

	bestTc := sorted[0].TestCase
	fmt.Printf("  精英保留: %d | 变异(源自top%d): %d | 交叉: %d | 随机: %d\n",
		elite, mutParentPool, 120, 40, casesPerIter-elite-120-40)
	fmt.Printf("  当前最优种子: %s (Boll P%d M%.1f",
		sorted[0].TestCase.Name, bestTc.BollPeriod, bestTc.BollMultiplier)
	if bestTc.UseBreakout {
		fmt.Printf(" + Break P%d", bestTc.BreakoutPeriod)
	}
	if bestTc.UseRSI {
		fmt.Printf(" + RSI P%d", bestTc.RSIPeriod)
	}
	fmt.Println(")")

	return cases
}

func mutate(p batchtest.TestCase, rng *rand.Rand, staleCount int) batchtest.TestCase {
	c := p

	scale := 1.0
	if staleCount >= 5 {
		scale = 1.5
	} else if staleCount >= 10 {
		scale = 2.0
	}

	if c.UseBoll {
		if rng.Float64() < 0.4 {
			c.BollPeriod += int(float64(rng.Intn(5)-2) * scale)
		}
		if rng.Float64() < 0.4 {
			c.BollMultiplier += (rng.Float64() - 0.5) * 0.4 * scale
		}
		if rng.Float64() < 0.2 {
			c.BollWeight *= 0.8 + rng.Float64()*0.4
		}
	}

	if c.UseBreakout {
		if rng.Float64() < 0.4 {
			c.BreakoutPeriod += int(float64(rng.Intn(7)-3) * scale)
		}
		if rng.Float64() < 0.2 {
			c.BreakoutWeight *= 0.8 + rng.Float64()*0.4
		}
	}

	if c.UseRSI {
		if rng.Float64() < 0.4 {
			c.RSIPeriod += int(float64(rng.Intn(5)-2) * scale)
		}
		if rng.Float64() < 0.25 {
			c.RSIOverbought += (rng.Float64() - 0.5) * 6 * scale
			c.RSIOversold += (rng.Float64() - 0.5) * 6 * scale
		}
		if rng.Float64() < 0.2 {
			c.RSIWeight *= 0.8 + rng.Float64()*0.4
		}
	}

	if c.UseATR && rng.Float64() < 0.3 {
		c.ATRPeriod += int(float64(rng.Intn(5)-2) * scale)
	}
	if c.UseVolume && rng.Float64() < 0.3 {
		c.VolumePeriod += int(float64(rng.Intn(7)-3) * scale)
	}
	if c.UseMACD && rng.Float64() < 0.25 {
		c.MACDFast += rng.Intn(3) - 1
		c.MACDSlow += rng.Intn(5) - 2
	}

	toggleProb := 0.08
	if staleCount >= 5 {
		toggleProb = 0.15
	}
	if rng.Float64() < toggleProb {
		switch rng.Intn(7) {
		case 0:
			c.UseATR = !c.UseATR
			if c.UseATR {
				c.ATRPeriod = 7 + rng.Intn(8)
				c.ATRWeight = -1
			}
		case 1:
			c.UseVolume = !c.UseVolume
			if c.UseVolume {
				c.VolumePeriod = 10 + rng.Intn(15)
				c.VolumeWeight = -1
			}
		case 2:
			if !c.UseRSI {
				c.UseRSI = true
				c.RSIPeriod = 7
				c.RSIOverbought = 75
				c.RSIOversold = 25
				c.RSIWeight = 1
			}
		case 3:
			if !c.UseBoll {
				c.UseBoll = true
				c.BollPeriod = 15
				c.BollMultiplier = 2.0
				c.BollWeight = 1
			}
		case 4:
			if !c.UseBreakout {
				c.UseBreakout = true
				c.BreakoutPeriod = 10
				c.BreakoutWeight = -1
			}
		case 5:
			c.UseSession = !c.UseSession
			if c.UseSession {
				c.SessionWeight = -1
			}
		case 6:
			c.UseMACross = !c.UseMACross
			if c.UseMACross {
				c.MACrossShort = 5
				c.MACrossLong = 20
				c.MACrossWeight = 1
			}
		}
	}

	return c
}

func crossover(p1, p2 batchtest.TestCase, rng *rand.Rand) batchtest.TestCase {
	c := p1
	if rng.Float64() < 0.5 {
		c.UseBoll = p2.UseBoll
		c.BollPeriod = p2.BollPeriod
		c.BollMultiplier = p2.BollMultiplier
		c.BollWeight = p2.BollWeight
	}
	if rng.Float64() < 0.5 {
		c.UseBreakout = p2.UseBreakout
		c.BreakoutPeriod = p2.BreakoutPeriod
		c.BreakoutWeight = p2.BreakoutWeight
	}
	if rng.Float64() < 0.5 {
		c.UseRSI = p2.UseRSI
		c.RSIPeriod = p2.RSIPeriod
		c.RSIOverbought = p2.RSIOverbought
		c.RSIOversold = p2.RSIOversold
		c.RSIWeight = p2.RSIWeight
	}
	if rng.Float64() < 0.5 {
		c.UseATR = p2.UseATR
		c.ATRPeriod = p2.ATRPeriod
		c.ATRWeight = p2.ATRWeight
	}
	if rng.Float64() < 0.5 {
		c.UseVolume = p2.UseVolume
		c.VolumePeriod = p2.VolumePeriod
		c.VolumeWeight = p2.VolumeWeight
	}
	if rng.Float64() < 0.5 {
		c.UseMA = p2.UseMA
		c.MaShort = p2.MaShort
		c.MaLong = p2.MaLong
		c.MaWeight = p2.MaWeight
	}
	if rng.Float64() < 0.5 {
		c.UseTrend = p2.UseTrend
		c.TrendN = p2.TrendN
		c.TrendWeight = p2.TrendWeight
	}
	if rng.Float64() < 0.5 {
		c.UseMACD = p2.UseMACD
		c.MACDFast = p2.MACDFast
		c.MACDSlow = p2.MACDSlow
		c.MACDSignal = p2.MACDSignal
		c.MACDWeight = p2.MACDWeight
	}
	if rng.Float64() < 0.5 {
		c.UseSession = p2.UseSession
		c.SessionWeight = p2.SessionWeight
	}
	if rng.Float64() < 0.5 {
		c.UseMACross = p2.UseMACross
		c.MACrossShort = p2.MACrossShort
		c.MACrossLong = p2.MACrossLong
		c.MACrossWeight = p2.MACrossWeight
		c.MACrossWindow = p2.MACrossWindow
		c.MACrossPreempt = p2.MACrossPreempt
	}
	return c
}

func randomCase(rng *rand.Rand) batchtest.TestCase {
	c := batchtest.TestCase{}
	if rng.Float64() < 0.8 {
		c.UseBoll = true
		c.BollPeriod = 5 + rng.Intn(25)
		c.BollMultiplier = roundTo1(0.8 + rng.Float64()*2.2)
		c.BollWeight = roundTo1(0.5 + rng.Float64()*3)
	}
	if rng.Float64() < 0.65 {
		c.UseBreakout = true
		c.BreakoutPeriod = 3 + rng.Intn(28)
		c.BreakoutWeight = roundTo1(-0.5 - rng.Float64()*3)
	}
	if rng.Float64() < 0.4 {
		c.UseRSI = true
		c.RSIPeriod = 5 + rng.Intn(12)
		c.RSIOverbought = 65 + float64(rng.Intn(16))
		c.RSIOversold = 15 + float64(rng.Intn(16))
		c.RSIWeight = roundTo1(0.5 + rng.Float64()*3)
	}
	if rng.Float64() < 0.15 {
		c.UseATR = true
		c.ATRPeriod = 5 + rng.Intn(16)
		c.ATRWeight = roundTo1(-0.5 - rng.Float64()*2)
	}
	if rng.Float64() < 0.1 {
		c.UseVolume = true
		c.VolumePeriod = 5 + rng.Intn(26)
		c.VolumeWeight = roundTo1(-0.5 - rng.Float64()*2)
	}
	if !c.UseBoll && !c.UseBreakout && !c.UseRSI {
		c.UseBoll = true
		c.BollPeriod = 10 + rng.Intn(10)
		c.BollMultiplier = roundTo1(1.5 + rng.Float64())
		c.BollWeight = 1
	}
	return c
}

func genName(tc batchtest.TestCase) string {
	var p []string
	if tc.UseBoll {
		p = append(p, fmt.Sprintf("Bo%dM%.1f", tc.BollPeriod, tc.BollMultiplier))
	}
	if tc.UseBreakout {
		p = append(p, fmt.Sprintf("Br%d", tc.BreakoutPeriod))
	}
	if tc.UseRSI {
		p = append(p, fmt.Sprintf("R%d", tc.RSIPeriod))
	}
	if tc.UseATR {
		p = append(p, fmt.Sprintf("A%d", tc.ATRPeriod))
	}
	if tc.UseVolume {
		p = append(p, fmt.Sprintf("V%d", tc.VolumePeriod))
	}
	if tc.UseMACD {
		p = append(p, "MACD")
	}
	if tc.UseMA {
		p = append(p, "MA")
	}
	if tc.UseTrend {
		p = append(p, fmt.Sprintf("Tr%d", tc.TrendN))
	}
	if tc.UsePriceVsMA {
		p = append(p, "PvMA")
	}
	if tc.UseSession {
		p = append(p, "Sess")
	}
	if tc.UseMACross {
		p = append(p, fmt.Sprintf("MACr%d_%d", tc.MACrossShort, tc.MACrossLong))
	}
	if len(p) == 0 {
		return "Empty"
	}
	return strings.Join(p, "+")
}

func ensureValid(tc batchtest.TestCase) batchtest.TestCase {
	if tc.UseBoll {
		tc.BollPeriod = clampInt(tc.BollPeriod, 3, 50)
		tc.BollMultiplier = roundTo1(clampFloat(tc.BollMultiplier, 0.5, 3.5))
		if tc.BollWeight == 0 {
			tc.BollWeight = 1
		}
	}
	if tc.UseBreakout {
		tc.BreakoutPeriod = clampInt(tc.BreakoutPeriod, 2, 50)
		if tc.BreakoutWeight == 0 {
			tc.BreakoutWeight = -1
		}
	}
	if tc.UseRSI {
		tc.RSIPeriod = clampInt(tc.RSIPeriod, 3, 30)
		tc.RSIOverbought = math.Round(clampFloat(tc.RSIOverbought, 55, 85))
		tc.RSIOversold = math.Round(clampFloat(tc.RSIOversold, 15, 45))
		if tc.RSIOverbought <= tc.RSIOversold {
			tc.RSIOverbought = 70
			tc.RSIOversold = 30
		}
		if tc.RSIWeight == 0 {
			tc.RSIWeight = 1
		}
	}
	if tc.UseATR {
		tc.ATRPeriod = clampInt(tc.ATRPeriod, 3, 30)
		if tc.ATRWeight == 0 {
			tc.ATRWeight = -1
		}
	}
	if tc.UseVolume {
		tc.VolumePeriod = clampInt(tc.VolumePeriod, 3, 40)
		if tc.VolumeWeight == 0 {
			tc.VolumeWeight = -1
		}
	}
	if tc.UseMACD {
		tc.MACDFast = clampInt(tc.MACDFast, 5, 20)
		tc.MACDSlow = clampInt(tc.MACDSlow, 15, 40)
		tc.MACDSignal = clampInt(tc.MACDSignal, 5, 15)
		if tc.MACDWeight == 0 {
			tc.MACDWeight = 1
		}
	}
	if tc.UseMA {
		tc.MaShort = clampInt(tc.MaShort, 3, 20)
		tc.MaLong = clampInt(tc.MaLong, 10, 60)
		if tc.MaLong <= tc.MaShort {
			tc.MaLong = tc.MaShort + 10
		}
		if tc.MaWeight == 0 {
			tc.MaWeight = 1
		}
	}
	if tc.UseMACross {
		tc.MACrossShort = clampInt(tc.MACrossShort, 2, 30)
		tc.MACrossLong = clampInt(tc.MACrossLong, 5, 250)
		if tc.MACrossLong <= tc.MACrossShort {
			tc.MACrossLong = tc.MACrossShort + 5
		}
		tc.MACrossWindow = clampInt(tc.MACrossWindow, 0, 5)
		tc.MACrossPreempt = clampFloat(tc.MACrossPreempt, 0, 0.01)
		if tc.MACrossWeight == 0 {
			tc.MACrossWeight = 1
		}
	}
	return tc
}

// ======================== ANALYSIS ========================

type analyzed struct {
	total     int
	top10     []batchtest.TestResult
	top10Avg  float64
	avgSigAcc float64
	maxSigAcc float64
	minSigAcc float64
}

func analyzeResults(results []batchtest.TestResult) analyzed {
	a := analyzed{total: len(results)}
	if len(results) == 0 {
		return a
	}

	sorted := make([]batchtest.TestResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SignalAccuracy > sorted[j].SignalAccuracy
	})

	a.maxSigAcc = sorted[0].SignalAccuracy
	a.minSigAcc = sorted[len(sorted)-1].SignalAccuracy
	sum := 0.0
	for _, r := range sorted {
		sum += r.SignalAccuracy
	}
	a.avgSigAcc = sum / float64(len(sorted))

	n := 10
	if n > len(sorted) {
		n = len(sorted)
	}
	a.top10 = sorted[:n]

	top10sum := 0.0
	for _, r := range a.top10 {
		top10sum += r.SignalAccuracy
	}
	a.top10Avg = top10sum / float64(len(a.top10))

	return a
}

func writeAnalysisLog(path string, iter int, excelFile string, a analyzed, elapsed time.Duration) {
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()

	w := func(s string, args ...interface{}) { fmt.Fprintf(f, s+"\n", args...) }
	w("Iteration: %d", iter)
	w("Time: %s", time.Now().Format("2006-01-02 15:04:05"))
	w("Excel: %s", excelFile)
	w("Duration: %s", elapsed.Round(time.Millisecond))
	w("Cases: %d", a.total)
	w("Signal Accuracy: best=%.2f%% top10avg=%.2f%% avg=%.2f%% worst=%.2f%%",
		a.maxSigAcc*100, a.top10Avg*100, a.avgSigAcc*100, a.minSigAcc*100)
	w("")
	w("=== TOP 10 ===")
	for i, r := range a.top10 {
		rate := 0.0
		if r.Total > 0 {
			rate = float64(r.SignalCount) / float64(r.Total) * 100
		}
		w("%2d. #%-3d %-40s Acc=%.2f%% Rate=%.1f%% Str=%.3f Avg=%+.3f",
			i+1, r.TestCase.ID, r.TestCase.Name, r.SignalAccuracy*100, rate, r.AvgAbsScore, r.AvgScore)
	}
}

func printTrend(summaries []iterSummary, globalBestAcc float64, globalBestName string) {
	fmt.Println("\n============ 正确率趋势 ============")
	fmt.Printf("%-5s  %-9s  %-9s  %-9s  %-7s\n", "轮次", "最佳", "Top10均", "全局均", "耗时")
	fmt.Println("--------------------------------------------")
	for _, s := range summaries {
		marker := ""
		if s.bestAcc == globalBestAcc {
			marker = " ★"
		}
		fmt.Printf("%-5d  %7.2f%%  %7.2f%%  %7.2f%%  %5.1fs%s\n",
			s.iter, s.bestAcc*100, s.top10Avg*100, s.avgAcc*100, s.elapsed.Seconds(), marker)
	}
	fmt.Println("--------------------------------------------")
	if len(summaries) >= 2 {
		first := summaries[0]
		last := summaries[len(summaries)-1]
		delta := last.bestAcc - first.bestAcc
		fmt.Printf("最佳变化: %+.2f%% (第1轮 %.2f%% → 第%d轮 %.2f%%)\n",
			delta*100, first.bestAcc*100, last.iter, last.bestAcc*100)
	}
	name := globalBestName
	if len(name) > 35 {
		name = name[:35]
	}
	fmt.Printf("全局最佳: %.2f%% (%s)\n", globalBestAcc*100, name)
	fmt.Println("=====================================")
}

func writeTrendFile(path string, summaries []iterSummary, globalBestAcc float64, globalBestName string) {
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()

	w := func(s string, args ...interface{}) { fmt.Fprintf(f, s+"\n", args...) }
	w("进化式自动调参 趋势记录")
	w("生成时间: %s", time.Now().Format("2006-01-02 15:04:05"))
	w("总轮数: %d", len(summaries))
	w("")
	w("%-5s  %-9s  %-9s  %-9s  %-7s  %s", "轮次", "最佳", "Top10均", "全局均", "耗时", "最佳策略")
	for _, s := range summaries {
		w("%-5d  %7.2f%%  %7.2f%%  %7.2f%%  %5.1fs  %s",
			s.iter, s.bestAcc*100, s.top10Avg*100, s.avgAcc*100, s.elapsed.Seconds(), s.bestName)
	}
	w("")
	if len(summaries) >= 2 {
		first := summaries[0]
		last := summaries[len(summaries)-1]
		delta := last.bestAcc - first.bestAcc
		w("最佳变化: %+.2f%% (第1轮 %.2f%% → 第%d轮 %.2f%%)",
			delta*100, first.bestAcc*100, last.iter, last.bestAcc*100)
	}
	w("全局最佳: %.2f%% (%s)", globalBestAcc*100, globalBestName)
}

// ======================== HELPERS ========================

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func roundTo1(v float64) float64 {
	return math.Round(v*10) / 10
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
