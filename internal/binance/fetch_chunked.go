package binance

import (
	"context"
	"fmt"
	"sort"
	"time"

	"quantitative-trading-app/internal/factor"

	"github.com/binance/binance-connector-go/clients/derivativestradingusdsfutures/src/restapi/models"
)

// FetchKlinesChunked 分页获取大量 K 线数据。
// 每次请求 perReq 根（上限 1500），共请求 chunks 次，使用 endTime 向前翻页。
// delayMs 为请求间隔（毫秒），防止频率过高。
// progressFn 可选，用于回调进度（当前第几轮, 总轮数, 已获取根数）。
func FetchKlinesChunked(
	symbol, interval string,
	perReq int64, chunks int, delayMs int,
	progressFn func(round, totalRounds, fetched int),
	cancelCh <-chan struct{},
) ([]*factor.KLine, error) {
	if perReq > 1500 {
		perReq = 1500
	}
	if perReq <= 0 {
		perReq = 1000
	}

	intervalParam := models.ContinuousContractKlineCandlestickDataIntervalParameterInterval15m
	switch interval {
	case "1m":
		intervalParam = models.ContinuousContractKlineCandlestickDataIntervalParameterInterval1m
	case "5m":
		intervalParam = models.ContinuousContractKlineCandlestickDataIntervalParameterInterval5m
	case "15m":
		intervalParam = models.ContinuousContractKlineCandlestickDataIntervalParameterInterval15m
	case "1h":
		intervalParam = models.ContinuousContractKlineCandlestickDataIntervalParameterInterval1h
	}

	allMap := make(map[int64]*factor.KLine)
	var endTime *int64

	for round := 0; round < chunks; round++ {
		select {
		case <-cancelCh:
			return mapToSlice(allMap), nil
		default:
		}

		req := Client().RestApi.MarketDataAPI.KlineCandlestickData(context.Background()).
			Symbol(symbol).
			Interval(intervalParam).
			Limit(perReq)

		if endTime != nil {
			req = req.EndTime(*endTime)
		}

		resp, err := req.Execute()
		if err != nil {
			return nil, fmt.Errorf("第 %d 次请求失败: %w", round+1, err)
		}
		if resp == nil || resp.Data.Items == nil || len(resp.Data.Items) == 0 {
			break
		}

		var minOpenTime int64
		for _, item := range resp.Data.Items {
			if len(item.Items) < 11 {
				continue
			}
			kl := &factor.KLine{
				OpenTime:            parseInt64(item.Items[0]),
				Open:                parseFloat(item.Items[1]),
				High:                parseFloat(item.Items[2]),
				Low:                 parseFloat(item.Items[3]),
				Close:               parseFloat(item.Items[4]),
				Volume:              parseFloat(item.Items[5]),
				CloseTime:           parseInt64(item.Items[6]),
				QuoteAssetVolume:    parseFloat(item.Items[7]),
				NumberOfTrades:      parseInt64(item.Items[8]),
				TakerBuyVolume:      parseFloat(item.Items[9]),
				TakerBuyQuoteVolume: parseFloat(item.Items[10]),
			}
			allMap[kl.OpenTime] = kl
			if minOpenTime == 0 || kl.OpenTime < minOpenTime {
				minOpenTime = kl.OpenTime
			}
		}

		if progressFn != nil {
			progressFn(round+1, chunks, len(allMap))
		}

		// Next page: endTime = min openTime - 1
		et := minOpenTime - 1
		endTime = &et

		if delayMs > 0 && round < chunks-1 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}

	return mapToSlice(allMap), nil
}

func mapToSlice(m map[int64]*factor.KLine) []*factor.KLine {
	result := make([]*factor.KLine, 0, len(m))
	for _, kl := range m {
		result = append(result, kl)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].OpenTime < result[j].OpenTime
	})
	return result
}
