package coze

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"quantitative-trading-app/internal/factor"

	"github.com/xuri/excelize/v2"
)

const cozeSheetName = "预测结果"

var cozeHeaders = []string{
	"时间", "预测时间", "交易对", "当前价", "市场结构", "场景",
}

// cozeExcelDir 返回 docs/coze 目录，不存在则创建
func cozeExcelDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(wd, "docs", "coze")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录 docs/coze 失败: %w", err)
	}
	return dir, nil
}

// AppendResultToExcel 将一次 Coze 预测结果追加到当日 Excel 表（docs/coze/coze_YYYY-MM-DD.xlsx），第一列为当前时间 YYYY-MM-DD HH:mm:ss
func AppendResultToExcel(res *CozeStructuredResult) error {
	if res == nil {
		return nil
	}
	dir, err := cozeExcelDir()
	if err != nil {
		return err
	}
	date := time.Now().Format("2006-01-02")
	filename := "coze_" + date + ".xlsx"
	path := filepath.Join(dir, filename)

	var f *excelize.File
	isNew := false
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f = excelize.NewFile()
		f.SetSheetName("Sheet1", cozeSheetName)
		for i, h := range cozeHeaders {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			_ = f.SetCellValue(cozeSheetName, cell, h)
		}
		isNew = true
	} else {
		f, err = excelize.OpenFile(path)
		if err != nil {
			return fmt.Errorf("打开 Excel 失败: %w", err)
		}
		defer f.Close()
	}

	rows, err := f.GetRows(cozeSheetName)
	if err != nil {
		return err
	}
	nextRow := len(rows) + 1

	nowStr := time.Now().Format("2006-01-02 15:04:05")
	scenariosStr := ""
	if len(res.RawAnswer) > 0 && len(res.Scenarios) == 0 {
		scenariosStr = res.RawAnswer
	} else if len(res.Scenarios) > 0 {
		b, _ := json.Marshal(res.Scenarios)
		scenariosStr = string(b)
	}

	row := []interface{}{
		nowStr,
		res.Timestamp,
		res.Symbol,
		res.CurrentPrice,
		res.MarketStructure,
		scenariosStr,
	}
	for i, v := range row {
		cell, _ := excelize.CoordinatesToCellName(i+1, nextRow)
		_ = f.SetCellValue(cozeSheetName, cell, v)
	}

	if isNew {
		if err := f.SaveAs(path); err != nil {
			return fmt.Errorf("保存 Excel 失败: %w", err)
		}
	} else {
		if err := f.Save(); err != nil {
			return fmt.Errorf("保存 Excel 失败: %w", err)
		}
	}
	return nil
}

const backtestSheetName = "回测记录"

var backtestHeaders = []string{
	"窗口序号", "Coze回复",
	"Open1", "Close1", "Open2", "Close2", "Open3", "Close3", "Open4", "Close4", "Open5", "Close5",
}

// AppendBacktestRowToExcel 将回测一行（Coze 结果 + 验证 5 根 K 线）追加到当前目录下 backtest_YYYY-MM-DD.xlsx。
// path 为完整路径；若文件不存在会创建并写表头。rowIndex 为窗口序号（从 1 开始）。
func AppendBacktestRowToExcel(path string, rowIndex int, res *CozeStructuredResult, verify5 []*factor.KLine) error {
	var f *excelize.File
	isNew := false
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f = excelize.NewFile()
		f.SetSheetName("Sheet1", backtestSheetName)
		for i, h := range backtestHeaders {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			_ = f.SetCellValue(backtestSheetName, cell, h)
		}
		isNew = true
	} else {
		var err error
		f, err = excelize.OpenFile(path)
		if err != nil {
			return fmt.Errorf("打开 Excel 失败: %w", err)
		}
		defer f.Close()
	}

	rows, err := f.GetRows(backtestSheetName)
	if err != nil {
		return err
	}
	nextRow := len(rows) + 1

	scenariosStr := ""
	if res != nil {
		if len(res.RawAnswer) > 0 && len(res.Scenarios) == 0 {
			scenariosStr = res.RawAnswer
		} else if len(res.Scenarios) > 0 {
			b, _ := json.Marshal(res.Scenarios)
			scenariosStr = string(b)
		}
	}

	row := []interface{}{rowIndex, scenariosStr}
	for i := 0; i < 5; i++ {
		if i < len(verify5) && verify5[i] != nil {
			row = append(row, verify5[i].Open, verify5[i].Close)
		} else {
			row = append(row, "", "")
		}
	}
	for i, v := range row {
		cell, _ := excelize.CoordinatesToCellName(i+1, nextRow)
		_ = f.SetCellValue(backtestSheetName, cell, v)
	}

	if isNew {
		if err := f.SaveAs(path); err != nil {
			return fmt.Errorf("保存 Excel 失败: %w", err)
		}
	} else {
		if err := f.Save(); err != nil {
			return fmt.Errorf("保存 Excel 失败: %w", err)
		}
	}
	return nil
}
