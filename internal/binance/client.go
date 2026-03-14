package binance

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"quantitative-trading-app/internal/config"
	"quantitative-trading-app/internal/market"

	binanceclient "github.com/binance/binance-connector-go/clients/derivativestradingusdsfutures"
	"github.com/binance/binance-connector-go/clients/derivativestradingusdsfutures/src/restapi/models"
	"github.com/binance/binance-connector-go/common/v2/common"
)

var (
	binanceClient *binanceclient.BinanceDerivativesTradingUsdsFuturesClient
	clientOnce    sync.Once
)

const (
	maxKlinesPerRequest = int64(1000)
	fixedChunkDelayMs   = 300
)

func parseProxyConfig(proxyURL string) (common.ProxyConfig, bool) {
	proxyURL = strings.TrimSpace(proxyURL)
	if proxyURL == "" {
		return common.ProxyConfig{}, false
	}
	u, err := url.Parse(proxyURL)
	if err != nil {
		return common.ProxyConfig{}, false
	}
	host := u.Hostname()
	port := 0
	if p := u.Port(); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}
	if host == "" {
		return common.ProxyConfig{}, false
	}
	protocol := u.Scheme
	if protocol == "" {
		protocol = "http"
	}
	pc := common.ProxyConfig{
		Host:     host,
		Port:     port,
		Protocol: protocol,
	}
	if u.User != nil {
		pc.Auth.Username = u.User.Username()
		pc.Auth.Password, _ = u.User.Password()
	}
	return pc, true
}

// InitClient 根据 .env 初始化 Binance USDT 永续合约客户端，仅执行一次。
func InitClient() {
	clientOnce.Do(func() {
		_ = config.Load()
		baseURL := config.Get(config.KeyBaseURL, config.BinanceMainnetBaseURL)
		symbol := config.Get(config.KeySymbol, "BTCUSDT")

		timeoutSec := 30
		if s := config.Get(config.KeyRequestTimeout, "30"); s != "" {
			if n, err := strconv.Atoi(s); err == nil && n > 0 {
				timeoutSec = n
			}
		}
		timeout := time.Duration(timeoutSec) * time.Second

		opts := []common.ConfigurationRestAPIOption{
			common.WithBasePath(baseURL),
			common.WithTimeout(timeout),
		}

		proxyURL := config.Get(config.KeyProxy, "")
		if proxyURL == "" {
			proxyURL = config.Get("HTTPS_PROXY", "")
		}
		if proxyURL == "" {
			proxyURL = config.Get("HTTP_PROXY", "")
		}
		if pc, ok := parseProxyConfig(proxyURL); ok {
			opts = append(opts, common.WithProxy(pc))
		}

		cfg := common.NewConfigurationRestAPI(opts...)
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

// FetchKlinesOpts 分页拉取时的可选控制参数。
type FetchKlinesOpts struct {
	ProgressFn func(round, totalRounds, fetched int)
	CancelCh   <-chan struct{}
}

// FetchKlines 获取合约 K 线数据，按时间升序返回。
// symbol 交易对，interval 周期（1m/5m/15m/1h 等），limit 根数。
// 当 limit 大于单次上限时会自动分页拉取，并使用固定轮间延迟。
func FetchKlines(symbol, interval string, limit int64) ([]*market.KLine, error) {
	return fetchKlinesWithOpts(symbol, interval, limit, nil)
}

func fetchKlinesWithOpts(symbol, interval string, limit int64, opts *FetchKlinesOpts) ([]*market.KLine, error) {
	if symbol == "" {
		symbol = config.Get(config.KeySymbol, "BTCUSDT")
	}
	if interval == "" {
		interval = "15m"
	}
	intervalParam := parseIntervalParam(interval)

	if limit <= 0 {
		limit = 100
	}
	perReq := limit
	if perReq > maxKlinesPerRequest {
		perReq = maxKlinesPerRequest
	}
	chunks := int((limit + maxKlinesPerRequest - 1) / maxKlinesPerRequest)
	if chunks < 1 {
		chunks = 1
	}
	var progressFn func(round, totalRounds, fetched int)
	var cancelCh <-chan struct{}
	if opts != nil {
		progressFn = opts.ProgressFn
		cancelCh = opts.CancelCh
	}
	klines, err := fetchKlinesChunked(symbol, intervalParam, perReq, chunks, fixedChunkDelayMs, progressFn, cancelCh)
	if err != nil {
		return nil, err
	}
	if int64(len(klines)) > limit {
		klines = klines[len(klines)-int(limit):]
	}
	return klines, nil
}

func parseIntervalParam(interval string) models.ContinuousContractKlineCandlestickDataIntervalParameter {
	switch interval {
	case "1m":
		return models.ContinuousContractKlineCandlestickDataIntervalParameterInterval1m
	case "3m":
		return models.ContinuousContractKlineCandlestickDataIntervalParameterInterval3m
	case "5m":
		return models.ContinuousContractKlineCandlestickDataIntervalParameterInterval5m
	case "15m":
		return models.ContinuousContractKlineCandlestickDataIntervalParameterInterval15m
	case "30m":
		return models.ContinuousContractKlineCandlestickDataIntervalParameterInterval30m
	case "1h":
		return models.ContinuousContractKlineCandlestickDataIntervalParameterInterval1h
	case "4h":
		return models.ContinuousContractKlineCandlestickDataIntervalParameterInterval4h
	case "1d":
		return models.ContinuousContractKlineCandlestickDataIntervalParameterInterval1d
	default:
		return models.ContinuousContractKlineCandlestickDataIntervalParameterInterval15m
	}
}

func fetchKlinesChunked(
	symbol string,
	intervalParam models.ContinuousContractKlineCandlestickDataIntervalParameter,
	perReq int64, chunks int, delayMs int,
	progressFn func(round, totalRounds, fetched int),
	cancelCh <-chan struct{},
) ([]*market.KLine, error) {
	allMap := make(map[int64]*market.KLine)
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

		parsed := parseKlineItems(resp.Data.Items)
		var minOpenTime int64
		for _, kl := range parsed {
			allMap[kl.OpenTime] = kl
			if minOpenTime == 0 || kl.OpenTime < minOpenTime {
				minOpenTime = kl.OpenTime
			}
		}

		if progressFn != nil {
			progressFn(round+1, chunks, len(allMap))
		}

		et := minOpenTime - 1
		endTime = &et

		if delayMs > 0 && round < chunks-1 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}

	return mapToSlice(allMap), nil
}

func parseKlineItems(items []models.KlineCandlestickDataResponseItem) []*market.KLine {
	out := make([]*market.KLine, 0, len(items))
	for _, item := range items {
		if len(item.Items) < 11 {
			continue
		}
		out = append(out, &market.KLine{
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
	return out
}

func mapToSlice(m map[int64]*market.KLine) []*market.KLine {
	result := make([]*market.KLine, 0, len(m))
	for _, kl := range m {
		result = append(result, kl)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].OpenTime < result[j].OpenTime
	})
	return result
}
