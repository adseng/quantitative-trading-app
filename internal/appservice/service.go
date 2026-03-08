package appservice

import (
	"context"
	"fmt"

	"quantitative-trading-app/internal/backtestlog"
	"quantitative-trading-app/internal/batchtest"
	"quantitative-trading-app/internal/batchtest/cases"
	"quantitative-trading-app/internal/binance"
	"quantitative-trading-app/internal/coze"
	"quantitative-trading-app/internal/factor"
)

type EventEmitter func(eventName string, payload any)

type Service struct {
	batchRunner  *batchtest.BatchRunner
	cancelFetch  chan struct{}
	cachedKlines []*factor.KLine
	cozeService  *coze.Service
	emit         EventEmitter
}

func New() *Service {
	return &Service{
		batchRunner: batchtest.NewBatchRunner(),
		cozeService: coze.NewService(),
	}
}

func (s *Service) BindRuntime(ctx context.Context, emit EventEmitter) {
	s.emit = emit
	s.cozeService.BindRuntime(ctx, func(eventName string, payload any) {
		emit(eventName, payload)
	})
}

func (s *Service) FetchKlines(symbol, interval string, limit int64) ([]factor.KLine, error) {
	return s.cozeService.FetchKlines(symbol, interval, limit)
}

func (s *Service) BacktestEmotion(klines []factor.KLine, useMA, useTrend bool, maShort, maLong, trendN int, maWeight, trendWeight float64) (factor.BacktestResultSummary, error) {
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

func (s *Service) BacktestEmotionV2(klines []factor.KLine,
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
		UseMA:           useMA,
		MaShort:         maShort,
		MaLong:          maLong,
		MaWeight:        maWeight,
		UseTrend:        useTrend,
		TrendN:          trendN,
		TrendWeight:     trendWeight,
		UseRSI:          useRSI,
		RSIPeriod:       rsiPeriod,
		RSIOverbought:   rsiOverbought,
		RSIOversold:     rsiOversold,
		RSIWeight:       rsiWeight,
		UseMACD:         useMACD,
		MACDFast:        macdFast,
		MACDSlow:        macdSlow,
		MACDSignal:      macdSignal,
		MACDWeight:      macdWeight,
		UseBoll:         useBoll,
		BollPeriod:      bollPeriod,
		BollMultiplier:  bollMultiplier,
		BollWeight:      bollWeight,
		UseBreakout:     useBreakout,
		BreakoutPeriod:  breakoutPeriod,
		BreakoutWeight:  breakoutWeight,
		UsePriceVsMA:    usePriceVsMA,
		PriceVsMAPeriod: priceVsMAPeriod,
		PriceVsMAWeight: priceVsMAWeight,
		UseATR:          useATR,
		ATRPeriod:       atrPeriod,
		ATRWeight:       atrWeight,
		UseVolume:       useVolume,
		VolumePeriod:    volumePeriod,
		VolumeWeight:    volumeWeight,
		UseSession:      useSession,
		SessionWeight:   sessionWeight,
		UseMACross:      useMACross,
		MACrossShort:    macrossShort,
		MACrossLong:     macrossLong,
		MACrossWeight:   macrossWeight,
		MACrossWindow:   macrossWindow,
		MACrossPreempt:  macrossPreempt,
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

func (s *Service) LogBacktestResult(
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

func (s *Service) StartBatchTest() error {
	if s.batchRunner.IsRunning() {
		return nil
	}
	if s.batchRunner.NextIndex() >= s.batchRunner.TotalCases() {
		s.batchRunner.ResetForNewBatch()
	}

	s.cancelFetch = make(chan struct{})
	emit := func(evt batchtest.ProgressEvent) {
		if s.emit != nil {
			s.emit("batch:progress", evt)
		}
	}

	go func() {
		if len(s.cachedKlines) > 0 {
			emit(batchtest.ProgressEvent{
				Phase:   "fetching",
				Message: fmt.Sprintf("使用缓存 K 线数据（%d 根），继续回测...", len(s.cachedKlines)),
			})
			s.batchRunner.Run(s.cachedKlines, emit)
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
			CancelCh: s.cancelFetch,
		})
		if err != nil {
			emit(batchtest.ProgressEvent{Phase: "error", Message: "获取 K 线失败: " + err.Error()})
			return
		}

		emit(batchtest.ProgressEvent{
			Phase:   "fetching",
			Message: fmt.Sprintf("K 线获取完成，共 %d 根，开始回测...", len(klines)),
		})

		s.cachedKlines = klines
		s.batchRunner.Run(klines, emit)
	}()

	return nil
}

func (s *Service) StopBatchTest() {
	if s.cancelFetch != nil {
		select {
		case <-s.cancelFetch:
		default:
			close(s.cancelFetch)
		}
	}
	s.batchRunner.Stop()
}

func (s *Service) GetBatchTestProgress() map[string]interface{} {
	return map[string]interface{}{
		"running":    s.batchRunner.IsRunning(),
		"nextIndex":  s.batchRunner.NextIndex(),
		"totalCases": s.batchRunner.TotalCases(),
	}
}

func (s *Service) GetBatchTestResults(maxN int) []cases.TestResult {
	if maxN <= 0 {
		maxN = 200
	}
	return s.batchRunner.GetLastResults(maxN)
}

func (s *Service) CozePredictStructured(symbol, interval string, count int) (*coze.CozeStructuredResult, error) {
	return s.cozeService.PredictStructured(symbol, interval, count)
}
