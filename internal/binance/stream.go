package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"quantitative-trading-app/internal/config"
	"quantitative-trading-app/internal/factor"

	"github.com/gorilla/websocket"
	"golang.org/x/net/proxy"
)

// 默认 WS base（futures）。也支持用 .env.local 的 BINANCE_WS_URL 覆盖：
// - wss://fstream.binance.com
// - wss://fstream.binance.com/ws
const defaultStreamBaseURL = "wss://fstream.binance.com"

// KlineStream 订阅 K 线 WebSocket，通过 OnKline 回调推送
type KlineStream struct {
	symbol   string
	interval string
	conn     *websocket.Conn
	onKline  func(*factor.KLine)
	onStatus func(string)
	onError  func(error)
	stopCh   chan struct{}
	mu       sync.Mutex
}

type wsKlineEvent struct {
	EventType string    `json:"e"`
	Kline     wsKlineK  `json:"k"`
}

type wsKlineK struct {
	OpenTime  int64  `json:"t"`
	CloseTime int64  `json:"T"`
	Symbol    string `json:"s"`
	Interval  string `json:"i"`
	Open      string `json:"o"`
	Close     string `json:"c"`
	High      string `json:"h"`
	Low       string `json:"l"`
	Volume    string `json:"v"`
	Closed    bool   `json:"x"`
}

func parseFloatStr(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func resolveWSBaseURL() string {
	base := strings.TrimSpace(config.Get(config.KeyWSURL, ""))
	if base == "" {
		base = defaultStreamBaseURL
	}
	base = strings.TrimRight(base, "/")
	// 兼容传入已带 /ws 的情况
	if strings.HasSuffix(base, "/ws") {
		return base
	}
	return base + "/ws"
}

// StartKlineStream 启动 K 线 WebSocket 流，symbol 如 btcusdt，interval 如 15m
func StartKlineStream(symbol, interval string, onKline func(*factor.KLine), onStatus func(string), onError func(error)) (*KlineStream, error) {
	symbol = strings.ToLower(symbol)
	if symbol == "" {
		symbol = "btcusdt"
	}
	if interval == "" {
		interval = "15m"
	}

	// 正确订阅 URL:
	// wss://fstream.binance.com/ws/btcusdt@kline_15m
	wsURL := fmt.Sprintf("%s/%s@kline_%s", resolveWSBaseURL(), symbol, interval)

	proxyURL := config.Get(config.KeyProxy, "")
	if proxyURL == "" {
		proxyURL = config.Get("HTTPS_PROXY", "")
	}
	if proxyURL == "" {
		proxyURL = config.Get("HTTP_PROXY", "")
	}

	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	if proxyURL != "" {
		pu, err := url.Parse(proxyURL)
		if err == nil {
			switch strings.ToLower(pu.Scheme) {
			case "socks5", "socks5h":
				// gorilla/websocket 的 Proxy 只支持 HTTP 代理；socks5 需要自定义 NetDialContext
				d, derr := proxy.FromURL(pu, proxy.Direct)
				if derr == nil {
					dialer.NetDialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
						_ = ctx // socks dialer 不支持 ctx，保留签名兼容
						return d.Dial(network, addr)
					}
				}
			default:
				dialer.Proxy = http.ProxyURL(pu)
			}
		}
	}

	if onStatus != nil {
		onStatus("dialing")
	}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("websocket dial failed (url=%s proxy=%s): %w", wsURL, proxyURL, err)
	}

	s := &KlineStream{
		symbol:   symbol,
		interval: interval,
		conn:     conn,
		onKline:  onKline,
		onStatus: onStatus,
		onError:  onError,
		stopCh:   make(chan struct{}),
	}

	if onStatus != nil {
		onStatus("connected")
	}
	go s.readLoop()
	return s, nil
}

func (s *KlineStream) readLoop() {
	defer s.conn.Close()
	first := true
	for {
		select {
		case <-s.stopCh:
			return
		default:
		}
		_, msg, err := s.conn.ReadMessage()
		if err != nil {
			log.Printf("[binance/stream] read error: %v", err)
			s.mu.Lock()
			onErr := s.onError
			s.mu.Unlock()
			if onErr != nil {
				onErr(err)
			}
			return
		}
		var evt wsKlineEvent
		if err := json.Unmarshal(msg, &evt); err != nil {
			continue
		}
		if evt.EventType != "kline" {
			continue
		}
		if first {
			first = false
			s.mu.Lock()
			onStatus := s.onStatus
			s.mu.Unlock()
			if onStatus != nil {
				onStatus("receiving")
			}
		}
		k := evt.Kline
		kl := &factor.KLine{
			OpenTime:  k.OpenTime,
			CloseTime: k.CloseTime,
			Open:      parseFloatStr(k.Open),
			High:      parseFloatStr(k.High),
			Low:       parseFloatStr(k.Low),
			Close:     parseFloatStr(k.Close),
			Volume:    parseFloatStr(k.Volume),
		}
		s.mu.Lock()
		fn := s.onKline
		s.mu.Unlock()
		if fn != nil {
			fn(kl)
		}
		// 仅在新 K 线收盘时打日志，便于确认 WebSocket 有数据
		if evt.Kline.Closed {
			log.Printf("[binance/stream] kline closed openTime=%d close=%.2f", kl.OpenTime, kl.Close)
		}
	}
}

// Stop 停止 WebSocket 流
func (s *KlineStream) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-s.stopCh:
	default:
		close(s.stopCh)
	}
	_ = s.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}
