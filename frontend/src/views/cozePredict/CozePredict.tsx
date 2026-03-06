import { useCallback, useEffect, useRef, useState } from 'react'
import { Button, Card, Input, InputNumber, message, Select, Table, Tag } from 'antd'
import * as echarts from 'echarts'
import { EventsOn } from '@wails/runtime/runtime'
import { CozePredictStructured, FetchKlines, SetCozeKlineCount, StartKlineStream, StopKlineStream } from '@wails/go/main/App'
import { factor } from '@wails/go/models'

type KLine = factor.KLine

interface CozeScenario {
  direction: string
  probability: number
  setup_logic: string
  trigger_condition: string
  entry_price?: number | null
  stop_loss?: number | null
  take_profit_1?: number | null
  take_profit_2?: number | null
  risk_reward_ratio?: number | null
  action: string
}

interface CozeResult {
  timestamp: string
  symbol: string
  current_price: number
  market_structure: string
  scenarios: CozeScenario[]
  rawAnswer?: string
}

const upColor = '#0ECB81'
const downColor = '#F6465D'

/** 币安风格：价格千分位，保留小数 */
function formatPrice(v: number, digits: number = 2): string {
  const fixed = v.toFixed(digits)
  const [intPart, decPart] = fixed.split('.')
  const withComma = intPart.replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  return decPart != null ? `${withComma}.${decPart}` : withComma
}

/** 成交量简短格式：K / M */
function formatVol(v: number): string {
  if (v >= 1e6) return `${(v / 1e6).toFixed(2)}M`
  if (v >= 1e3) return `${(v / 1e3).toFixed(3)}K`
  return v.toFixed(2)
}

function calcMA(dayCount: number, data: KLine[]) {
  const result: (number | null)[] = []
  for (let i = 0; i < data.length; i++) {
    if (i < dayCount - 1) {
      result.push(null)
      continue
    }
    let sum = 0
    for (let j = 0; j < dayCount; j++) sum += data[i - j].close ?? 0
    result.push(Number((sum / dayCount).toFixed(2)))
  }
  return result
}

function calcVolMA(period: number, data: KLine[]) {
  const result: (number | null)[] = []
  for (let i = 0; i < data.length; i++) {
    if (i < period - 1) {
      result.push(null)
      continue
    }
    let sum = 0
    for (let j = 0; j < period; j++) sum += data[i - j].volume ?? 0
    result.push(Number((sum / period).toFixed(2)))
  }
  return result
}

