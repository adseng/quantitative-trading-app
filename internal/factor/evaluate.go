package factor

// FactorConfig 因子配置，统一传参。
type FactorConfig struct {
	UseMA     bool
	MaShort   int
	MaLong    int
	MaWeight  float64

	UseTrend    bool
	TrendN      int
	TrendWeight float64

	UseRSI        bool
	RSIPeriod     int
	RSIOverbought float64
	RSIOversold   float64
	RSIWeight     float64

	UseMACD    bool
	MACDFast   int
	MACDSlow   int
	MACDSignal int
	MACDWeight float64

	UseBoll        bool
	BollPeriod     int
	BollMultiplier float64
	BollWeight     float64

	UseBreakout    bool
	BreakoutPeriod int
	BreakoutWeight float64

	UsePriceVsMA    bool
	PriceVsMAPeriod int
	PriceVsMAWeight float64

	UseATR    bool
	ATRPeriod int
	ATRWeight float64

	UseVolume    bool
	VolumePeriod int
	VolumeWeight float64

	UseSession    bool
	SessionWeight float64

	// 均线金叉/死叉（事件型 + 时间容错 + 预判）
	UseMACross      bool
	MACrossShort    int
	MACrossLong     int
	MACrossWeight   float64
	MACrossWindow   int     // 容错根数，0=仅当根有交叉
	MACrossPreempt  float64 // 预判阈值，0=关闭
}

// MinHistory 计算该配置所需的最小 K 线数量。
func (cfg *FactorConfig) MinHistory() int {
	m := 2
	if cfg.UseMA && cfg.MaLong+2 > m {
		m = cfg.MaLong + 2
	}
	if cfg.UseTrend && cfg.TrendN+2 > m {
		m = cfg.TrendN + 2
	}
	if cfg.UseRSI && cfg.RSIPeriod+2 > m {
		m = cfg.RSIPeriod + 2
	}
	if cfg.UseMACD {
		need := cfg.MACDSlow + cfg.MACDSignal + 2
		if need > m {
			m = need
		}
	}
	if cfg.UseBoll && cfg.BollPeriod+2 > m {
		m = cfg.BollPeriod + 2
	}
	if cfg.UseBreakout && cfg.BreakoutPeriod+2 > m {
		m = cfg.BreakoutPeriod + 2
	}
	if cfg.UsePriceVsMA && cfg.PriceVsMAPeriod+2 > m {
		m = cfg.PriceVsMAPeriod + 2
	}
	if cfg.UseATR && cfg.ATRPeriod+2 > m {
		m = cfg.ATRPeriod + 2
	}
	if cfg.UseVolume && cfg.VolumePeriod+2 > m {
		m = cfg.VolumePeriod + 2
	}
	if cfg.UseMACross {
		need := cfg.MACrossLong + 2 + cfg.MACrossWindow
		if need > m {
			m = need
		}
	}
	return m
}

// EvaluateWithConfig 使用 FactorConfig 执行所有启用的因子。
func EvaluateWithConfig(kl *KLineHistory, cfg *FactorConfig) *SignalContext {
	ctx := NewSignalContext(kl)
	if cfg.UseMA {
		ctx.FactorMA(cfg.MaShort, cfg.MaLong, cfg.MaWeight)
	}
	if cfg.UseTrend {
		ctx.FactorTrend(cfg.TrendN, cfg.TrendWeight)
	}
	if cfg.UseRSI {
		ctx.FactorRSI(cfg.RSIPeriod, cfg.RSIOverbought, cfg.RSIOversold, cfg.RSIWeight)
	}
	if cfg.UseMACD {
		ctx.FactorMACD(cfg.MACDFast, cfg.MACDSlow, cfg.MACDSignal, cfg.MACDWeight)
	}
	if cfg.UseBoll {
		ctx.FactorBoll(cfg.BollPeriod, cfg.BollMultiplier, cfg.BollWeight)
	}
	if cfg.UseBreakout {
		ctx.FactorBreakout(cfg.BreakoutPeriod, cfg.BreakoutWeight)
	}
	if cfg.UsePriceVsMA {
		ctx.FactorPriceVsMA(cfg.PriceVsMAPeriod, cfg.PriceVsMAWeight)
	}
	if cfg.UseATR {
		ctx.FactorATR(cfg.ATRPeriod, cfg.ATRWeight)
	}
	if cfg.UseVolume {
		ctx.FactorVolume(cfg.VolumePeriod, cfg.VolumeWeight)
	}
	if cfg.UseSession {
		ctx.FactorSession(cfg.SessionWeight)
	}
	if cfg.UseMACross {
		w, p := cfg.MACrossWindow, cfg.MACrossPreempt
		if w < 0 {
			w = 0
		}
		ctx.FactorMACross(cfg.MACrossShort, cfg.MACrossLong, cfg.MACrossWeight, w, p)
	}
	return ctx
}

