package main

import (
	"context"
	"fmt"

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
	cozeService  *coze.Service
}

func NewApp() *App {
	return &App{
		batchRunner: batchtest.NewBatchRunner(),
		cozeService: coze.NewService(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	_ = config.Load()
	binance.InitClient()
	a.cozeService.BindRuntime(ctx, func(eventName string, payload any) {
		wailsRuntime.EventsEmit(ctx, eventName, payload)
	})
}

// FetchKlines 从币安合约获取 K 线数据，按时间升序返回。
func (a *App) FetchKlines(symbol, interval string, limit int64) ([]factor.KLine, error) {
	return a.cozeService.FetchKlines(symbol, interval, limit)
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
			Chunks:  100,
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

func (a *App) StartKlineStream(symbol, interval string, limit int64) error {
	return a.cozeService.StartKlineStream(symbol, interval, limit)
}

func (a *App) StopKlineStream() {
	a.cozeService.StopKlineStream()
}

func (a *App) CozePredictStructured(symbol, interval string, count int) (*coze.CozeStructuredResult, error) {
	return a.cozeService.PredictStructured(symbol, interval, count)
}
