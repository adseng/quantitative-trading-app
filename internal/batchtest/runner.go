package batchtest

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"quantitative-trading-app/internal/batchtest/cases"
	"quantitative-trading-app/internal/factor"

	"github.com/xuri/excelize/v2"
)

// BatchRunner 批量回测运行器，支持 start/stop/resume。
// 进度通过 Excel 持久化：启动时读取已完成的用例 ID，跳过已完成部分。
type BatchRunner struct {
	mu          sync.Mutex
	running     bool
	cancelCh    chan struct{}
	nextIdx     int // 下一个要运行的 case 索引（用于 resume）
	testCases   []cases.TestCase
	results     []cases.TestResult
	completedIDs map[int]bool // 从 Excel 读取的已完成 ID 集合
	excelFile   string        // 当前批次的 Excel 文件名（含时间戳）
}

// NewBatchRunner 创建批量运行器，自动从最新 Excel 恢复进度。
func NewBatchRunner() *BatchRunner {
	r := &BatchRunner{
		testCases: cases.GenerateTestCases(),
	}
	r.findOrCreateExcelFile()
	r.loadProgressFromExcel()
	return r
}

// NewBatchRunnerWithCases 创建使用自定义用例的运行器（不恢复进度，新建 Excel）。
func NewBatchRunnerWithCases(testCases []cases.TestCase) *BatchRunner {
	return &BatchRunner{
		testCases: testCases,
		excelFile: "batch_test_log_" + time.Now().Format("20060102150405") + ".xlsx",
	}
}

// excelDir 返回 Excel 输出目录（docs/test/），不存在则自动创建。
func excelDir() string {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	dir := filepath.Join(wd, "docs", "test")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

// findOrCreateExcelFile 查找 docs/test/ 下最新的 batch_test_log_*.xlsx，
// 如果没有则生成一个带当前时间戳的新文件名。
func (r *BatchRunner) findOrCreateExcelFile() {
	dir := excelDir()

	entries, err := os.ReadDir(dir)
	if err != nil {
		r.excelFile = "batch_test_log_" + time.Now().Format("20060102150405") + ".xlsx"
		return
	}

	latest := ""
	for _, e := range entries {
		name := e.Name()
		if len(name) > 19 && name[:15] == "batch_test_log_" && name[len(name)-5:] == ".xlsx" {
			if name > latest {
				latest = name
			}
		}
	}

	if latest != "" {
		r.excelFile = latest
	} else {
		r.excelFile = "batch_test_log_" + time.Now().Format("20060102150405") + ".xlsx"
	}
}

// ExcelFile 返回当前使用的 Excel 文件名。
func (r *BatchRunner) ExcelFile() string {
	return r.excelFile
}

// ResetForNewBatch 重置运行器：生成新文件名、清空进度，用于全新一轮测试。
func (r *BatchRunner) ResetForNewBatch() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.excelFile = "batch_test_log_" + time.Now().Format("20060102150405") + ".xlsx"
	r.nextIdx = 0
	r.results = nil
	r.completedIDs = nil
	r.testCases = cases.GenerateTestCases()
}

// loadProgressFromExcel 读取当前 Excel 文件，收集已完成的用例 ID，
// 计算 nextIdx 使其跳过所有已完成用例。
func (r *BatchRunner) loadProgressFromExcel() {
	path := filepath.Join(excelDir(), r.excelFile)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return
	}

	f, err := excelize.OpenFile(path)
	if err != nil {
		return
	}
	defer f.Close()

	rows, err := f.GetRows(batchSheetName)
	if err != nil || len(rows) <= 1 {
		return
	}

	r.completedIDs = make(map[int]bool)
	for i := 1; i < len(rows); i++ { // 跳过表头
		if len(rows[i]) == 0 {
			continue
		}
		id, err := strconv.Atoi(rows[i][0])
		if err != nil {
			continue
		}
		r.completedIDs[id] = true
	}

	// 从头扫描 testCases，找到第一个未完成的索引
	for idx, tc := range r.testCases {
		if !r.completedIDs[tc.ID] {
			r.nextIdx = idx
			return
		}
	}
	// 全部已完成
	r.nextIdx = len(r.testCases)
}

