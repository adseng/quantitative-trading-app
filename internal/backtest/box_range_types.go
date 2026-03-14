package backtest

import (
	"time"

	"quantitative-trading-app/internal/market"
	"quantitative-trading-app/internal/strategy"
)

type RunBoxRangeRequest struct {
	DataPath         string                          `json:"dataPath"`
	StrategyName     string                          `json:"strategyName"`
	Params           strategy.BoxRangeReversalParams `json:"params"`
	InitialBalance   float64                         `json:"initialBalance"`
	PositionSizeUSDT float64                         `json:"positionSizeUSDT"`
	ResultPath       string                          `json:"resultPath"`
}

func DefaultBoxRangeRunRequest() RunBoxRangeRequest {
	return RunBoxRangeRequest{
		StrategyName:     strategy.BoxRangeReversalName,
		Params:           strategy.DefaultBoxRangeReversalParams(),
		InitialBalance:   10000,
		PositionSizeUSDT: 100,
	}
}

func (r RunBoxRangeRequest) Normalize() RunBoxRangeRequest {
	defaults := DefaultBoxRangeRunRequest()
	if r.StrategyName == "" {
		r.StrategyName = defaults.StrategyName
	}
	r.Params = r.Params.Normalize()
	if r.InitialBalance <= 0 {
		r.InitialBalance = defaults.InitialBalance
	}
	if r.PositionSizeUSDT <= 0 {
		r.PositionSizeUSDT = defaults.PositionSizeUSDT
	}
	return r
}

type BoxRangeReport struct {
	StrategyName     string                          `json:"strategyName"`
	DataPath         string                          `json:"dataPath"`
	ResultPath       string                          `json:"resultPath"`
	GeneratedAt      string                          `json:"generatedAt"`
	InitialBalance   float64                         `json:"initialBalance"`
	PositionSizeUSDT float64                         `json:"positionSizeUSDT"`
	Params           strategy.BoxRangeReversalParams `json:"params"`
	Klines           []market.KLine                  `json:"klines"`
	Signals          []strategy.Signal               `json:"signals"`
	Trades           []Trade                         `json:"trades"`
	Summary          Summary                         `json:"summary"`
}

func NewBoxRangeReport(req RunBoxRangeRequest, klines []market.KLine, signals []strategy.Signal) BoxRangeReport {
	return BoxRangeReport{
		StrategyName:     req.StrategyName,
		DataPath:         req.DataPath,
		ResultPath:       req.ResultPath,
		GeneratedAt:      time.Now().Format(time.RFC3339),
		InitialBalance:   req.InitialBalance,
		PositionSizeUSDT: req.PositionSizeUSDT,
		Params:           req.Params,
		Klines:           klines,
		Signals:          signals,
	}
}
