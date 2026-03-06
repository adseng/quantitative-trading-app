package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"quantitative-trading-app/internal/backtestlog"
	"quantitative-trading-app/internal/batchtest"
	"quantitative-trading-app/internal/batchtest/cases"
	"quantitative-trading-app/internal/binance"
	"quantitative-trading-app/internal/config"
	"quantitative-trading-app/internal/coze"
	"quantitative-trading-app/internal/factor"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx          context.Context
	batchRunner  *batchtest.BatchRunner
	cancelFetch  chan struct{}
	cachedKlines []*factor.KLine // 缓存已获取的 K 线，resume 时复用

	// Coze 预测页：K 线流（REST 轮询，每 100ms 拉最新 1 根）
	klineStream        *binance.KlineStream
	klinePollStop      chan struct{}
	klineBuf           []*factor.KLine
	klineBufMu         sync.Mutex
	klineSymbol        string
	klineInterval      string
	klineLimit         int64
	cozeKlineCount int // 发给豆包的 K 线根数，默认 50
}

func NewApp() *App {
	return &App{
		batchRunner:    batchtest.NewBatchRunner(),
		cozeKlineCount: 50,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	_ = config.Load()
	binance.InitClient()
}

// FetchKlines 从币安合约获取 K 线数据，按时间升序返回。
func (a *App) FetchKlines(symbol, interval string, limit int64) ([]factor.KLine, error) {
	kl, err := binance.FetchKlines(symbol, interval, limit, nil)
	if err != nil || kl == nil {
		return nil, err
	}
	out := make([]factor.KLine, len(kl))
	for i, k := range kl {
		out[i] = *k
	}
	return out, nil
}

// BacktestEmotion 使用基础因子配置进行回测。
func (a *App) BacktestEmotion(klines []factor.KLine, useMA, useTrend bool, maShort, maLong, trendN int, maWeight, trendWeight float64) (factor.BacktestResultSummary, error) {
	minLen := maLong + 2
	if len(klines) < minLen {
		return factor.BacktestResultSummary{Results: []factor.BacktestResult{}}, nil
	}
	if !useMA && !useTrend {
		return factor.BacktestResultSummary{Results: []factor.BacktestResult{}}, nil
	}
	ptr := make([]*factor.KLine, len(klines))
	for i := range klines {
		ptr[i] = &klines[i]
	}
	return *factor.Backtest(ptr, maShort, maLong, trendN, useMA, useTrend, maWeight, trendWeight), nil
}

// BacktestEmotionV2 使用完整因子配置（10 因子）进行回测。
func (a *App) BacktestEmotionV2(klines []factor.KLine,
	useMA bool, maShort, maLong int, maWeight float64,
	useTrend bool, trendN int, trendWeight float64,
	useRSI bool, rsiPeriod int, rsiOverbought, rsiOversold, rsiWeight float64,
	useMACD bool, macdFast, macdSlow, macdSignal int, macdWeight float64,
	useBoll bool, bollPeriod int, bollMultiplier, bollWeight float64,
	useBreakout bool, breakoutPeriod int, breakoutWeight float64,
	usePriceVsMA bool, priceVsMAPeriod int, priceVsMAWeight float64,
	useATR bool, atrPeriod int, atrWeight float64,
	useVolume bool, volumePeriod int, volumeWeight float64,
	useSession bool, sessionWeight float64,
	useMACross bool, macrossShort, macrossLong int, macrossWeight float64, macrossWindow int, macrossPreempt float64,
) (factor.BacktestResultSummary, error) {
	cfg := &factor.FactorConfig{
		UseMA: useMA, MaShort: maShort, MaLong: maLong, MaWeight: maWeight,
		UseTrend: useTrend, TrendN: trendN, TrendWeight: trendWeight,
		UseRSI: useRSI, RSIPeriod: rsiPeriod, RSIOverbought: rsiOverbought, RSIOversold: rsiOversold, RSIWeight: rsiWeight,
		UseMACD: useMACD, MACDFast: macdFast, MACDSlow: macdSlow, MACDSignal: macdSignal, MACDWeight: macdWeight,
		UseBoll: useBoll, BollPeriod: bollPeriod, BollMultiplier: bollMultiplier, BollWeight: bollWeight,
		UseBreakout: useBreakout, BreakoutPeriod: breakoutPeriod, BreakoutWeight: breakoutWeight,
		UsePriceVsMA: usePriceVsMA, PriceVsMAPeriod: priceVsMAPeriod, PriceVsMAWeight: priceVsMAWeight,
		UseATR: useATR, ATRPeriod: atrPeriod, ATRWeight: atrWeight,
		UseVolume: useVolume, VolumePeriod: volumePeriod, VolumeWeight: volumeWeight,
		UseSession: useSession, SessionWeight: sessionWeight,
		UseMACross: useMACross, MACrossShort: macrossShort, MACrossLong: macrossLong, MACrossWeight: macrossWeight,
			MACrossWindow: macrossWindow, MACrossPreempt: macrossPreempt,
	}
	minLen := cfg.MinHistory()
	if len(klines) < minLen {
		return factor.BacktestResultSummary{Results: []factor.BacktestResult{}}, nil
	}
	anyEnabled := useMA || useTrend || useRSI || useMACD || useBoll || useBreakout || usePriceVsMA || useATR || useVolume || useSession || useMACross
	if !anyEnabled {
		return factor.BacktestResultSummary{Results: []factor.BacktestResult{}}, nil
	}
	ptr := make([]*factor.KLine, len(klines))
	for i := range klines {
		ptr[i] = &klines[i]
	}
	return *factor.BacktestWithConfigDetailed(ptr, cfg), nil
}

// LogBacktestResult 将回测参数和正确率追加到 Excel 文件。
func (a *App) LogBacktestResult(
	symbol, interval string, limit int64,
	useMA, useTrend bool, maShort, maLong, trendN int, maWeight, trendWeight float64,
	accuracy float64, correct, total int,
) error {
	return backtestlog.LogBacktestResult(
		symbol, interval, limit,
		useMA, maShort, maLong, maWeight,
		useTrend, trendN, trendWeight,
		accuracy, correct, total,
	)
}

// StartBatchTest 开始或继续批量回测。
// 首次运行时分页获取 K 线（约 60s），之后 resume 时直接复用缓存数据。
// 通过 Wails 事件 "batch:progress" 推送进度。
func (a *App) StartBatchTest() error {
	if a.batchRunner.IsRunning() {
		return nil
	}
	if a.batchRunner.NextIndex() >= a.batchRunner.TotalCases() {
		a.batchRunner.ResetForNewBatch()
	}

	a.cancelFetch = make(chan struct{})
	emit := func(evt batchtest.ProgressEvent) {
		wailsRuntime.EventsEmit(a.ctx, "batch:progress", evt)
	}

	go func() {
		// 如果已有缓存 K 线，直接复用，不重新拉取
		if len(a.cachedKlines) > 0 {
			emit(batchtest.ProgressEvent{
				Phase:   "fetching",
				Message: fmt.Sprintf("使用缓存 K 线数据（%d 根），继续回测...", len(a.cachedKlines)),
			})
			a.batchRunner.Run(a.cachedKlines, emit)
			return
		}

		emit(batchtest.ProgressEvent{Phase: "fetching", Message: "首次运行，开始获取 K 线数据（100轮 × 1000根）..."})

		klines, err := binance.FetchKlines("BTCUSDT", "15m", 0, &binance.FetchKlinesOpts{
			PerReq:  1000,
			Chunks: 100,
			DelayMs: 600,
			ProgressFn: func(round, totalRounds, fetched int) {
				emit(batchtest.ProgressEvent{
					Phase:   "fetching",
					Message: fmt.Sprintf("获取 K 线: %d/%d 轮, 已获取 %d 根", round, totalRounds, fetched),
				})
			},
			CancelCh: a.cancelFetch,
		})
		if err != nil {
			emit(batchtest.ProgressEvent{Phase: "error", Message: "获取 K 线失败: " + err.Error()})
			return
		}

		emit(batchtest.ProgressEvent{
			Phase:   "fetching",
			Message: fmt.Sprintf("K 线获取完成，共 %d 根，开始回测...", len(klines)),
		})

		a.cachedKlines = klines
		a.batchRunner.Run(klines, emit)
	}()

	return nil
}

// StopBatchTest 停止批量回测。
func (a *App) StopBatchTest() {
	if a.cancelFetch != nil {
		select {
		case <-a.cancelFetch:
		default:
			close(a.cancelFetch)
		}
	}
	a.batchRunner.Stop()
}

// GetBatchTestProgress 获取当前进度。
func (a *App) GetBatchTestProgress() map[string]interface{} {
	return map[string]interface{}{
		"running":    a.batchRunner.IsRunning(),
		"nextIndex":  a.batchRunner.NextIndex(),
		"totalCases": a.batchRunner.TotalCases(),
	}
}

// GetBatchTestResults 获取最近的回测结果（用于页面初始化展示，从 Excel 加载历史数据）。
func (a *App) GetBatchTestResults(maxN int) []cases.TestResult {
	if maxN <= 0 {
		maxN = 200
	}
	return a.batchRunner.GetLastResults(maxN)
}

// intervalMs 返回周期毫秒数，用于判断 K 线是否连续
func intervalMs(interval string) int64 {
	switch interval {
	case "1m":
		return 60 * 1000
	case "3m":
		return 3 * 60 * 1000
	case "5m":
		return 5 * 60 * 1000
	case "15m":
		return 15 * 60 * 1000
	case "30m":
		return 30 * 60 * 1000
	case "1h":
		return 60 * 60 * 1000
	case "4h":
		return 4 * 60 * 60 * 1000
	case "1d":
		return 24 * 60 * 60 * 1000
	default:
		return 15 * 60 * 1000
	}
}

// StartKlineStream 获取 limit 根 K 线，启动 REST 轮询（每 100ms 拉最新 1 根）。
// 若最新 1 根与当前不连续则重新拉取 limit 根。事件: kline:snapshot, kline:update, kline:status
func (a *App) StartKlineStream(symbol, interval string, limit int64) error {
	if symbol == "" {
		symbol = "BTCUSDT"
	}
	if interval == "" {
		interval = "15m"
	}
	if limit <= 0 {
		limit = 1000
	}
	if limit > 1500 {
		limit = 1500
	}
	a.StopKlineStream()

	klines, err := binance.FetchKlines(symbol, interval, limit, nil)
	if err != nil {
		return err
	}
	if len(klines) == 0 {
		return fmt.Errorf("未获取到 K 线")
	}

	a.klineBufMu.Lock()
	a.klineBuf = klines
	a.klineSymbol = symbol
	a.klineInterval = interval
	a.klineLimit = limit
	a.klineBufMu.Unlock()

	snapshot := make([]factor.KLine, len(klines))
	for i, k := range klines {
		snapshot[i] = *k
	}
	wailsRuntime.EventsEmit(a.ctx, "kline:snapshot", map[string]interface{}{
		"klines":   snapshot,
		"symbol":   symbol,
		"interval": interval,
	})
	wailsRuntime.EventsEmit(a.ctx, "kline:status", map[string]interface{}{"status": "polling"})

	a.klinePollStop = make(chan struct{})
	go a.runKlinePoll()
	return nil
}

func (a *App) runKlinePoll() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-a.klinePollStop:
			return
		case <-ticker.C:
			a.klineBufMu.Lock()
			symbol := a.klineSymbol
			interval := a.klineInterval
			buf := a.klineBuf
			a.klineBufMu.Unlock()
			if symbol == "" || len(buf) == 0 {
				continue
			}
			latest, err := binance.FetchKlines(symbol, interval, 1, nil)
			if err != nil || len(latest) == 0 {
				continue
			}
			kl := latest[0]
			lastOpen := buf[len(buf)-1].OpenTime
			ms := intervalMs(interval)
			a.klineBufMu.Lock()
			buf = a.klineBuf
			a.klineBufMu.Unlock()
			if kl.OpenTime == lastOpen {
				// 更新最后一根
				a.klineBufMu.Lock()
				for i := range a.klineBuf {
					if a.klineBuf[i].OpenTime == kl.OpenTime {
						a.klineBuf[i] = kl
						break
					}
				}
				a.klineBufMu.Unlock()
				payload := map[string]interface{}{
					"openTime":  kl.OpenTime, "closeTime": kl.CloseTime,
					"open": kl.Open, "high": kl.High, "low": kl.Low, "close": kl.Close, "volume": kl.Volume,
				}
				wailsRuntime.EventsEmit(a.ctx, "kline:update", payload)
			} else if kl.OpenTime == lastOpen+ms {
				// 下一根，追加
				a.klineBufMu.Lock()
				a.klineBuf = append(a.klineBuf, kl)
				cap := int(a.klineLimit)
				if cap <= 0 {
					cap = 1000
				}
				if len(a.klineBuf) > cap {
					a.klineBuf = a.klineBuf[len(a.klineBuf)-cap:]
				}
				a.klineBufMu.Unlock()
				payload := map[string]interface{}{
					"openTime":  kl.OpenTime, "closeTime": kl.CloseTime,
					"open": kl.Open, "high": kl.High, "low": kl.Low, "close": kl.Close, "volume": kl.Volume,
				}
				wailsRuntime.EventsEmit(a.ctx, "kline:update", payload)
			} else {
				// 不连续，重新拉 limit 根
				a.klineBufMu.Lock()
				limit := a.klineLimit
				a.klineBufMu.Unlock()
				if limit <= 0 {
					limit = 1000
				}
				full, err := binance.FetchKlines(symbol, interval, limit, nil)
				if err != nil || len(full) == 0 {
					continue
				}
				a.klineBufMu.Lock()
				a.klineBuf = full
				a.klineBufMu.Unlock()
				snap := make([]factor.KLine, len(full))
				for i, k := range full {
					snap[i] = *k
				}
				wailsRuntime.EventsEmit(a.ctx, "kline:snapshot", map[string]interface{}{
					"klines": snap, "symbol": symbol, "interval": interval,
				})
			}
		}
	}
}

