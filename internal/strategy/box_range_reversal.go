package strategy

import (
	"fmt"
	"math"

	"quantitative-trading-app/internal/market"
)

func EvaluateBoxRangeReversal(klines []market.KLine, rawParams BoxRangeReversalParams) []Signal {
	params := rawParams.Normalize()
	minBars := params.ImpulseLookback + params.ConsolidationLookback + 2
	if params.ATRPeriod > minBars {
		minBars = params.ATRPeriod + params.ConsolidationLookback + 2
	}
	if len(klines) < minBars {
		return nil
	}

	atr := computeATR(klines, params.ATRPeriod)
	signals := make([]Signal, 0)
	nextAllowedIndex := 0

	for triggerIndex := minBars - 1; triggerIndex < len(klines)-1; triggerIndex++ {
		if triggerIndex < nextAllowedIndex {
			continue
		}

		box, ok := detectRangeBox(klines, atr, triggerIndex, params)
		if !ok {
			continue
		}

		trigger := klines[triggerIndex]
		if signal, ok := findRangeLongSignal(klines, atr, trigger, triggerIndex, box, params); ok {
			signals = append(signals, signal)
			nextAllowedIndex = signal.TriggerIndex + params.CooldownBars + 1
			continue
		}
		if signal, ok := findRangeShortSignal(klines, atr, trigger, triggerIndex, box, params); ok {
			signals = append(signals, signal)
			nextAllowedIndex = signal.TriggerIndex + params.CooldownBars + 1
		}
	}

	return signals
}

type rangeBox struct {
	High                 float64
	Low                  float64
	Mid                  float64
	ImpulseMovePercent   float64
	ImpulseVolumeAverage float64
	ConsolidationVolume  float64
	ImpulseATRAverage    float64
	ConsolidationATR     float64
	UpperTouches         int
	LowerTouches         int
}

func detectRangeBox(klines []market.KLine, atr []float64, endIndex int, params BoxRangeReversalParams) (rangeBox, bool) {
	consolidationStart := endIndex - params.ConsolidationLookback + 1
	impulseEnd := consolidationStart - 1
	impulseStart := impulseEnd - params.ImpulseLookback + 1
	if impulseStart < 0 || consolidationStart <= 0 {
		return rangeBox{}, false
	}

	boxHigh := highestHigh(klines, endIndex, params.ConsolidationLookback)
	boxLow := lowestLow(klines, endIndex, params.ConsolidationLookback)
	if boxHigh <= boxLow {
		return rangeBox{}, false
	}
	boxWidthPercent := (boxHigh - boxLow) / maxNonZero((boxHigh+boxLow)/2)
	if boxWidthPercent < params.MinBoxWidthPercent || boxWidthPercent > params.MaxBoxWidthPercent {
		return rangeBox{}, false
	}

	impulseMovePercent := math.Abs(klines[impulseEnd].Close-klines[impulseStart].Open) / maxNonZero(klines[impulseStart].Open)
	if impulseMovePercent < params.MinImpulsePercent {
		return rangeBox{}, false
	}

	impulseATRAvg := averageATRRange(atr, impulseStart, impulseEnd)
	consolidationATRAvg := averageATRRange(atr, consolidationStart, endIndex)
	if impulseATRAvg <= 0 || consolidationATRAvg <= 0 {
		return rangeBox{}, false
	}
	if consolidationATRAvg/impulseATRAvg > params.ConsolidationATRRatio {
		return rangeBox{}, false
	}

	impulseRangePercent := (highestHigh(klines, impulseEnd, params.ImpulseLookback) - lowestLow(klines, impulseEnd, params.ImpulseLookback)) / maxNonZero(klines[impulseStart].Open)
	if impulseRangePercent/maxNonZero(consolidationATRAvg/maxNonZero(klines[endIndex].Close)) < params.MinImpulseATRRatio {
		return rangeBox{}, false
	}

	impulseVolumeAvg := averageVolume(klines, impulseStart, impulseEnd)
	consolidationVolumeAvg := averageVolume(klines, consolidationStart, endIndex)
	if impulseVolumeAvg <= 0 || consolidationVolumeAvg <= 0 {
		return rangeBox{}, false
	}
	if consolidationVolumeAvg/impulseVolumeAvg > params.ConsolidationVolumeRatio {
		return rangeBox{}, false
	}

	upperTouches, lowerTouches := countBoundaryTouches(klines, consolidationStart, endIndex, boxHigh, boxLow, params.EdgeTolerancePercent)
	if upperTouches < params.MinBoundaryTouches || lowerTouches < params.MinBoundaryTouches {
		return rangeBox{}, false
	}

	return rangeBox{
		High:                 boxHigh,
		Low:                  boxLow,
		Mid:                  (boxHigh + boxLow) / 2,
		ImpulseMovePercent:   impulseMovePercent,
		ImpulseVolumeAverage: impulseVolumeAvg,
		ConsolidationVolume:  consolidationVolumeAvg,
		ImpulseATRAverage:    impulseATRAvg,
		ConsolidationATR:     consolidationATRAvg,
		UpperTouches:         upperTouches,
		LowerTouches:         lowerTouches,
	}, true
}

