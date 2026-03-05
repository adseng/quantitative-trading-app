package factor

// SignalContext 信号评分上下文
// 各因子产出信号 0/1/-1，影响力=权重×信号，BullScore/BearScore 为各因子影响力之和
type SignalContext struct {
	BullScore float64       // 看涨总积分 = 各因子看涨影响力之和
	BearScore float64       // 看跌总积分 = 各因子看跌影响力之和
	KLine     *KLineHistory // 当前 K 线及历史数据
	Weights   map[string]float64
}

// NewSignalContext 创建信号评分上下文。
// 参数 kl 为当前 K 线及历史数据，若为 nil 则创建空上下文。
func NewSignalContext(kl *KLineHistory) *SignalContext {
	return &SignalContext{
		BullScore: 0,
		BearScore: 0,
		KLine:     kl,
		Weights:   make(map[string]float64),
	}
}

// AddBull 增加看涨积分。
// score 为增量，支持链式调用。
func (e *SignalContext) AddBull(score float64) *SignalContext {
	e.BullScore += score
	return e
}

// AddBear 增加看跌积分。
// score 为增量，支持链式调用。
func (e *SignalContext) AddBear(score float64) *SignalContext {
	e.BearScore += score
	return e
}

// NetScore 返回 BullScore - BearScore，正=看涨强度，负=看跌强度。
func (e *SignalContext) NetScore() float64 {
	return e.BullScore - e.BearScore
}

// Prediction 根据 BullScore 与 BearScore 判断预测方向。
// 返回 1=看涨，-1=看跌，0=持平（两者相等时）。
func (e *SignalContext) Prediction() int {
	if e.BullScore > e.BearScore {
		return 1
	}
	if e.BullScore < e.BearScore {
		return -1
	}
	return 0
}
