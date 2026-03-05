package cases

// TestCase defines a single backtest parameter combination.
type TestCase struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	UseMA    bool    `json:"useMA"`
	MaShort  int     `json:"maShort"`
	MaLong   int     `json:"maLong"`
	MaWeight float64 `json:"maWeight"`

	UseTrend    bool    `json:"useTrend"`
	TrendN      int     `json:"trendN"`
	TrendWeight float64 `json:"trendWeight"`

	UseRSI        bool    `json:"useRSI"`
	RSIPeriod     int     `json:"rsiPeriod"`
	RSIOverbought float64 `json:"rsiOverbought"`
	RSIOversold   float64 `json:"rsiOversold"`
	RSIWeight     float64 `json:"rsiWeight"`

	UseMACD    bool    `json:"useMACD"`
	MACDFast   int     `json:"macdFast"`
	MACDSlow   int     `json:"macdSlow"`
	MACDSignal int     `json:"macdSignal"`
	MACDWeight float64 `json:"macdWeight"`

	UseBoll        bool    `json:"useBoll"`
	BollPeriod     int     `json:"bollPeriod"`
	BollMultiplier float64 `json:"bollMultiplier"`
	BollWeight     float64 `json:"bollWeight"`

	UseBreakout    bool    `json:"useBreakout"`
	BreakoutPeriod int     `json:"breakoutPeriod"`
	BreakoutWeight float64 `json:"breakoutWeight"`

	UsePriceVsMA    bool    `json:"usePriceVsMA"`
	PriceVsMAPeriod int     `json:"priceVsMAPeriod"`
	PriceVsMAWeight float64 `json:"priceVsMAWeight"`

	UseATR    bool    `json:"useATR"`
	ATRPeriod int     `json:"atrPeriod"`
	ATRWeight float64 `json:"atrWeight"`

	UseVolume    bool    `json:"useVolume"`
	VolumePeriod int     `json:"volumePeriod"`
	VolumeWeight float64 `json:"volumeWeight"`

	UseSession    bool    `json:"useSession"`
	SessionWeight float64 `json:"sessionWeight"`

	UseMACross     bool    `json:"useMACross"`
	MACrossShort   int     `json:"macrossShort"`
	MACrossLong    int     `json:"macrossLong"`
	MACrossWeight  float64 `json:"macrossWeight"`
	MACrossWindow  int     `json:"macrossWindow"`
	MACrossPreempt float64 `json:"macrossPreempt"`
}

// TestResult holds one test case result.
type TestResult struct {
	TestCase       TestCase `json:"testCase"`
	Accuracy       float64  `json:"accuracy"`
	Correct        int      `json:"correct"`
	Total          int      `json:"total"`
	SignalCount    int      `json:"signalCount"`
	SignalAccuracy float64  `json:"signalAccuracy"`
	AvgScore       float64  `json:"avgScore"`
	AvgAbsScore    float64  `json:"avgAbsScore"`
}
