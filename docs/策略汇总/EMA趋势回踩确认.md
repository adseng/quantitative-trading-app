# EMA 趋势回踩确认策略

## 1. 策略定位
- 目标：只做已经形成趋势的二次介入，不追第一根突破大阳/大阴。
- 适用品种：波动连续、趋势段明显的合约或高流动性币种。
- 适用周期：`5m`、`15m`、`1h`。

## 2. 核心思路
- 先用双 EMA 判断趋势方向。
- 再用最近 `N` 根高低点判断是否发生有效突破。
- 突破后不立刻追单，等待价格在限定根数内回踩 `突破位 / 快 EMA`。
- 回踩结束后由确认 K 线重新顺趋势收回，下一根 K 线开盘入场。

## 3. 多头规则
### 趋势成立
- `EMA(fast) > EMA(slow)`。
- `EMA(slow)` 当前值不低于上一根，说明慢趋势未走平转弱。

### 突破成立
- 当前 K 线为阳线。
- 当前收盘价大于前 `breakoutLookback` 根 K 线最高价。
- 当前收盘价位于快 EMA 上方。

### 回踩确认
- 在突破后的 `pullbackLookahead` 根 K 线内寻找第一根确认 K 线。
- 该 K 线最低价必须回到 `突破位 / 快 EMA` 附近，允许 `pullbackTolerancePercent` 容差。
- 该 K 线收盘重新站回快 EMA 上方。
- 该 K 线收盘仍高于突破位。
- 该 K 线本身为阳线。

### 入场与风控
- 入场：确认 K 线之后下一根 K 线开盘价做多。
- 止损：`min(确认K线最低价, 快EMA)` 再减去 `ATR * stopATRMultiplier`。
- 止盈：按固定盈亏比 `riskRewardRatio` 计算。

## 4. 空头规则
### 趋势成立
- `EMA(fast) < EMA(slow)`。
- `EMA(slow)` 当前值不高于上一根。

### 突破成立
- 当前 K 线为阴线。
- 当前收盘价小于前 `breakoutLookback` 根 K 线最低价。
- 当前收盘价位于快 EMA 下方。

### 回踩确认
- 在突破后的 `pullbackLookahead` 根 K 线内寻找第一根确认 K 线。
- 该 K 线最高价必须反抽到 `突破位 / 快 EMA` 附近，允许 `pullbackTolerancePercent` 容差。
- 该 K 线收盘重新回到快 EMA 下方。
- 该 K 线收盘仍低于突破位。
- 该 K 线本身为阴线。

### 入场与风控
- 入场：确认 K 线之后下一根 K 线开盘价做空。
- 止损：`max(确认K线最高价, 快EMA)` 再加上 `ATR * stopATRMultiplier`。
- 止盈：按固定盈亏比 `riskRewardRatio` 计算。

## 5. 参数说明
| 参数 | 含义 | 默认值 |
| --- | --- | --- |
| `fastPeriod` | 快 EMA 周期 | `20` |
| `slowPeriod` | 慢 EMA 周期 | `60` |
| `breakoutLookback` | 判断突破的回看窗口 | `20` |
| `pullbackLookahead` | 突破后允许等待回踩的 K 线数 | `5` |
| `pullbackTolerancePercent` | 回踩到突破位/EMA 的容差 | `0.003` |
| `atrPeriod` | ATR 波动计算周期 | `14` |
| `stopATRMultiplier` | 止损缓冲倍数 | `1.0` |
| `cooldownBars` | 每次信号后的冷却根数 | `3` |
| `riskRewardRatio` | 固定盈亏比 | `1.5` |

## 6. 首版落地约束
- 不做成交量过滤。
- 不做更高周期共振。
- 不做分批止盈。
- 支持多仓位并行，由回测引擎统一管理。
- 冷却时间只约束“新突破信号”的生成，不强制平掉已有仓位。

## 7. 推荐起步参数
### 偏稳健
- `fastPeriod=20`
- `slowPeriod=60`
- `breakoutLookback=20`
- `pullbackLookahead=5`
- `pullbackTolerancePercent=0.003`
- `atrPeriod=14`
- `stopATRMultiplier=1.0`
- `cooldownBars=3`
- `riskRewardRatio=1.5`

### 偏激进
- `fastPeriod=12`
- `slowPeriod=36`
- `breakoutLookback=12`
- `pullbackLookahead=4`
- `pullbackTolerancePercent=0.004`
- `atrPeriod=10`
- `stopATRMultiplier=0.8`
- `cooldownBars=1`
- `riskRewardRatio=1.3`

## 8. 后续可扩展项
- 增加成交量确认，过滤无量突破。
- 增加 `EMA(slow)` 斜率阈值，过滤横盘假趋势。
- 增加回踩后“确认实体强度”过滤。
- 增加分批止盈和移动止损。
