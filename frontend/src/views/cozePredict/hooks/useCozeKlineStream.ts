import { useCallback, useMemo, useState } from 'react'

import { loadLocalKlines } from '@/api/wails'
import type { KLine } from '../types'

const defaultLimit = 100000

function normalizeLimit(limit: number): number {
  const value = Number(limit) || defaultLimit
  return Math.max(10, value)
}

function nowText(): string {
  return new Date().toLocaleTimeString('zh-CN')
}

export function useCozeKlineStream(interval: string, limit: number) {
  const normalizedLimit = useMemo(() => normalizeLimit(limit), [limit])
  const [klines, setKlines] = useState<KLine[]>([])
  const [loading, setLoading] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [streamStatus, setStreamStatus] = useState<string | null>(null)
  const [lastUpdate, setLastUpdate] = useState<string | null>(null)

  const loadFromFile = useCallback(async (path: string) => {
    setLoading(true)
    setLoadError(null)
    try {
      const data = await loadLocalKlines(path)
      setKlines(data)
      setLastUpdate(nowText())
      setStreamStatus(`本地加载 ${data.length} 根`)
    } catch (error: any) {
      const errMsg = error?.message || '本地加载 K 线失败'
      setLoadError(errMsg)
    } finally {
      setLoading(false)
    }
  }, [])

  return {
    klines,
    loading,
    loadError,
    streamStatus,
    lastUpdate,
    limitNum: normalizedLimit,
    loadFromFile,
  }
}
