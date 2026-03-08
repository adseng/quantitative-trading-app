import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { message } from 'antd'
import { FetchKlines } from '@wails/go/main/App'
import { factor } from '@wails/go/models'

import type { KLine } from '../types'

const defaultLimit = 1000
const maxLimit = 1500
const pollIntervalMs = 1000

function normalizeLimit(limit: number): number {
  const value = Number(limit) || defaultLimit
  return Math.min(maxLimit, Math.max(1, value))
}

function nowText(): string {
  return new Date().toLocaleTimeString('zh-CN')
}

function createKLine(input: unknown): KLine {
  return factor.KLine.createFrom(input)
}

function intervalToMs(interval: string): number {
  switch (interval) {
    case '1m':
      return 60 * 1000
    case '3m':
      return 3 * 60 * 1000
    case '5m':
      return 5 * 60 * 1000
    case '15m':
      return 15 * 60 * 1000
    case '30m':
      return 30 * 60 * 1000
    case '1h':
      return 60 * 60 * 1000
    case '4h':
      return 4 * 60 * 60 * 1000
    case '1d':
      return 24 * 60 * 60 * 1000
    default:
      return 15 * 60 * 1000
  }
}

export function useCozeKlineStream(symbol: string, interval: string, limit: number) {
  const normalizedLimit = useMemo(() => normalizeLimit(limit), [limit])

  const [klines, setKlines] = useState<KLine[]>([])
  const [loading, setLoading] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [streaming, setStreaming] = useState(false)
  const [streamStatus, setStreamStatus] = useState<string | null>(null)
  const [lastUpdate, setLastUpdate] = useState<string | null>(null)

  const requestIdRef = useRef(0)
  const pollRequestIdRef = useRef(0)
  const streamIntentRef = useRef(false)
  const pollTimerRef = useRef<number | null>(null)
  const pollInFlightRef = useRef(false)
  const marketVersionRef = useRef(0)
  const lastPollWarningRef = useRef<{ message: string; time: number } | null>(null)
  const mountedRef = useRef(true)
  const klinesRef = useRef<KLine[]>([])
  const symbolRef = useRef(symbol)
  const intervalRef = useRef(interval)
  const limitRef = useRef(normalizedLimit)

  klinesRef.current = klines
  symbolRef.current = symbol
  intervalRef.current = interval
  limitRef.current = normalizedLimit

  const clearPollTimer = useCallback(() => {
    if (pollTimerRef.current != null) {
      window.clearInterval(pollTimerRef.current)
      pollTimerRef.current = null
    }
  }, [])

  const mergeLatestKline = useCallback((nextKline: KLine, activeInterval: string, activeLimit: number) => {
    if (!mountedRef.current) return
    setKlines((prev) => {
      if (prev.length === 0) return [nextKline]
      const last = prev[prev.length - 1]
      if (last?.openTime === nextKline.openTime) {
        const next = [...prev]
        next[next.length - 1] = nextKline
        return next
      }
      const expectedStep = intervalToMs(activeInterval)
      if (last?.openTime != null && nextKline.openTime > last.openTime + expectedStep) {
        return prev
      }
      if (last?.openTime != null && nextKline.openTime > last.openTime) {
        const next = [...prev, nextKline]
        return next.length > activeLimit ? next.slice(-activeLimit) : next
      }
      const idx = prev.findIndex((item) => item.openTime === nextKline.openTime)
      if (idx >= 0) {
        const next = [...prev]
        next[idx] = nextKline
        return next
      }
      return prev
    })
    setLastUpdate(nowText())
    if (streamIntentRef.current) setStreamStatus('polling')
  }, [])

  const loadKlines = useCallback(async () => {
    const marketVersion = marketVersionRef.current
    const requestId = ++requestIdRef.current
    setLoading(true)
    setLoadError(null)

    try {
      const data = await FetchKlines(symbol, interval, normalizedLimit)
      if (requestId !== requestIdRef.current || marketVersion !== marketVersionRef.current || !mountedRef.current) return
      const list = Array.isArray(data) ? data.map((item: unknown) => createKLine(item)) : []
      setKlines(list)
      setLastUpdate(nowText())
    } catch (error: any) {
      if (requestId !== requestIdRef.current || marketVersion !== marketVersionRef.current || !mountedRef.current) return
      const errMsg = error?.message || '加载 K 线失败'
      setLoadError(errMsg)
    } finally {
      if (requestId === requestIdRef.current && marketVersion === marketVersionRef.current && mountedRef.current) {
        setLoading(false)
      }
    }
  }, [interval, normalizedLimit, symbol])

  const pollLatestKline = useCallback(async () => {
    if (!mountedRef.current || !streamIntentRef.current || pollInFlightRef.current) return

    pollInFlightRef.current = true
    const pollRequestId = ++pollRequestIdRef.current
    const marketVersion = marketVersionRef.current
    const activeSymbol = symbolRef.current
    const activeInterval = intervalRef.current
    const activeLimit = limitRef.current

    try {
      const data = await FetchKlines(activeSymbol, activeInterval, 1)
      if (
        !mountedRef.current ||
        !streamIntentRef.current ||
        pollRequestId !== pollRequestIdRef.current ||
        marketVersion !== marketVersionRef.current
      ) {
        return
      }
      const latest = Array.isArray(data) && data.length > 0 ? createKLine(data[0]) : null
      if (!latest) return

      const currentKlines = klinesRef.current
      const last = currentKlines[currentKlines.length - 1]
      const expectedStep = intervalToMs(activeInterval)
      const needsResync = !last || latest.openTime > last.openTime + expectedStep
      if (needsResync) {
        await loadKlines()
      } else {
        mergeLatestKline(latest, activeInterval, activeLimit)
      }
      setLoadError(null)
      setStreamStatus('polling')
    } catch (error: any) {
      if (
        !mountedRef.current ||
        !streamIntentRef.current ||
        pollRequestId !== pollRequestIdRef.current ||
        marketVersion !== marketVersionRef.current
      ) {
        return
      }
      const errMsg = error?.message || '实时轮询失败'
      setStreamStatus('retrying')
      const lastWarning = lastPollWarningRef.current
      const now = Date.now()
      if (!lastWarning || lastWarning.message !== errMsg || now - lastWarning.time > 5000) {
        message.warning(errMsg)
        lastPollWarningRef.current = { message: errMsg, time: now }
      }
    } finally {
      pollInFlightRef.current = false
    }
  }, [loadKlines, mergeLatestKline])

  useEffect(() => {
    marketVersionRef.current += 1
    requestIdRef.current += 1
    pollRequestIdRef.current += 1
    pollInFlightRef.current = false
    lastPollWarningRef.current = null
    void loadKlines()
    if (streamIntentRef.current) {
      setStreamStatus('polling')
      void pollLatestKline()
    }
  }, [interval, loadKlines, normalizedLimit, pollLatestKline, symbol])

  useEffect(() => {
    return () => {
      mountedRef.current = false
      requestIdRef.current++
      pollRequestIdRef.current++
      streamIntentRef.current = false
      clearPollTimer()
    }
  }, [clearPollTimer])

  const startStream = useCallback(async () => {
    if (streamIntentRef.current) return
    streamIntentRef.current = true
    setStreaming(true)
    setStreamStatus('polling')
    clearPollTimer()
    pollTimerRef.current = window.setInterval(() => {
      void pollLatestKline()
    }, pollIntervalMs)
    await pollLatestKline()
  }, [clearPollTimer, pollLatestKline])

  const stopStream = useCallback(async () => {
    streamIntentRef.current = false
    clearPollTimer()
    pollRequestIdRef.current++
    pollInFlightRef.current = false
    setStreaming(false)
    setStreamStatus(null)
  }, [clearPollTimer])

  return {
    klines,
    loading,
    loadError,
    streaming,
    streamStatus,
    lastUpdate,
    limitNum: normalizedLimit,
    reloadKlines: loadKlines,
    startStream,
    stopStream,
  }
}