export default function CozePredict() {
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstance = useRef<echarts.ECharts | null>(null)
  const renderRaf = useRef<number | null>(null)

  const [klines, setKlines] = useState<KLine[]>([])
  const [streaming, setStreaming] = useState(false)
  const [results, setResults] = useState<CozeResult[]>([])
  const [predicting, setPredicting] = useState(false)
  const [lastUpdate, setLastUpdate] = useState<string | null>(null)
  const [streamStatus, setStreamStatus] = useState<string | null>(null)
  const [cozeStatus, setCozeStatus] = useState<string | null>(null)
  const [cozePreview, setCozePreview] = useState<string | null>(null)
  const [symbol, setSymbol] = useState('BTCUSDT')
  const [interval, setInterval] = useState('15m')
  const [limit, setLimit] = useState(1000)
  const [cozeKlineCount, setCozeKlineCountState] = useState(50)
  /** 定时预测间隔（分钟），仅当实时流启动时生效 */
  const [cozeIntervalMinutes, setCozeIntervalMinutes] = useState(2)
  /** 当前时间窗（滑块）：显示的 K 线根数，默认 100 */
  const [visibleBars, setVisibleBars] = useState(100)

  const klinesRef = useRef<KLine[]>([])
  const predictingRef = useRef(false)
  const symbolRef = useRef(symbol)
  const cozeKlineCountRef = useRef(cozeKlineCount)
  const dataZoomRef = useRef<{ start: number; end: number } | null>(null)
  const chartUpdateTimeoutRef = useRef<number | null>(null)
  const lastChartUpdateRef = useRef(0)
  const fromThrottleRef = useRef(false)
  const [chartUpdateTrigger, setChartUpdateTrigger] = useState(0)
  klinesRef.current = klines
  predictingRef.current = predicting
  symbolRef.current = symbol
  cozeKlineCountRef.current = cozeKlineCount

  useEffect(() => {
    SetCozeKlineCount(cozeKlineCount)
  }, [cozeKlineCount])

  useEffect(() => {
    dataZoomRef.current = null
  }, [symbol, interval, limit, visibleBars])

  useEffect(() => {
    const count = Math.min(Math.max(Number(limit) || 1000, 1), 1500)
    FetchKlines(symbol, interval, count)
      .then((data) => {
        const list = Array.isArray(data) ? data.map((d: unknown) => factor.KLine.createFrom(d)) : []
        setKlines(list)
      })
      .catch(() => {})
  }, [symbol, interval, limit])

  const mergeKline = useCallback((kl: KLine) => {
    setKlines((prev) => {
      const k = factor.KLine.createFrom(kl)
      const idx = prev.findIndex((p) => p.openTime === k.openTime)
      let next: KLine[]
      if (idx >= 0) {
        next = [...prev]
        next[idx] = k
      } else {
        next = [...prev, k].sort((a, b) => a.openTime - b.openTime)
        const cap = Math.min(Math.max(Number(limit) || 1000, 1), 1500)
        if (next.length > cap) next = next.slice(-cap)
      }
      return next
    })
    setLastUpdate(new Date().toLocaleTimeString('zh-CN'))
  }, [limit])

  useEffect(() => {
    const unsubSnapshot = EventsOn('kline:snapshot', (data: { klines?: KLine[] } | unknown) => {
      const obj = data && typeof data === 'object' && !Array.isArray(data) ? data as { klines?: unknown[] } : {}
      const list = obj?.klines ?? []
      setKlines(list.map((d: unknown) => factor.KLine.createFrom(d)))
      setLastUpdate(new Date().toLocaleTimeString('zh-CN'))
    })
    const unsubUpdate = EventsOn('kline:update', (...args: unknown[]) => {
      const raw = args != null && args.length > 0 ? args[0] : null
      const payload = raw != null && typeof raw === 'object' ? (Array.isArray(raw) ? raw[0] : raw) : null
      if (payload) mergeKline(factor.KLine.createFrom(payload))
    })
    const unsubStatus = EventsOn('kline:status', (...args: unknown[]) => {
      const raw = args != null && args.length > 0 ? args[0] : null
      const payload = raw != null && typeof raw === 'object' ? (Array.isArray(raw) ? raw[0] : raw) : null
      const status = payload && typeof (payload as any).status === 'string' ? (payload as any).status : null
      if (status) {
        setStreamStatus(status)
        if (status === 'stopped') setStreaming(false)
      }
    })
    const unsubErr = EventsOn('kline:error', (...args: unknown[]) => {
      const raw = args != null && args.length > 0 ? args[0] : null
      const payload = raw != null && typeof raw === 'object' ? (Array.isArray(raw) ? raw[0] : raw) : null
      const errMsg = payload && typeof (payload as any).error === 'string' ? (payload as any).error : null
      if (errMsg) message.error(errMsg)
      setStreaming(false)
      setStreamStatus('error')
    })
    const unsubCoze = EventsOn('coze:result', (data: CozeResult) => {
      setResults((prev) => [data, ...prev].slice(0, 20))
      // 用一行做预览：优先 market_structure，其次 rawAnswer
      const preview =
        (data?.market_structure ? `结构: ${data.market_structure}` : '') ||
        (data?.rawAnswer ? data.rawAnswer.replace(/\s+/g, ' ').slice(0, 120) : '')
      if (preview) setCozePreview(preview)
    })
    const unsubCozeStatus = EventsOn('coze:status', (...args: unknown[]) => {
      const raw = args != null && args.length > 0 ? args[0] : null
      const payload = raw != null && typeof raw === 'object' ? (Array.isArray(raw) ? raw[0] : raw) : null
      const status = payload && typeof (payload as any).status === 'string' ? (payload as any).status : null
      const msg = payload && typeof (payload as any).message === 'string' ? (payload as any).message : null
      setCozeStatus(status === 'requesting' ? '请求中...' : status === 'error' ? `失败: ${msg || ''}` : status === 'done' ? '完成' : status ?? null)
    })
    return () => {
      unsubSnapshot?.()
      unsubUpdate?.()
      unsubStatus?.()
      unsubErr?.()
      unsubCoze?.()
      unsubCozeStatus?.()
    }
  }, [mergeKline])

  useEffect(() => {
    if (klines.length === 0 || !chartRef.current) return
    if (!chartInstance.current) chartInstance.current = echarts.init(chartRef.current)
    // 实时流时节流：避免每次推送都重绘，导致时间窗和 tooltip 被重置
    const throttleMs = 800
    if (streaming && !fromThrottleRef.current && lastChartUpdateRef.current && Date.now() - lastChartUpdateRef.current < throttleMs) {
      if (chartUpdateTimeoutRef.current != null) window.clearTimeout(chartUpdateTimeoutRef.current)
      chartUpdateTimeoutRef.current = window.setTimeout(() => {
        chartUpdateTimeoutRef.current = null
        fromThrottleRef.current = true
        setChartUpdateTrigger((t) => t + 1)
      }, throttleMs - (Date.now() - lastChartUpdateRef.current))
      return () => {
        if (chartUpdateTimeoutRef.current != null) {
          window.clearTimeout(chartUpdateTimeoutRef.current)
          chartUpdateTimeoutRef.current = null
        }
      }
    }
    fromThrottleRef.current = false
    if (renderRaf.current) cancelAnimationFrame(renderRaf.current)
    renderRaf.current = requestAnimationFrame(() => {
      const times = klines.map((k) => new Date(k.openTime).toISOString().slice(0, 16).replace('T', ' '))
      const values = klines.map((k) => [k.open, k.close, k.low, k.high])
      const volValues = klines.map((k) => k.volume ?? 0)
      const ma7 = calcMA(7, klines)
      const ma25 = calcMA(25, klines)
      const ma99 = calcMA(99, klines)
      const volMA5 = calcVolMA(5, klines)
      const volMA10 = calcVolMA(10, klines)

      const last = klines[klines.length - 1]
      const lastPrice = last?.close ?? 0
      const priceDigits = lastPrice < 1 ? 6 : lastPrice < 100 ? 4 : 2
      const lastOpen = last?.open ?? 0
      const lastChg = lastOpen ? ((lastPrice - lastOpen) / lastOpen) * 100 : 0
      const lastChgColor = lastPrice >= lastOpen ? upColor : downColor

      const lastMA7 = typeof ma7[ma7.length - 1] === 'number' ? (ma7[ma7.length - 1] as number) : null
      const lastMA25 = typeof ma25[ma25.length - 1] === 'number' ? (ma25[ma25.length - 1] as number) : null
      const lastMA99 = typeof ma99[ma99.length - 1] === 'number' ? (ma99[ma99.length - 1] as number) : null

      const bars = Math.max(1, visibleBars)
      const defaultZoomStart = klines.length <= bars ? 0 : 100 - (bars / klines.length) * 100
      const defaultZoomEnd = 100
      // 首次渲染或切换品种/周期后（ref 已清空）用默认时间窗（最后约 100 根）；否则保留用户拖动的时间窗
      const useDefaultZoom = lastChartUpdateRef.current === 0 || dataZoomRef.current == null
      let start = defaultZoomStart
      let end = defaultZoomEnd
      if (!useDefaultZoom) {
        const prevOpt = chartInstance.current?.getOption() as any
        const dz = Array.isArray(prevOpt?.dataZoom) ? prevOpt.dataZoom.find((z: any) => z.xAxisIndex != null) : null
        if (dz?.start != null && dz?.end != null) {
          start = dz.start
          end = dz.end
        } else if (dataZoomRef.current) {
          start = dataZoomRef.current.start
          end = dataZoomRef.current.end
        }
      }
      dataZoomRef.current = { start, end }
      const isFirstRender = lastChartUpdateRef.current === 0

      // 时间轴：仅显示 HH:mm
      const formatTimeLabel = (index: number) => {
        const t = klines[index]?.openTime
        if (t == null) return ''
        const d = new Date(t)
        const h = d.getUTCHours()
        const m = d.getUTCMinutes()
        return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}`
      }

      chartInstance.current?.setOption(
        {
          backgroundColor: '#0b0e11',
          animation: false,
          axisPointer: {
            link: [{ xAxisIndex: 'all' }],
            label: { backgroundColor: '#2b3139' },
          },
          tooltip: {
            trigger: 'axis',
            axisPointer: { type: 'cross' },
            backgroundColor: 'rgba(18, 22, 28, 0.92)',
            borderColor: '#2b3139',
            textStyle: { color: '#eaecef' },
            extraCssText: 'box-shadow: 0 2px 10px rgba(0,0,0,0.35);',
            formatter: (params: any[]) => {
              const p = params?.find((x) => x?.seriesType === 'candlestick') ?? params?.[0]
              const idx = p?.dataIndex
              if (idx == null || !klines[idx]) return ''
              const k = klines[idx]
              const open = k.open ?? 0
              const close = k.close ?? 0
              const high = k.high ?? 0
              const low = k.low ?? 0
              const vol = k.volume ?? 0
              const chg = open ? ((close - open) / open) * 100 : 0
              const amplitude = open ? ((high - low) / open) * 100 : 0
              const color = close >= open ? upColor : downColor
              const timeStr = times[idx].replace(/-/g, '/')
              const vMa7 = idx < ma7.length ? ma7[idx] : null
              const vMa25 = idx < ma25.length ? ma25[idx] : null
              const vMa99 = idx < ma99.length ? ma99[idx] : null
              const vVolMA5 = idx < volMA5.length ? volMA5[idx] : null
              const vVolMA10 = idx < volMA10.length ? volMA10[idx] : null
              const volBtc = vol
              const volUsdt = vol * close
              const lines: string[] = [
                `<div style="font-weight:600;margin-bottom:6px;">${timeStr}</div>`,
                `开 ${formatPrice(open)}  高 ${formatPrice(high)}`,
                `低 ${formatPrice(low)}  收 <span style="color:${color};font-weight:600">${formatPrice(close)}</span>`,
                `涨跌幅 <span style="color:${color};font-weight:600">${chg >= 0 ? '+' : ''}${chg.toFixed(2)}%</span>  振幅 ${amplitude.toFixed(2)}%`,
                `MA(7) <span style="color:#f0b90b">${vMa7 != null ? formatPrice(vMa7) : '-'}</span>  MA(25) <span style="color:#c994ff">${vMa25 != null ? formatPrice(vMa25) : '-'}</span>  MA(99) <span style="color:#a78bfa">${vMa99 != null ? formatPrice(vMa99) : '-'}</span>`,
                `Vol(BTC) ${formatVol(volBtc)}  Vol(USDT) ${formatVol(volUsdt)}`,
              ]
              if (vVolMA5 != null || vVolMA10 != null) {
                lines.push(`Vol MA5 ${formatVol(vVolMA5 ?? 0)}  Vol MA10 ${formatVol(vVolMA10 ?? 0)}`)
              }
              return lines.join('<br/>')
            },
          },
          grid: [
            // 顶部留出一行给 MA/价格信息（更像币安）
            { left: 54, right: 72, top: 44, height: 334 },
            { left: 54, right: 72, top: 392, height: 90 },
          ],
          xAxis: [
            {
              type: 'category',
              data: times,
              boundaryGap: true,
              axisLine: { lineStyle: { color: '#2b3139' } },
              axisLabel: {
                color: '#848e9c',
                hideOverlap: true,
                formatter: (value: string, index: number) => formatTimeLabel(index),
              },
              splitLine: { show: false },
              axisTick: { show: false },
              min: 'dataMin',
              max: 'dataMax',
            },
            {
              type: 'category',
              gridIndex: 1,
              data: times,
              boundaryGap: true,
              axisLine: { lineStyle: { color: '#2b3139' } },
              axisLabel: { show: false },
              splitLine: { show: false },
              axisTick: { show: false },
              min: 'dataMin',
              max: 'dataMax',
            },
          ],
          yAxis: [
            {
              scale: true,
              position: 'right',
              axisLine: { lineStyle: { color: '#2b3139' } },
              axisLabel: { color: '#848e9c', formatter: (v: number) => Number(v).toFixed(priceDigits) },
              splitLine: { lineStyle: { color: '#1e2329' } },
            },
            {
              gridIndex: 1,
              scale: true,
              position: 'right',
              axisLine: { lineStyle: { color: '#2b3139' } },
              axisLabel: { color: '#848e9c', formatter: (v: number) => (v >= 1e6 ? `${(v / 1e6).toFixed(1)}M` : v >= 1e3 ? `${(v / 1e3).toFixed(1)}K` : `${Math.round(v)}`) },
              splitLine: { show: false },
            },
          ],
          dataZoom: [
            { type: 'inside', xAxisIndex: [0, 1], start, end },
            {
              type: 'slider',
              xAxisIndex: [0, 1],
              bottom: 8,
              height: 22,
              start,
              end,
              brushSelect: false,
              fillerColor: 'rgba(132, 142, 156, 0.15)',
              borderColor: '#2b3139',
              handleStyle: { color: '#2b3139' },
              textStyle: { color: '#848e9c' },
            },
          ],
          graphic: [
            {
              type: 'text',
              left: 'center',
              top: 150,
              style: {
                text: 'BINANCE',
                fontSize: 42,
                fontWeight: 700,
                fill: 'rgba(132, 142, 156, 0.08)',
              },
              silent: true,
            },
            // 左上角：与币安一致 - 最后一根 K 的开高收低、涨跌幅、振幅、MA
            {
              type: 'group',
              left: 54,
              top: 14,
              silent: true,
              children: [
                {
                  type: 'text',
                  left: 0,
                  top: 0,
                  style: {
                    text: `开 ${formatPrice(lastOpen)}  高 ${formatPrice(last?.high ?? 0)}  低 ${formatPrice(last?.low ?? 0)}  收 ${formatPrice(lastPrice)}`,
                    fill: '#eaecef',
                    fontSize: 12,
                    fontWeight: 500,
                  },
                },
                {
                  type: 'text',
                  left: 0,
                  top: 18,
                  style: {
                    text: `涨跌幅 ${lastChg >= 0 ? '+' : ''}${lastChg.toFixed(2)}%  振幅 ${lastOpen ? (((last?.high ?? 0) - (last?.low ?? 0)) / lastOpen * 100).toFixed(2) : '0'}%`,
                    fill: lastChgColor,
                    fontSize: 12,
                    fontWeight: 500,
                  },
                },
                {
                  type: 'text',
                  left: 0,
                  top: 36,
                  style: {
                    text: `MA(7) ${lastMA7 != null ? formatPrice(lastMA7) : '-'}`,
                    fill: '#f0b90b',
                    fontSize: 12,
                    fontWeight: 500,
                  },
                },
                {
                  type: 'text',
                  left: 120,
                  top: 36,
                  style: {
                    text: `MA(25) ${lastMA25 != null ? formatPrice(lastMA25) : '-'}`,
                    fill: '#c994ff',
                    fontSize: 12,
                    fontWeight: 500,
                  },
                },
                {
                  type: 'text',
                  left: 255,
                  top: 36,
                  style: {
                    text: `MA(99) ${lastMA99 != null ? formatPrice(lastMA99) : '-'}`,
                    fill: '#a78bfa',
                    fontSize: 12,
                    fontWeight: 500,
                  },
                },
                {
                  type: 'text',
                  left: 0,
                  top: 54,
                  style: {
                    text: `最新 ${formatPrice(lastPrice)}  (${lastChg >= 0 ? '+' : ''}${lastChg.toFixed(2)}%)`,
                    fill: lastChgColor,
                    fontSize: 12,
                    fontWeight: 600,
                  },
                },
              ],
            },
            // 成交量区左侧：Vol(BTC) Vol(USDT)，与币安一致
            {
              type: 'group',
              left: 54,
              top: 398,
              silent: true,
              children: [
                {
                  type: 'text',
                  left: 0,
                  top: 0,
                  style: {
                    text: `Vol(BTC) ${formatVol(last?.volume ?? 0)}  Vol(USDT) ${formatVol((last?.volume ?? 0) * (last?.close ?? 0))}`,
                    fill: '#848e9c',
                    fontSize: 12,
                    fontWeight: 500,
                  },
                },
              ],
            },
          ],
          series: [
            {
              name: 'K',
              type: 'candlestick',
              data: values,
              itemStyle: {
                color: upColor,
                color0: downColor,
                borderColor: upColor,
                borderColor0: downColor,
              },
              markLine: {
                symbol: ['none', 'none'],
                label: {
                  show: true,
                  formatter: () => `${lastPrice.toFixed(priceDigits)}`,
                  color: '#eaecef',
                  backgroundColor: '#2b3139',
                  padding: [3, 6],
                  borderRadius: 3,
                },
                lineStyle: { color: '#848e9c', type: 'dashed', width: 1, opacity: 0.8 },
                data: [{ yAxis: lastPrice }],
              },
            },
            {
              name: 'MA7',
              type: 'line',
              data: ma7,
              smooth: true,
              showSymbol: false,
              connectNulls: true,
              lineStyle: { width: 1.5, color: '#f0b90b' },
              tooltip: { show: false },
            },
            {
              name: 'MA25',
              type: 'line',
              data: ma25,
              smooth: true,
              showSymbol: false,
              connectNulls: true,
              lineStyle: { width: 1.5, color: '#c994ff' },
              tooltip: { show: false },
            },
            {
              name: 'MA99',
              type: 'line',
              data: ma99,
              smooth: true,
              showSymbol: false,
              connectNulls: true,
              lineStyle: { width: 1.5, color: '#a78bfa' },
              tooltip: { show: false },
            },
            {
              name: 'VOL',
              type: 'bar',
              xAxisIndex: 1,
              yAxisIndex: 1,
              data: volValues,
              barWidth: '60%',
              itemStyle: {
                color: (p: any) => {
                  const idx = p?.dataIndex ?? 0
                  const k = klines[idx]
                  return k && (k.close ?? 0) >= (k.open ?? 0) ? upColor : downColor
                },
                opacity: 0.85,
              },
              tooltip: { show: false },
            },
            {
              name: 'Vol MA5',
              type: 'line',
              xAxisIndex: 1,
              yAxisIndex: 1,
              data: volMA5,
              smooth: true,
              showSymbol: false,
              connectNulls: true,
              lineStyle: { width: 1, color: '#38bdf8' },
              tooltip: { show: false },
            },
            {
              name: 'Vol MA10',
              type: 'line',
              xAxisIndex: 1,
              yAxisIndex: 1,
              data: volMA10,
              smooth: true,
              showSymbol: false,
              connectNulls: true,
              lineStyle: { width: 1, color: '#fb7185' },
              tooltip: { show: false },
            },
          ],
        },
        !isFirstRender
      )
      chartInstance.current?.resize()
      lastChartUpdateRef.current = Date.now()
    })
    chartInstance.current.resize()
    return () => {
      if (chartUpdateTimeoutRef.current != null) {
        window.clearTimeout(chartUpdateTimeoutRef.current)
        chartUpdateTimeoutRef.current = null
      }
    }
  }, [klines, streaming, chartUpdateTrigger, visibleBars])

  useEffect(() => {
    const chart = chartInstance.current
    const onResize = () => chart?.resize()
    window.addEventListener('resize', onResize)
    return () => {
      if (renderRaf.current) cancelAnimationFrame(renderRaf.current)
      window.removeEventListener('resize', onResize)
      chart?.dispose()
      chartInstance.current = null
    }
  }, [])

  const handleStart = () => {
    const count = Math.min(Math.max(Number(limit) || 1000, 1), 1500)
    StartKlineStream(symbol, interval, count)
      .then(() => {
        setStreaming(true)
        setStreamStatus('polling')
        message.success('K 线轮询已启动（每 100ms）')
      })
      .catch((err) => {
        message.error(err?.message || '启动失败')
        setStreaming(false)
        setStreamStatus('error')
      })
  }

  const handleStop = () => {
    StopKlineStream()
    setStreaming(false)
    setStreamStatus(null)
    message.info('已停止')
  }

  const doPredict = useCallback(() => {
    const arr = klinesRef.current
    if (arr.length < 5) {
      message.warning('至少需要 5 根 K 线')
      return
    }
    if (predictingRef.current) return
    setPredicting(true)
    const sym = symbolRef.current
    const count = Math.min(500, Math.max(1, cozeKlineCountRef.current))
    CozePredictStructured(arr.map((k) => factor.KLine.createFrom(k)), sym, count)
      .then(() => {})
      .catch((err) => message.error(err?.message || '预测失败'))
      .finally(() => setPredicting(false))
  }, [])

  const handlePredict = () => {
    doPredict()
  }

  // 启动实时流时启动定时器，按间隔调用 Coze 预测；停止流时清除定时器
  useEffect(() => {
    if (!streaming) return
    const ms = Math.max(1, cozeIntervalMinutes) * 60 * 1000
    const id = window.setInterval(() => {
      doPredict()
    }, ms)
    return () => window.clearInterval(id)
  }, [streaming, cozeIntervalMinutes, doPredict])

  /** 与 agent 约定一致：三情景固定顺序 LONG / SHORT / SIDEWAYS */
  const SCENARIO_ORDER = ['LONG', 'SHORT', 'SIDEWAYS'] as const
  const sortScenarios = (list: CozeScenario[] | undefined) => {
    if (!list?.length) return []
    return [...list].sort((a, b) => SCENARIO_ORDER.indexOf(a.direction as any) - SCENARIO_ORDER.indexOf(b.direction as any))
  }

  const scenarioColumns = [
    { title: '方向', dataIndex: 'direction', key: 'direction', width: 80 },
    { title: '概率', dataIndex: 'probability', key: 'probability', width: 60, render: (v: number) => `${v}%` },
    { title: '入场', dataIndex: 'entry_price', key: 'entry_price', width: 90, render: (v: number | null) => (v != null ? v.toFixed(2) : '-') },
    { title: '止损', dataIndex: 'stop_loss', key: 'stop_loss', width: 90, render: (v: number | null) => (v != null ? v.toFixed(2) : '-') },
    { title: '止盈1', dataIndex: 'take_profit_1', key: 'take_profit_1', width: 90, render: (v: number | null) => (v != null ? v.toFixed(2) : '-') },
    { title: '止盈2', dataIndex: 'take_profit_2', key: 'take_profit_2', width: 90, render: (v: number | null) => (v != null ? v.toFixed(2) : '-') },
    { title: '风报比', dataIndex: 'risk_reward_ratio', key: 'risk_reward_ratio', width: 70, render: (v: number | null) => (v != null ? v.toFixed(1) : '-') },
    {
      title: '动作',
      dataIndex: 'action',
      key: 'action',
      width: 80,
      render: (action: string) => {
        const map: Record<string, { color: string }> = { EXECUTE: { color: 'green' }, WAIT: { color: 'orange' }, SKIP: { color: 'default' } }
        const cfg = map[action] ?? { color: 'default' }
        return <Tag color={cfg.color}>{action || '-'}</Tag>
      },
    },
    { title: '触发条件', dataIndex: 'trigger_condition', key: 'trigger_condition', ellipsis: true },
  ]

  const limitNum = Math.min(Math.max(Number(limit) || 1000, 1), 1500)

  return (
    <div className="max-w-7xl mx-auto p-4 space-y-4">
      <h1 className="text-xl font-medium text-[#242f57]">Coze K 线预测</h1>
      <Card title="K 线图">
        <div className="flex flex-wrap items-center gap-3 mb-3">
          <span className="text-sm text-gray-600">交易对</span>
          <Input
            value={symbol}
            onChange={(e) => setSymbol((e.target.value || '').trim().toUpperCase() || 'BTCUSDT')}
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
            onChange={(v) => setLimit(v == null ? 1000 : Number(v))}
            style={{ width: 100 }}
          />
          <span className="text-sm text-gray-600">时间窗(根)</span>
          <InputNumber
            min={10}
            max={500}
            value={visibleBars}
            onChange={(v) => setVisibleBars(v == null ? 100 : Math.min(500, Math.max(10, Number(v))))}
            style={{ width: 88 }}
          />
          <Button type="primary" onClick={handleStart} disabled={streaming}>
            启动实时流
          </Button>
          <Button onClick={handleStop} disabled={!streaming}>
            停止
          </Button>
          <span className="text-sm text-gray-600">定时预测间隔(分钟)</span>
          <InputNumber
            min={1}
            max={60}
            value={cozeIntervalMinutes}
            onChange={(v) => setCozeIntervalMinutes(v == null ? 2 : Math.min(60, Math.max(1, Number(v))))}
            style={{ width: 88 }}
          />
          <span className="text-sm text-gray-500 self-center">
            {klines.length} 根 · 轮询: {streamStatus ?? '-'} · 最后更新: {lastUpdate ?? '-'} · Coze: {cozeStatus ?? '-'}
            {streaming ? ` · 定时: 每 ${cozeIntervalMinutes} 分钟` : ''}
          </span>
        </div>
        <div
          ref={chartRef}
          style={{
            width: '100%',
            minHeight: 520,
            height: 520,
            background: '#0b0e11',
            border: '1px solid #1e2329',
            borderRadius: 6,
          }}
        />
      </Card>


      <Card title="Coze 预测" style={{ marginTop: 16 }}>
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-sm text-gray-600">给豆包的数据量</span>
          <InputNumber
            min={1}
            max={500}
            value={cozeKlineCount}
            onChange={(v) => {
              const n = v == null ? 50 : Math.min(500, Math.max(1, Number(v)))
              setCozeKlineCountState(n)
              SetCozeKlineCount(n)
            }}
            style={{ width: 80 }}
          />
          <Button onClick={handlePredict} loading={predicting} disabled={klines.length < 5}>
            手动预测
          </Button>
          <span className="text-sm text-gray-600">
            当前状态: <span className="font-medium">{cozeStatus ?? '-'}</span>
          </span>
          {cozeStatus === '请求中...' && (cozePreview || results[0]?.market_structure || results[0]?.rawAnswer) ? (
            <span className="text-sm text-gray-500 truncate max-w-[720px]">
              {cozePreview ?? results[0]?.market_structure ?? results[0]?.rawAnswer?.replace(/\s+/g, ' ').slice(0, 120)}
            </span>
          ) : null}
        </div>
      </Card>

      <Card title="预测结果 (最近 20 次)">
        {results.length === 0 ? (
          <div className="text-gray-500 py-8 text-center">暂无预测结果</div>
        ) : (
          <div className="space-y-4">
            {results.map((r, i) => (
              <div key={i} className="border rounded p-3 bg-gray-50/50">
                <div className="flex gap-4 mb-2 text-sm">
                  <span>{r.timestamp}</span>
                  <span>{r.symbol}</span>
                  <span>当前价: {r.current_price?.toFixed(2)}</span>
                  <span className="text-gray-600">{r.market_structure}</span>
                </div>
                {r.rawAnswer && !r.scenarios?.length ? (
                  <pre className="text-xs overflow-auto max-h-32">{r.rawAnswer}</pre>
                ) : (
                  <Table
                    dataSource={sortScenarios(r.scenarios)}
                    columns={scenarioColumns}
                    rowKey="direction"
                    size="small"
                    pagination={false}
                  />
                )}
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  )
}
