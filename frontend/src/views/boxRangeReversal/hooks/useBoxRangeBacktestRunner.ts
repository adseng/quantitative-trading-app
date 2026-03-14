import { useCallback, useState } from 'react'

import { runBoxRangeBacktest } from '@/api/wails'
import type { BoxRangeBacktestReport, RunBoxRangeBacktestRequest } from '@/views/cozePredict/types'

export function useBoxRangeBacktestRunner() {
  const [running, setRunning] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [report, setReport] = useState<BoxRangeBacktestReport | null>(null)

  const execute = useCallback(async (req: RunBoxRangeBacktestRequest) => {
    setRunning(true)
    setError(null)
    try {
      const nextReport = await runBoxRangeBacktest(req)
      setReport(nextReport)
      return nextReport
    } catch (err: any) {
      const message = err?.message || '箱体震荡反转回测失败'
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