func findRangeLongSignal(klines []market.KLine, atr []float64, trigger market.KLine, triggerIndex int, box rangeBox, params BoxRangeReversalParams) (Signal, bool) {
	tolerance := box.Low * params.EdgeTolerancePercent
	if trigger.Low > box.Low+tolerance {
		return Signal{}, false
	}
	if !trigger.IsBullish() || trigger.Close <= box.Low || trigger.Close >= box.Mid {
		return Signal{}, false
	}
	if wickBodyRatioLong(trigger) < params.MinRejectWickBodyRatio {
		return Signal{}, false
	}

	entryIndex := triggerIndex + 1
	entry := klines[entryIndex]
	stopLoss := box.Low - atr[triggerIndex]*params.StopATRMultiplier
	takeProfit := interpolateTakeProfit(DirectionLong, box.Low, box.Mid, box.High, params.TakeProfitFactor)
	if stopLoss >= entry.Open || takeProfit <= entry.Open {
		return Signal{}, false
	}

	return Signal{
		StrategyName:    BoxRangeReversalName,
		Direction:       DirectionLong,
		K1Index:         triggerIndex,
		TriggerIndex:    triggerIndex,
		EntryIndex:      entryIndex,
		K1OpenTime:      trigger.OpenTime,
		TriggerTime:     trigger.OpenTime,
		EntryTime:       entry.OpenTime,
		BoxHigh:         box.High,
		BoxLow:          box.Low,
		EntryPrice:      entry.Open,
		StopLoss:        stopLoss,
		TakeProfit:      takeProfit,
		RiskRewardRatio: 0,
		ConfirmBarOpen:  trigger.Open,
		ConfirmBarClose: trigger.Close,
		ConfirmBarLow:   trigger.Low,
		ConfirmBarHigh:  trigger.High,
		Reason:          fmt.Sprintf("大波动后缩量横盘，下沿反转做多，目标因子 %.2f", params.TakeProfitFactor),
	}, true
}

func findRangeShortSignal(klines []market.KLine, atr []float64, trigger market.KLine, triggerIndex int, box rangeBox, params BoxRangeReversalParams) (Signal, bool) {
	tolerance := box.High * params.EdgeTolerancePercent
	if trigger.High < box.High-tolerance {
		return Signal{}, false
	}
	if !trigger.IsBearish() || trigger.Close >= box.High || trigger.Close <= box.Mid {
		return Signal{}, false
	}
	if wickBodyRatioShort(trigger) < params.MinRejectWickBodyRatio {
		return Signal{}, false
	}

	entryIndex := triggerIndex + 1
	entry := klines[entryIndex]
	stopLoss := box.High + atr[triggerIndex]*params.StopATRMultiplier
	takeProfit := interpolateTakeProfit(DirectionShort, box.Low, box.Mid, box.High, params.TakeProfitFactor)
	if stopLoss <= entry.Open || takeProfit >= entry.Open {
		return Signal{}, false
	}

	return Signal{
		StrategyName:    BoxRangeReversalName,
		Direction:       DirectionShort,
		K1Index:         triggerIndex,
		TriggerIndex:    triggerIndex,
		EntryIndex:      entryIndex,
		K1OpenTime:      trigger.OpenTime,
		TriggerTime:     trigger.OpenTime,
		EntryTime:       entry.OpenTime,
		BoxHigh:         box.High,
		BoxLow:          box.Low,
		EntryPrice:      entry.Open,
		StopLoss:        stopLoss,
		TakeProfit:      takeProfit,
		RiskRewardRatio: 0,
		ConfirmBarOpen:  trigger.Open,
		ConfirmBarClose: trigger.Close,
		ConfirmBarLow:   trigger.Low,
		ConfirmBarHigh:  trigger.High,
		Reason:          fmt.Sprintf("大波动后缩量横盘，上沿反转做空，目标因子 %.2f", params.TakeProfitFactor),
	}, true
}

func averageVolume(klines []market.KLine, start, end int) float64 {
	if start < 0 {
		start = 0
	}
	if end < start {
		return 0
	}
	total := 0.0
	count := 0
	for i := start; i <= end; i++ {
		total += klines[i].Volume
		count++
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func averageATRRange(atr []float64, start, end int) float64 {
	if start < 0 {
		start = 0
	}
	if end < start {
		return 0
	}
	total := 0.0
	count := 0
	for i := start; i <= end; i++ {
		total += atr[i]
		count++
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func countBoundaryTouches(klines []market.KLine, start, end int, boxHigh, boxLow, tolerancePercent float64) (int, int) {
	upperTol := boxHigh * tolerancePercent
	lowerTol := boxLow * tolerancePercent
	upperTouches := 0
	lowerTouches := 0
	for i := start; i <= end; i++ {
		if klines[i].High >= boxHigh-upperTol {
			upperTouches++
		}
		if klines[i].Low <= boxLow+lowerTol {
			lowerTouches++
		}
	}
	return upperTouches, lowerTouches
}

func interpolateTakeProfit(direction Direction, boxLow, mid, boxHigh, factor float64) float64 {
	if direction == DirectionLong {
		return mid + (boxHigh-mid)*factor
	}
	return mid - (mid-boxLow)*factor
}
