// siminvest 策略投资收益模拟
//
// 假设 1000U 本金，使用 1/2/3/4 因子的最优解，获取最新 200 根 15 分钟 K 线，
// 逐根回放并按信号投资（pred=1 做多，pred=-1 做空），统计投资次数和最终收益。
package main

import (
	"fmt"
	"os"
	"time"

	"quantitative-trading-app/internal/binance"
	"quantitative-trading-app/internal/config"
	"quantitative-trading-app/internal/factor"
)

const (
	initBalance = 1000.0
	klinesCount = 800
	symbol      = "BTCUSDT"
	interval    = "15m"
)

// 1~4 因子最优配置（来自 docs/多因子回测结论.md）
var strategies = []struct {
	name string
	cfg  factor.FactorConfig
}{
	{
		name: "1因子_Boll_P13_M2.4",
		cfg: factor.FactorConfig{
			UseBoll: true, BollPeriod: 13, BollMultiplier: 2.4, BollWeight: 1,
		},
	},
	{
		name: "2因子_Bo13M2.2+Br20",
		cfg: factor.FactorConfig{
			UseBoll: true, BollPeriod: 13, BollMultiplier: 2.2, BollWeight: 1,
			UseBreakout: true, BreakoutPeriod: 20, BreakoutWeight: 1,
		},
	},
	{
		name: "3因子_R7+Bo13M2.0+Br15",
		cfg: factor.FactorConfig{
			UseRSI: true, RSIPeriod: 7, RSIOverbought: 80, RSIOversold: 20, RSIWeight: 1,
			UseBoll: true, BollPeriod: 13, BollMultiplier: 2.0, BollWeight: 1,
			UseBreakout: true, BreakoutPeriod: 15, BreakoutWeight: 1,
		},
	},
	{
		name: "4因子_R7+Bo13M2.0+Br12+A20",
		cfg: factor.FactorConfig{
			UseRSI: true, RSIPeriod: 7, RSIOverbought: 80, RSIOversold: 20, RSIWeight: 1,
			UseBoll: true, BollPeriod: 13, BollMultiplier: 2.0, BollWeight: 1,
			UseBreakout: true, BreakoutPeriod: 12, BreakoutWeight: 1,
			UseATR: true, ATRPeriod: 20, ATRWeight: 1,
		},
	},
}

func main() {
	_ = config.Load()
	binance.InitClient()

	fmt.Println("=== 策略投资收益模拟 ===")
	fmt.Printf("本金: %.0f U，K线: 最新 %d 根 %s\n\n", initBalance, klinesCount, interval)

	klines, err := binance.FetchKlines(symbol, interval, klinesCount, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取K线失败: %v\n", err)
		os.Exit(1)
	}
	if len(klines) < 50 {
		fmt.Fprintf(os.Stderr, "K线数量不足: %d\n", len(klines))
		os.Exit(1)
	}
	startTime := time.UnixMilli(klines[0].OpenTime).Format("2006-01-02 15:04")
	endTime := time.UnixMilli(klines[len(klines)-1].OpenTime).Format("2006-01-02 15:04")
	fmt.Printf("已获取 %d 根K线 (开盘时间: %s ~ %s)\n\n", len(klines), startTime, endTime)

	fmt.Printf("%-30s %8s %12s %12s\n", "策略", "投资次数", "最终余额(U)", "收益率")
	fmt.Println("--------------------------------------------------------------------------------")

	for _, s := range strategies {
		trades, finalBal := simulate(klines, &s.cfg)
		pct := (finalBal - initBalance) / initBalance * 100
		fmt.Printf("%-30s %8d %12.2f %+11.2f%%\n", s.name, trades, finalBal, pct)
	}
	fmt.Println("--------------------------------------------------------------------------------")
}

// simulate 回放 K 线，按信号投资，返回投资次数和最终余额
func simulate(klines []*factor.KLine, cfg *factor.FactorConfig) (trades int, finalBalance float64) {
	minLen := cfg.MinHistory()
	if len(klines) < minLen+1 {
		return 0, initBalance
	}

	window := minLen + 10
	if window > len(klines) {
		window = len(klines)
	}
	buf := make([]*factor.KLine, window)
	hist := &factor.KLineHistory{}

	balance := initBalance
	trades = 0

	for i := minLen - 1; i < len(klines)-1; i++ {
		factor.FillHistoryWindow(klines, i, window, buf, hist)
		ctx := factor.EvaluateWithConfig(hist, cfg)
		pred := ctx.Prediction()
		if pred == 0 {
			continue
		}

		// 有信号：在 close[i] 入场，close[i+1] 出场
		c0 := klines[i].Close
		c1 := klines[i+1].Close
		if c0 <= 0 {
			continue
		}

		var ret float64
		if pred == 1 {
			ret = (c1 - c0) / c0 // 做多收益率
		} else {
			ret = (c0 - c1) / c0 // 做空收益率
		}

		balance *= (1 + ret)
		trades++
	}

	return trades, balance
}
