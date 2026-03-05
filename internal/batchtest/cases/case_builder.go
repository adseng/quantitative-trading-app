package cases

// caseBuilder 用于收集策略用例，供各 section 文件调用。
type caseBuilder struct {
	cases []TestCase
	id    int
}

func (b *caseBuilder) add(name string, c TestCase) {
	b.id++
	c.ID = b.id
	c.Name = name
	b.cases = append(b.cases, c)
}

func (b *caseBuilder) boll(period int, mult float64) TestCase {
	return TestCase{UseBoll: true, BollPeriod: period, BollMultiplier: mult, BollWeight: 1}
}

func (b *caseBuilder) brk(period int) TestCase {
	return TestCase{UseBreakout: true, BreakoutPeriod: period, BreakoutWeight: -1}
}

func (b *caseBuilder) rsi(period int, ob, os float64) TestCase {
	return TestCase{UseRSI: true, RSIPeriod: period, RSIOverbought: ob, RSIOversold: os, RSIWeight: 1}
}

func (b *caseBuilder) ma(short, long int, weight float64) TestCase {
	return TestCase{UseMA: true, MaShort: short, MaLong: long, MaWeight: weight}
}

func (b *caseBuilder) trend(n int, weight float64) TestCase {
	return TestCase{UseTrend: true, TrendN: n, TrendWeight: weight}
}

func (b *caseBuilder) macd(fast, slow, signal int, weight float64) TestCase {
	return TestCase{UseMACD: true, MACDFast: fast, MACDSlow: slow, MACDSignal: signal, MACDWeight: weight}
}

func (b *caseBuilder) priceVsMA(period int, weight float64) TestCase {
	return TestCase{UsePriceVsMA: true, PriceVsMAPeriod: period, PriceVsMAWeight: weight}
}

func (b *caseBuilder) atr(period int, weight float64) TestCase {
	return TestCase{UseATR: true, ATRPeriod: period, ATRWeight: weight}
}

func (b *caseBuilder) volume(period int, weight float64) TestCase {
	return TestCase{UseVolume: true, VolumePeriod: period, VolumeWeight: weight}
}

func (b *caseBuilder) session(weight float64) TestCase {
	return TestCase{UseSession: true, SessionWeight: weight}
}

func (b *caseBuilder) macross(short, long int, weight float64, window int, preempt float64) TestCase {
	return TestCase{
		UseMACross: true, MACrossShort: short, MACrossLong: long,
		MACrossWeight: weight, MACrossWindow: window, MACrossPreempt: preempt,
	}
}

func (b *caseBuilder) result() []TestCase {
	const maxCases = 200 // 单次不超过200组，避免卡住
	if len(b.cases) > maxCases {
		return b.cases[:maxCases]
	}
	return b.cases
}
