package main

import (
	"context"

	"quantitative-trading-app/internal/appservice"
	"quantitative-trading-app/internal/batchtest/cases"
	"quantitative-trading-app/internal/config"
	"quantitative-trading-app/internal/coze"
	"quantitative-trading-app/internal/factor"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	service *appservice.Service
}

func NewApp() *App {
	return &App{
		service: appservice.New(),
	}
}

func (a *App) startup(ctx context.Context) {
	_ = config.Load()
	a.service.BindRuntime(ctx, func(eventName string, payload any) {
		wailsRuntime.EventsEmit(ctx, eventName, payload)
	})
}

// Market Data

// FetchKlines 从币安合约获取 K 线数据，按时间升序返回。
func (a *App) FetchKlines(symbol, interval string, limit int64) ([]factor.KLine, error) {
	return a.service.FetchKlines(symbol, interval, limit)
}

// Factor Backtests

// BacktestEmotion 使用基础因子配置进行回测。
func (a *App) BacktestEmotion(klines []factor.KLine, useMA, useTrend bool, maShort, maLong, trendN int, maWeight, trendWeight float64) (factor.BacktestResultSummary, error) {
	return a.service.BacktestEmotion(klines, useMA, useTrend, maShort, maLong, trendN, maWeight, trendWeight)
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
	return a.service.BacktestEmotionV2(
		klines,
		useMA, maShort, maLong, maWeight,
		useTrend, trendN, trendWeight,
		useRSI, rsiPeriod, rsiOverbought, rsiOversold, rsiWeight,
		useMACD, macdFast, macdSlow, macdSignal, macdWeight,
		useBoll, bollPeriod, bollMultiplier, bollWeight,
		useBreakout, breakoutPeriod, breakoutWeight,
		usePriceVsMA, priceVsMAPeriod, priceVsMAWeight,
		useATR, atrPeriod, atrWeight,
		useVolume, volumePeriod, volumeWeight,
		useSession, sessionWeight,
		useMACross, macrossShort, macrossLong, macrossWeight, macrossWindow, macrossPreempt,
	)
}

// LogBacktestResult 将回测参数和正确率追加到 Excel 文件。
func (a *App) LogBacktestResult(
	symbol, interval string, limit int64,
	useMA, useTrend bool, maShort, maLong, trendN int, maWeight, trendWeight float64,
	accuracy float64, correct, total int,
) error {
	return a.service.LogBacktestResult(symbol, interval, limit, useMA, useTrend, maShort, maLong, trendN, maWeight, trendWeight, accuracy, correct, total)
}

// Batch Testing

// StartBatchTest 开始或继续批量回测。
// 首次运行时分页获取 K 线（约 60s），之后 resume 时直接复用缓存数据。
// 通过 Wails 事件 "batch:progress" 推送进度。
func (a *App) StartBatchTest() error {
	return a.service.StartBatchTest()
}

// StopBatchTest 停止批量回测。
func (a *App) StopBatchTest() {
	a.service.StopBatchTest()
}

// GetBatchTestProgress 获取当前进度。
func (a *App) GetBatchTestProgress() map[string]interface{} {
	return a.service.GetBatchTestProgress()
}

// GetBatchTestResults 获取最近的回测结果（用于页面初始化展示，从 Excel 加载历史数据）。
func (a *App) GetBatchTestResults(maxN int) []cases.TestResult {
	return a.service.GetBatchTestResults(maxN)
}

// CozePredict

func (a *App) CozePredictStructured(symbol, interval string, count int) (*coze.CozeStructuredResult, error) {
	return a.service.CozePredictStructured(symbol, interval, count)
}
