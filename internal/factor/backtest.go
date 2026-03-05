package factor

// BacktestResult 单次回测结果
type BacktestResult struct {
	Index     int           `json:"index"`     // K 线索引
	OpenTime  int64         `json:"openTime"`  // 预测所基于的 K 线开盘时间
	Actual    int           `json:"actual"`    // 实际方向 1涨 -1跌 0平
	Predicted int           `json:"predicted"` // 预测方向
	Correct   bool          `json:"correct"`   // 是否预测正确
	Factors   *FactorDetail `json:"factors"`   // 各因子计算结果
}

// BacktestResultSummary 回测汇总
type BacktestResultSummary struct {
	Results        []BacktestResult `json:"results"`
	Total          int              `json:"total"`
	Correct        int              `json:"correct"`
	Accuracy       float64          `json:"accuracy"`       // 正确率 = correct / total
	SignalCount    int              `json:"signalCount"`    // 有效预测次数（pred != 0）
	SignalAccuracy float64          `json:"signalAccuracy"` // 信号正确率 = correct / signalCount
	AvgScore       float64          `json:"avgScore"`       // 平均净分(BullScore-BearScore)，正=偏看涨
	AvgAbsScore    float64          `json:"avgAbsScore"`    // 平均|净分|，信号强度/置信度
}

// Backtest 对 K 线序列进行回测，用 klines[0:i+1] 预测 klines[i+1] 方向，统计正确率。
//
// 参数：
//   - klines: K 线数组，按时间升序，至少 maLong+2 根
//   - maShort, maLong, trendN: 均线周期及趋势统计根数
//   - useMA, useTrend: 是否启用均线因子、趋势因子
//   - maWeight, trendWeight: 各因子权重
//
// 返回包含每条预测结果及正确率的汇总。
func Backtest(klines []*KLine, maShort, maLong, trendN int, useMA, useTrend bool, maWeight, trendWeight float64) *BacktestResultSummary {
	if len(klines) < maLong+2 {
		return &BacktestResultSummary{Results: []BacktestResult{}}
	}

	var results []BacktestResult
	correct := 0

	for i := maLong; i < len(klines)-1; i++ {
		hist := KLinesToHistory(klines[:i+1])
		ctx := Evaluate(hist, maShort, maLong, trendN, useMA, useTrend, maWeight, trendWeight)
		pred := ctx.Prediction()
		factors := ComputeFactorDetail(hist, maShort, maLong, trendN, useMA, useTrend, maWeight, trendWeight)

		actual := 0
		if klines[i+1].Close > klines[i].Close {
			actual = 1
		} else if klines[i+1].Close < klines[i].Close {
			actual = -1
		}

		ok := (pred != 0 && pred == actual)
		if ok {
			correct++
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

	return &BacktestResultSummary{
		Results:  results,
		Total:    total,
		Correct:  correct,
		Accuracy: acc,
	}
}
