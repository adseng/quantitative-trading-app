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

func RunEMA(req RunEMARequest) (*EMAReport, error) {
	req = req.Normalize()
	klines, err := datafile.LoadKlines(req.DataPath)
	if err != nil {
		return nil, err
	}
	if len(klines) == 0 {
		return nil, fmt.Errorf("no klines loaded from %s", req.DataPath)
	}
	return RunEMAOnKlines(req, marketPointersToValues(klines))
}

func RunEMAOnKlines(req RunEMARequest, klines []market.KLine) (*EMAReport, error) {
	req = req.Normalize()
	if len(klines) < 3 {
		return nil, fmt.Errorf("not enough klines for backtest")
	}

	signals := strategy.EvaluateEMATrendPullback(klines, req.Params)
	report := NewEMAReport(req, klines, signals)
	report.Trades, report.Summary = simulateTrades(klines, signals, req.InitialBalance, req.PositionSizeUSDT)
	return &report, nil
}

func DefaultEMAResultPath(strategyName string) string {
	if strategyName == "" {
		strategyName = strategy.EMATrendPullbackName
	}
	return filepath.ToSlash(filepath.Join(datafile.DefaultDir, fmt.Sprintf("test-%s.txt", strategyName)))
}

func SaveEMAReport(path string, report *EMAReport) error {
	if path == "" {
		path = report.ResultPath
	}
	if path == "" {
		path = DefaultEMAResultPath(report.StrategyName)
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
