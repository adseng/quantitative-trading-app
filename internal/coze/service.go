package coze

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

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
	klinePollInterval              = time.Second
)

type EventEmitter func(eventName string, payload any)

type Service struct {
	ctx  context.Context
	emit EventEmitter

	klineSessionID string
	klinePollStop  chan struct{}
	klineBuf       []*factor.KLine
	klineBufMu     sync.Mutex
	klineSymbol    string
	klineInterval  string
	klineLimit     int64
}

type klineSnapshotEvent struct {
	SessionID string         `json:"sessionId"`
	Symbol    string         `json:"symbol"`
	Interval  string         `json:"interval"`
	Limit     int64          `json:"limit"`
	Source    string         `json:"source"`
	Klines    []factor.KLine `json:"klines"`
}

type klineUpdateEvent struct {
	SessionID string       `json:"sessionId"`
	Symbol    string       `json:"symbol"`
	Interval  string       `json:"interval"`
	Limit     int64        `json:"limit"`
	Source    string       `json:"source"`
	Kline     factor.KLine `json:"kline"`
}

type klineStatusEvent struct {
	SessionID string `json:"sessionId"`
	Symbol    string `json:"symbol"`
	Interval  string `json:"interval"`
	Limit     int64  `json:"limit"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
}

type klineErrorEvent struct {
	SessionID string `json:"sessionId"`
	Symbol    string `json:"symbol"`
	Interval  string `json:"interval"`
	Limit     int64  `json:"limit"`
	Error     string `json:"error"`
	Retryable bool   `json:"retryable"`
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

func (s *Service) StartKlineStream(symbol, interval string, limit int64) error {
	symbol = normalizeSymbol(symbol)
	interval = normalizeInterval(interval)
	limit = clampKlineLimit(limit)
	s.StopKlineStream()

	klines, err := binance.FetchKlines(symbol, interval, limit, nil)
	if err != nil {
		return err
	}
	if len(klines) == 0 {
		return fmt.Errorf("未获取到 K 线")
	}

	sessionID := s.nextKlineSessionID(symbol)
	stopCh := make(chan struct{})

	s.klineBufMu.Lock()
	s.klineSessionID = sessionID
	s.klinePollStop = stopCh
	s.klineBufMu.Unlock()

	s.setKlineBuffer(symbol, interval, limit, klines)
	s.emitKlineSnapshot(sessionID, symbol, interval, "rest_snapshot", limit, klines)
	s.emitKlineStatus(sessionID, symbol, interval, limit, "polling", "")

	go s.runKlinePoll(stopCh, sessionID, symbol, interval, limit)
	return nil
}

func (s *Service) StopKlineStream() {
	s.klineBufMu.Lock()
	stopCh := s.klinePollStop
	sessionID := s.klineSessionID
	symbol := s.klineSymbol
	interval := s.klineInterval
	limit := s.klineLimit
	s.klinePollStop = nil
	s.klineSessionID = ""
	s.klineBufMu.Unlock()

	if stopCh != nil {
		select {
		case <-stopCh:
		default:
			close(stopCh)
		}
	}
	if sessionID != "" || symbol != "" {
		s.emitKlineStatus(sessionID, symbol, interval, limit, "stopped", "")
	}
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

func (s *Service) runKlinePoll(stopCh chan struct{}, sessionID, symbol, interval string, limit int64) {
	ticker := time.NewTicker(klinePollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			buf, ok := s.getKlineBuffer(symbol, interval)
			if !ok {
				full, err := binance.FetchKlines(symbol, interval, limit, nil)
				if err != nil || len(full) == 0 {
					if err == nil {
						err = fmt.Errorf("未获取到 K 线")
					}
					s.emitKlineError(sessionID, symbol, interval, limit, err, true)
					continue
				}
				s.setKlineBuffer(symbol, interval, limit, full)
				s.emitKlineSnapshot(sessionID, symbol, interval, "rest_resync", limit, full)
				continue
			}
			if symbol == "" || len(buf) == 0 {
				continue
			}

			latest, err := binance.FetchKlines(symbol, interval, 1, nil)
			if err != nil || len(latest) == 0 {
				if err == nil {
					err = fmt.Errorf("未获取到最新 K 线")
				}
				s.emitKlineError(sessionID, symbol, interval, limit, err, true)
				continue
			}

			kl := latest[0]
			lastOpen := buf[len(buf)-1].OpenTime
			ms := intervalMs(interval)

			if kl.OpenTime == lastOpen {
				for i := range buf {
					if buf[i] != nil && buf[i].OpenTime == kl.OpenTime {
						buf[i] = kl
						break
					}
				}
				s.setKlineBuffer(symbol, interval, limit, buf)
				s.emitKlineUpdate(sessionID, symbol, interval, "rest_poll", limit, kl)
				continue
			}

			if kl.OpenTime == lastOpen+ms {
				buf = append(buf, kl)
				cap := int(limit)
				if cap <= 0 {
					cap = int(defaultKlineLimit)
				}
				if len(buf) > cap {
					buf = buf[len(buf)-cap:]
				}
				s.setKlineBuffer(symbol, interval, limit, buf)
				s.emitKlineUpdate(sessionID, symbol, interval, "rest_poll", limit, kl)
				continue
			}

			full, err := binance.FetchKlines(symbol, interval, limit, nil)
			if err != nil || len(full) == 0 {
				if err == nil {
					err = fmt.Errorf("未获取到 K 线")
				}
				s.emitKlineError(sessionID, symbol, interval, limit, err, true)
				continue
			}
			s.setKlineBuffer(symbol, interval, limit, full)
			s.emitKlineSnapshot(sessionID, symbol, interval, "rest_resync", limit, full)
		}
	}
}

func (s *Service) nextKlineSessionID(symbol string) string {
	return fmt.Sprintf("%s-%d", strings.ToLower(symbol), time.Now().UnixNano())
}

func (s *Service) setKlineBuffer(symbol, interval string, limit int64, klines []*factor.KLine) {
	s.klineBufMu.Lock()
	defer s.klineBufMu.Unlock()
	s.klineSymbol = symbol
	s.klineInterval = interval
	s.klineLimit = limit
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

func (s *Service) emitKlineSnapshot(sessionID, symbol, interval, source string, limit int64, klines []*factor.KLine) {
	s.emitEvent("kline:snapshot", klineSnapshotEvent{
		SessionID: sessionID,
		Symbol:    symbol,
		Interval:  interval,
		Limit:     limit,
		Source:    source,
		Klines:    flattenKlines(klines),
	})
}

func (s *Service) emitKlineUpdate(sessionID, symbol, interval, source string, limit int64, kl *factor.KLine) {
	if kl == nil {
		return
	}
	s.emitEvent("kline:update", klineUpdateEvent{
		SessionID: sessionID,
		Symbol:    symbol,
		Interval:  interval,
		Limit:     limit,
		Source:    source,
		Kline:     *kl,
	})
}

func (s *Service) emitKlineStatus(sessionID, symbol, interval string, limit int64, status, message string) {
	s.emitEvent("kline:status", klineStatusEvent{
		SessionID: sessionID,
		Symbol:    symbol,
		Interval:  interval,
		Limit:     limit,
		Status:    status,
		Message:   message,
	})
}

func (s *Service) emitKlineError(sessionID, symbol, interval string, limit int64, err error, retryable bool) {
	if err == nil {
		return
	}
	s.emitEvent("kline:error", klineErrorEvent{
		SessionID: sessionID,
		Symbol:    symbol,
		Interval:  interval,
		Limit:     limit,
		Error:     err.Error(),
		Retryable: retryable,
	})
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
