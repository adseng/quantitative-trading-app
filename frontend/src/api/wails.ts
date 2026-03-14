import type { BacktestReport, EMABacktestReport, KLine, RunBacktestRequest, RunEMABacktestRequest } from '@/views/cozePredict/types'

declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          LoadLocalKlines: (path: string) => Promise<KLine[]>
          RunBacktest: (req: RunBacktestRequest) => Promise<BacktestReport>
          RunEMABacktest: (req: RunEMABacktestRequest) => Promise<EMABacktestReport>
        }
      }
    }
  }
}

function appApi() {
  const app = window.go?.main?.App
  if (!app) {
    throw new Error('Wails runtime is not ready')
  }
  return app
}

export async function loadLocalKlines(path: string): Promise<KLine[]> {
  const data = await appApi().LoadLocalKlines(path)
  return Array.isArray(data) ? data : []
}

export async function runBacktest(req: RunBacktestRequest): Promise<BacktestReport> {
  return appApi().RunBacktest(req)
}

export async function runEMABacktest(req: RunEMABacktestRequest): Promise<EMABacktestReport> {
  return appApi().RunEMABacktest(req)
}