// Evaluate 向下兼容旧签名。
func Evaluate(kl *KLineHistory, maShort, maLong, trendN int, useMA, useTrend bool, maWeight, trendWeight float64) *SignalContext {
	return EvaluateWithConfig(kl, &FactorConfig{
		UseMA: useMA, MaShort: maShort, MaLong: maLong, MaWeight: maWeight,
		UseTrend: useTrend, TrendN: trendN, TrendWeight: trendWeight,
	})
}

// BacktestWithConfigDetailed 用 FactorConfig 进行详细回测，返回逐条结果（用于 UI 表格）。
func BacktestWithConfigDetailed(klines []*KLine, cfg *FactorConfig) *BacktestResultSummary {
	minLen := cfg.MinHistory()
	if len(klines) < minLen {
		return &BacktestResultSummary{Results: []BacktestResult{}}
	}

	startIdx := minLen - 2
	if startIdx < 1 {
		startIdx = 1
	}
	window := minLen + 10

	buf := make([]*KLine, window)
	hist := &KLineHistory{}

	var results []BacktestResult
	correct := 0
	signalCount := 0

	for i := startIdx; i < len(klines)-1; i++ {
		FillHistoryWindow(klines, i, window, buf, hist)
		ctx := EvaluateWithConfig(hist, cfg)
		pred := ctx.Prediction()
		factors := ComputeFactorDetailV2(hist, cfg)

		actual := 0
		if klines[i+1].Close > klines[i].Close {
			actual = 1
		} else if klines[i+1].Close < klines[i].Close {
			actual = -1
		}

		ok := false
		if pred != 0 {
			signalCount++
			if pred == actual {
				ok = true
				correct++
			}
		}

		results = append(results, BacktestResult{
			Index:     i,
			OpenTime:  klines[i].OpenTime,
			Actual:    actual,
			Predicted: pred,
			Correct:   ok,
			Factors:   factors,
		})
	}

	total := len(results)
	acc := 0.0
	if total > 0 {
		acc = float64(correct) / float64(total)
	}
	sigAcc := 0.0
	if signalCount > 0 {
		sigAcc = float64(correct) / float64(signalCount)
	}

	return &BacktestResultSummary{
		Results:        results,
		Total:          total,
		Correct:        correct,
		Accuracy:       acc,
		SignalCount:    signalCount,
		SignalAccuracy: sigAcc,
	}
}

// BacktestWithConfig 用 FactorConfig 进行回测，仅返回汇总（不含逐条 Factors 以提高性能）。
func BacktestWithConfig(klines []*KLine, cfg *FactorConfig) *BacktestResultSummary {
	minLen := cfg.MinHistory()
	if len(klines) < minLen {
		return &BacktestResultSummary{Results: []BacktestResult{}}
	}

	startIdx := minLen - 2
	if startIdx < 1 {
		startIdx = 1
	}
	window := minLen + 10

	buf := make([]*KLine, window)
	hist := &KLineHistory{}

	correct := 0
	total := 0
	signalCount := 0
	scoreSum := 0.0
	absScoreSum := 0.0

	for i := startIdx; i < len(klines)-1; i++ {
		FillHistoryWindow(klines, i, window, buf, hist)
		ctx := EvaluateWithConfig(hist, cfg)
		pred := ctx.Prediction()
		ns := ctx.NetScore()

		actual := 0
		if klines[i+1].Close > klines[i].Close {
			actual = 1
		} else if klines[i+1].Close < klines[i].Close {
			actual = -1
		}

		if pred != 0 {
			signalCount++
			scoreSum += ns
			if ns < 0 {
				absScoreSum += -ns
			} else {
				absScoreSum += ns
			}
			if pred == actual {
				correct++
			}
		}
		total++
	}

	acc := 0.0
	if total > 0 {
		acc = float64(correct) / float64(total)
	}
	sigAcc := 0.0
	avgScore := 0.0
	avgAbsScore := 0.0
	if signalCount > 0 {
		sigAcc = float64(correct) / float64(signalCount)
		avgScore = scoreSum / float64(signalCount)
		avgAbsScore = absScoreSum / float64(signalCount)
	}

	return &BacktestResultSummary{
		Total:          total,
		Correct:        correct,
		Accuracy:       acc,
		SignalCount:    signalCount,
		SignalAccuracy: sigAcc,
		AvgScore:       avgScore,
		AvgAbsScore:    avgAbsScore,
	}
}
