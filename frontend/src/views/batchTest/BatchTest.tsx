import { StartBatchTest, StopBatchTest, GetBatchTestProgress, GetBatchTestResults } from '../../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'
import { useEffect, useState } from 'react'
import { Button, Progress, Table, Tag, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'

const { Title, Text } = Typography

interface TestCaseInfo {
  id: number
  name: string
  useMA: boolean
  maShort: number
  maLong: number
  maWeight: number
  useTrend: boolean
  trendN: number
  trendWeight: number
  useRSI: boolean
  rsiPeriod: number
  rsiOverbought: number
  rsiOversold: number
  rsiWeight: number
  useMACD: boolean
  macdFast: number
  macdSlow: number
  macdSignal: number
  macdWeight: number
  useBoll: boolean
  bollPeriod: number
  bollMultiplier: number
  bollWeight: number
  useBreakout: boolean
  breakoutPeriod: number
  breakoutWeight: number
  usePriceVsMA: boolean
  priceVsMAPeriod: number
  priceVsMAWeight: number
  useATR: boolean
  atrPeriod: number
  atrWeight: number
  useVolume: boolean
  volumePeriod: number
  volumeWeight: number
  useSession: boolean
  sessionWeight: number
  useMACross?: boolean
}

interface TestResultRow {
  testCase: TestCaseInfo
  accuracy: number
  correct: number
  total: number
  signalCount: number
  signalAccuracy: number
  avgScore: number
  avgAbsScore: number
}

interface ProgressEvent {
  phase: 'fetching' | 'testing' | 'done' | 'stopped' | 'error'
  message: string
  caseIndex: number
  totalCases: number
  current?: TestResultRow
  recent?: TestResultRow[]
}

function accuracyColor(acc: number): string {
  if (acc >= 0.70) return '#1a9850'
  if (acc >= 0.65) return '#33a02c'
  if (acc >= 0.60) return '#66bd63'
  if (acc >= 0.575) return '#a6d96a'
  if (acc >= 0.55) return '#d9ef8b'
  return ''
}

const columns: ColumnsType<TestResultRow> = [
  { title: '#', width: 45, render: (_, r) => r.testCase.id },
  { title: '名称', dataIndex: ['testCase', 'name'], width: 150, ellipsis: true },
  {
    title: '因子',
    width: 180,
    render: (_, r) => {
      const tags: string[] = []
      if (r.testCase.useMA) tags.push('MA')
      if (r.testCase.useTrend) tags.push('Trend')
      if (r.testCase.useRSI) tags.push('RSI')
      if (r.testCase.useMACD) tags.push('MACD')
      if (r.testCase.useBoll) tags.push('Boll')
      if (r.testCase.useBreakout) tags.push('突破')
      if (r.testCase.usePriceVsMA) tags.push('价MA')
      if (r.testCase.useATR) tags.push('ATR')
      if (r.testCase.useVolume) tags.push('量价')
      if (r.testCase.useSession) tags.push('时段')
      if (r.testCase.useMACross) tags.push('金叉')
      return tags.map((t) => <Tag key={t} color="blue">{t}</Tag>)
    },
  },
  { title: '有效预测', dataIndex: 'signalCount', width: 80, align: 'right' },
  {
    title: '信号正确率',
    width: 100,
    align: 'right',
    render: (_, r) => {
      const pct = (r.signalAccuracy * 100).toFixed(1) + '%'
      const bg = accuracyColor(r.signalAccuracy)
      return bg
        ? <span style={{ background: bg, color: '#fff', padding: '2px 8px', borderRadius: 4, fontWeight: 600 }}>{pct}</span>
        : <span>{pct}</span>
    },
    sorter: (a, b) => a.signalAccuracy - b.signalAccuracy,
  },
  { title: '正确数', dataIndex: 'correct', width: 65, align: 'right' },
  { title: '总数', dataIndex: 'total', width: 65, align: 'right' },
  {
    title: '平均净分',
    width: 85,
    align: 'right',
    render: (_, r) => {
      const v = r.avgScore
      const color = v > 0 ? '#c41d7f' : v < 0 ? '#08979c' : '#999'
      return <span style={{ color, fontWeight: 500 }}>{v > 0 ? '+' : ''}{v.toFixed(3)}</span>
    },
    sorter: (a, b) => a.avgScore - b.avgScore,
  },
  {
    title: '信号强度',
    width: 80,
    align: 'right',
    render: (_, r) => {
      const v = r.avgAbsScore
      const color = v >= 2 ? '#531dab' : v >= 1 ? '#1d39c4' : '#999'
      return <span style={{ color, fontWeight: v >= 1 ? 600 : 400 }}>{v.toFixed(3)}</span>
    },
    sorter: (a, b) => a.avgAbsScore - b.avgAbsScore,
  },
]

export default function BatchTest() {
  const [running, setRunning] = useState(false)
  const [phase, setPhase] = useState('')
  const [messageText, setMessageText] = useState('')
  const [progress, setProgress] = useState(0)
  const [total, setTotal] = useState(200)
  const [recent, setRecent] = useState<TestResultRow[]>([])

  useEffect(() => {
    GetBatchTestProgress().then((info: { nextIndex?: number; totalCases?: number; running?: boolean }) => {
      const next = info.nextIndex ?? 0
      const tot = info.totalCases ?? 200
      setProgress(next)
      setTotal(tot)
      if (info.running) {
        setRunning(true)
      } else if (next > 0 && next < tot) {
        setPhase('stopped')
        setMessageText(`上次已完成 ${next}/${tot}，点击继续`)
        GetBatchTestResults(200).then((rows) => {
          if (Array.isArray(rows) && rows.length > 0) setRecent(rows as TestResultRow[])
        }).catch(() => {})
      } else if (next >= tot && tot > 0) {
        setPhase('done')
        setMessageText('全部 200 组已完成')
        GetBatchTestResults(200).then((rows) => {
          if (Array.isArray(rows) && rows.length > 0) setRecent(rows as TestResultRow[])
        }).catch(() => {})
      }
    })

    const cleanup = EventsOn('batch:progress', (evt: ProgressEvent) => {
      setPhase(evt.phase)
      setMessageText(evt.message)
      if (evt.caseIndex !== undefined) setProgress(evt.caseIndex)
      if (evt.totalCases) setTotal(evt.totalCases)
      if (evt.recent) setRecent(evt.recent)
      if (evt.phase === 'done' || evt.phase === 'stopped' || evt.phase === 'error') {
        setRunning(false)
      }
    })

    return () => {
      EventsOff('batch:progress')
    }
  }, [])

  const handleStart = async () => {
    const isResume = progress > 0 && phase !== 'done'
    setRunning(true)
    setPhase('fetching')
    setMessageText(isResume ? '继续中...' : '启动中...')
    if (!isResume) setRecent([])
    try {
      await StartBatchTest()
    } catch (e: unknown) {
      setMessageText('启动失败: ' + (e as Error).toString())
      setRunning(false)
    }
  }

  const handleStop = () => {
    StopBatchTest()
  }

  const pct = total > 0 ? Math.round((progress / total) * 100) : 0

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <Title level={4} style={{ margin: 0 }}>批量回测 · 200组参数(正/负权重探索)</Title>
        <div className="flex items-center gap-3">
          <Text type="secondary">BTCUSDT · 15m · 10万根K线</Text>
          {!running ? (
            <Button type="primary" size="large" onClick={handleStart}>
              {progress > 0 && phase !== 'done' ? '继续测试' : '开始测试'}
            </Button>
          ) : (
            <Button danger size="large" onClick={handleStop}>
              停止
            </Button>
          )}
        </div>
      </div>

      <div className="rounded-lg border border-gray-200 p-4 space-y-2">
        <div className="flex items-center justify-between">
          <Text strong>
            {phase === 'fetching' && '阶段: 获取K线数据'}
            {phase === 'testing' && '阶段: 回测中'}
            {phase === 'done' && '阶段: 全部完成'}
            {phase === 'stopped' && '阶段: 已暂停'}
            {phase === 'error' && '阶段: 出错'}
            {!phase && '等待开始'}
          </Text>
          <Text type="secondary">{progress}/{total}</Text>
        </div>
        <Progress percent={pct} status={running ? 'active' : phase === 'done' ? 'success' : 'normal'} />
        <Text type="secondary" style={{ fontSize: 12 }}>{messageText}</Text>
      </div>

      <div>
        <Title level={5}>最近 200 条结果</Title>
        <Table
          dataSource={[...recent].reverse()}
          columns={columns}
          rowKey={(r) => r.testCase.id}
          pagination={false}
          size="small"
          scroll={{ y: 500 }}
          rowClassName={(r) => r.signalAccuracy > 0.55 ? 'bg-green-50' : ''}
        />
      </div>

      <Text type="secondary" style={{ fontSize: 12 }}>
        V4方案：Boll P13 M2.4 约59.71% · MA/Trend/MACD/PMA用负权重 · RSI/Boll用正权重 · Break/ATR/Vol/Sess用负权重 · 信号正确率&gt;55% 标绿
      </Text>
    </div>
  )
}
