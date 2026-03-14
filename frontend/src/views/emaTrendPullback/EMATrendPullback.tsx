import { useMemo, useState } from 'react'
import { Alert, Button, Card, Descriptions, InputNumber, Select, Statistic } from 'antd'

import { KlineChart } from '@/views/cozePredict/components/KlineChart'
import { PredictionResults } from '@/views/cozePredict/components/PredictionResults'
import { useCozeKlineStream } from '@/views/cozePredict/hooks/useCozeKlineStream'
import type { EMATrendPullbackParams } from '@/views/cozePredict/types'
import { useEMABacktestRunner } from './hooks/useEMABacktestRunner'

const STRATEGY_NAME = 'ema-trend-pullback-confirmation'

function normalizeVisibleBars(value: number | null): number {
  if (value == null) return 100
  return Math.min(500, Math.max(10, Number(value)))
}

export default function EMATrendPullback() {
  const [interval, setInterval] = useState('15m')
  const [limit, setLimit] = useState(100000)
  const [visibleBars, setVisibleBars] = useState(100)
  const [initialBalance, setInitialBalance] = useState(10000)
  const [positionSizeUSDT, setPositionSizeUSDT] = useState(100)
  const [params, setParams] = useState<EMATrendPullbackParams>({
    fastPeriod: 20,
    slowPeriod: 60,
    breakoutLookback: 20,
    pullbackLookahead: 5,
    pullbackTolerancePercent: 0.003,
    atrPeriod: 14,
    stopATRMultiplier: 1,
    cooldownBars: 3,
    riskRewardRatio: 1.5,
  })

  const {
    klines,
    loading,
    loadError,
    streamStatus,
    lastUpdate,
    limitNum,
    loadFromFile,
  } = useCozeKlineStream(interval, limit)

  const { running, error: backtestError, report, execute } = useEMABacktestRunner()

  const summaryText = useMemo(() => {
    const parts = [
      `${klines.length} 根`,
      `加载: ${loading ? '进行中' : '完成'}`,
      `轮询: ${streamStatus ?? '-'}`,
      `最后更新: ${lastUpdate ?? '-'}`,
    ]
    return parts.join(' · ')
  }, [klines.length, lastUpdate, loading, streamStatus])

  const dataPath = useMemo(() => `docs/db/data-k-${interval}-${limitNum}.txt`, [interval, limitNum])
  const resultPath = useMemo(() => `docs/db/test-${STRATEGY_NAME}.txt`, [])
  const activeKlines = report?.klines?.length ? report.klines : klines

  const runBacktest = async () => {
    await execute({
      dataPath,
      strategyName: STRATEGY_NAME,
      resultPath,
      initialBalance,
      positionSizeUSDT,
      params,
    })
  }

  const updateParams = <K extends keyof EMATrendPullbackParams>(key: K, value: EMATrendPullbackParams[K]) => {
    setParams((current) => ({ ...current, [key]: value }))
  }

  return (
    <div className="max-w-7xl mx-auto p-4 space-y-4">
      <h1 className="text-xl font-medium text-[#242f57]">EMA 趋势回踩确认策略</h1>

      <Card title="数据源">
        <div className="space-y-3">
          <div className="flex flex-wrap items-center gap-3">
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
            <span className="text-sm text-gray-600">根数</span>
            <InputNumber min={10} value={limitNum} onChange={(value) => setLimit(value == null ? 100000 : Number(value))} style={{ width: 120 }} />
            <Button onClick={() => void loadFromFile(dataPath)} loading={loading}>
              加载本地文件
            </Button>
          </div>

          <div className="text-sm text-gray-600">
            数据文件: <span className="font-mono text-gray-800">{dataPath}</span>
          </div>
          <div className="text-sm text-gray-600">
            结果文件: <span className="font-mono text-gray-800">{resultPath}</span>
          </div>
        </div>
      </Card>

      <Card title="回测参数">
        <div className="flex flex-wrap items-center gap-3">
          <span className="text-sm text-gray-600">快 EMA</span>
          <InputNumber min={2} max={100} value={params.fastPeriod} onChange={(value) => updateParams('fastPeriod', value == null ? 20 : Number(value))} />
          <span className="text-sm text-gray-600">慢 EMA</span>
          <InputNumber min={5} max={300} value={params.slowPeriod} onChange={(value) => updateParams('slowPeriod', value == null ? 60 : Number(value))} />
          <span className="text-sm text-gray-600">突破窗口</span>
          <InputNumber min={2} max={200} value={params.breakoutLookback} onChange={(value) => updateParams('breakoutLookback', value == null ? 20 : Number(value))} />
          <span className="text-sm text-gray-600">回踩观察</span>
          <InputNumber min={1} max={20} value={params.pullbackLookahead} onChange={(value) => updateParams('pullbackLookahead', value == null ? 5 : Number(value))} />
          <span className="text-sm text-gray-600">回踩容差</span>
          <InputNumber min={0} max={0.02} step={0.0001} value={params.pullbackTolerancePercent} onChange={(value) => updateParams('pullbackTolerancePercent', value == null ? 0.003 : Number(value))} />
          <span className="text-sm text-gray-600">ATR 周期</span>
          <InputNumber min={2} max={100} value={params.atrPeriod} onChange={(value) => updateParams('atrPeriod', value == null ? 14 : Number(value))} />
          <span className="text-sm text-gray-600">止损 ATR 倍数</span>
          <InputNumber min={0.2} max={5} step={0.1} value={params.stopATRMultiplier} onChange={(value) => updateParams('stopATRMultiplier', value == null ? 1 : Number(value))} />
          <span className="text-sm text-gray-600">冷却 K 线</span>
          <InputNumber min={0} max={20} value={params.cooldownBars} onChange={(value) => updateParams('cooldownBars', value == null ? 3 : Number(value))} />
          <span className="text-sm text-gray-600">盈亏比</span>
          <InputNumber min={0.5} max={5} step={0.1} value={params.riskRewardRatio} onChange={(value) => updateParams('riskRewardRatio', value == null ? 1.5 : Number(value))} />
          <span className="text-sm text-gray-600">初始资金</span>
          <InputNumber min={100} step={100} value={initialBalance} onChange={(value) => setInitialBalance(value == null ? 10000 : Number(value))} />
          <span className="text-sm text-gray-600">单次下单</span>
          <InputNumber min={10} step={10} value={positionSizeUSDT} onChange={(value) => setPositionSizeUSDT(value == null ? 100 : Number(value))} />
          <span className="text-sm text-gray-600">时间窗(根)</span>
          <InputNumber
            min={10}
            max={500}
            value={visibleBars}
            onChange={(value) => setVisibleBars(normalizeVisibleBars(value))}
            style={{ width: 90 }}
          />
          <Button type="primary" onClick={() => void runBacktest()} loading={running}>
            运行回测
          </Button>
          <span className="text-sm text-gray-500 self-center">{summaryText}</span>
        </div>
        {backtestError ? <Alert type="error" showIcon message={backtestError} className="mt-3" /> : null}
      </Card>

      <Card title="K 线图">
        {loadError ? <Alert type="error" showIcon message={loadError} className="mb-3" /> : null}
        <KlineChart klines={activeKlines} streaming={false} visibleBars={visibleBars} signals={report?.signals} trades={report?.trades} />
      </Card>

      <div className="grid grid-cols-4 gap-4">
        <Card><Statistic title="信号数" value={report?.summary.totalSignals ?? 0} /></Card>
        <Card><Statistic title="成交单数" value={report?.summary.executedTrades ?? 0} /></Card>
        <Card><Statistic title="胜率" value={(report?.summary.winRate ?? 0) * 100} precision={2} suffix="%" /></Card>
        <Card><Statistic title="收益率" value={(report?.summary.roi ?? 0) * 100} precision={2} suffix="%" /></Card>
      </div>

      {report ? (
        <Card title="回测摘要">
          <Descriptions column={2} size="small">
            <Descriptions.Item label="策略">{report.strategyName}</Descriptions.Item>
            <Descriptions.Item label="生成时间">{report.generatedAt}</Descriptions.Item>
            <Descriptions.Item label="数据文件">{report.dataPath}</Descriptions.Item>
            <Descriptions.Item label="结果文件">{report.resultPath}</Descriptions.Item>
            <Descriptions.Item label="初始资金">{report.initialBalance.toFixed(2)} USDT</Descriptions.Item>
            <Descriptions.Item label="单次下单">{report.positionSizeUSDT.toFixed(2)} USDT</Descriptions.Item>
            <Descriptions.Item label="最终余额">{report.summary.finalBalance.toFixed(2)} USDT</Descriptions.Item>
            <Descriptions.Item label="总盈亏">{report.summary.totalPnL.toFixed(2)} USDT</Descriptions.Item>
            <Descriptions.Item label="最大回撤">{(report.summary.maxDrawdown * 100).toFixed(2)}%</Descriptions.Item>
            <Descriptions.Item label="跳过信号">{report.summary.skippedSignals}</Descriptions.Item>
          </Descriptions>
        </Card>
      ) : null}

      <Card title="订单结果">
        <PredictionResults results={report?.trades ?? []} />
      </Card>
    </div>
  )
}
