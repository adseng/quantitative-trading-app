package strategy

import (
	"fmt"
	"math"

	"quantitative-trading-app/internal/market"
)

func EvaluateBoxPullback(klines []market.KLine, rawParams BoxPullbackParams) []Signal {
	params := rawParams.Normalize()
	minBars := maxInt(params.LookaheadN+2, maxInt(params.K1StrengthLookback+1, params.TrendMAPeriod+1))
	if len(klines) < minBars {
		return nil
	}

	signals := make([]Signal, 0)
	nextAllowedK1 := 0

	for i := 0; i < len(klines)-params.LookaheadN-1; i++ {
		if i < nextAllowedK1 {
			continue
		}

		k1 := klines[i]
		if !isStrongK1(klines, i, params) {
			continue
		}

		if k1.IsBullish() {
			signal, ok := findLongSignal(klines, i, params)
			if ok {
				signals = append(signals, signal)
				nextAllowedK1 = signal.TriggerIndex + params.CooldownBars + 1
			}
			continue
		}

		if k1.IsBearish() {
			signal, ok := findShortSignal(klines, i, params)
			if ok {
				signals = append(signals, signal)
				nextAllowedK1 = signal.TriggerIndex + params.CooldownBars + 1
			}
		}
	}

	return signals
}

func isStrongK1(klines []market.KLine, index int, params BoxPullbackParams) bool {
	k1 := klines[index]
	if !k1.IsBullish() && !k1.IsBearish() {
		return false
	}
	if k1.Range() <= 0 {
		return false
	}
	if k1.BodyPercent() < params.MinK1BodyPercent {
		return false
	}
	boxRangePercent := k1.Range() / maxNonZero(k1.Open)
	if boxRangePercent < params.MinBoxRangePercent || boxRangePercent > params.MaxBoxRangePercent {
		return false
	}
	avgBody := averageBodySize(klines, index, params.K1StrengthLookback)
	if avgBody <= 0 {
		return false
	}
	if k1.BodySize()/avgBody < params.MinK1BodyToAvgRatio {
		return false
	}
	return passesTrendFilter(klines, index, params.TrendMAPeriod, k1.IsBullish())
}

func findLongSignal(klines []market.KLine, k1Index int, params BoxPullbackParams) (Signal, bool) {
	k1 := klines[k1Index]
	boxHigh := k1.High
	boxLow := k1.Low
	tolerance := boxHigh * params.TouchTolerancePercent

	for offset := 1; offset <= params.LookaheadN; offset++ {
		triggerIndex := k1Index + offset
		trigger := klines[triggerIndex]
		if !trigger.IsBearish() {
			continue
		}
		if trigger.Low > boxHigh+tolerance || trigger.Low < boxLow-tolerance {
			continue
		}
		if trigger.Close <= boxHigh {
			continue
		}
		if wickBodyRatioLong(trigger) < params.MinConfirmWickBodyRatio {
			continue
		}

		entryIndex := triggerIndex + 1
		entry := klines[entryIndex]
		stopLoss := boxLow
		risk := entry.Open - stopLoss
		if risk <= 0 {
			continue
		}
		takeProfit := entry.Open + risk*params.RiskRewardRatio
		return Signal{
			StrategyName:    BoxPullbackName,
			Direction:       DirectionLong,
			K1Index:         k1Index,
			TriggerIndex:    triggerIndex,
			EntryIndex:      entryIndex,
			K1OpenTime:      k1.OpenTime,
			TriggerTime:     trigger.OpenTime,
			EntryTime:       entry.OpenTime,
			BoxHigh:         boxHigh,
			BoxLow:          boxLow,
			EntryPrice:      entry.Open,
			StopLoss:        stopLoss,
			TakeProfit:      takeProfit,
			RiskRewardRatio: params.RiskRewardRatio,
			ConfirmBarOpen:  trigger.Open,
			ConfirmBarClose: trigger.Close,
			ConfirmBarLow:   trigger.Low,
			ConfirmBarHigh:  trigger.High,
			Reason:          fmt.Sprintf("K1 阳线后 %d 根内回踩箱体并收回箱顶上方", offset),
		}, true
	}

	return Signal{}, false
}

