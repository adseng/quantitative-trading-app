package appservice

import (
	"context"
	"path/filepath"

	"quantitative-trading-app/internal/backtest"
	"quantitative-trading-app/internal/datafile"
	"quantitative-trading-app/internal/market"
)

type EventEmitter func(eventName string, payload any)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) BindRuntime(_ context.Context, _ EventEmitter) {}

func (s *Service) LoadLocalKlines(path string) ([]market.KLine, error) {
	klines, err := datafile.LoadKlines(path)
	if err != nil {
		return nil, err
	}
	return valuesFromPointers(klines), nil
}

func (s *Service) RunBacktest(req backtest.RunRequest) (backtest.Report, error) {
	req = req.Normalize()
	if req.ResultPath == "" {
		req.ResultPath = backtest.DefaultResultPath(req.StrategyName)
	}
	req.ResultPath = filepath.ToSlash(req.ResultPath)

	report, err := backtest.Run(req)
	if err != nil {
		return backtest.Report{}, err
	}
	if err := backtest.SaveReport(req.ResultPath, report); err != nil {
		return backtest.Report{}, err
	}
	return *report, nil
}

func (s *Service) RunEMABacktest(req backtest.RunEMARequest) (backtest.EMAReport, error) {
	req = req.Normalize()
	if req.ResultPath == "" {
		req.ResultPath = backtest.DefaultEMAResultPath(req.StrategyName)
	}
	req.ResultPath = filepath.ToSlash(req.ResultPath)

	report, err := backtest.RunEMA(req)
	if err != nil {
		return backtest.EMAReport{}, err
	}
	if err := backtest.SaveEMAReport(req.ResultPath, report); err != nil {
		return backtest.EMAReport{}, err
	}
	return *report, nil
}

func valuesFromPointers(items []*market.KLine) []market.KLine {
	out := make([]market.KLine, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, *item)
	}
	return out
}
