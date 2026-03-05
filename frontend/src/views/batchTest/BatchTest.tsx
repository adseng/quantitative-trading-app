import {
  StartBatchTest,
  StopBatchTest,
  GetBatchTestProgress,
  FetchKlines,
  BacktestEmotionV2,
  LogBacktestResult,
} from '../../../wailsjs/go/main/App'
import { factor } from '../../../wailsjs/go/models'
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'
import { useEffect, useRef, useState } from 'react'
import { Button, Card, Checkbox, Input, InputNumber, message, Progress, Select, Table, Tag, Typography } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import * as echarts from 'echarts'

const { Title, Text } = Typography

type KLine = factor.KLine
type BacktestResult = factor.BacktestResult
type BacktestResultSummary = factor.BacktestResultSummary

const INTERVAL_OPTIONS = [
  { value: '1m', label: '1 分钟' },
  { value: '5m', label: '5 分钟' },
  { value: '15m', label: '15 分钟' },
  { value: '1h', label: '1 小时' },
]

function klineToCandlestickData(k: KLine) {
  return [k.open, k.close, k.low, k.high]
}

function dirText(v: number) {
  return v === 1 ? '涨' : v === -1 ? '跌' : '平'
}

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

const batchColumns: ColumnsType<TestResultRow> = [
  { title: '#', width: 45, render: (_, r) => r.testCase.id },
  { title: '名称', dataIndex: ['testCase', 'name'], width: 150, ellipsis: true },
  {
    title: '因子',
    width: 200,
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
      return bg ? (
        <span style={{ background: bg, color: '#fff', padding: '2px 8px', borderRadius: 4, fontWeight: 600 }}>{pct}</span>
      ) : (
        <span>{pct}</span>
      )
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

const fmt = (v: number) => (v != null ? v.toFixed(2) : '-')

const singleBacktestColumns = [
  { title: '序号', dataIndex: 'index', key: 'index', width: 56 },
  { title: '时间', dataIndex: 'openTime', key: 'openTime', width: 150, render: (t: number) => new Date(t).toLocaleString('zh-CN') },
  { title: '实际', dataIndex: 'actual', key: 'actual', width: 44, render: dirText },
  { title: '预测', dataIndex: 'predicted', key: 'predicted', width: 44, render: dirText },
  { title: '正确', dataIndex: 'correct', key: 'correct', width: 44, render: (v: boolean) => (v ? '✓' : '✗') },
  { title: '均线', key: 'ma', width: 60, render: (_: unknown, r: BacktestResult) => r.factors ? fmt(r.factors.maScore) : '-' },
  { title: '趋势', key: 'trend', width: 60, render: (_: unknown, r: BacktestResult) => r.factors ? fmt(r.factors.trendContrib) : '-' },
  { title: 'RSI', key: 'rsi', width: 60, render: (_: unknown, r: BacktestResult) => r.factors ? fmt(r.factors.rsiContrib) : '-' },
  { title: 'MACD', key: 'macd', width: 60, render: (_: unknown, r: BacktestResult) => r.factors ? fmt(r.factors.macdContrib) : '-' },
  { title: 'Boll', key: 'boll', width: 60, render: (_: unknown, r: BacktestResult) => r.factors ? fmt(r.factors.bollContrib) : '-' },
  { title: '突破', key: 'brk', width: 60, render: (_: unknown, r: BacktestResult) => r.factors ? fmt(r.factors.breakoutContrib) : '-' },
  { title: '价MA', key: 'pma', width: 60, render: (_: unknown, r: BacktestResult) => r.factors ? fmt(r.factors.priceVsMAContrib) : '-' },
  { title: 'ATR', key: 'atr', width: 60, render: (_: unknown, r: BacktestResult) => r.factors ? fmt(r.factors.atrContrib) : '-' },
  { title: '量价', key: 'vol', width: 60, render: (_: unknown, r: BacktestResult) => r.factors ? fmt(r.factors.volumeContrib) : '-' },
  { title: '时段', key: 'ses', width: 60, render: (_: unknown, r: BacktestResult) => r.factors ? fmt(r.factors.sessionContrib) : '-' },
  { title: '金叉', key: 'macross', width: 60, render: (_: unknown, r: BacktestResult) => r.factors && 'macrossContrib' in r.factors ? fmt((r.factors as Record<string, number>).macrossContrib) : '-' },
  {
    title: '信号分数',
    key: 'score',
    width: 80,
    render: (_: unknown, r: BacktestResult) => (r.factors ? fmt((r.factors.bullScore ?? 0) - (r.factors.bearScore ?? 0)) : '-'),
  },
]

const F = ({
  label,
  checked,
  onChange,
  children,
}: {
  label: string
  checked: boolean
  onChange: (v: boolean) => void
  children: React.ReactNode
}) => (
  <div className="flex items-center gap-2 flex-wrap">
    <Checkbox checked={checked} onChange={(e) => onChange(e.target.checked)} />
    <span className="text-sm font-medium w-24">{label}</span>
    {children}
  </div>
)

const W = ({ value, onChange, disabled }: { value: number; onChange: (v: number) => void; disabled: boolean }) => (
  <>
    <InputNumber value={value} onChange={(v) => onChange(v ?? 0)} min={-5} max={5} step={0.1} disabled={disabled} style={{ width: 64 }} />
    <span className="text-xs text-gray-500">权重</span>
  </>
)

export default function BatchTest() {
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstance = useRef<echarts.ECharts | null>(null)

  const [symbol, setSymbol] = useState('BTCUSDT')
  const [interval, setInterval] = useState('15m')
  const [limit, setLimit] = useState(100)
  const [klines, setKlines] = useState<KLine[]>([])
  const [fetching, setFetching] = useState(false)
  const [backtestResult, setBacktestResult] = useState<BacktestResultSummary | null>(null)
  const [backtesting, setBacktesting] = useState(false)

  const [useMA, setUseMA] = useState(false)
  const [maShort, setMaShort] = useState(5)
  const [maLong, setMaLong] = useState(20)
  const [maWeight, setMaWeight] = useState(-1)
  const [useTrend, setUseTrend] = useState(false)
  const [trendN, setTrendN] = useState(10)
  const [trendWeight, setTrendWeight] = useState(-1)
  const [useRSI, setUseRSI] = useState(false)
  const [rsiPeriod, setRsiPeriod] = useState(7)
  const [rsiOverbought, setRsiOverbought] = useState(75)
  const [rsiOversold, setRsiOversold] = useState(30)
  const [rsiWeight, setRsiWeight] = useState(1)
  const [useMACD, setUseMACD] = useState(false)
  const [macdFast, setMacdFast] = useState(12)
  const [macdSlow, setMacdSlow] = useState(26)
  const [macdSignal, setMacdSignal] = useState(9)
  const [macdWeight, setMacdWeight] = useState(-1)
  const [useBoll, setUseBoll] = useState(true)
  const [bollPeriod, setBollPeriod] = useState(13)
  const [bollMultiplier, setBollMultiplier] = useState(2.4)
  const [bollWeight, setBollWeight] = useState(1)
  const [useBreakout, setUseBreakout] = useState(false)
  const [breakoutPeriod, setBreakoutPeriod] = useState(15)
  const [breakoutWeight, setBreakoutWeight] = useState(-1)
  const [usePriceVsMA, setUsePriceVsMA] = useState(false)
  const [priceVsMAPeriod, setPriceVsMAPeriod] = useState(20)
  const [priceVsMAWeight, setPriceVsMAWeight] = useState(-1)
  const [useATR, setUseATR] = useState(false)
  const [atrPeriod, setAtrPeriod] = useState(14)
  const [atrWeight, setAtrWeight] = useState(-1)
  const [useVolume, setUseVolume] = useState(false)
  const [volumePeriod, setVolumePeriod] = useState(20)
  const [volumeWeight, setVolumeWeight] = useState(-1)
  const [useSession, setUseSession] = useState(false)
  const [sessionWeight, setSessionWeight] = useState(-1)
  const [useMACross, setUseMACross] = useState(false)
  const [macrossShort, setMacrossShort] = useState(5)
  const [macrossLong, setMacrossLong] = useState(20)
  const [macrossWeight, setMacrossWeight] = useState(1)
  const [macrossWindow, setMacrossWindow] = useState(2)
  const [macrossPreempt, setMacrossPreempt] = useState(0)

  const [running, setRunning] = useState(false)
  const [phase, setPhase] = useState('')
  const [messageText, setMessageText] = useState('')
  const [progress, setProgress] = useState(0)
  const [total, setTotal] = useState(200)
  const [recent, setRecent] = useState<TestResultRow[]>([])
  const cleanupRef = useRef<(() => void) | null>(null)

  const handleFetchKlines = () => {
    setFetching(true)
    setBacktestResult(null)
    FetchKlines(symbol, interval, limit)
      .then((data) => {
        const list = Array.isArray(data) ? data.map((d: unknown) => factor.KLine.createFrom(d)) : []
        setKlines(list)
        message.success(`已获取 ${list.length} 根 K 线`)
      })
      .catch((err) => message.error(err?.message || '获取 K 线失败'))
      .finally(() => setFetching(false))
  }

  useEffect(() => {
    if (klines.length === 0 || !chartRef.current) return
    if (!chartInstance.current) {
      chartInstance.current = echarts.init(chartRef.current)
    }
    const times = klines.map((k) => new Date(k.openTime).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }))
    const option: echarts.EChartsOption = {
      grid: { left: '8%', right: '5%', top: '10%', bottom: '22%' },
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'cross' },
        formatter: (params: unknown) => {
          const p = (params as Array<{ dataIndex?: number }>)?.[0]
          if (p?.dataIndex == null || !klines[p.dataIndex]) return ''
          const k = klines[p.dataIndex]
          const { open, close, low, high } = k
          const change = open ? ((close - open) / open) * 100 : 0
          return [
            `<div style="font-weight:600">${new Date(k.openTime).toLocaleString('zh-CN')}</div>`,
            `开: ${(open ?? 0).toFixed(2)} &nbsp; 高: ${(high ?? 0).toFixed(2)}`,
            `收: ${(close ?? 0).toFixed(2)} &nbsp; 低: ${(low ?? 0).toFixed(2)}`,
            `涨跌: ${change >= 0 ? '+' : ''}${change.toFixed(2)}%`,
          ].join('<br/>')
        },
      },
      axisPointer: { link: [{ xAxisIndex: 'all' }], label: { backgroundColor: '#777' } },
      toolbox: {
        feature: {
          saveAsImage: { title: '保存图片' },
          dataZoom: { title: { zoom: '区域缩放', back: '还原' } },
          restore: { title: '还原' },
        },
        right: 12,
        top: 8,
      },
      dataZoom: [
        { type: 'inside', xAxisIndex: 0, start: 0, end: 100, zoomOnMouseWheel: true, moveOnMouseMove: true },
        { type: 'slider', xAxisIndex: 0, bottom: 8, start: 0, end: 100, height: 20 },
      ],
      xAxis: { type: 'category', data: times, axisLabel: { fontSize: 10 }, boundaryGap: true },
      yAxis: { type: 'value', scale: true, splitLine: { lineStyle: { type: 'dashed', opacity: 0.3 } } },
      series: [
        {
          type: 'candlestick',
          data: klines.map(klineToCandlestickData),
          itemStyle: { color: '#26a69a', color0: '#ef5350', borderColor: '#26a69a', borderColor0: '#ef5350' },
        },
      ],
    }
    chartInstance.current.setOption(option, true)
    chartInstance.current.resize()
  }, [klines])

  useEffect(() => {
    const chart = chartInstance.current
    const onResize = () => chart?.resize()
    window.addEventListener('resize', onResize)
    return () => {
      window.removeEventListener('resize', onResize)
      chart?.dispose()
      chartInstance.current = null
    }
  }, [])

  const handleBacktest = () => {
    const anyEnabled = useMA || useTrend || useRSI || useMACD || useBoll || useBreakout || usePriceVsMA || useATR || useVolume || useSession || useMACross
    if (!anyEnabled) {
      message.warning('请至少勾选一个指标因子')
      return
    }
    if (useMA && maShort >= maLong) {
      message.warning('均线短期周期应小于长期周期')
      return
    }
    if (useMACross && macrossShort >= macrossLong) {
      message.warning('金叉短期周期应小于长期周期')
      return
    }
    if (klines.length < 22) {
      message.warning('至少需要 22 根 K 线，请先获取 K 线')
      return
    }
    setBacktesting(true)
    BacktestEmotionV2(
      klines,
      useMA,
      maShort,
      maLong,
      maWeight,
      useTrend,
      trendN,
      trendWeight,
      useRSI,
      rsiPeriod,
      rsiOverbought,
      rsiOversold,
      rsiWeight,
      useMACD,
      macdFast,
      macdSlow,
      macdSignal,
      macdWeight,
      useBoll,
      bollPeriod,
      bollMultiplier,
      bollWeight,
      useBreakout,
      breakoutPeriod,
      breakoutWeight,
      usePriceVsMA,
      priceVsMAPeriod,
      priceVsMAWeight,
      useATR,
      atrPeriod,
      atrWeight,
      useVolume,
      volumePeriod,
      volumeWeight,
      useSession,
      sessionWeight,
      useMACross,
      macrossShort,
      macrossLong,
      macrossWeight,
      macrossWindow,
      macrossPreempt,
    )
      .then((result) => {
        const summary = factor.BacktestResultSummary.createFrom(result)
        setBacktestResult(summary)
        LogBacktestResult(
          symbol,
          interval,
          limit,
          useMA,
          useTrend,
          maShort,
          maLong,
          trendN,
          maWeight,
          trendWeight,
          summary.accuracy ?? 0,
          summary.correct ?? 0,
          summary.total ?? 0,
        ).catch((err) => message.warning('记录日志失败: ' + (err?.message || '未知错误')))
      })
      .catch((err) => message.error(err?.message || '回测失败'))
      .finally(() => setBacktesting(false))
  }

  useEffect(() => {
    GetBatchTestProgress().then((info: { nextIndex?: number; totalCases?: number; running?: boolean }) => {
      setProgress(info.nextIndex ?? 0)
      setTotal(info.totalCases ?? 200)
      if (info.running) {
        setRunning(true)
      } else if ((info.nextIndex ?? 0) > 0 && (info.nextIndex ?? 0) < (info.totalCases ?? 200)) {
        setPhase('stopped')
        setMessageText(`上次已完成 ${info.nextIndex}/${info.totalCases}，点击继续`)
      } else if ((info.nextIndex ?? 0) >= (info.totalCases ?? 200)) {
        setPhase('done')
        setMessageText('全部 200 组已完成')
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
    cleanupRef.current = cleanup

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
    <div className="max-w-6xl mx-auto p-4 space-y-4">
      <Title level={4}>批量回测 · 多因子探索</Title>

      <Card title="单策略回测 · 数据与因子配置">
        <div className="space-y-4">
          <div className="flex gap-2 flex-wrap items-end">
            <div>
              <label className="block mb-1 text-sm text-gray-600">交易对</label>
              <Input value={symbol} onChange={(e) => setSymbol(e.target.value)} placeholder="BTCUSDT" style={{ width: 120 }} />
            </div>
            <div>
              <label className="block mb-1 text-sm text-gray-600">周期</label>
              <Select value={interval} onChange={setInterval} options={INTERVAL_OPTIONS} style={{ width: 100 }} />
            </div>
            <div>
              <label className="block mb-1 text-sm text-gray-600">数量</label>
              <InputNumber value={limit} onChange={(v) => setLimit(v ?? 100)} min={22} max={1500} style={{ width: 80 }} />
            </div>
            <Button onClick={handleFetchKlines} loading={fetching}>
              获取 K 线
            </Button>
            <Text type="secondary">（单策略回测用 · 默认 Boll P13 M2.4）</Text>
          </div>

          {klines.length > 0 && (
            <div className="space-y-2">
              <Text strong>K 线: {symbol} · {interval} · {klines.length} 根</Text>
            </div>
          )}
        </div>
      </Card>

      {klines.length > 0 && (
        <Card title="K 线图">
          <div ref={chartRef} style={{ width: '100%', minHeight: 400, height: 400 }} />
        </Card>
      )}

      {klines.length >= 22 && (
        <Card title="因子配置 · 单策略回测">
          <div className="space-y-2 mb-4">
            <F label="均线因子" checked={useMA} onChange={setUseMA}>
              <span className="text-xs text-gray-500">短/长</span>
              <InputNumber value={maShort} onChange={(v) => setMaShort(v ?? 5)} min={1} max={99} disabled={!useMA} style={{ width: 52 }} />
              <InputNumber value={maLong} onChange={(v) => setMaLong(v ?? 20)} min={2} max={200} disabled={!useMA} style={{ width: 52 }} />
              <W value={maWeight} onChange={setMaWeight} disabled={!useMA} />
            </F>
            <F label="趋势因子" checked={useTrend} onChange={setUseTrend}>
              <span className="text-xs text-gray-500">N</span>
              <InputNumber value={trendN} onChange={(v) => setTrendN(v ?? 10)} min={1} max={99} disabled={!useTrend} style={{ width: 52 }} />
              <W value={trendWeight} onChange={setTrendWeight} disabled={!useTrend} />
            </F>
            <F label="RSI因子" checked={useRSI} onChange={setUseRSI}>
              <span className="text-xs text-gray-500">周期</span>
              <InputNumber value={rsiPeriod} onChange={(v) => setRsiPeriod(v ?? 14)} min={2} max={100} disabled={!useRSI} style={{ width: 52 }} />
              <span className="text-xs text-gray-500">超买</span>
              <InputNumber value={rsiOverbought} onChange={(v) => setRsiOverbought(v ?? 70)} min={50} max={100} disabled={!useRSI} style={{ width: 52 }} />
              <span className="text-xs text-gray-500">超卖</span>
              <InputNumber value={rsiOversold} onChange={(v) => setRsiOversold(v ?? 30)} min={0} max={50} disabled={!useRSI} style={{ width: 52 }} />
              <W value={rsiWeight} onChange={setRsiWeight} disabled={!useRSI} />
            </F>
            <F label="MACD因子" checked={useMACD} onChange={setUseMACD}>
              <span className="text-xs text-gray-500">快/慢/信号</span>
              <InputNumber value={macdFast} onChange={(v) => setMacdFast(v ?? 12)} min={1} max={99} disabled={!useMACD} style={{ width: 52 }} />
              <InputNumber value={macdSlow} onChange={(v) => setMacdSlow(v ?? 26)} min={2} max={200} disabled={!useMACD} style={{ width: 52 }} />
              <InputNumber value={macdSignal} onChange={(v) => setMacdSignal(v ?? 9)} min={1} max={99} disabled={!useMACD} style={{ width: 52 }} />
              <W value={macdWeight} onChange={setMacdWeight} disabled={!useMACD} />
            </F>
            <F label="布林带因子" checked={useBoll} onChange={setUseBoll}>
              <span className="text-xs text-gray-500">周期</span>
              <InputNumber value={bollPeriod} onChange={(v) => setBollPeriod(v ?? 20)} min={2} max={200} disabled={!useBoll} style={{ width: 52 }} />
              <span className="text-xs text-gray-500">倍数</span>
              <InputNumber value={bollMultiplier} onChange={(v) => setBollMultiplier(v ?? 2)} min={0.1} max={5} step={0.1} disabled={!useBoll} style={{ width: 64 }} />
              <W value={bollWeight} onChange={setBollWeight} disabled={!useBoll} />
            </F>
            <F label="突破因子" checked={useBreakout} onChange={setUseBreakout}>
              <span className="text-xs text-gray-500">周期</span>
              <InputNumber value={breakoutPeriod} onChange={(v) => setBreakoutPeriod(v ?? 20)} min={2} max={200} disabled={!useBreakout} style={{ width: 52 }} />
              <W value={breakoutWeight} onChange={setBreakoutWeight} disabled={!useBreakout} />
            </F>
            <F label="价格vs均线" checked={usePriceVsMA} onChange={setUsePriceVsMA}>
              <span className="text-xs text-gray-500">周期</span>
              <InputNumber value={priceVsMAPeriod} onChange={(v) => setPriceVsMAPeriod(v ?? 20)} min={2} max={200} disabled={!usePriceVsMA} style={{ width: 52 }} />
              <W value={priceVsMAWeight} onChange={setPriceVsMAWeight} disabled={!usePriceVsMA} />
            </F>
            <F label="ATR波动率" checked={useATR} onChange={setUseATR}>
              <span className="text-xs text-gray-500">周期</span>
              <InputNumber value={atrPeriod} onChange={(v) => setAtrPeriod(v ?? 14)} min={2} max={100} disabled={!useATR} style={{ width: 52 }} />
              <W value={atrWeight} onChange={setAtrWeight} disabled={!useATR} />
            </F>
            <F label="量价配合" checked={useVolume} onChange={setUseVolume}>
              <span className="text-xs text-gray-500">均量周期</span>
              <InputNumber value={volumePeriod} onChange={(v) => setVolumePeriod(v ?? 20)} min={2} max={200} disabled={!useVolume} style={{ width: 52 }} />
              <W value={volumeWeight} onChange={setVolumeWeight} disabled={!useVolume} />
            </F>
            <F label="时段因子" checked={useSession} onChange={setUseSession}>
              <W value={sessionWeight} onChange={setSessionWeight} disabled={!useSession} />
            </F>
            <F label="金叉/死叉" checked={useMACross} onChange={setUseMACross}>
              <span className="text-xs text-gray-500">短/长</span>
              <InputNumber value={macrossShort} onChange={(v) => setMacrossShort(v ?? 5)} min={2} max={50} disabled={!useMACross} style={{ width: 52 }} />
              <InputNumber value={macrossLong} onChange={(v) => setMacrossLong(v ?? 20)} min={3} max={250} disabled={!useMACross} style={{ width: 52 }} />
              <W value={macrossWeight} onChange={setMacrossWeight} disabled={!useMACross} />
              <span className="text-xs text-gray-500">容错</span>
              <InputNumber value={macrossWindow} onChange={(v) => setMacrossWindow(v ?? 0)} min={0} max={5} disabled={!useMACross} style={{ width: 52 }} title="左侧几根内有交叉仍算有效" />
              <span className="text-xs text-gray-500">预判</span>
              <InputNumber value={macrossPreempt} onChange={(v) => setMacrossPreempt(v ?? 0)} min={0} max={0.01} step={0.001} disabled={!useMACross} style={{ width: 64 }} title="0=关" />
            </F>
            <div className="pt-1 flex items-center gap-3">
              <Button type="primary" onClick={handleBacktest} loading={backtesting}>
                因子回测
              </Button>
              <span className="text-xs text-gray-400">记录到 backtest_log.xlsx</span>
            </div>
          </div>
          {backtestResult && (
            <div>
              <div className="mb-2 text-lg">
                正确率:{' '}
                <span className="font-medium text-blue-600">{((backtestResult.accuracy ?? 0) * 100).toFixed(1)}%</span>
                {' '}({backtestResult.correct}/{backtestResult.total})
              </div>
              <Table
                dataSource={backtestResult.results}
                columns={singleBacktestColumns}
                rowKey="index"
                pagination={{ pageSize: 20 }}
                size="small"
                scroll={{ x: 1200 }}
              />
            </div>
          )}
        </Card>
      )}

      <Card title="批量回测 · 200组参数">
        <div className="flex items-center justify-between mb-4">
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
            <Text type="secondary">
              {progress}/{total}
            </Text>
          </div>
          <Progress percent={pct} status={running ? 'active' : phase === 'done' ? 'success' : 'normal'} />
          <Text type="secondary" style={{ fontSize: 12 }}>
            {messageText}
          </Text>
        </div>
        <div className="mt-4">
          <Title level={5}>最近 200 条结果</Title>
          <Table
            dataSource={[...recent].reverse()}
            columns={batchColumns}
            rowKey={(r) => r.testCase.id}
            pagination={false}
            size="small"
            scroll={{ y: 500 }}
            rowClassName={(r) => (r.signalAccuracy > 0.55 ? 'bg-green-50' : '')}
          />
        </div>
      </Card>

      <Text type="secondary" style={{ fontSize: 12 }}>
        V4方案： Boll P13 M2.4 约 59.71% · MA/Trend/MACD/PMA 负权重 · RSI/Boll 正权重 · Break/ATR/Vol/Sess 负权重 · 信号正确率 &gt; 55% 标绿
      </Text>
    </div>
  )
}