func findShortSignal(klines []market.KLine, k1Index int, params BoxPullbackParams) (Signal, bool) {
	k1 := klines[k1Index]
	boxHigh := k1.High
	boxLow := k1.Low
	tolerance := math.Abs(boxLow) * params.TouchTolerancePercent

	for offset := 1; offset <= params.LookaheadN; offset++ {
		triggerIndex := k1Index + offset
		trigger := klines[triggerIndex]
		if !trigger.IsBullish() {
			continue
		}
		if trigger.High < boxLow-tolerance || trigger.High > boxHigh+tolerance {
			continue
		}
		if trigger.Close >= boxLow {
			continue
		}
		if wickBodyRatioShort(trigger) < params.MinConfirmWickBodyRatio {
			continue
		}

		entryIndex := triggerIndex + 1
		entry := klines[entryIndex]
		stopLoss := boxHigh
		risk := stopLoss - entry.Open
		if risk <= 0 {
			continue
		}
		takeProfit := entry.Open - risk*params.RiskRewardRatio
		return Signal{
			StrategyName:    BoxPullbackName,
			Direction:       DirectionShort,
			K1Index:         k1Index,
			TriggerIndex:    triggerIndex,
			EntryIndex:      entryIndex,
			K1OpenTime:      k1.OpenTime,
			TriggerTime:     trigger.OpenTime,
			EntryTime:       entry.OpenTime,
			BoxHigh:         boxHigh,
			BoxLow:          boxLow,
			EntryPrice:      entry.Open,
			StopLoss:        stopLoss,
			TakeProfit:      takeProfit,
			RiskRewardRatio: params.RiskRewardRatio,
			ConfirmBarOpen:  trigger.Open,
			ConfirmBarClose: trigger.Close,
			ConfirmBarLow:   trigger.Low,
			ConfirmBarHigh:  trigger.High,
			Reason:          fmt.Sprintf("K1 阴线后 %d 根内反抽箱体并收回箱底下方", offset),
		}, true
	}

	return Signal{}, false
}

func wickBodyRatioLong(k market.KLine) float64 {
	body := math.Abs(k.Close - k.Open)
	if body == 0 {
		return 0
	}
	lowerWick := math.Min(k.Open, k.Close) - k.Low
	if lowerWick < 0 {
		lowerWick = 0
	}
	return lowerWick / body
}

func wickBodyRatioShort(k market.KLine) float64 {
	body := math.Abs(k.Close - k.Open)
	if body == 0 {
		return 0
	}
	upperWick := k.High - math.Max(k.Open, k.Close)
	if upperWick < 0 {
		upperWick = 0
	}
	return upperWick / body
}

func averageBodySize(klines []market.KLine, endIndex, lookback int) float64 {
	start := endIndex - lookback
	if start < 0 {
		start = 0
	}
	total := 0.0
	count := 0
	for i := start; i < endIndex; i++ {
		total += klines[i].BodySize()
		count++
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func passesTrendFilter(klines []market.KLine, index, period int, wantLong bool) bool {
	if period <= 1 || index < period {
		return false
	}
	currentMA := simpleMA(klines, index, period)
	prevMA := simpleMA(klines, index-1, period)
	price := klines[index].Close
	if wantLong {
		return price >= currentMA && currentMA >= prevMA
	}
	return price <= currentMA && currentMA <= prevMA
}

func simpleMA(klines []market.KLine, endIndex, period int) float64 {
	start := endIndex - period + 1
	if start < 0 {
		start = 0
	}
	sum := 0.0
	count := 0
	for i := start; i <= endIndex; i++ {
		sum += klines[i].Close
		count++
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxNonZero(v float64) float64 {
	if math.Abs(v) < 1e-9 {
		return 1
	}
	return math.Abs(v)
}
