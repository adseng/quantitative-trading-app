import { coze, factor } from '@wails/go/models'

export type KLine = factor.KLine
export type CozeResult = coze.CozeStructuredResult
export type CozeScenario = coze.CozeScenario

export interface KlineSnapshotEvent {
  sessionId: string
  symbol: string
  interval: string
  limit: number
  source: string
  klines: KLine[]
}

export interface KlineUpdateEvent {
  sessionId: string
  symbol: string
  interval: string
  limit: number
  source: string
  kline: KLine
}

export interface KlineStatusEvent {
  sessionId: string
  symbol: string
  interval: string
  limit: number
  status: string
  message?: string
}

export interface KlineErrorEvent {
  sessionId: string
  symbol: string
  interval: string
  limit: number
  error: string
  retryable: boolean
}

export interface CozeStatusEvent {
  status: string
  message?: string
  symbol: string
  interval: string
  count: number
}
