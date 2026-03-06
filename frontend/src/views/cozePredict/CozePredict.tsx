import { useEffect, useMemo, useState } from 'react'
import { Alert, Button, Card, Input, InputNumber, Select } from 'antd'

import { KlineChart } from './components/KlineChart'
import { PredictionResults } from './components/PredictionResults'
import { useCozePrediction } from './hooks/useCozePrediction'
import { useCozeKlineStream } from './hooks/useCozeKlineStream'

function normalizeVisibleBars(value: number | null): number {
  if (value == null) return 100
  return Math.min(500, Math.max(10, Number(value)))
}

function normalizePredictInterval(value: number | null): number {
  if (value == null) return 2
  return Math.min(60, Math.max(1, Number(value)))
}

export default function CozePredict() {
  const [symbol, setSymbol] = useState('BTCUSDT')
  const [interval, setInterval] = useState('15m')
  const [limit, setLimit] = useState(1000)
  const [visibleBars, setVisibleBars] = useState(100)
  const [cozeKlineCount, setCozeKlineCount] = useState(20)
  const [cozeIntervalMinutes, setCozeIntervalMinutes] = useState(2)

  const {
    klines,
    loading,
    loadError,
    streaming,
    streamStatus,
    lastUpdate,
    limitNum,
    startStream,
    stopStream,
  } = useCozeKlineStream(symbol, interval, limit)

  const {
    results,
    predicting,
    cozeStatus,
    cozePreview,
    cozeKlineCount: normalizedCozeKlineCount,
    triggerPredict,
  } = useCozePrediction(symbol, interval, cozeKlineCount)

  const streamActive = streaming || streamStatus === 'starting' || streamStatus === 'retrying'
  const summaryText = useMemo(() => {
    const parts = [
      `${klines.length} 根`,
      `加载: ${loading ? '进行中' : '完成'}`,
      `轮询: ${streamStatus ?? '-'}`,
      `最后更新: ${lastUpdate ?? '-'}`,
      `Coze: ${cozeStatus ?? '-'}`,
    ]
    if (streaming) parts.push(`定时: 每 ${cozeIntervalMinutes} 分钟`)
    return parts.join(' · ')
  }, [cozeIntervalMinutes, cozeStatus, klines.length, lastUpdate, loading, streamStatus, streaming])

  useEffect(() => {
    if (!streaming) return
    const intervalMs = normalizePredictInterval(cozeIntervalMinutes) * 60 * 1000
    const timer = window.setInterval(() => {
      void triggerPredict()
    }, intervalMs)
    return () => window.clearInterval(timer)
  }, [cozeIntervalMinutes, streaming, triggerPredict])

  return (
    <div className="max-w-7xl mx-auto p-4 space-y-4">
      <h1 className="text-xl font-medium text-[#242f57]">Coze K 线预测</h1>

      <Card title="K 线图">
        <div className="flex flex-wrap items-center gap-3 mb-3">
          <span className="text-sm text-gray-600">交易对</span>
          <Input
            value={symbol}
            onChange={(event) => setSymbol((event.target.value || '').trim().toUpperCase() || 'BTCUSDT')}
            placeholder="BTCUSDT"
            style={{ width: 140 }}
          />
          <span className="text-sm text-gray-600">周期</span>
          <Select
            value={interval}
            onChange={setInterval}
            options={[
              { value: '1m', label: '1 分钟' },
              { value: '5m', label: '5 分钟' },
              { value: '15m', label: '15 分钟' },
              { value: '30m', label: '30 分钟' },
              { value: '1h', label: '1 小时' },
              { value: '4h', label: '4 小时' },
              { value: '1d', label: '1 天' },
            ]}
            style={{ width: 120 }}
          />
          <span className="text-sm text-gray-600">数量</span>
          <InputNumber
            min={1}
            max={1500}
            value={limitNum}
            onChange={(value) => setLimit(value == null ? 1000 : Number(value))}
            style={{ width: 100 }}
          />
          <span className="text-sm text-gray-600">时间窗(根)</span>
          <InputNumber
            min={10}
            max={500}
            value={visibleBars}
            onChange={(value) => setVisibleBars(normalizeVisibleBars(value))}
            style={{ width: 88 }}
          />
          <Button type="primary" onClick={() => void startStream()} disabled={streamActive}>
            启动实时流
          </Button>
          <Button onClick={() => void stopStream()} disabled={!streamActive}>
            停止
          </Button>
          <span className="text-sm text-gray-600">定时预测间隔(分钟)</span>
          <InputNumber
            min={1}
            max={60}
            value={cozeIntervalMinutes}
            onChange={(value) => setCozeIntervalMinutes(normalizePredictInterval(value))}
            style={{ width: 88 }}
          />
          <span className="text-sm text-gray-500 self-center">{summaryText}</span>
        </div>

        {loadError ? <Alert type="error" showIcon message={loadError} className="mb-3" /> : null}
        <KlineChart klines={klines} streaming={streaming} visibleBars={visibleBars} />
      </Card>

      <Card title="Coze 预测" style={{ marginTop: 16 }}>
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-sm text-gray-600">给豆包的数据量</span>
          <InputNumber
            min={5}
            max={500}
            value={normalizedCozeKlineCount}
            onChange={(value) => setCozeKlineCount(value == null ? 20 : Number(value))}
            style={{ width: 88 }}
          />
          <Button onClick={() => void triggerPredict()} loading={predicting} disabled={klines.length < 5}>
            手动预测
          </Button>
          <span className="text-sm text-gray-600">
            当前状态: <span className="font-medium">{cozeStatus ?? '-'}</span>
          </span>
          {cozeStatus === '请求中...' && cozePreview ? (
            <span className="text-sm text-gray-500 truncate max-w-[720px]">{cozePreview}</span>
          ) : null}
        </div>
      </Card>

      <Card title="预测结果 (最近 20 次)">
        <PredictionResults results={results} />
      </Card>
    </div>
  )
}
