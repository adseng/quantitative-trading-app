package main

import (
	"context"

	"quantitative-trading-app/internal/appservice"
	"quantitative-trading-app/internal/backtest"
	"quantitative-trading-app/internal/config"
	"quantitative-trading-app/internal/market"

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

// LoadLocalKlines 从本地数据文件加载 K 线。
func (a *App) LoadLocalKlines(path string) ([]market.KLine, error) {
	return a.service.LoadLocalKlines(path)
}

// RunBacktest 从本地文件加载 K 线并运行订单级回测。
func (a *App) RunBacktest(req backtest.RunRequest) (backtest.Report, error) {
	return a.service.RunBacktest(req)
}

// RunEMABacktest 从本地文件加载 K 线并运行 EMA 趋势回踩确认策略回测。
func (a *App) RunEMABacktest(req backtest.RunEMARequest) (backtest.EMAReport, error) {
	return a.service.RunEMABacktest(req)
}
