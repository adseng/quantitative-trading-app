package factor

// FactorMACD MACD因子。
// 基于回测结论：MACD线 < 信号线时下一根更易上涨，故输出正向看涨信号。
// MACD > 信号线 → 看跌；MACD < 信号线 → 看涨。
// fast, slow, signalN: EMA周期（常用12, 26, 9），weight: 权重（多因子用）
func (e *SignalContext) FactorMACD(fast, slow, signalN int, weight float64) *SignalContext {
	if e.KLine == nil {
		return e
	}
	prices := e.KLine.ClosePrices()
	need := slow + signalN
	if len(prices) < need {
		return e
	}

	// prices is newest-first, reverse for EMA calculation (oldest first)
	reversed := make([]float64, len(prices))
	for i, p := range prices {
		reversed[len(prices)-1-i] = p
	}

	emaFast := calcEMA(reversed, fast)
	emaSlow := calcEMA(reversed, slow)

	if len(emaFast) == 0 || len(emaSlow) == 0 {
		return e
	}

	// Build MACD line: align by using the last N values where both exist
	macdLen := len(emaFast)
	if l := len(emaSlow); l < macdLen {
		macdLen = l
	}
	macdLine := make([]float64, macdLen)
	for i := 0; i < macdLen; i++ {
		fi := len(emaFast) - macdLen + i
		si := len(emaSlow) - macdLen + i
		macdLine[i] = emaFast[fi] - emaSlow[si]
	}

	if len(macdLine) < signalN {
		return e
	}

	signalLine := calcEMA(macdLine, signalN)
	if len(signalLine) == 0 {
		return e
	}

	lastMACD := macdLine[len(macdLine)-1]
	lastSignal := signalLine[len(signalLine)-1]

	if lastMACD < lastSignal {
		e.AddBull(weight)
	} else if lastMACD > lastSignal {
		e.AddBear(weight)
	}
	return e
}

// calcEMA computes EMA values from a time-series (oldest first). Returns empty if not enough data.
func calcEMA(prices []float64, period int) []float64 {
	if len(prices) < period || period <= 0 {
		return nil
	}
	multiplier := 2.0 / float64(period+1)
	result := make([]float64, 0, len(prices)-period+1)

	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	sma := sum / float64(period)
	result = append(result, sma)

	for i := period; i < len(prices); i++ {
		ema := (prices[i]-result[len(result)-1])*multiplier + result[len(result)-1]
		result = append(result, ema)
	}
	return result
}
