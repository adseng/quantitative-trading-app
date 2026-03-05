package factor

// KLine 单根 K 线，与 Binance API 返回值对应
// 参考: https://developers.binance.com/docs/zh-CN/derivatives/usds-margined-futures/market-data/rest-api/Kline-Candlestick-Data
type KLine struct {
	OpenTime            int64   `json:"openTime"`
	Open                float64 `json:"open"`
	High                float64 `json:"high"`
	Low                 float64 `json:"low"`
	Close               float64 `json:"close"`
	Volume              float64 `json:"volume"`
	CloseTime           int64   `json:"closeTime"`
	QuoteAssetVolume    float64 `json:"quoteAssetVolume"`
	NumberOfTrades      int64   `json:"numberOfTrades"`
	TakerBuyVolume      float64 `json:"takerBuyVolume"`
	TakerBuyQuoteVolume float64 `json:"takerBuyQuoteVolume"`
}

// KLineHistory 当前 K 线及历史数据，用于因子计算
type KLineHistory struct {
	Current  *KLine   // 当前 K 线
	History  []*KLine // 历史 K 线，按时间升序，最近的在前（History[0] 是上一根）
}

// ClosePrices 提取 History 中每条 K 线的收盘价，顺序与 History 一致。
func (h *KLineHistory) ClosePrices() []float64 {
	if h == nil || len(h.History) == 0 {
		return nil
	}
	prices := make([]float64, len(h.History))
	for i, k := range h.History {
		prices[i] = k.Close
	}
	return prices
}

// Volumes 提取 History 中每条 K 线的成交量，顺序与 History 一致。
func (h *KLineHistory) Volumes() []float64 {
	if h == nil || len(h.History) == 0 {
		return nil
	}
	vols := make([]float64, len(h.History))
	for i, k := range h.History {
		vols[i] = k.Volume
	}
	return vols
}

// KLinesToHistory 从 K 线切片构建 KLineHistory。
//
// klines 按时间升序 [最旧,...,最新]；返回的 History 逆序存放 [最新,次新,...]，
// 即 History[0] 为最近一根，便于因子按时间从近到远计算。
func KLinesToHistory(klines []*KLine) *KLineHistory {
	if len(klines) == 0 {
		return nil
	}
	n := len(klines)
	hist := make([]*KLine, n)
	for i := 0; i < n; i++ {
		hist[i] = klines[n-1-i]
	}
	return &KLineHistory{
		Current: hist[0],
		History: hist,
	}
}

// FillHistoryWindow 将 klines[end-windowSize+1 .. end] 以逆序填入 buf，
// 并设置 out 的 Current/History。buf 和 out 应在循环外预分配，循环内零分配。
func FillHistoryWindow(klines []*KLine, end, windowSize int, buf []*KLine, out *KLineHistory) {
	start := end - windowSize + 1
	if start < 0 {
		start = 0
	}
	n := end - start + 1
	for i := 0; i < n; i++ {
		buf[i] = klines[end-i]
	}
	out.Current = buf[0]
	out.History = buf[:n]
}
