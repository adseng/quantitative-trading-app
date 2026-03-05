package backtestlog

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xuri/excelize/v2"
)

const sheetName = "Sheet1"
const defaultFileName = "backtest_log.xlsx"

var headers = []string{
	"时间", "交易对", "周期", "数量",
	"均线启用", "均线短", "均线长", "均线权重",
	"趋势启用", "趋势N", "趋势权重",
	"正确率", "正确数", "总数",
}

// LogBacktestResult 将回测参数和结果追加到 Excel 文件。
// 文件保存在当前工作目录下的 backtest_log.xlsx。
func LogBacktestResult(
	symbol, interval string, limit int64,
	useMA bool, maShort, maLong int, maWeight float64,
	useTrend bool, trendN int, trendWeight float64,
	accuracy float64, correct, total int,
) error {
	path, err := getLogPath()
	if err != nil {
		return err
	}

	var f *excelize.File
	isNew := false
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f = excelize.NewFile()
		f.SetSheetName("Sheet1", sheetName)
		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			_ = f.SetCellValue(sheetName, cell, h)
		}
		isNew = true
	} else {
		f, err = excelize.OpenFile(path)
		if err != nil {
			return fmt.Errorf("打开日志文件失败: %w", err)
		}
		defer f.Close()
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return err
	}
	nextRow := len(rows) + 1

	useMAStr := "否"
	if useMA {
		useMAStr = "是"
	}
	useTrendStr := "否"
	if useTrend {
		useTrendStr = "是"
	}

	row := []interface{}{
		time.Now().Format("2006-01-02 15:04:05"),
		symbol,
		interval,
		limit,
		useMAStr,
		maShort,
		maLong,
		maWeight,
		useTrendStr,
		trendN,
		trendWeight,
		fmt.Sprintf("%.1f%%", accuracy*100),
		correct,
		total,
	}
	for i, v := range row {
		cell, _ := excelize.CoordinatesToCellName(i+1, nextRow)
		_ = f.SetCellValue(sheetName, cell, v)
	}

	if isNew {
		if err := f.SaveAs(path); err != nil {
			return fmt.Errorf("保存日志失败: %w", err)
		}
	} else {
		if err := f.Save(); err != nil {
			return fmt.Errorf("保存日志失败: %w", err)
		}
	}
	return nil
}

func getLogPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, defaultFileName), nil
}
