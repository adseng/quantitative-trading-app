import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { message } from 'antd'
import { EventsOn } from '@wails/runtime/runtime'
import { CozePredictStructured } from '@wails/go/main/App'
import { coze } from '@wails/go/models'

import type { CozeResult, CozeStatusEvent } from '../types'

function normalizeCount(count: number): number {
  const value = Number(count) || 20
  return Math.min(500, Math.max(5, value))
}

function buildPreview(result: CozeResult | null | undefined): string | null {
  if (!result) return null
  if (result.market_structure) return `结构: ${result.market_structure}`
  if (result.rawAnswer) return result.rawAnswer.replace(/\s+/g, ' ').slice(0, 120)
  return null
}

function formatStatus(event: CozeStatusEvent): string | null {
  if (!event?.status) return null
  if (event.status === 'requesting') return '请求中...'
  if (event.status === 'done') return '完成'
  if (event.status === 'error') return `失败: ${event.message || ''}`
  return event.status
}

export function useCozePrediction(symbol: string, interval: string, count: number) {
  const normalizedCount = useMemo(() => normalizeCount(count), [count])

  const [results, setResults] = useState<CozeResult[]>([])
  const [predicting, setPredicting] = useState(false)
  const [cozeStatus, setCozeStatus] = useState<string | null>(null)
  const [cozePreview, setCozePreview] = useState<string | null>(null)

  const predictingRef = useRef(false)
  const requestIdRef = useRef(0)
  const marketVersionRef = useRef(0)
  const mountedRef = useRef(true)
  const symbolRef = useRef(symbol)
  const intervalRef = useRef(interval)
  const countRef = useRef(normalizedCount)

  symbolRef.current = symbol
  intervalRef.current = interval
  countRef.current = normalizedCount

  const appendResult = useCallback((result: CozeResult) => {
    setResults((prev) => [result, ...prev].slice(0, 20))
    const preview = buildPreview(result)
    if (preview) setCozePreview(preview)
  }, [])

  useEffect(() => {
    marketVersionRef.current += 1
    requestIdRef.current += 1
    predictingRef.current = false
    setPredicting(false)
    setResults([])
    setCozeStatus(null)
    setCozePreview(null)
  }, [interval, symbol])

  useEffect(() => {
    const unsubscribe = EventsOn('coze:status', (event: CozeStatusEvent) => {
      if (!mountedRef.current) return
      if (!event || event.symbol !== symbolRef.current || event.interval !== intervalRef.current) return
      setCozeStatus(formatStatus(event))
    })
    return () => {
      mountedRef.current = false
      unsubscribe?.()
    }
  }, [])

  const triggerPredict = useCallback(async () => {
    if (predictingRef.current || !mountedRef.current) return null

    const requestId = ++requestIdRef.current
    const marketVersion = marketVersionRef.current
    predictingRef.current = true
    setPredicting(true)
    try {
      const result = coze.CozeStructuredResult.createFrom(
        await CozePredictStructured(symbolRef.current, intervalRef.current, countRef.current),
      )
      if (!mountedRef.current || requestId !== requestIdRef.current || marketVersion !== marketVersionRef.current) return null
      appendResult(result)
      setCozeStatus('完成')
      return result
    } catch (error: any) {
      if (!mountedRef.current || requestId !== requestIdRef.current || marketVersion !== marketVersionRef.current) return null
      const errMsg = error?.message || '预测失败'
      setCozeStatus(`失败: ${errMsg}`)
      message.error(errMsg)
      return null
    } finally {
      if (requestId === requestIdRef.current && marketVersion === marketVersionRef.current) {
        predictingRef.current = false
        if (mountedRef.current) setPredicting(false)
      }
    }
  }, [appendResult])

  return {
    results,
    predicting,
    cozeStatus,
    cozePreview,
    cozeKlineCount: normalizedCount,
    triggerPredict,
  }
}