// IsRunning 是否正在运行。
func (r *BatchRunner) IsRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.running
}

// NextIndex 当前进度。
func (r *BatchRunner) NextIndex() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.nextIdx
}

// TotalCases 总用例数。
func (r *BatchRunner) TotalCases() int {
	return len(r.testCases)
}

// Results 返回已完成的测试结果。
func (r *BatchRunner) Results() []cases.TestResult {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.results
}

// GetLastResults 返回最近的结果：若内存中有则直接返回，否则从当前 Excel 加载（用于页面初始化展示）。
func (r *BatchRunner) GetLastResults(maxN int) []cases.TestResult {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.results) > 0 {
		return recentN(r.results, maxN)
	}
	loaded := r.loadResultsFromExcel(maxN)
	return loaded
}

// loadResultsFromExcel 从当前 Excel 文件加载最近 maxN 条结果（不加锁，调用方需持有锁）。
func (r *BatchRunner) loadResultsFromExcel(maxN int) []cases.TestResult {
	path := filepath.Join(excelDir(), r.excelFile)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	rows, err := f.GetRows(batchSheetName)
	if err != nil || len(rows) <= 1 {
		return nil
	}
	idToCase := make(map[int]cases.TestCase)
	for _, c := range r.testCases {
		idToCase[c.ID] = c
	}
	var all []cases.TestResult
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 50 {
			continue
		}
		id, _ := strconv.Atoi(row[0])
		tc, ok := idToCase[id]
		if !ok {
			continue
		}
		acc, _ := parsePct(safeCol(row, 44))       // 信号正确率 "59.7%"
		correct, _ := strconv.Atoi(safeCol(row, 45))   // 正确数
		signalCount, _ := strconv.Atoi(safeCol(row, 46)) // 有效预测
		total, _ := strconv.Atoi(safeCol(row, 47))     // 总数
		avgScore, _ := strconv.ParseFloat(safeCol(row, 48), 64)   // 平均净分
		avgAbsScore, _ := strconv.ParseFloat(safeCol(row, 49), 64) // 信号强度
		accuracy := 0.0
		if total > 0 {
			accuracy = float64(correct) / float64(total)
		}
		all = append(all, cases.TestResult{
			TestCase:       tc,
			Accuracy:       accuracy,
			Correct:        correct,
			Total:          total,
			SignalCount:    signalCount,
			SignalAccuracy: acc,
			AvgScore:       avgScore,
			AvgAbsScore:    avgAbsScore,
		})
	}
	return recentN(all, maxN)
}

