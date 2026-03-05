package factor

import (
	"math"
	"time"
)

// FactorDetail 单次回测时各因子的计算结果
type FactorDetail struct {
	MaScore        float64 `json:"maScore"`
	TrendContrib   float64 `json:"trendContrib"`
	RsiContrib     float64 `json:"rsiContrib"`
	MacdContrib    float64 `json:"macdContrib"`
	BollContrib    float64 `json:"bollContrib"`
	BreakoutContrib float64 `json:"breakoutContrib"`
	PriceVsMAContrib float64 `json:"priceVsMAContrib"`
	AtrContrib     float64 `json:"atrContrib"`
	VolumeContrib  float64 `json:"volumeContrib"`
	SessionContrib   float64 `json:"sessionContrib"`
	MACrossContrib   float64 `json:"macrossContrib"`
	BullScore        float64 `json:"bullScore"`
	BearScore      float64 `json:"bearScore"`
}

// ComputeFactorDetailV2 使用 FactorConfig 计算各因子影响力。
func ComputeFactorDetailV2(kl *KLineHistory, cfg *FactorConfig) *FactorDetail {
	fd := &FactorDetail{}
	if kl == nil || kl.History == nil {
		return fd
	}

	// 均线因子
	if cfg.UseMA {
		prices := kl.ClosePrices()
		if len(prices) >= cfg.MaLong {
			shortMA := avg(prices[:cfg.MaShort])
			longMA := avg(prices[:cfg.MaLong])
			if shortMA > longMA {
				fd.MaScore = cfg.MaWeight
			} else if shortMA < longMA {
				fd.MaScore = -cfg.MaWeight
			}
		}
	}

	// 趋势因子
	if cfg.UseTrend {
		need := cfg.TrendN + 1
		if len(kl.History) >= need {
			upCount, downCount := 0, 0
			for i := 0; i < cfg.TrendN; i++ {
				if kl.History[i].Close > kl.History[i+1].Close {
					upCount++
				} else if kl.History[i].Close < kl.History[i+1].Close {
					downCount++
				}
			}
			diff := upCount - downCount
			if diff > 0 {
				fd.TrendContrib = cfg.TrendWeight
			} else if diff < 0 {
				fd.TrendContrib = -cfg.TrendWeight
			}
		}
	}

	// RSI 因子
	if cfg.UseRSI {
		prices := kl.ClosePrices()
		rsi := ComputeRSI(prices, cfg.RSIPeriod)
		if rsi < cfg.RSIOversold {
			fd.RsiContrib = cfg.RSIWeight
		} else if rsi > cfg.RSIOverbought {
			fd.RsiContrib = -cfg.RSIWeight
		}
	}

	// MACD 因子
	if cfg.UseMACD {
		prices := kl.ClosePrices()
		need := cfg.MACDSlow + cfg.MACDSignal
		if len(prices) >= need {
			reversed := make([]float64, len(prices))
			for i, p := range prices {
				reversed[len(prices)-1-i] = p
			}
			emaFast := calcEMA(reversed, cfg.MACDFast)
			emaSlow := calcEMA(reversed, cfg.MACDSlow)
			if len(emaFast) > 0 && len(emaSlow) > 0 {
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
				if len(macdLine) >= cfg.MACDSignal {
					signalLine := calcEMA(macdLine, cfg.MACDSignal)
					if len(signalLine) > 0 {
						lastMACD := macdLine[len(macdLine)-1]
						lastSignal := signalLine[len(signalLine)-1]
						if lastMACD > lastSignal {
							fd.MacdContrib = cfg.MACDWeight
						} else if lastMACD < lastSignal {
							fd.MacdContrib = -cfg.MACDWeight
						}
					}
				}
			}
		}
	}

	// 布林带因子
	if cfg.UseBoll {
		prices := kl.ClosePrices()
		if len(prices) >= cfg.BollPeriod {
			slice := prices[:cfg.BollPeriod]
			middle := avg(slice)
			var variance float64
			for _, p := range slice {
				d := p - middle
				variance += d * d
			}
			stddev := math.Sqrt(variance / float64(cfg.BollPeriod))
			upper := middle + cfg.BollMultiplier*stddev
			lower := middle - cfg.BollMultiplier*stddev
			if prices[0] < lower {
				fd.BollContrib = cfg.BollWeight
			} else if prices[0] > upper {
				fd.BollContrib = -cfg.BollWeight
			}
		}
	}

	// 突破因子
	if cfg.UseBreakout && len(kl.History) >= cfg.BreakoutPeriod+1 {
		currentClose := kl.History[0].Close
		var highest, lowest float64
		for i := 1; i <= cfg.BreakoutPeriod; i++ {
			h := kl.History[i].High
			l := kl.History[i].Low
			if i == 1 || h > highest {
				highest = h
			}
			if i == 1 || l < lowest {
				lowest = l
			}
		}
		if currentClose > highest {
			fd.BreakoutContrib = cfg.BreakoutWeight
		} else if currentClose < lowest {
			fd.BreakoutContrib = -cfg.BreakoutWeight
		}
	}

	// 价格vs均线因子
	if cfg.UsePriceVsMA {
		prices := kl.ClosePrices()
		if len(prices) >= cfg.PriceVsMAPeriod {
			sma := avg(prices[:cfg.PriceVsMAPeriod])
			if prices[0] > sma {
				fd.PriceVsMAContrib = cfg.PriceVsMAWeight
			} else if prices[0] < sma {
				fd.PriceVsMAContrib = -cfg.PriceVsMAWeight
			}
		}
	}

	// ATR 波动率因子
	if cfg.UseATR && len(kl.History) >= cfg.ATRPeriod+1 {
		trs := make([]float64, cfg.ATRPeriod)
		for i := 0; i < cfg.ATRPeriod; i++ {
			cur := kl.History[i]
			prev := kl.History[i+1]
			hl := cur.High - cur.Low
			hc := math.Abs(cur.High - prev.Close)
			lc := math.Abs(cur.Low - prev.Close)
			tr := hl
			if hc > tr {
				tr = hc
			}
			if lc > tr {
				tr = lc
			}
			trs[i] = tr
		}
		currentATR := trs[0]
		avgATR := 0.0
		for _, v := range trs {
			avgATR += v
		}
		avgATR /= float64(cfg.ATRPeriod)
		if currentATR > avgATR {
			priceChange := kl.History[0].Close - kl.History[1].Close
			if priceChange > 0 {
				fd.AtrContrib = cfg.ATRWeight
			} else if priceChange < 0 {
				fd.AtrContrib = -cfg.ATRWeight
			}
		}
	}

	// 量价配合因子
	if cfg.UseVolume {
		vols := kl.Volumes()
		if len(vols) >= cfg.VolumePeriod+1 {
			currentVol := vols[0]
			var avgVol float64
			for i := 1; i <= cfg.VolumePeriod; i++ {
				avgVol += vols[i]
			}
			avgVol /= float64(cfg.VolumePeriod)
			if currentVol > avgVol && len(kl.History) >= 2 {
				priceChange := kl.History[0].Close - kl.History[1].Close
				if priceChange > 0 {
					fd.VolumeContrib = cfg.VolumeWeight
				} else if priceChange < 0 {
					fd.VolumeContrib = -cfg.VolumeWeight
				}
			}
		}
	}

	// 时段因子
	if cfg.UseSession && len(kl.History) >= 2 {
		t := time.UnixMilli(kl.History[0].OpenTime).UTC()
		hour := t.Hour()
		if hour >= 8 { // 欧盘/美盘才出信号
			priceChange := kl.History[0].Close - kl.History[1].Close
			if priceChange > 0 {
				fd.SessionContrib = cfg.SessionWeight
			} else if priceChange < 0 {
				fd.SessionContrib = -cfg.SessionWeight
			}
		}
	}

	// 金叉/死叉因子（事件型+容错+预判）
	if cfg.UseMACross {
		prices := kl.ClosePrices()
		w, p := cfg.MACrossWindow, cfg.MACrossPreempt
		if w < 0 {
			w = 0
		}
		sig := detectMACrossSignal(prices, cfg.MACrossShort, cfg.MACrossLong, w, p)
		if sig > 0 {
			fd.MACrossContrib = cfg.MACrossWeight
		} else if sig < 0 {
			fd.MACrossContrib = -cfg.MACrossWeight
		}
	}

	// 汇总
	contribs := []float64{
		fd.MaScore, fd.TrendContrib, fd.RsiContrib, fd.MacdContrib, fd.BollContrib,
		fd.BreakoutContrib, fd.PriceVsMAContrib, fd.AtrContrib, fd.VolumeContrib, fd.SessionContrib, fd.MACrossContrib,
	}
	for _, c := range contribs {
		if c > 0 {
			fd.BullScore += c
		} else if c < 0 {
			fd.BearScore += -c
		}
	}
	return fd
}

// ComputeFactorDetail 向下兼容旧签名。
func ComputeFactorDetail(kl *KLineHistory, maShort, maLong, trendN int, useMA, useTrend bool, maWeight, trendWeight float64) *FactorDetail {
	return ComputeFactorDetailV2(kl, &FactorConfig{
		UseMA: useMA, MaShort: maShort, MaLong: maLong, MaWeight: maWeight,
		UseTrend: useTrend, TrendN: trendN, TrendWeight: trendWeight,
	})
}
