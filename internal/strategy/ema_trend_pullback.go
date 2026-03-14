package strategy

import (
	"fmt"
	"math"

	"quantitative-trading-app/internal/market"
)

func EvaluateEMATrendPullback(klines []market.KLine, rawParams EMATrendPullbackParams) []Signal {
	params := rawParams.Normalize()
	minBars := maxInt(params.SlowPeriod+2, maxInt(params.BreakoutLookback+2, params.ATRPeriod+2))
	if len(klines) < minBars+params.PullbackLookahead+1 {
		return nil
	}

	emaFast := computeEMA(klines, params.FastPeriod)
	emaSlow := computeEMA(klines, params.SlowPeriod)
	atr := computeATR(klines, params.ATRPeriod)

	signals := make([]Signal, 0)
	nextAllowedBreakout := 0

	for i := minBars; i < len(klines)-params.PullbackLookahead-1; i++ {
		if i < nextAllowedBreakout {
			continue
		}

		if isLongBreakout(klines, emaFast, emaSlow, i, params) {
			signal, ok := findEMALongPullback(klines, emaFast, atr, i, params)
			if ok {
				signals = append(signals, signal)
				nextAllowedBreakout = signal.TriggerIndex + params.CooldownBars + 1
			}
			continue
		}

		if isShortBreakout(klines, emaFast, emaSlow, i, params) {
			signal, ok := findEMAShortPullback(klines, emaFast, atr, i, params)
			if ok {
				signals = append(signals, signal)
				nextAllowedBreakout = signal.TriggerIndex + params.CooldownBars + 1
			}
		}
	}

	return signals
}

func isLongBreakout(klines []market.KLine, emaFast, emaSlow []float64, index int, params EMATrendPullbackParams) bool {
	bar := klines[index]
	breakoutLevel := highestHigh(klines, index-1, params.BreakoutLookback)
	return bar.IsBullish() &&
		emaFast[index] > emaSlow[index] &&
		emaSlow[index] >= emaSlow[index-1] &&
		bar.Close > breakoutLevel &&
		bar.Close > emaFast[index]
}

func isShortBreakout(klines []market.KLine, emaFast, emaSlow []float64, index int, params EMATrendPullbackParams) bool {
	bar := klines[index]
	breakoutLevel := lowestLow(klines, index-1, params.BreakoutLookback)
	return bar.IsBearish() &&
		emaFast[index] < emaSlow[index] &&
		emaSlow[index] <= emaSlow[index-1] &&
		bar.Close < breakoutLevel &&
		bar.Close < emaFast[index]
}

func findEMALongPullback(klines []market.KLine, emaFast, atr []float64, breakoutIndex int, params EMATrendPullbackParams) (Signal, bool) {
	breakout := klines[breakoutIndex]
	breakoutLevel := highestHigh(klines, breakoutIndex-1, params.BreakoutLookback)
	for offset := 1; offset <= params.PullbackLookahead; offset++ {
		triggerIndex := breakoutIndex + offset
		trigger := klines[triggerIndex]
		pullbackZoneLow := math.Min(breakoutLevel, emaFast[triggerIndex]) * (1 - params.PullbackTolerancePercent)
		pullbackZoneHigh := math.Max(breakoutLevel, emaFast[triggerIndex]) * (1 + params.PullbackTolerancePercent)
		touched := trigger.Low <= pullbackZoneHigh && trigger.Low >= pullbackZoneLow
		if !touched {
			continue
		}
		if !trigger.IsBullish() || trigger.Close < emaFast[triggerIndex] || trigger.Close < breakoutLevel {
			continue
		}
		entryIndex := triggerIndex + 1
		entry := klines[entryIndex]
		stopLoss := math.Min(trigger.Low, emaFast[triggerIndex]) - atr[triggerIndex]*params.StopATRMultiplier
		risk := entry.Open - stopLoss
		if risk <= 0 {
			continue
		}
		takeProfit := entry.Open + risk*params.RiskRewardRatio
		return Signal{
			StrategyName:    EMATrendPullbackName,
			Direction:       DirectionLong,
			K1Index:         breakoutIndex,
			TriggerIndex:    triggerIndex,
			EntryIndex:      entryIndex,
			K1OpenTime:      breakout.OpenTime,
			TriggerTime:     trigger.OpenTime,
			EntryTime:       entry.OpenTime,
			BoxHigh:         breakoutLevel,
			BoxLow:          stopLoss,
			EntryPrice:      entry.Open,
			StopLoss:        stopLoss,
			TakeProfit:      takeProfit,
			RiskRewardRatio: params.RiskRewardRatio,
			ConfirmBarOpen:  trigger.Open,
			ConfirmBarClose: trigger.Close,
			ConfirmBarLow:   trigger.Low,
			ConfirmBarHigh:  trigger.High,
			Reason:          fmt.Sprintf("EMA 趋势向上，突破后在 %d 根内回踩确认做多", offset),
		}, true
	}
	return Signal{}, false
}

