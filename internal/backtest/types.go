package backtest

import (
	"time"

	"quantitative-trading-app/internal/market"
	"quantitative-trading-app/internal/strategy"
)

type RunRequest struct {
	DataPath         string                     `json:"dataPath"`
	StrategyName     string                     `json:"strategyName"`
	Params           strategy.BoxPullbackParams `json:"params"`
	InitialBalance   float64                    `json:"initialBalance"`
	PositionSizeUSDT float64                    `json:"positionSizeUSDT"`
	ResultPath       string                     `json:"resultPath"`
}

func DefaultRunRequest() RunRequest {
	return RunRequest{
		StrategyName:     strategy.BoxPullbackName,
		Params:           strategy.DefaultBoxPullbackParams(),
		InitialBalance:   10000,
		PositionSizeUSDT: 100,
	}
}

func (r RunRequest) Normalize() RunRequest {
	defaults := DefaultRunRequest()
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

type Trade struct {
	ID             int                `json:"id"`
	StrategyName   string             `json:"strategyName"`
	Direction      strategy.Direction `json:"direction"`
	SignalIndex    int                `json:"signalIndex"`
	EntryIndex     int                `json:"entryIndex"`
	ExitIndex      int                `json:"exitIndex"`
	SignalTime     int64              `json:"signalTime"`
	EntryTime      int64              `json:"entryTime"`
	ExitTime       int64              `json:"exitTime"`
	EntryPrice     float64            `json:"entryPrice"`
	ExitPrice      float64            `json:"exitPrice"`
	StopLoss       float64            `json:"stopLoss"`
	TakeProfit     float64            `json:"takeProfit"`
	Quantity       float64            `json:"quantity"`
	OrderValueUSDT float64            `json:"orderValueUSDT"`
	PnL            float64            `json:"pnl"`
	PnLPercent     float64            `json:"pnlPercent"`
	BalanceAfter   float64            `json:"balanceAfter"`
	ExitReason     string             `json:"exitReason"`
	HoldBars       int                `json:"holdBars"`
}

type Summary struct {
	TotalSignals   int     `json:"totalSignals"`
	ExecutedTrades int     `json:"executedTrades"`
	Wins           int     `json:"wins"`
	Losses         int     `json:"losses"`
	SkippedSignals int     `json:"skippedSignals"`
	WinRate        float64 `json:"winRate"`
	FinalBalance   float64 `json:"finalBalance"`
	TotalPnL       float64 `json:"totalPnL"`
	ROI            float64 `json:"roi"`
	MaxDrawdown    float64 `json:"maxDrawdown"`
}

type Report struct {
	StrategyName     string                     `json:"strategyName"`
	DataPath         string                     `json:"dataPath"`
	ResultPath       string                     `json:"resultPath"`
	GeneratedAt      string                     `json:"generatedAt"`
	InitialBalance   float64                    `json:"initialBalance"`
	PositionSizeUSDT float64                    `json:"positionSizeUSDT"`
	Params           strategy.BoxPullbackParams `json:"params"`
	Klines           []market.KLine             `json:"klines"`
	Signals          []strategy.Signal          `json:"signals"`
	Trades           []Trade                    `json:"trades"`
	Summary          Summary                    `json:"summary"`
}

func NewReport(req RunRequest, klines []market.KLine, signals []strategy.Signal) Report {
	return Report{
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
