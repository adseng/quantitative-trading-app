# 量化交易工作台

基于 `Wails + Go + React` 的量化交易桌面工具。当前版本聚焦一条最小可用链路：

`Binance K 线抓取 -> 本地数据文件 -> 箱体突破回踩确认策略 -> 订单级回测 -> K 线图查看结果`

## 当前能力

- 保留桌面应用骨架与交互式 K 线图
- 保留 Binance 历史 K 线抓取能力
- 新增本地数据加载与订单级回测
- 首个策略为 `箱体突破回踩确认`
- 回测规则默认使用：
  - 初始资金 `10000 USDT`
  - 单次下单 `100 USDT`
  - 盈亏比 `2:1`

## 项目结构

```text
quantitative-trading/
├── internal/
│   ├── appservice/       # Wails 前后端桥接
│   ├── backtest/         # 订单级回测引擎与结果导出
│   ├── binance/          # Binance 行情访问
│   ├── config/           # .env 配置读取
│   ├── datafile/         # K 线 JSONL 文件读写
│   ├── market/           # K 线领域模型
│   └── strategy/         # 策略实现
├── cmd/
│   ├── fetchdata/        # 拉取 Binance K 线到本地
│   └── backtest/         # 从本地数据文件执行回测
└── frontend/             # React UI（Wails 桌面应用）
```

## 配置

复制 `.env.example` 为 `.env.local` 或 `.env`，按需填写：

| 变量 | 说明 |
|------|------|
| `BINANCE_BASE_URL` | 币安 REST 地址，默认正式网 |
| `BINANCE_PROXY` | 代理地址，如 `http://127.0.0.1:7897` |
| `BINANCE_SYMBOL` | 默认交易对，默认 `BTCUSDT` |
| `BINANCE_REQUEST_TIMEOUT` | 请求超时秒数 |

K 线抓取使用公开行情接口，不需要 API Key。

## 快速开始

```bash
# 1. 安装依赖
go mod download
cd frontend && pnpm install

# 2. 拉取 K 线到本地
go run ./cmd/fetchdata -symbol BTCUSDT -interval 15m -count 10000

# 3. 运行回测
go run ./cmd/backtest -data docs/db/data-k-15m-10000.txt

# 4. 启动桌面应用
wails dev
```

## 常用命令

```bash
# 自定义输出路径抓取 K 线
go run ./cmd/fetchdata -interval 1h -count 5000 -out docs/db/data-k-1h-5000.txt

# 自定义回测参数
go run ./cmd/backtest `
  -data docs/db/data-k-15m-10000.txt `
  -lookahead 5 `
  -min-k1-body 0.003 `
  -touch-tolerance 0.001 `
  -confirm-wick-ratio 1.2 `
  -cooldown 3 `
  -rr 2
```

## 输出文件

- K 线数据默认写入 `docs/db/data-k-{interval}-{count}.txt`
- 回测结果默认写入 `docs/db/test-box-pullback-confirmation.txt`

## 相关文档

- [整理计划](docs/整理plan.md)
- [箱体突破回踩确认](docs/策略汇总/箱体突破回踩确认.md)