func safeCol(row []string, idx int) string {
	if idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

func parsePct(s string) (float64, error) {
	s = strings.TrimSuffix(strings.TrimSpace(s), "%")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return f / 100, nil
}

// ExcelDir 返回 Excel 输出目录。
func ExcelDirPath() string {
	return excelDir()
}

// Stop 停止运行。
func (r *BatchRunner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.running && r.cancelCh != nil {
		close(r.cancelCh)
		r.running = false
	}
}

// ProgressEvent 进度事件数据。
type ProgressEvent struct {
	Phase      string             `json:"phase"`      // "fetching" | "testing" | "done" | "error"
	Message    string             `json:"message"`
	CaseIndex  int                `json:"caseIndex"`
	TotalCases int                `json:"totalCases"`
	Current    *cases.TestResult  `json:"current,omitempty"`
	Recent     []cases.TestResult `json:"recent,omitempty"` // 最近 200 条
}

// Run 开始或继续批量测试。
// klines 为已获取的 K 线数据；onProgress 回调进度事件。
func (r *BatchRunner) Run(klines []*factor.KLine, onProgress func(ProgressEvent)) {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return
	}
	r.running = true
	r.cancelCh = make(chan struct{})
	r.mu.Unlock()

	defer func() {
		r.mu.Lock()
		r.running = false
		r.mu.Unlock()
	}()

	cancelCh := r.cancelCh
	total := len(r.testCases)

	for r.nextIdx < total {
		select {
		case <-cancelCh:
			onProgress(ProgressEvent{Phase: "stopped", Message: "已停止", CaseIndex: r.nextIdx, TotalCases: total, Recent: recentN(r.results, 200)})
			return
		default:
		}

		tc := r.testCases[r.nextIdx]

		cfg := &factor.FactorConfig{
			UseMA: tc.UseMA, MaShort: tc.MaShort, MaLong: tc.MaLong, MaWeight: tc.MaWeight,
			UseTrend: tc.UseTrend, TrendN: tc.TrendN, TrendWeight: tc.TrendWeight,
			UseRSI: tc.UseRSI, RSIPeriod: tc.RSIPeriod, RSIOverbought: tc.RSIOverbought, RSIOversold: tc.RSIOversold, RSIWeight: tc.RSIWeight,
			UseMACD: tc.UseMACD, MACDFast: tc.MACDFast, MACDSlow: tc.MACDSlow, MACDSignal: tc.MACDSignal, MACDWeight: tc.MACDWeight,
			UseBoll: tc.UseBoll, BollPeriod: tc.BollPeriod, BollMultiplier: tc.BollMultiplier, BollWeight: tc.BollWeight,
			UseBreakout: tc.UseBreakout, BreakoutPeriod: tc.BreakoutPeriod, BreakoutWeight: tc.BreakoutWeight,
			UsePriceVsMA: tc.UsePriceVsMA, PriceVsMAPeriod: tc.PriceVsMAPeriod, PriceVsMAWeight: tc.PriceVsMAWeight,
			UseATR: tc.UseATR, ATRPeriod: tc.ATRPeriod, ATRWeight: tc.ATRWeight,
			UseVolume: tc.UseVolume, VolumePeriod: tc.VolumePeriod, VolumeWeight: tc.VolumeWeight,
			UseSession: tc.UseSession, SessionWeight: tc.SessionWeight,
			UseMACross: tc.UseMACross, MACrossShort: tc.MACrossShort, MACrossLong: tc.MACrossLong, MACrossWeight: tc.MACrossWeight,
				MACrossWindow: tc.MACrossWindow, MACrossPreempt: tc.MACrossPreempt,
		}

		summary := factor.BacktestWithConfig(klines, cfg)

		tr := cases.TestResult{
			TestCase:       tc,
			Accuracy:       summary.Accuracy,
			Correct:        summary.Correct,
			Total:          summary.Total,
			SignalCount:    summary.SignalCount,
			SignalAccuracy: summary.SignalAccuracy,
			AvgScore:       summary.AvgScore,
			AvgAbsScore:    summary.AvgAbsScore,
		}
		r.results = append(r.results, tr)
		r.nextIdx++

		_ = r.writeOneToExcel(tr)

		onProgress(ProgressEvent{
			Phase:      "testing",
			Message:    fmt.Sprintf("[%d/%d] %s → 信号正确率%.1f%% 强度%.2f", r.nextIdx, total, tc.Name, summary.SignalAccuracy*100, summary.AvgAbsScore),
			CaseIndex:  r.nextIdx,
			TotalCases: total,
			Current:    &tr,
			Recent:     recentN(r.results, 200),
		})
	}

	onProgress(ProgressEvent{Phase: "done", Message: "全部完成", CaseIndex: total, TotalCases: total, Recent: recentN(r.results, 200)})
}

func recentN(results []cases.TestResult, n int) []cases.TestResult {
	if len(results) <= n {
		return results
	}
	return results[len(results)-n:]
}

const batchSheetName = "Sheet1"

