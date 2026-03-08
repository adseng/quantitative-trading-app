package coze

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"quantitative-trading-app/internal/binance"
	"quantitative-trading-app/internal/factor"
)

const (
	defaultKlineLimit        int64 = 1000
	maxKlineLimit            int64 = 1500
	defaultPredictCount            = 50
	minPredictCount                = 5
	maxPredictCount                = 500
	defaultPredictFetchLimit int64 = 100
)

type EventEmitter func(eventName string, payload any)

type Service struct {
	ctx  context.Context
	emit EventEmitter

	klineBuf      []*factor.KLine
	klineBufMu    sync.Mutex
	klineSymbol   string
	klineInterval string
}

type cozeStatusEvent struct {
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
	Symbol   string `json:"symbol"`
	Interval string `json:"interval"`
	Count    int    `json:"count"`
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) BindRuntime(ctx context.Context, emit EventEmitter) {
	s.klineBufMu.Lock()
	defer s.klineBufMu.Unlock()
	s.ctx = ctx
	s.emit = emit
}

func (s *Service) FetchKlines(symbol, interval string, limit int64) ([]factor.KLine, error) {
	symbol = normalizeSymbol(symbol)
	interval = normalizeInterval(interval)
	limit = clampKlineLimit(limit)

	klines, err := binance.FetchKlines(symbol, interval, limit, nil)
	if err != nil || klines == nil {
		return nil, err
	}

	s.setKlineBuffer(symbol, interval, limit, klines)
	return flattenKlines(klines), nil
}

func (s *Service) PredictStructured(symbol, interval string, count int) (*CozeStructuredResult, error) {
	symbol = normalizeSymbol(symbol)
	interval = normalizeInterval(interval)
	count = clampPredictCount(count)
	if count < minPredictCount {
		return nil, fmt.Errorf("至少需要 %d 根 K 线", minPredictCount)
	}

	klines, ok := s.getKlineBuffer(symbol, interval)
	if !ok || len(klines) < count {
		fetchLimit := int64(count)
		if fetchLimit < defaultPredictFetchLimit {
			fetchLimit = defaultPredictFetchLimit
		}

		latest, err := binance.FetchKlines(symbol, interval, fetchLimit, nil)
		if err != nil {
			return nil, err
		}
		if len(latest) < minPredictCount {
			return nil, fmt.Errorf("至少需要 %d 根 K 线", minPredictCount)
		}
		s.setKlineBuffer(symbol, interval, fetchLimit, latest)
		klines = cloneKlinePointers(latest)
	}

	statusFn := func(status, message string) {
		s.emitEvent("coze:status", cozeStatusEvent{
			Status:   status,
			Message:  message,
			Symbol:   symbol,
			Interval: interval,
			Count:    count,
		})
	}

	return PredictStructuredWithNotify(s.ctx, klines, symbol, count, statusFn, nil)
}

func (s *Service) setKlineBuffer(symbol, interval string, _ int64, klines []*factor.KLine) {
	s.klineBufMu.Lock()
	defer s.klineBufMu.Unlock()
	s.klineSymbol = symbol
	s.klineInterval = interval
	s.klineBuf = cloneKlinePointers(klines)
}

func (s *Service) getKlineBuffer(symbol, interval string) ([]*factor.KLine, bool) {
	s.klineBufMu.Lock()
	defer s.klineBufMu.Unlock()
	if len(s.klineBuf) == 0 {
		return nil, false
	}
	if symbol != "" && s.klineSymbol != symbol {
		return nil, false
	}
	if interval != "" && s.klineInterval != interval {
		return nil, false
	}
	return cloneKlinePointers(s.klineBuf), true
}

func (s *Service) emitEvent(eventName string, payload any) {
	if s.emit != nil {
		s.emit(eventName, payload)
	}
}

func clampKlineLimit(limit int64) int64 {
	if limit <= 0 {
		return defaultKlineLimit
	}
	if limit > maxKlineLimit {
		return maxKlineLimit
	}
	return limit
}

func clampPredictCount(count int) int {
	if count <= 0 {
		return defaultPredictCount
	}
	if count > maxPredictCount {
		return maxPredictCount
	}
	return count
}

func normalizeSymbol(symbol string) string {
	symbol = strings.TrimSpace(strings.ToUpper(symbol))
	if symbol == "" {
		return "BTCUSDT"
	}
	return symbol
}

func normalizeInterval(interval string) string {
	interval = strings.TrimSpace(interval)
	if interval == "" {
		return "15m"
	}
	return interval
}

func cloneKlinePointers(src []*factor.KLine) []*factor.KLine {
	out := make([]*factor.KLine, len(src))
	for i, k := range src {
		if k == nil {
			continue
		}
		cp := *k
		out[i] = &cp
	}
	return out
}

func flattenKlines(src []*factor.KLine) []factor.KLine {
	out := make([]factor.KLine, len(src))
	for i, k := range src {
		if k != nil {
			out[i] = *k
		}
	}
	return out
}
