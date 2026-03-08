import { coze, factor } from '@wails/go/models'

export type KLine = factor.KLine
export type CozeResult = coze.CozeStructuredResult
export type CozeScenario = coze.CozeScenario

export interface CozeStatusEvent {
  status: string
  message?: string
  symbol: string
  interval: string
  count: number
}
