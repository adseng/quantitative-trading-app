package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Load 加载 .env.local，若文件不存在则忽略
func Load() error {
	return godotenv.Load(".env.local")
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
	KeyAPIKey     = "BINANCE_API_KEY"
	KeySecretKey  = "BINANCE_SECRET_KEY"
	KeyBaseURL    = "BINANCE_BASE_URL"
	KeyWSURL      = "BINANCE_WS_URL"
	KeySymbol     = "BINANCE_SYMBOL"
)

// BinanceTestnetBaseURL USDT 合约测试网 REST 地址
// 来源: https://developers.binance.com/docs/derivatives/usds-margined-futures/general-info
const BinanceTestnetBaseURL = "https://demo-fapi.binance.com"

// BinanceTestnetWSURL USDT 合约测试网 WebSocket 地址
const BinanceTestnetWSURL = "wss://fstream.binancefuture.com"
