package binance

import (
	"context"
	"strconv"
	"sync"

	"quantitative-trading-app/internal/config"
	"quantitative-trading-app/internal/factor"

	binanceclient "github.com/binance/binance-connector-go/clients/derivativestradingusdsfutures"
	"github.com/binance/binance-connector-go/clients/derivativestradingusdsfutures/src/restapi/models"
	"github.com/binance/binance-connector-go/common/v2/common"
)

var (
	binanceClient *binanceclient.BinanceDerivativesTradingUsdsFuturesClient
	clientOnce    sync.Once
)

// InitClient 根据 .env.local 初始化 Binance USDT 永续合约客户端，仅执行一次。
func InitClient() {
	clientOnce.Do(func() {
		_ = config.Load()
		baseURL := config.Get(config.KeyBaseURL, config.BinanceTestnetBaseURL)
		symbol := config.Get(config.KeySymbol, "BTCUSDT")

		cfg := common.NewConfigurationRestAPI()
		cfg.BasePath = baseURL
		// 行情接口无需签名
		binanceClient = binanceclient.NewBinanceDerivativesTradingUsdsFuturesClient(
			binanceclient.WithRestAPI(cfg),
		)
		_ = symbol
	})
}

// Client 返回已初始化的 Binance 合约客户端；若未初始化则先调用 InitClient。
func Client() *binanceclient.BinanceDerivativesTradingUsdsFuturesClient {
	if binanceClient == nil {
		InitClient()
	}
	return binanceClient
}

// parseFloat 从 ItemInner 解析 float64（String 字段）
func parseFloat(v models.KlineCandlestickDataResponseItemInner) float64 {
	if v.String == nil {
		return 0
	}
	f, _ := strconv.ParseFloat(*v.String, 64)
	return f
}

// parseInt64 从 ItemInner 解析 int64（Int64 字段）
func parseInt64(v models.KlineCandlestickDataResponseItemInner) int64 {
	if v.Int64 == nil {
		return 0
	}
	return *v.Int64
}

// FetchKlines 获取合约 K 线数据，按时间升序。
// symbol 交易对，interval 周期（1m/5m/15m/1h 等），limit 根数（上限 1500）。
func FetchKlines(symbol, interval string, limit int64) ([]*factor.KLine, error) {
	if symbol == "" {
		symbol = config.Get(config.KeySymbol, "BTCUSDT")
	}
	if interval == "" {
		interval = "15m"
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 1500 {
		limit = 1500
	}

	intervalParam := models.ContinuousContractKlineCandlestickDataIntervalParameterInterval15m
	switch interval {
	case "1m":
		intervalParam = models.ContinuousContractKlineCandlestickDataIntervalParameterInterval1m
	case "3m":
		intervalParam = models.ContinuousContractKlineCandlestickDataIntervalParameterInterval3m
	case "5m":
		intervalParam = models.ContinuousContractKlineCandlestickDataIntervalParameterInterval5m
	case "15m":
		intervalParam = models.ContinuousContractKlineCandlestickDataIntervalParameterInterval15m
	case "30m":
		intervalParam = models.ContinuousContractKlineCandlestickDataIntervalParameterInterval30m
	case "1h":
		intervalParam = models.ContinuousContractKlineCandlestickDataIntervalParameterInterval1h
	case "4h":
		intervalParam = models.ContinuousContractKlineCandlestickDataIntervalParameterInterval4h
	case "1d":
		intervalParam = models.ContinuousContractKlineCandlestickDataIntervalParameterInterval1d
	}

	resp, err := Client().RestApi.MarketDataAPI.KlineCandlestickData(context.Background()).
		Symbol(symbol).
		Interval(intervalParam).
		Limit(limit).
		Execute()
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Data.Items == nil {
		return nil, nil
	}

	klines := make([]*factor.KLine, 0, len(resp.Data.Items))
	for _, item := range resp.Data.Items {
		if len(item.Items) < 11 {
			continue
		}
		// [0]openTime [1]open [2]high [3]low [4]close [5]volume [6]closeTime
		// [7]quoteAssetVolume [8]numberOfTrades [9]takerBuyVolume [10]takerBuyQuoteVolume
		klines = append(klines, &factor.KLine{
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
		})
	}
	return klines, nil
}
