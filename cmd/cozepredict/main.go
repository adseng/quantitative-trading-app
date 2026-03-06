package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"quantitative-trading-app/internal/binance"
	"quantitative-trading-app/internal/config"
	"quantitative-trading-app/internal/coze"
	"quantitative-trading-app/internal/factor"
)

const (
	klineLimit   = 200
	windowSize   = 25 // 每窗 25 条：20 条给 Coze，5 条验证
	cozeKlines   = 20
	verifyKlines = 5
	windowStep   = 5
	interval     = "15m"

	capitalU = 1000.0
	leverage = 1.0
)

// backtestRow 单窗回测记录：Coze 结果 + 验证 5 根 K 线
type backtestRow struct {
	Index   int
	Res     *coze.CozeStructuredResult
	Verify5 []*factor.KLine
}

func main() {
	if err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	symbol := config.Get(config.KeySymbol, "BTCUSDT")
	ctx := context.Background()

	// 1. 拉取最近 200 根 K 线
	klines, err := binance.FetchKlines(symbol, interval, klineLimit, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取 K 线失败: %v\n", err)
		os.Exit(1)
	}
	if len(klines) < windowSize {
		fmt.Fprintf(os.Stderr, "K 线不足 %d 根，当前 %d 根\n", windowSize, len(klines))
		os.Exit(1)
	}

	// 输出目录：与可执行文件同级（cmd/cozepredict）；go run 时在临时目录，则用 cwd/cmd/cozepredict
	scriptDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取工作目录失败: %v\n", err)
		os.Exit(1)
	}
	if execPath, err := os.Executable(); err == nil {
		dir := filepath.Dir(execPath)
		// 避免 go run 时写到系统临时目录
		if !strings.Contains(dir, "go-build") {
			scriptDir = dir
		} else {
			scriptDir = filepath.Join(scriptDir, "cmd", "cozepredict")
		}
	}
	excelPath := filepath.Join(scriptDir, "backtest_"+time.Now().Format("2006-01-02")+".xlsx")

	var rows []backtestRow
	// 2. 滑动窗口：每窗 20 条送 Coze，5 条验证；先记录到内存并写 Excel
	for start := 0; start+windowSize <= len(klines); start += windowStep {
		idx := len(rows) + 1
		res, err := coze.PredictStructured(ctx, klines[start:start+cozeKlines], symbol, cozeKlines)
		if err != nil {
			fmt.Fprintf(os.Stderr, "窗口 %d Coze 调用失败: %v\n", idx, err)
			rows = append(rows, backtestRow{Index: idx, Res: nil, Verify5: nil})
			// 仍写 Excel（Coze 为空，验证 5 根照写）
			verify5 := make([]*factor.KLine, verifyKlines)
			for i := 0; i < verifyKlines; i++ {
				verify5[i] = klines[start+cozeKlines+i]
			}
			_ = coze.AppendBacktestRowToExcel(excelPath, idx, nil, verify5)
			continue
		}
		verify5 := klines[start+cozeKlines : start+windowSize]
		rows = append(rows, backtestRow{Index: idx, Res: res, Verify5: verify5})
		if err := coze.AppendBacktestRowToExcel(excelPath, idx, res, verify5); err != nil {
			fmt.Fprintf(os.Stderr, "窗口 %d 写 Excel 失败: %v\n", idx, err)
		}
	}

	// 3. 四档阈值验证：55%、60%、65%、70%
	thresholds := []int{55, 60, 65, 70}
	type roundStat struct {
		Threshold int
		Deals     int // 出手次数
		Correct   int // 正确次数
		TotalPnlU float64
	}
	var stats []roundStat
	for _, th := range thresholds {
		var deals, correct int
		var pnlSum float64
		for _, r := range rows {
			if r.Res == nil || len(r.Verify5) < verifyKlines {
				continue
			}
			open0 := r.Verify5[0].Open
			close5 := r.Verify5[verifyKlines-1].Close
			// 选出概率 >= 阈值且最高的场景作为当窗方向
			best, ok := bestScenarioAtThreshold(r.Res.Scenarios, th)
			if !ok {
				continue
			}
			deals++
			isLong := isDirectionLong(best.Direction)
			var correctThis bool
			var pnl float64
			if isLong {
				correctThis = close5 > open0
				pnl = (close5 - open0) / open0 * capitalU * leverage
			} else {
				correctThis = close5 < open0
				pnl = (open0 - close5) / open0 * capitalU * leverage
			}
			if correctThis {
				correct++
			}
			pnlSum += pnl
		}
		stats = append(stats, roundStat{Threshold: th, Deals: deals, Correct: correct, TotalPnlU: pnlSum})
	}

	// 4. 写入 测试结果.txt（与脚本同级目录）
	resultPath := filepath.Join(scriptDir, "测试结果.txt")
	f, err := os.Create(resultPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建 测试结果.txt 失败: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	fmt.Fprintf(f, "回测结果 %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(f, "交易对=%s 周期=%s 总窗数=%d 本金=%.0fU 杠杆=%.0fx\n\n", symbol, interval, len(rows), capitalU, leverage)
	for _, s := range stats {
		correctRate := 0.0
		if s.Deals > 0 {
			correctRate = float64(s.Correct) / float64(s.Deals) * 100
		}
		fmt.Fprintf(f, "阈值 %d%%：出手次数=%d 总次数=%d 正确率=%.2f%% 盈收=%.2f U\n",
			s.Threshold, s.Deals, s.Deals, correctRate, s.TotalPnlU)
	}
	fmt.Printf("回测完成。Excel: %s\n结果: %s\n", excelPath, resultPath)
}

// bestScenarioAtThreshold 从 Scenarios 中选出概率 >= threshold 且概率最高的一条（仅看涨/看跌，不含 SIDEWAYS）；若没有则返回 (nil, false)。
func bestScenarioAtThreshold(scenarios []coze.CozeScenario, threshold int) (coze.CozeScenario, bool) {
	var best *coze.CozeScenario
	for i := range scenarios {
		s := &scenarios[i]
		if s.Probability < threshold {
			continue
		}
		if !isTradableDirection(s.Direction) {
			continue
		}
		if best == nil || s.Probability > best.Probability {
			best = s
		}
	}
	if best == nil {
		return coze.CozeScenario{}, false
	}
	return *best, true
}

func isTradableDirection(direction string) bool {
	return isDirectionLong(direction) || isDirectionShort(direction)
}

func isDirectionLong(direction string) bool {
	d := strings.TrimSpace(strings.ToLower(direction))
	return strings.Contains(d, "涨") || d == "long" || strings.Contains(d, "多")
}

func isDirectionShort(direction string) bool {
	d := strings.TrimSpace(strings.ToLower(direction))
	return strings.Contains(d, "跌") || d == "short" || strings.Contains(d, "空")
}