var batchHeaders = []string{
	"ID", "名称", "时间",
	"MA启用", "MA短", "MA长", "MA权重",
	"趋势启用", "趋势N", "趋势权重",
	"RSI启用", "RSI周期", "RSI超买", "RSI超卖", "RSI权重",
	"MACD启用", "MACD快", "MACD慢", "MACD信号", "MACD权重",
	"Boll启用", "Boll周期", "Boll倍数", "Boll权重",
	"突破启用", "突破周期", "突破权重",
	"价格MA启用", "价格MA周期", "价格MA权重",
	"ATR启用", "ATR周期", "ATR权重",
	"量价启用", "量价周期", "量价权重",
	"时段启用", "时段权重",
	"金叉启用", "金叉短", "金叉长", "金叉权重", "金叉窗", "金叉预判",
	"信号正确率", "正确数", "有效预测", "总数", "平均净分", "信号强度", "标记",
}

func (r *BatchRunner) writeOneToExcel(tr cases.TestResult) error {
	path := filepath.Join(excelDir(), r.excelFile)

	var f *excelize.File
	isNew := false
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f = excelize.NewFile()
		f.SetSheetName("Sheet1", batchSheetName)
		for i, h := range batchHeaders {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			_ = f.SetCellValue(batchSheetName, cell, h)
		}
		isNew = true
	} else {
		f, err = excelize.OpenFile(path)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	rows, _ := f.GetRows(batchSheetName)
	nextRow := len(rows) + 1

	tc := tr.TestCase
	boolStr := func(b bool) string {
		if b {
			return "是"
		}
		return "否"
	}
	row := []interface{}{
		tc.ID, tc.Name, time.Now().Format("2006-01-02 15:04:05"),
		boolStr(tc.UseMA), tc.MaShort, tc.MaLong, tc.MaWeight,
		boolStr(tc.UseTrend), tc.TrendN, tc.TrendWeight,
		boolStr(tc.UseRSI), tc.RSIPeriod, tc.RSIOverbought, tc.RSIOversold, tc.RSIWeight,
		boolStr(tc.UseMACD), tc.MACDFast, tc.MACDSlow, tc.MACDSignal, tc.MACDWeight,
		boolStr(tc.UseBoll), tc.BollPeriod, tc.BollMultiplier, tc.BollWeight,
		boolStr(tc.UseBreakout), tc.BreakoutPeriod, tc.BreakoutWeight,
		boolStr(tc.UsePriceVsMA), tc.PriceVsMAPeriod, tc.PriceVsMAWeight,
		boolStr(tc.UseATR), tc.ATRPeriod, tc.ATRWeight,
		boolStr(tc.UseVolume), tc.VolumePeriod, tc.VolumeWeight,
		boolStr(tc.UseSession), tc.SessionWeight,
		boolStr(tc.UseMACross), tc.MACrossShort, tc.MACrossLong, tc.MACrossWeight, tc.MACrossWindow, tc.MACrossPreempt,
		fmt.Sprintf("%.1f%%", tr.SignalAccuracy*100),
		tr.Correct, tr.SignalCount, tr.Total,
		fmt.Sprintf("%.3f", tr.AvgScore),
		fmt.Sprintf("%.3f", tr.AvgAbsScore),
		"",
	}
	for i, v := range row {
		cell, _ := excelize.CoordinatesToCellName(i+1, nextRow)
		_ = f.SetCellValue(batchSheetName, cell, v)
	}

	// 信号正确率 > 55% 标绿
	if tr.SignalAccuracy > 0.55 {
		markCell, _ := excelize.CoordinatesToCellName(len(batchHeaders), nextRow)
		_ = f.SetCellValue(batchSheetName, markCell, "★")
		var color string
		switch {
		case tr.SignalAccuracy >= 0.70:
			color = "2ECC71"
		case tr.SignalAccuracy >= 0.65:
			color = "58D68D"
		case tr.SignalAccuracy >= 0.60:
			color = "82E0AA"
		case tr.SignalAccuracy >= 0.575:
			color = "A9DFBF"
		default:
			color = "C6EFCE"
		}
		sid, _ := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{color}},
		})
		_ = f.SetCellStyle(batchSheetName, markCell, markCell, sid)
	}

	if isNew {
		return f.SaveAs(path)
	}
	return f.Save()
}
