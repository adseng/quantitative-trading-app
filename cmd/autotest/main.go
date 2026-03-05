package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"quantitative-trading-app/internal/batchtest"
	"quantitative-trading-app/internal/batchtest/cases"
	"quantitative-trading-app/internal/datafile"
	"quantitative-trading-app/internal/factor"
)

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
	dataPath := flag.String("data", datafile.DefaultPath, "path to kline data file")
	flag.Parse()

	fmt.Println("=== 策略回测测试 ===")
	fmt.Printf("数据文件: %s\n\n", *dataPath)

	fmt.Println("[加载数据] 读取本地K线文件...")
	klines, err := datafile.LoadKlines(*dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载失败: %v\n先运行: go run ./cmd/fetchdata\n", err)
		os.Exit(1)
	}
	fmt.Printf("[加载数据] 完成，共 %d 根K线 (%.0f天)\n\n", len(klines), float64(len(klines))*15/60/24)

	// 仅调用 GenerateTestCases，不做进化。优化流程：执行脚本 → 分析结果 → 更新 case_*.go → 再执行
	cases := cases.GenerateTestCases()
	fmt.Printf("[策略] GenerateTestCases 生成 %d 组用例\n\n", len(cases))

	s, _ := runOneIteration(1, klines, cases)

	printTrend([]iterSummary{s}, s.bestAcc, s.bestName)
	trendFile := filepath.Join(batchtest.ExcelDirPath(),
		fmt.Sprintf("trend_%s.txt", time.Now().Format("20060102150405")))
	writeTrendFile(trendFile, []iterSummary{s}, s.bestAcc, s.bestName)
	fmt.Printf("\n趋势文件: %s\n", trendFile)
}

