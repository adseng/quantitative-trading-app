package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Load 加载 .env.local 或 .env，若文件不存在则忽略
func Load() error {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")
	return nil
}

// Get 获取环境变量，缺省时返回 def
func Get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Binance 相关配置
const (
	KeyAPIKey        = "BINANCE_API_KEY"
	KeySecretKey     = "BINANCE_SECRET_KEY"
	KeyBaseURL       = "BINANCE_BASE_URL"
	KeyWSURL         = "BINANCE_WS_URL"
	KeySymbol        = "BINANCE_SYMBOL"
	KeyProxy         = "BINANCE_PROXY" // 代理地址，如 http://127.0.0.1:7890 或 socks5://127.0.0.1:1080
	KeyRequestTimeout = "BINANCE_REQUEST_TIMEOUT" // 请求超时秒数，默认 30
)

// BinanceTestnetBaseURL USDT 合约测试网 REST 地址
// 来源: https://developers.binance.com/docs/derivatives/usds-margined-futures/general-info
const BinanceTestnetBaseURL = "https://demo-fapi.binance.com"

// BinanceMainnetBaseURL USDT 合约正式网 REST 地址（行情接口公开，无需 API 密钥）
const BinanceMainnetBaseURL = "https://fapi.binance.com"

// BinanceTestnetWSURL USDT 合约测试网 WebSocket 地址
const BinanceTestnetWSURL = "wss://fstream.binancefuture.com"
