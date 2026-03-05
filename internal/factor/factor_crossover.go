package factor

// detectMACrossSignal 检测金叉/死叉信号：1=金叉 -1=死叉 0=无。
// 供 FactorMACross 与 ComputeFactorDetailV2 共用。
func detectMACrossSignal(prices []float64, shortPeriod, longPeriod int, window int, preempt float64) int {
	if shortPeriod >= longPeriod || len(prices) < longPeriod+2+window {
		return 0
	}
	if window < 0 {
		window = 0
	}

	for k := 0; k <= window; k++ {
		currShort := avg(prices[k : k+shortPeriod])
		currLong := avg(prices[k : k+longPeriod])
		prevShort := avg(prices[k+1 : k+1+shortPeriod])
		prevLong := avg(prices[k+1 : k+1+longPeriod])
		if prevShort <= prevLong && currShort > currLong {
			return 1
		}
		if prevShort >= prevLong && currShort < currLong {
			return -1
		}
	}

	if preempt > 0 {
		currShort := avg(prices[:shortPeriod])
		currLong := avg(prices[:longPeriod])
		prevShort := avg(prices[1 : 1+shortPeriod])
		prevLong := avg(prices[1 : 1+longPeriod])
		ratio := 0.0
		if currLong != 0 {
			ratio = (currShort - currLong) / currLong
		}
		if prevShort < prevLong && currShort < currLong && ratio > -preempt && currShort > prevShort {
			return 1
		}
		if prevShort > prevLong && currShort > currLong && ratio < preempt && currShort < prevShort {
			return -1
		}
	}
	return 0
}

// FactorMACross 均线金叉/死叉因子（事件型 + 时间容错 + 预判）
// 回测结论：死叉时下一根更易上涨，故输出正向看涨信号。
// 金叉 → 看跌；死叉 → 看涨；否则 0
//
// MA 有延后性，因此：
//   - 时间容错(window): 当前 K 线「左侧」最近 window 根内若有金叉/死叉，仍记为有效
//   - 预判(preempt): 当短期均线接近长期且趋势朝交叉方向，提前发出信号
//
// 参数：
//   - shortPeriod, longPeriod: 短期、长期周期
//   - weight: 权重
//   - window: 容错根数，0=仅当根，1=当前或前1根有交叉即可，2=当前或前2根，类推
//   - preempt: 预判阈值，0=关闭。>0 时，当 (short-long)/long 在 ±preempt 内且趋势朝向交叉则出信号
func (e *SignalContext) FactorMACross(shortPeriod, longPeriod int, weight float64, window int, preempt float64) *SignalContext {
	if e.KLine == nil {
		return e
	}
	prices := e.KLine.ClosePrices()
	sig := detectMACrossSignal(prices, shortPeriod, longPeriod, window, preempt)
	if sig < 0 {
		e.AddBull(weight)
	} else if sig > 0 {
		e.AddBear(weight)
	}
	return e
}
