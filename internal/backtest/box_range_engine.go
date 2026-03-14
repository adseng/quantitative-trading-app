package backtest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"quantitative-trading-app/internal/datafile"
	"quantitative-trading-app/internal/market"
	"quantitative-trading-app/internal/strategy"
)

func RunBoxRange(req RunBoxRangeRequest) (*BoxRangeReport, error) {
	req = req.Normalize()
	klines, err := datafile.LoadKlines(req.DataPath)
	if err != nil {
		return nil, err
	}
	if len(klines) == 0 {
		return nil, fmt.Errorf("no klines loaded from %s", req.DataPath)
	}
	return RunBoxRangeOnKlines(req, marketPointersToValues(klines))
}

func RunBoxRangeOnKlines(req RunBoxRangeRequest, klines []market.KLine) (*BoxRangeReport, error) {
	req = req.Normalize()
	if len(klines) < 3 {
		return nil, fmt.Errorf("not enough klines for backtest")
	}

	signals := strategy.EvaluateBoxRangeReversal(klines, req.Params)
	report := NewBoxRangeReport(req, klines, signals)
	report.Trades, report.Summary = simulateTrades(klines, signals, req.InitialBalance, req.PositionSizeUSDT)
	return &report, nil
}

func DefaultBoxRangeResultPath(strategyName string) string {
	if strategyName == "" {
		strategyName = strategy.BoxRangeReversalName
	}
	return filepath.ToSlash(filepath.Join(datafile.DefaultDir, fmt.Sprintf("test-%s.txt", strategyName)))
}

func SaveBoxRangeReport(path string, report *BoxRangeReport) error {
	if path == "" {
		path = report.ResultPath
	}
	if path == "" {
		path = DefaultBoxRangeResultPath(report.StrategyName)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}