// StopKlineStream 停止 K 线轮询，并通知前端状态为已停止
func (a *App) StopKlineStream() {
	if a.klinePollStop != nil {
		select {
		case <-a.klinePollStop:
		default:
			close(a.klinePollStop)
		}
		a.klinePollStop = nil
	}
	if a.klineStream != nil {
		a.klineStream.Stop()
		a.klineStream = nil
	}
	wailsRuntime.EventsEmit(a.ctx, "kline:status", map[string]interface{}{"status": "stopped"})
}

// SetCozeKlineCount 设置发给豆包的 K 线根数（定时预测与手动预测均使用），默认 50
func (a *App) SetCozeKlineCount(count int) {
	a.klineBufMu.Lock()
	defer a.klineBufMu.Unlock()
	if count <= 0 {
		count = 50
	}
	a.cozeKlineCount = count
}

// emitCozeStatus / emitCozeResult 供 coze 包回调，推送状态与结果到前端
func (a *App) emitCozeStatus(status, message string) {
	wailsRuntime.EventsEmit(a.ctx, "coze:status", map[string]interface{}{"status": status, "message": message})
}
func (a *App) emitCozeResult(res *coze.CozeStructuredResult) {
	wailsRuntime.EventsEmit(a.ctx, "coze:result", res)
}

// CozePredictStructured 手动调用 Coze 结构化预测（传入当前 K 线，count 为发给豆包的 K 线根数，≤0 用 50）
func (a *App) CozePredictStructured(klines []factor.KLine, symbol string, count int) (*coze.CozeStructuredResult, error) {
	if symbol == "" {
		symbol = "BTCUSDT"
	}
	if count <= 0 {
		count = 50
	}
	ptr := make([]*factor.KLine, len(klines))
	for i := range klines {
		ptr[i] = &klines[i]
	}
	return coze.PredictStructuredWithNotify(a.ctx, ptr, symbol, count, a.emitCozeStatus, a.emitCozeResult)
}