func runOneIteration(iter int, klines []*factor.KLine, testCases []cases.TestCase) (iterSummary, []cases.TestResult) {
	runner := batchtest.NewBatchRunnerWithCases(testCases)
	fmt.Printf("[回测] 开始 %d 组参数回测...\n", len(testCases))
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
	printPerGroupBest(results)

	logFile := filepath.Join(batchtest.ExcelDirPath(),
		fmt.Sprintf("analysis_iter%d_%s.txt", iter, time.Now().Format("20060102150405")))
	writeAnalysisLog(logFile, iter, runner.ExcelFile(), a, elapsed, results)

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

func printFactorDistribution(results []cases.TestResult) {
	sorted := make([]cases.TestResult, len(results))
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

// ======================== ANALYSIS ========================

type analyzed struct {
	total     int
	top10     []cases.TestResult
	top10Avg  float64
	avgSigAcc float64
	maxSigAcc float64
	minSigAcc float64
}

func analyzeResults(results []cases.TestResult) analyzed {
	a := analyzed{total: len(results)}
	if len(results) == 0 {
		return a
	}

	sorted := make([]cases.TestResult, len(results))
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

func strategyGroup(tc cases.TestCase) string {
	var parts []string
	if tc.UseBoll {
		parts = append(parts, "Bo")
	}
	if tc.UseRSI {
		parts = append(parts, "R")
	}
	if tc.UseBreakout {
		parts = append(parts, "Br")
	}
	if tc.UseATR {
		parts = append(parts, "A")
	}
	if tc.UseMA {
		parts = append(parts, "M")
	}
	if tc.UseTrend {
		parts = append(parts, "T")
	}
	if tc.UsePriceVsMA {
		parts = append(parts, "Pv")
	}
	if tc.UseVolume {
		parts = append(parts, "V")
	}
	if tc.UseMACD {
		parts = append(parts, "Macd")
	}
	if tc.UseMACross {
		parts = append(parts, "Mc")
	}
	if tc.UseSession {
		parts = append(parts, "Sess")
	}
	if len(parts) == 0 {
		return "other"
	}
	return strings.Join(parts, "+")
}

func printPerGroupBest(results []cases.TestResult) {
	groupBest := make(map[string]cases.TestResult)
	for _, r := range results {
		g := strategyGroup(r.TestCase)
		prev, ok := groupBest[g]
		if !ok || r.SignalAccuracy > prev.SignalAccuracy {
			groupBest[g] = r
		}
	}
	// sort groups by name for stable output
	var keys []string
	for k := range groupBest {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Println("[分析] 各策略组最优:")
	for _, g := range keys {
		r := groupBest[g]
		rate := 0.0
		if r.Total > 0 {
			rate = float64(r.SignalCount) / float64(r.Total) * 100
		}
		fmt.Printf("    %-18s %-42s %.2f%% (信号率%.1f%%)\n",
			g, r.TestCase.Name, r.SignalAccuracy*100, rate)
	}
}

func writeAnalysisLog(path string, iter int, excelFile string, a analyzed, elapsed time.Duration, results []cases.TestResult) {
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

	// 各组最优
	groupBest := make(map[string]cases.TestResult)
	for _, r := range results {
		g := strategyGroup(r.TestCase)
		prev, ok := groupBest[g]
		if !ok || r.SignalAccuracy > prev.SignalAccuracy {
			groupBest[g] = r
		}
	}
	var keys []string
	for k := range groupBest {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	w("")
	w("=== 各组策略最优参数 ===")
	for _, g := range keys {
		r := groupBest[g]
		rate := 0.0
		if r.Total > 0 {
			rate = float64(r.SignalCount) / float64(r.Total) * 100
		}
		w("%-18s %-42s Acc=%.2f%% Rate=%.1f%% | %s", g, r.TestCase.Name, r.SignalAccuracy*100, rate, formatTestCaseParams(r.TestCase))
	}
}

func formatTestCaseParams(tc cases.TestCase) string {
	var parts []string
	if tc.UseBoll {
		parts = append(parts, fmt.Sprintf("Bo P%d M%.1f w%.1f", tc.BollPeriod, tc.BollMultiplier, tc.BollWeight))
	}
	if tc.UseRSI {
		parts = append(parts, fmt.Sprintf("RSI P%d %.0f/%.0f w%.1f", tc.RSIPeriod, tc.RSIOverbought, tc.RSIOversold, tc.RSIWeight))
	}
	if tc.UseBreakout {
		parts = append(parts, fmt.Sprintf("Br P%d w%.1f", tc.BreakoutPeriod, tc.BreakoutWeight))
	}
	if tc.UseATR {
		parts = append(parts, fmt.Sprintf("ATR P%d w%.1f", tc.ATRPeriod, tc.ATRWeight))
	}
	if tc.UseMA {
		parts = append(parts, fmt.Sprintf("MA %d/%d w%.1f", tc.MaShort, tc.MaLong, tc.MaWeight))
	}
	if tc.UseTrend {
		parts = append(parts, fmt.Sprintf("T N%d w%.1f", tc.TrendN, tc.TrendWeight))
	}
	if tc.UsePriceVsMA {
		parts = append(parts, fmt.Sprintf("Pv P%d w%.1f", tc.PriceVsMAPeriod, tc.PriceVsMAWeight))
	}
	if tc.UseVolume {
		parts = append(parts, fmt.Sprintf("V P%d w%.1f", tc.VolumePeriod, tc.VolumeWeight))
	}
	if tc.UseMACD {
		parts = append(parts, fmt.Sprintf("Macd %d/%d/%d w%.1f", tc.MACDFast, tc.MACDSlow, tc.MACDSignal, tc.MACDWeight))
	}
	if tc.UseMACross {
		parts = append(parts, fmt.Sprintf("Mc %d/%d w%.1f", tc.MACrossShort, tc.MACrossLong, tc.MACrossWeight))
	}
	if tc.UseSession {
		parts = append(parts, fmt.Sprintf("Sess w%.1f", tc.SessionWeight))
	}
	if len(parts) == 0 {
		return "-"
	}
	s := ""
	for i, p := range parts {
		if i > 0 {
			s += " "
		}
		s += p
	}
	return s
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
