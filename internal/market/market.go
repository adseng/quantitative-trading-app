package market

import "math"

// KLine 单根 K 线，与 Binance API 返回值保持一致。
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

func (k KLine) IsBullish() bool {
	return k.Close > k.Open
}

func (k KLine) IsBearish() bool {
	return k.Close < k.Open
}

func (k KLine) BodySize() float64 {
	return math.Abs(k.Close - k.Open)
}

func (k KLine) Range() float64 {
	return k.High - k.Low
}

func (k KLine) BodyPercent() float64 {
	if k.Open == 0 {
		return 0
	}
	return k.BodySize() / k.Open
}

func (k KLine) LowerWick() float64 {
	return minFloat(k.Open, k.Close) - k.Low
}

func (k KLine) UpperWick() float64 {
	return k.High - maxFloat(k.Open, k.Close)
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