func findEMAShortPullback(klines []market.KLine, emaFast, atr []float64, breakoutIndex int, params EMATrendPullbackParams) (Signal, bool) {
	breakout := klines[breakoutIndex]
	breakoutLevel := lowestLow(klines, breakoutIndex-1, params.BreakoutLookback)
	for offset := 1; offset <= params.PullbackLookahead; offset++ {
		triggerIndex := breakoutIndex + offset
		trigger := klines[triggerIndex]
		pullbackZoneLow := math.Min(breakoutLevel, emaFast[triggerIndex]) * (1 - params.PullbackTolerancePercent)
		pullbackZoneHigh := math.Max(breakoutLevel, emaFast[triggerIndex]) * (1 + params.PullbackTolerancePercent)
		touched := trigger.High >= pullbackZoneLow && trigger.High <= pullbackZoneHigh
		if !touched {
			continue
		}
		if !trigger.IsBearish() || trigger.Close > emaFast[triggerIndex] || trigger.Close > breakoutLevel {
			continue
		}
		entryIndex := triggerIndex + 1
		entry := klines[entryIndex]
		stopLoss := math.Max(trigger.High, emaFast[triggerIndex]) + atr[triggerIndex]*params.StopATRMultiplier
		risk := stopLoss - entry.Open
		if risk <= 0 {
			continue
		}
		takeProfit := entry.Open - risk*params.RiskRewardRatio
		return Signal{
			StrategyName:    EMATrendPullbackName,
			Direction:       DirectionShort,
			K1Index:         breakoutIndex,
			TriggerIndex:    triggerIndex,
			EntryIndex:      entryIndex,
			K1OpenTime:      breakout.OpenTime,
			TriggerTime:     trigger.OpenTime,
			EntryTime:       entry.OpenTime,
			BoxHigh:         stopLoss,
			BoxLow:          breakoutLevel,
			EntryPrice:      entry.Open,
			StopLoss:        stopLoss,
			TakeProfit:      takeProfit,
			RiskRewardRatio: params.RiskRewardRatio,
			ConfirmBarOpen:  trigger.Open,
			ConfirmBarClose: trigger.Close,
			ConfirmBarLow:   trigger.Low,
			ConfirmBarHigh:  trigger.High,
			Reason:          fmt.Sprintf("EMA 趋势向下，跌破后在 %d 根内反抽确认做空", offset),
		}, true
	}
	return Signal{}, false
}

func computeEMA(klines []market.KLine, period int) []float64 {
	result := make([]float64, len(klines))
	if len(klines) == 0 {
		return result
	}
	multiplier := 2.0 / float64(period+1)
	result[0] = klines[0].Close
	for i := 1; i < len(klines); i++ {
		result[i] = (klines[i].Close-result[i-1])*multiplier + result[i-1]
	}
	return result
}

func computeATR(klines []market.KLine, period int) []float64 {
	result := make([]float64, len(klines))
	if len(klines) == 0 {
		return result
	}
	for i := 1; i < len(klines); i++ {
		tr := trueRange(klines[i], klines[i-1].Close)
		if i == 1 {
			result[i] = tr
			continue
		}
		result[i] = ((result[i-1] * float64(period-1)) + tr) / float64(period)
	}
	if len(result) > 1 && result[0] == 0 {
		result[0] = result[1]
	}
	return result
}

func trueRange(current market.KLine, prevClose float64) float64 {
	return math.Max(current.High-current.Low, math.Max(math.Abs(current.High-prevClose), math.Abs(current.Low-prevClose)))
}

func highestHigh(klines []market.KLine, endIndex, lookback int) float64 {
	start := endIndex - lookback + 1
	if start < 0 {
		start = 0
	}
	high := klines[start].High
	for i := start + 1; i <= endIndex; i++ {
		if klines[i].High > high {
			high = klines[i].High
		}
	}
	return high
}

func lowestLow(klines []market.KLine, endIndex, lookback int) float64 {
	start := endIndex - lookback + 1
	if start < 0 {
		start = 0
	}
	low := klines[start].Low
	for i := start + 1; i <= endIndex; i++ {
		if klines[i].Low < low {
			low = klines[i].Low
		}
	}
	return low
}
