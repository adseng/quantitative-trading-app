import { useCallback, useState } from 'react'

import { runBacktest } from '@/api/wails'
import type { BacktestReport, RunBacktestRequest } from '../types'

export function useBacktestRunner() {
  const [running, setRunning] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [report, setReport] = useState<BacktestReport | null>(null)

  const execute = useCallback(async (req: RunBacktestRequest) => {
    setRunning(true)
    setError(null)
    try {
      const nextReport = await runBacktest(req)
      setReport(nextReport)
      return nextReport
    } catch (err: any) {
      const message = err?.message || '回测失败'
      setError(message)
      throw err
    } finally {
      setRunning(false)
    }
  }, [])

  return {
    running,
    error,
    report,
    execute,
    setReport,
  }
}
