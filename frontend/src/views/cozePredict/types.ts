export interface KLine {
  openTime: number
  open: number
  high: number
  low: number
  close: number
  volume: number
  closeTime: number
  quoteAssetVolume: number
  numberOfTrades: number
  takerBuyVolume: number
  takerBuyQuoteVolume: number
}

export type Direction = 'LONG' | 'SHORT'

export interface BoxPullbackParams {
  lookaheadN: number
  minK1BodyPercent: number
  k1StrengthLookback: number
  minK1BodyToAvgRatio: number
  trendMAPeriod: number
  minBoxRangePercent: number
  maxBoxRangePercent: number
  touchTolerancePercent: number
  minConfirmWickBodyRatio: number
  cooldownBars: number
  riskRewardRatio: number
}

export interface EMATrendPullbackParams {
  fastPeriod: number
  slowPeriod: number
  breakoutLookback: number
  pullbackLookahead: number
  pullbackTolerancePercent: number
  atrPeriod: number
  stopATRMultiplier: number
  cooldownBars: number
  riskRewardRatio: number
}

export interface StrategySignal {
  strategyName: string
  direction: Direction
  k1Index: number
  triggerIndex: number
  entryIndex: number
  k1OpenTime: number
  triggerTime: number
  entryTime: number
  boxHigh: number
  boxLow: number
  entryPrice: number
  stopLoss: number
  takeProfit: number
  riskRewardRatio: number
  confirmBarOpen: number
  confirmBarClose: number
  confirmBarLow: number
  confirmBarHigh: number
  reason: string
}

export interface BacktestTrade {
  id: number
  strategyName: string
  direction: Direction
  signalIndex: number
  entryIndex: number
  exitIndex: number
  signalTime: number
  entryTime: number
  exitTime: number
  entryPrice: number
  exitPrice: number
  stopLoss: number
  takeProfit: number
  quantity: number
  orderValueUSDT: number
  pnl: number
  pnlPercent: number
  balanceAfter: number
  exitReason: string
  holdBars: number
}

export interface BacktestSummary {
  totalSignals: number
  executedTrades: number
  wins: number
  losses: number
  skippedSignals: number
  winRate: number
  finalBalance: number
  totalPnL: number
  roi: number
  maxDrawdown: number
}

export interface BacktestReport {
  strategyName: string
  dataPath: string
  resultPath: string
  generatedAt: string
  initialBalance: number
  positionSizeUSDT: number
  params: BoxPullbackParams
  klines: KLine[]
  signals: StrategySignal[]
  trades: BacktestTrade[]
  summary: BacktestSummary
}

export interface EMABacktestReport {
  strategyName: string
  dataPath: string
  resultPath: string
  generatedAt: string
  initialBalance: number
  positionSizeUSDT: number
  params: EMATrendPullbackParams
  klines: KLine[]
  signals: StrategySignal[]
  trades: BacktestTrade[]
  summary: BacktestSummary
}

export interface RunBacktestRequest {
  dataPath: string
  strategyName: string
  params: BoxPullbackParams
  initialBalance: number
  positionSizeUSDT: number
  resultPath: string
}

export interface RunEMABacktestRequest {
  dataPath: string
  strategyName: string
  params: EMATrendPullbackParams
  initialBalance: number
  positionSizeUSDT: number
  resultPath: string
}
