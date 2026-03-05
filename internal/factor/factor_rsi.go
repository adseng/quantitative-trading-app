package factor

// FactorRSI RSI因子：RSI < oversold → 看涨(信号1)；RSI > overbought → 看跌(信号-1)。
// period: RSI周期（常用14），overbought/oversold: 超买超卖阈值（常用70/30），weight: 权重。
func (e *SignalContext) FactorRSI(period int, overbought, oversold, weight float64) *SignalContext {
	if e.KLine == nil || len(e.KLine.History) < period+1 {
		return e
	}
	prices := e.KLine.ClosePrices()
	if len(prices) < period+1 {
		return e
	}

	// Calculate RSI using prices[0..period] (most recent first)
	var gainSum, lossSum float64
	for i := 0; i < period; i++ {
		diff := prices[i] - prices[i+1] // newer - older (History is newest first)
		if diff > 0 {
			gainSum += diff
		} else {
			lossSum += -diff
		}
	}
	avgGain := gainSum / float64(period)
	avgLoss := lossSum / float64(period)

	if avgLoss == 0 {
		// RSI = 100, overbought
		if 100 > overbought {
			e.AddBear(weight)
		}
		return e
	}

	rs := avgGain / avgLoss
	rsi := 100.0 - 100.0/(1.0+rs)

	if rsi < oversold {
		e.AddBull(weight)
	} else if rsi > overbought {
		e.AddBear(weight)
	}
	return e
}

// ComputeRSI calculates RSI value for display purposes.
func ComputeRSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50
	}
	var gainSum, lossSum float64
	for i := 0; i < period; i++ {
		diff := prices[i] - prices[i+1]
		if diff > 0 {
			gainSum += diff
		} else {
			lossSum += -diff
		}
	}
	avgGain := gainSum / float64(period)
	avgLoss := lossSum / float64(period)
	if avgLoss == 0 {
		return 100
	}
	rs := avgGain / avgLoss
	return 100.0 - 100.0/(1.0+rs)
}
