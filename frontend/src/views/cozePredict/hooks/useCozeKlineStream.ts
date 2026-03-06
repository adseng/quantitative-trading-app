import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { message } from 'antd'
import { EventsOn } from '@wails/runtime/runtime'
import { FetchKlines, StartKlineStream, StopKlineStream } from '@wails/go/main/App'
import { factor } from '@wails/go/models'

import type { KLine, KlineErrorEvent, KlineSnapshotEvent, KlineStatusEvent, KlineUpdateEvent } from '../types'

const defaultLimit = 1000
const maxLimit = 1500

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

export function useCozeKlineStream(symbol: string, interval: string, limit: number) {
  const normalizedLimit = useMemo(() => normalizeLimit(limit), [limit])

  const [klines, setKlines] = useState<KLine[]>([])
  const [loading, setLoading] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [streaming, setStreaming] = useState(false)
  const [streamStatus, setStreamStatus] = useState<string | null>(null)
  const [lastUpdate, setLastUpdate] = useState<string | null>(null)

  const requestIdRef = useRef(0)
  const streamIntentRef = useRef(false)
  const sessionIdRef = useRef<string | null>(null)
  const symbolRef = useRef(symbol)
  const intervalRef = useRef(interval)

  symbolRef.current = symbol
  intervalRef.current = interval

  const matchesCurrentMarket = useCallback((payload: { symbol?: string; interval?: string } | null | undefined) => {
    if (!payload) return false
    return payload.symbol === symbolRef.current && payload.interval === intervalRef.current
  }, [])

  const shouldAcceptStreamEvent = useCallback((payload: { sessionId?: string; symbol?: string; interval?: string } | null | undefined) => {
    if (!matchesCurrentMarket(payload)) return false
    const sessionId = payload?.sessionId
    if (!sessionIdRef.current) {
      if (streamIntentRef.current && sessionId) {
        sessionIdRef.current = sessionId
        return true
      }
      return !streamIntentRef.current
    }
    return sessionId === sessionIdRef.current
  }, [matchesCurrentMarket])

  const mergeKline = useCallback((nextKline: KLine) => {
    setKlines((prev) => {
      const idx = prev.findIndex((item) => item.openTime === nextKline.openTime)
      if (idx >= 0) {
        const next = [...prev]
        next[idx] = nextKline
        return next
      }
      const next = [...prev, nextKline].sort((a, b) => a.openTime - b.openTime)
      return next.length > normalizedLimit ? next.slice(-normalizedLimit) : next
    })
    setLastUpdate(nowText())
    if (streamIntentRef.current) setStreamStatus('polling')
  }, [normalizedLimit])

  const loadKlines = useCallback(async () => {
    const requestId = ++requestIdRef.current
    setLoading(true)
    setLoadError(null)

    try {
      const data = await FetchKlines(symbol, interval, normalizedLimit)
      if (requestId !== requestIdRef.current) return
      const list = Array.isArray(data) ? data.map((item: unknown) => createKLine(item)) : []
      setKlines(list)
      setLastUpdate(nowText())
    } catch (error: any) {
      if (requestId !== requestIdRef.current) return
      const errMsg = error?.message || '加载 K 线失败'
      setLoadError(errMsg)
    } finally {
      if (requestId === requestIdRef.current) {
        setLoading(false)
      }
    }
  }, [interval, normalizedLimit, symbol])

  useEffect(() => {
    void loadKlines()
  }, [loadKlines])

  useEffect(() => {
    const unsubSnapshot = EventsOn('kline:snapshot', (event: KlineSnapshotEvent) => {
      if (!shouldAcceptStreamEvent(event)) return
      const list = Array.isArray(event?.klines) ? event.klines.map((item: unknown) => createKLine(item)) : []
      setKlines(list)
      setLastUpdate(nowText())
      if (streamIntentRef.current) {
        setStreaming(true)
        setStreamStatus('polling')
      }
    })

    const unsubUpdate = EventsOn('kline:update', (event: KlineUpdateEvent) => {
      if (!shouldAcceptStreamEvent(event) || !event?.kline) return
      mergeKline(createKLine(event.kline))
      if (streamIntentRef.current) setStreaming(true)
    })

    const unsubStatus = EventsOn('kline:status', (event: KlineStatusEvent) => {
      if (!matchesCurrentMarket(event)) return
      if (event?.status === 'polling') {
        sessionIdRef.current = event.sessionId || sessionIdRef.current
        setStreaming(true)
      } else if (event?.status === 'stopped') {
        if (sessionIdRef.current && event?.sessionId && event.sessionId !== sessionIdRef.current) return
        sessionIdRef.current = null
        streamIntentRef.current = false
        setStreaming(false)
      }
      setStreamStatus(event?.status ?? null)
    })

    const unsubError = EventsOn('kline:error', (event: KlineErrorEvent) => {
      if (!shouldAcceptStreamEvent(event)) return
      const errMsg = event?.error || '实时流异常'
      setStreamStatus(event?.retryable ? 'retrying' : 'error')
      message.warning(errMsg)
      if (!event?.retryable) {
        streamIntentRef.current = false
        sessionIdRef.current = null
        setStreaming(false)
      }
    })

    return () => {
      unsubSnapshot?.()
      unsubUpdate?.()
      unsubStatus?.()
      unsubError?.()
    }
  }, [matchesCurrentMarket, mergeKline, shouldAcceptStreamEvent])

  useEffect(() => {
    if (!streamIntentRef.current) return
    setStreamStatus('starting')
    void StartKlineStream(symbol, interval, normalizedLimit).catch((error: any) => {
      const errMsg = error?.message || '启动实时流失败'
      streamIntentRef.current = false
      sessionIdRef.current = null
      setStreaming(false)
      setStreamStatus('error')
      message.error(errMsg)
    })
  }, [interval, normalizedLimit, symbol])

  useEffect(() => {
    return () => {
      if (!streamIntentRef.current) return
      streamIntentRef.current = false
      void StopKlineStream()
    }
  }, [])

  const startStream = useCallback(async () => {
    streamIntentRef.current = true
    setStreamStatus('starting')
    try {
      await StartKlineStream(symbolRef.current, intervalRef.current, normalizeLimit(normalizedLimit))
    } catch (error: any) {
      const errMsg = error?.message || '启动实时流失败'
      streamIntentRef.current = false
      sessionIdRef.current = null
      setStreaming(false)
      setStreamStatus('error')
      message.error(errMsg)
    }
  }, [normalizedLimit])

  const stopStream = useCallback(async () => {
    streamIntentRef.current = false
    sessionIdRef.current = null
    setStreaming(false)
    setStreamStatus(null)
    await StopKlineStream()
  }, [])

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
