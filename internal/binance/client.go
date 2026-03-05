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
	"quantitative-trading-app/internal/factor"

	binanceclient "github.com/binance/binance-connector-go/clients/derivativestradingusdsfutures"
	"github.com/binance/binance-connector-go/clients/derivativestradingusdsfutures/src/restapi/models"
	"github.com/binance/binance-connector-go/common/v2/common"
)

var (
	binanceClient *binanceclient.BinanceDerivativesTradingUsdsFuturesClient
	clientOnce    sync.Once
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

// FetchKlinesOpts 分页拉取时的可选参数；nil 时单次请求（limit 上限 1500）。
type FetchKlinesOpts struct {
	PerReq     int64  // 每轮请求根数，默认 1500
	Chunks     int    // 请求轮数
	DelayMs    int    // 轮间延迟（毫秒）
	ProgressFn func(round, totalRounds, fetched int)
	CancelCh   <-chan struct{}
}

// FetchKlines 获取合约 K 线数据，按时间升序返回。
// symbol 交易对，interval 周期（1m/5m/15m/1h 等），limit 根数。
// 当 opts 为 nil 且 limit<=1500 时单次请求；当 opts 非 nil 或 limit>1500 时分页请求。
func FetchKlines(symbol, interval string, limit int64, opts *FetchKlinesOpts) ([]*factor.KLine, error) {
	if symbol == "" {
		symbol = config.Get(config.KeySymbol, "BTCUSDT")
	}
	if interval == "" {
		interval = "15m"
	}
	intervalParam := parseIntervalParam(interval)

	if opts == nil {
		if limit <= 0 {
			limit = 100
		}
		if limit > 1500 {
			limit = 1500
		}
		return fetchKlinesSingle(symbol, intervalParam, limit)
	}

	perReq := opts.PerReq
	if perReq <= 0 {
		perReq = 1000
	}
	if perReq > 1500 {
		perReq = 1500
	}
	chunks := opts.Chunks
	if chunks <= 0 {
		chunks = 1
	}
	return fetchKlinesChunked(symbol, intervalParam, perReq, chunks, opts.DelayMs, opts.ProgressFn, opts.CancelCh)
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

func fetchKlinesSingle(symbol string, intervalParam models.ContinuousContractKlineCandlestickDataIntervalParameter, limit int64) ([]*factor.KLine, error) {
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
	return parseKlineItems(resp.Data.Items), nil
}

func fetchKlinesChunked(
	symbol string,
	intervalParam models.ContinuousContractKlineCandlestickDataIntervalParameter,
	perReq int64, chunks int, delayMs int,
	progressFn func(round, totalRounds, fetched int),
	cancelCh <-chan struct{},
) ([]*factor.KLine, error) {
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

func parseKlineItems(items []models.KlineCandlestickDataResponseItem) []*factor.KLine {
	out := make([]*factor.KLine, 0, len(items))
	for _, item := range items {
		if len(item.Items) < 11 {
			continue
		}
		out = append(out, &factor.KLine{
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
