import { useCallback, useState } from 'react'

import { runEMABacktest } from '@/api/wails'
import type { EMABacktestReport, RunEMABacktestRequest } from '@/views/cozePredict/types'

export function useEMABacktestRunner() {
  const [running, setRunning] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [report, setReport] = useState<EMABacktestReport | null>(null)

  const execute = useCallback(async (req: RunEMABacktestRequest) => {
    setRunning(true)
    setError(null)
    try {
      const nextReport = await runEMABacktest(req)
      setReport(nextReport)
      return nextReport
    } catch (err: any) {
      const message = err?.message || 'EMA 回测失败'
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
