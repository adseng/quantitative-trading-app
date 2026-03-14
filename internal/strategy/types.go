package strategy

type Direction string

const (
	DirectionLong  Direction = "LONG"
	DirectionShort Direction = "SHORT"
)

const (
	BoxPullbackName      = "box-pullback-confirmation"
	EMATrendPullbackName = "ema-trend-pullback-confirmation"
)

type BoxPullbackParams struct {
	LookaheadN              int     `json:"lookaheadN"`
	MinK1BodyPercent        float64 `json:"minK1BodyPercent"`
	K1StrengthLookback      int     `json:"k1StrengthLookback"`
	MinK1BodyToAvgRatio     float64 `json:"minK1BodyToAvgRatio"`
	TrendMAPeriod           int     `json:"trendMAPeriod"`
	MinBoxRangePercent      float64 `json:"minBoxRangePercent"`
	MaxBoxRangePercent      float64 `json:"maxBoxRangePercent"`
	TouchTolerancePercent   float64 `json:"touchTolerancePercent"`
	MinConfirmWickBodyRatio float64 `json:"minConfirmWickBodyRatio"`
	CooldownBars            int     `json:"cooldownBars"`
	RiskRewardRatio         float64 `json:"riskRewardRatio"`
}

type EMATrendPullbackParams struct {
	FastPeriod               int     `json:"fastPeriod"`
	SlowPeriod               int     `json:"slowPeriod"`
	BreakoutLookback         int     `json:"breakoutLookback"`
	PullbackLookahead        int     `json:"pullbackLookahead"`
	PullbackTolerancePercent float64 `json:"pullbackTolerancePercent"`
	ATRPeriod                int     `json:"atrPeriod"`
	StopATRMultiplier        float64 `json:"stopATRMultiplier"`
	CooldownBars             int     `json:"cooldownBars"`
	RiskRewardRatio          float64 `json:"riskRewardRatio"`
}

func DefaultBoxPullbackParams() BoxPullbackParams {
	return BoxPullbackParams{
		LookaheadN:              5,
		MinK1BodyPercent:        0.003,
		K1StrengthLookback:      20,
		MinK1BodyToAvgRatio:     1.5,
		TrendMAPeriod:           50,
		MinBoxRangePercent:      0.002,
		MaxBoxRangePercent:      0.03,
		TouchTolerancePercent:   0.001,
		MinConfirmWickBodyRatio: 1.2,
		CooldownBars:            3,
		RiskRewardRatio:         2,
	}
}

func DefaultEMATrendPullbackParams() EMATrendPullbackParams {
	return EMATrendPullbackParams{
		FastPeriod:               20,
		SlowPeriod:               60,
		BreakoutLookback:         20,
		PullbackLookahead:        5,
		PullbackTolerancePercent: 0.003,
		ATRPeriod:                14,
		StopATRMultiplier:        1,
		CooldownBars:             3,
		RiskRewardRatio:          1.5,
	}
}

func (p BoxPullbackParams) Normalize() BoxPullbackParams {
	defaults := DefaultBoxPullbackParams()

	if p.LookaheadN <= 0 {
		p.LookaheadN = defaults.LookaheadN
	}
	if p.MinK1BodyPercent <= 0 {
		p.MinK1BodyPercent = defaults.MinK1BodyPercent
	}
	if p.K1StrengthLookback <= 0 {
		p.K1StrengthLookback = defaults.K1StrengthLookback
	}
	if p.MinK1BodyToAvgRatio <= 0 {
		p.MinK1BodyToAvgRatio = defaults.MinK1BodyToAvgRatio
	}
	if p.TrendMAPeriod <= 1 {
		p.TrendMAPeriod = defaults.TrendMAPeriod
	}
	if p.MinBoxRangePercent <= 0 {
		p.MinBoxRangePercent = defaults.MinBoxRangePercent
	}
	if p.MaxBoxRangePercent <= 0 {
		p.MaxBoxRangePercent = defaults.MaxBoxRangePercent
	}
	if p.MinBoxRangePercent > p.MaxBoxRangePercent {
		p.MinBoxRangePercent, p.MaxBoxRangePercent = defaults.MinBoxRangePercent, defaults.MaxBoxRangePercent
	}
	if p.TouchTolerancePercent < 0 {
		p.TouchTolerancePercent = defaults.TouchTolerancePercent
	}
	if p.MinConfirmWickBodyRatio <= 0 {
		p.MinConfirmWickBodyRatio = defaults.MinConfirmWickBodyRatio
	}
	if p.CooldownBars < 0 {
		p.CooldownBars = defaults.CooldownBars
	}
	if p.RiskRewardRatio <= 0 {
		p.RiskRewardRatio = defaults.RiskRewardRatio
	}

	return p
}

func (p EMATrendPullbackParams) Normalize() EMATrendPullbackParams {
	defaults := DefaultEMATrendPullbackParams()

	if p.FastPeriod <= 1 {
		p.FastPeriod = defaults.FastPeriod
	}
	if p.SlowPeriod <= p.FastPeriod {
		p.SlowPeriod = defaults.SlowPeriod
	}
	if p.BreakoutLookback <= 1 {
		p.BreakoutLookback = defaults.BreakoutLookback
	}
	if p.PullbackLookahead <= 0 {
		p.PullbackLookahead = defaults.PullbackLookahead
	}
	if p.PullbackTolerancePercent < 0 {
		p.PullbackTolerancePercent = defaults.PullbackTolerancePercent
	}
	if p.ATRPeriod <= 1 {
		p.ATRPeriod = defaults.ATRPeriod
	}
	if p.StopATRMultiplier <= 0 {
		p.StopATRMultiplier = defaults.StopATRMultiplier
	}
	if p.CooldownBars < 0 {
		p.CooldownBars = defaults.CooldownBars
	}
	if p.RiskRewardRatio <= 0 {
		p.RiskRewardRatio = defaults.RiskRewardRatio
	}

	return p
}

type Signal struct {
	StrategyName    string    `json:"strategyName"`
	Direction       Direction `json:"direction"`
	K1Index         int       `json:"k1Index"`
	TriggerIndex    int       `json:"triggerIndex"`
	EntryIndex      int       `json:"entryIndex"`
	K1OpenTime      int64     `json:"k1OpenTime"`
	TriggerTime     int64     `json:"triggerTime"`
	EntryTime       int64     `json:"entryTime"`
	BoxHigh         float64   `json:"boxHigh"`
	BoxLow          float64   `json:"boxLow"`
	EntryPrice      float64   `json:"entryPrice"`
	StopLoss        float64   `json:"stopLoss"`
	TakeProfit      float64   `json:"takeProfit"`
	RiskRewardRatio float64   `json:"riskRewardRatio"`
	ConfirmBarOpen  float64   `json:"confirmBarOpen"`
	ConfirmBarClose float64   `json:"confirmBarClose"`
	ConfirmBarLow   float64   `json:"confirmBarLow"`
	ConfirmBarHigh  float64   `json:"confirmBarHigh"`
	Reason          string    `json:"reason"`
}
