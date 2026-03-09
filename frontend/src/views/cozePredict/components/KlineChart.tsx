import { useEffect, useMemo, useRef, useState, type CSSProperties } from 'react'
import * as echarts from 'echarts'

import type { KLine } from '../types'

const upColor = '#0ECB81'
const downColor = '#F6465D'

function formatPrice(value: number, digits: number = 2): string {
  const fixed = value.toFixed(digits)
  const [intPart, decPart] = fixed.split('.')
  const withComma = intPart.replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  return decPart != null ? `${withComma}.${decPart}` : withComma
}

function formatVolume(value: number): string {
  const formatWithUnit = (scaled: number, unit: string) => `${trimTrailingZeros(scaled.toFixed(3))}${unit}`
  if (value >= 1e9) return formatWithUnit(value / 1e9, 'B')
  if (value >= 1e6) return formatWithUnit(value / 1e6, 'M')
  if (value >= 1e3) return formatWithUnit(value / 1e3, 'K')
  return trimTrailingZeros(value.toFixed(3))
}

function trimTrailingZeros(value: string): string {
  return value.replace(/\.?0+$/, '')
}

function formatTimeOnly(timestamp: number): string {
  const date = new Date(timestamp)
  const hour = date.getHours()
  const minute = date.getMinutes()
  return `${hour.toString().padStart(2, '0')}:${minute.toString().padStart(2, '0')}`
}

function formatDateOnly(timestamp: number): string {
  const date = new Date(timestamp)
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  return `${month}/${day}`
}

function formatDateTime(timestamp: number): string {
  const date = new Date(timestamp)
  const year = date.getFullYear()
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  const hour = `${date.getHours()}`.padStart(2, '0')
  const minute = `${date.getMinutes()}`.padStart(2, '0')
  return `${year}-${month}-${day} ${hour}:${minute}`
}

function formatXAxisLabel(klines: KLine[], index: number): string {
  const current = klines[index]
  if (!current) return ''
  if (index === 0) return formatTimeOnly(current.openTime)
  const previous = klines[index - 1]
  if (!previous) return formatTimeOnly(current.openTime)
  const currentDate = new Date(current.openTime)
  const prevDate = new Date(previous.openTime)
  const dayChanged =
    currentDate.getFullYear() !== prevDate.getFullYear() ||
    currentDate.getMonth() !== prevDate.getMonth() ||
    currentDate.getDate() !== prevDate.getDate()
  return dayChanged ? formatDateOnly(current.openTime) : formatTimeOnly(current.openTime)
}

function calcMA(period: number, data: KLine[]) {
  const result: (number | null)[] = []
  for (let i = 0; i < data.length; i++) {
    if (i < period - 1) {
      result.push(null)
      continue
    }
    let sum = 0
    for (let j = 0; j < period; j++) sum += data[i - j].close ?? 0
    result.push(Number((sum / period).toFixed(2)))
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

interface KlineChartProps {
  klines: KLine[]
  streaming: boolean
  visibleBars: number
}

export function KlineChart({ klines, streaming, visibleBars }: KlineChartProps) {
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstance = useRef<echarts.ECharts | null>(null)
  const renderRaf = useRef<number | null>(null)
  const dataZoomRef = useRef<{ start: number; end: number } | null>(null)
  const chartUpdateTimeoutRef = useRef<number | null>(null)
  const lastChartUpdateRef = useRef(0)
  const fromThrottleRef = useRef(false)
  const isHoveringRef = useRef(false)
  const pendingRefreshRef = useRef(false)
  const hoveredDataIndexRef = useRef<number | null>(null)
  const chartDataRef = useRef<ReturnType<typeof buildChartData> | null>(null)
  const klinesRef = useRef<KLine[]>([])
  const priceDigitsRef = useRef(2)
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null)
  const [chartUpdateTrigger, setChartUpdateTrigger] = useState(0)

  const chartData = useMemo(() => buildChartData(klines), [klines])
  chartDataRef.current = chartData
  klinesRef.current = klines
  const displayIndex =
    hoveredIndex != null && hoveredIndex >= 0 && hoveredIndex < klines.length ? hoveredIndex : Math.max(klines.length - 1, 0)
  const displayMetrics = useMemo(() => getDisplayMetrics(klines, chartData, displayIndex), [chartData, displayIndex, klines])
  const lastCompletedIndex = streaming ? Math.max(klines.length - 2, 0) : Math.max(klines.length - 1, 0)
  const latestVolMA7 = chartData.volMA7[lastCompletedIndex]
  const latestVolMA14 = chartData.volMA14[lastCompletedIndex]

  useEffect(() => {
    dataZoomRef.current = null
  }, [visibleBars])

  // Static options & event handlers — runs once on mount
  useEffect(() => {
    if (!chartRef.current) return
    const chart = echarts.init(chartRef.current)
    chartInstance.current = chart
    lastChartUpdateRef.current = 0

    // Formatters read from refs so they always reflect the latest data
    // without needing to be recreated on every update.
    chart.setOption({
      backgroundColor: 'transparent',
      animation: false,
      axisPointer: {
        show: true,
        link: [{ xAxisIndex: 'all' }],
        label: {
          backgroundColor: '#2b3139',
          color: '#eaecef',
          borderRadius: 2,
          padding: [3, 6],
        },
        lineStyle: {
          color: '#6b7280',
          type: 'dashed',
          width: 1,
          opacity: 0.9,
        },
      },
      tooltip: {
        show: true,
        trigger: 'axis',
        triggerOn: 'mousemove|click',
        formatter: (params: any) => {
          const nextIndex = resolveDataIndex(params, chartDataRef.current?.times ?? [])
          if (nextIndex != null) {
            hoveredDataIndexRef.current = nextIndex
            isHoveringRef.current = true
            setHoveredIndex((cur) => (cur === nextIndex ? cur : nextIndex))
          }
          return ''
        },
        backgroundColor: 'transparent',
        borderWidth: 0,
        padding: 0,
        textStyle: { color: 'transparent', fontSize: 0 },
        extraCssText: 'box-shadow:none;',
        axisPointer: { type: 'cross' },
      },
      grid: [
        { left: 54, right: 72, top: 38, height: 372 },
        { left: 54, right: 72, top: 426, height: 90 },
      ],
      xAxis: [
        {
          type: 'category',
          data: [],
          boundaryGap: true,
          axisLine: { lineStyle: { color: '#2b3139' } },
          axisLabel: { show: false },
          splitNumber: 7,
          axisPointer: {
            label: {
              formatter: (params: any) => {
                const index = resolveDataIndex({ axesInfo: [params] }, chartDataRef.current?.times ?? [])
                if (index == null) return ''
                return formatDateTime(klinesRef.current[index]?.openTime ?? 0)
              },
            },
          },
          splitLine: { show: false },
          axisTick: { show: false },
        },
        {
          type: 'category',
          gridIndex: 1,
          data: [],
          boundaryGap: true,
          axisLine: { lineStyle: { color: '#2b3139' } },
          axisLabel: {
            color: '#848e9c',
            fontSize: 11,
            hideOverlap: true,
            margin: 14,
            formatter: (_v: string, index: number) => formatXAxisLabel(klinesRef.current, index),
          },
          splitLine: { show: false },
          axisTick: { show: false },
        },
      ],
      yAxis: [
        {
          scale: true,
          position: 'right',
          axisLine: { lineStyle: { color: '#2b3139' } },
          axisLabel: {
            color: '#848e9c',
            fontSize: 11,
            formatter: (value: number) => Number(value).toFixed(priceDigitsRef.current),
          },
          axisPointer: {
            label: {
              formatter: ({ value }: { value: number }) => formatPrice(Number(value), priceDigitsRef.current),
            },
          },
          splitLine: { lineStyle: { color: '#1b1f27', width: 1 } },
        },
        {
          gridIndex: 1,
          scale: true,
          position: 'right',
          axisLine: { lineStyle: { color: '#2b3139' } },
          axisLabel: {
            color: '#848e9c',
            fontSize: 11,
            formatter: (value: number) =>
              value >= 1e6 ? `${(value / 1e6).toFixed(1)}M` : value >= 1e3 ? `${(value / 1e3).toFixed(1)}K` : `${Math.round(value)}`,
          },
          splitLine: { show: false },
        },
      ],
      dataZoom: [
        {
          type: 'inside',
          xAxisIndex: [0, 1],
          start: 0,
          end: 100,
          zoomOnMouseWheel: false,
          moveOnMouseMove: true,
          moveOnMouseWheel: true,
          preventDefaultMouseMove: true,
        },
        {
          type: 'slider',
          xAxisIndex: [0, 1],
          bottom: 8,
          height: 18,
          start: 0,
          end: 100,
          brushSelect: false,
          fillerColor: 'rgba(132, 142, 156, 0.12)',
          borderColor: '#2b3139',
          handleStyle: { color: '#6b7280', borderColor: '#6b7280' },
          moveHandleStyle: { color: '#6b7280' },
          textStyle: { color: '#6b7280', fontSize: 10 },
        },
      ],
      graphic: [],
      series: [
        {
          name: 'K',
          type: 'candlestick',
          data: [],
          itemStyle: {
            color: upColor,
            color0: downColor,
            borderColor: upColor,
            borderColor0: downColor,
          },
        },
        {
          name: 'MA7',
          type: 'line',
          data: [],
          smooth: true,
          showSymbol: false,
          connectNulls: true,
          lineStyle: { width: 1.5, color: '#f0b90b' },
          tooltip: { show: false },
        },
        {
          name: 'MA25',
          type: 'line',
          data: [],
          smooth: true,
          showSymbol: false,
          connectNulls: true,
          lineStyle: { width: 1.5, color: '#ED4D94' },
          tooltip: { show: false },
        },
        {
          name: 'MA99',
          type: 'line',
          data: [],
          smooth: true,
          showSymbol: false,
          connectNulls: true,
          lineStyle: { width: 1.5, color: '#7B61FF' },
          tooltip: { show: false },
        },
        {
          name: 'VOL',
          type: 'bar',
          xAxisIndex: 1,
          yAxisIndex: 1,
          data: [],
          barWidth: '60%',
          itemStyle: {
            color: (params: any) => {
              const idx = params?.dataIndex ?? 0
              const cur = klinesRef.current[idx]
              return cur && (cur.close ?? 0) >= (cur.open ?? 0) ? upColor : downColor
            },
            opacity: 0.85,
          },
          tooltip: { show: false },
        },
        {
          name: 'Vol MA7',
          type: 'line',
          xAxisIndex: 1,
          yAxisIndex: 1,
          data: [],
          smooth: true,
          showSymbol: false,
          connectNulls: true,
          lineStyle: { width: 1, color: '#38bdf8' },
          tooltip: { show: false },
        },
        {
          name: 'Vol MA14',
          type: 'line',
          xAxisIndex: 1,
          yAxisIndex: 1,
          data: [],
          smooth: true,
          showSymbol: false,
          connectNulls: true,
          lineStyle: { width: 1, color: '#fb7185' },
          tooltip: { show: false },
        },
      ],
    })

    chart.on('updateAxisPointer', (event: any) => {
      const nextIndex = resolveDataIndex(event, chartDataRef.current?.times ?? [])
      if (nextIndex == null) return
      hoveredDataIndexRef.current = nextIndex
      isHoveringRef.current = true
      setHoveredIndex((cur) => (cur === nextIndex ? cur : nextIndex))
    })

    chart.on('datazoom', (event: any) => {
      const payload = Array.isArray(event?.batch) ? event.batch[0] : event
      const s = Number(payload?.start)
      const e = Number(payload?.end)
      if (Number.isFinite(s) && Number.isFinite(e)) {
        dataZoomRef.current = { start: s, end: e }
        return
      }
      const sv = Number(payload?.startValue)
      const ev = Number(payload?.endValue)
      const total = klinesRef.current.length
      if (Number.isFinite(sv) && Number.isFinite(ev) && total > 1) {
        dataZoomRef.current = { start: (sv / total) * 100, end: ((ev + 1) / total) * 100 }
      }
    })

    chart.getZr().on('globalout', () => {
      hoveredDataIndexRef.current = null
      isHoveringRef.current = false
      chart.dispatchAction({ type: 'hideTip' })
      setHoveredIndex(null)
      if (pendingRefreshRef.current) {
        pendingRefreshRef.current = false
        fromThrottleRef.current = true
        setChartUpdateTrigger((v) => v + 1)
      }
    })

    const onResize = () => chart.resize()
    window.addEventListener('resize', onResize)

    return () => {
      if (renderRaf.current) cancelAnimationFrame(renderRaf.current)
      if (chartUpdateTimeoutRef.current != null) {
        window.clearTimeout(chartUpdateTimeoutRef.current)
        chartUpdateTimeoutRef.current = null
      }
      window.removeEventListener('resize', onResize)
      chart.dispose()
      chartInstance.current = null
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  // Data-only updates — partial setOption with just the changing parts
  useEffect(() => {
    const chart = chartInstance.current
    if (!chart || klines.length === 0) return

    if (streaming && isHoveringRef.current) {
      pendingRefreshRef.current = true
      return
    }

    const throttleMs = 800
    const elapsed = Date.now() - lastChartUpdateRef.current
    if (streaming && !fromThrottleRef.current && lastChartUpdateRef.current && elapsed < throttleMs) {
      if (chartUpdateTimeoutRef.current != null) window.clearTimeout(chartUpdateTimeoutRef.current)
      chartUpdateTimeoutRef.current = window.setTimeout(() => {
        chartUpdateTimeoutRef.current = null
        fromThrottleRef.current = true
        setChartUpdateTrigger((v) => v + 1)
      }, throttleMs - elapsed)
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
      const { times, values, volumes, ma7, ma25, ma99, volMA7, volMA14 } = chartData
      const latestClose = klines[klines.length - 1]?.close ?? 0
      const latestOpen = klines[klines.length - 1]?.open ?? 0
      const priceDigits = latestClose < 1 ? 6 : latestClose < 100 ? 4 : 2
      priceDigitsRef.current = priceDigits

      const rightGap = 4
      const totalBars = times.length + rightGap
      const xAxisMax = Math.max(times.length - 1 + rightGap, 0)
      const bars = Math.max(1, visibleBars)
      const defaultStart = totalBars <= bars ? 0 : 100 - (bars / totalBars) * 100
      const useDefaultZoom = lastChartUpdateRef.current === 0 || dataZoomRef.current == null
      let start = defaultStart
      let end = 100

      if (!useDefaultZoom) {
        const prevOpt = chart.getOption() as any
        const zoom = Array.isArray(prevOpt?.dataZoom)
          ? prevOpt.dataZoom.find((item: any) => item.xAxisIndex != null)
          : null
        if (zoom?.start != null && zoom?.end != null) {
          start = Number(zoom.start)
          end = Number(zoom.end)
        } else if (dataZoomRef.current) {
          start = dataZoomRef.current.start
          end = dataZoomRef.current.end
        }
      }

      dataZoomRef.current = { start, end }

      chart.setOption({
        xAxis: [
          { data: times, min: 0, max: xAxisMax },
          { data: times, min: 0, max: xAxisMax },
        ],
        dataZoom: [{ start, end }, { start, end }],
        series: [
          {
            data: values,
            markLine: {
              symbol: ['none', 'none'],
              label: {
                show: true,
                position: 'end',
                formatter: () => formatPrice(latestClose, priceDigits),
                color: '#ffffff',
                opacity: 1,
                fontSize: 12,
                fontWeight: 500,
                backgroundColor: latestClose >= latestOpen ? upColor : downColor,
                padding: [3, 3, 3, 3],
                borderRadius: 2,
                distance: -2,
              },
              lineStyle: {
                color: latestClose >= latestOpen ? upColor : downColor,
                type: 'dashed',
                width: 1,
                opacity: 0.6,
              },
              data: [{ yAxis: latestClose }],
            },
          },
          { data: ma7 },
          { data: ma25 },
          { data: ma99 },
          { data: volumes },
          { data: volMA7 },
          { data: volMA14 },
        ],
      })

      chart.resize()
      lastChartUpdateRef.current = Date.now()
    })

    return () => {
      if (chartUpdateTimeoutRef.current != null) {
        window.clearTimeout(chartUpdateTimeoutRef.current)
        chartUpdateTimeoutRef.current = null
      }
    }
  }, [chartData, chartUpdateTrigger, klines, streaming, visibleBars])

  return (
    <div
      style={{
        position: 'relative',
        width: '100%',
        minHeight: 552,
        height: 552,
        background: '#0b0e11',
        border: '1px solid #1e2329',
        borderRadius: 6,
        overflow: 'hidden',
      }}
    >
      <div style={overlayRowStyle(8, 54, 6)}>
        <span style={metricCellStyle(108, '#848e9c', 400)}>{displayMetrics.timeText}</span>
        <span style={metricCellStyle(78, '#eaecef')}>
          <span style={metricLabelStyle}>开 </span>
          <span style={metricValueStyle(displayMetrics.changeColor)}>{formatPrice(displayMetrics.open, displayMetrics.priceDigits)}</span>
        </span>
        <span style={metricCellStyle(78, '#eaecef')}>
          <span style={metricLabelStyle}>高 </span>
          <span style={metricValueStyle(displayMetrics.changeColor)}>{formatPrice(displayMetrics.high, displayMetrics.priceDigits)}</span>
        </span>
        <span style={metricCellStyle(78, '#eaecef')}>
          <span style={metricLabelStyle}>低 </span>
          <span style={metricValueStyle(displayMetrics.changeColor)}>{formatPrice(displayMetrics.low, displayMetrics.priceDigits)}</span>
        </span>
        <span style={metricCellStyle(92, displayMetrics.changeColor, 600)}>
          <span style={metricLabelStyle}>收 </span>
          <span style={metricValueStyle(displayMetrics.changeColor, 600)}>{formatPrice(displayMetrics.close, displayMetrics.priceDigits)}</span>
        </span>
        <span style={metricCellStyle(102, displayMetrics.changeColor, 600)}>
          <span style={metricLabelStyle}>涨跌幅 </span>
          <span style={metricValueStyle(displayMetrics.changeColor, 600)}>{`${displayMetrics.change >= 0 ? '+' : ''}${displayMetrics.change.toFixed(2)}%`}</span>
        </span>
        <span style={metricCellStyle(76, '#848e9c')}>
          <span style={metricLabelStyle}>振幅 </span>
          <span style={metricValueStyle(displayMetrics.changeColor)}>{`${displayMetrics.amplitude.toFixed(2)}%`}</span>
        </span>
      </div>
      <div style={overlayRowStyle(24, 54, 12)}>
        <span style={metricCellStyle(118, '#f0b90b')}>
          <span style={metricLabelStyle}>MA(7) </span>
          <span style={metricValueStyle('#f0b90b')}>{displayMetrics.ma7 != null ? formatPrice(displayMetrics.ma7, displayMetrics.priceDigits) : '-'}</span>
        </span>
        <span style={metricCellStyle(126, '#ED4D94')}>
          <span style={metricLabelStyle}>MA(25) </span>
          <span style={metricValueStyle('#ED4D94')}>{displayMetrics.ma25 != null ? formatPrice(displayMetrics.ma25, displayMetrics.priceDigits) : '-'}</span>
        </span>
        <span style={metricCellStyle(126, '#7B61FF')}>
          <span style={metricLabelStyle}>MA(99) </span>
          <span style={metricValueStyle('#7B61FF')}>{displayMetrics.ma99 != null ? formatPrice(displayMetrics.ma99, displayMetrics.priceDigits) : '-'}</span>
        </span>
      </div>
      <div style={overlayRowStyle(398, 54, 10)}>
        <span style={metricCellStyle(88, '#848e9c')}>
          <span style={metricLabelStyle}>Vol(BTC) </span>
          <span style={metricValueStyle('#eaecef')}>{formatVolume(displayMetrics.volume)}</span>
        </span>
        <span style={metricCellStyle(126, '#848e9c')}>
          <span style={metricLabelStyle}>Vol(USDT) </span>
          <span style={metricValueStyle('#eaecef')}>{formatVolume(displayMetrics.quoteVolume)}</span>
        </span>
        <span style={metricCellStyle(84, '#38bdf8')}>
          <span style={metricValueStyle('#38bdf8')}>{latestVolMA7 != null ? formatVolume(latestVolMA7) : '-'}</span>
        </span>
        <span style={metricCellStyle(88, '#fb7185')}>
          <span style={metricValueStyle('#fb7185')}>{latestVolMA14 != null ? formatVolume(latestVolMA14) : '-'}</span>
        </span>
      </div>
      <div
        ref={chartRef}
        style={{
          width: '100%',
          height: '100%',
        }}
      />
    </div>
  )
}

function buildChartData(klines: KLine[]) {
    const times = klines.map((item) => new Date(item.openTime).toISOString().slice(0, 16).replace('T', ' '))
    const values = klines.map((item) => [item.open, item.close, item.low, item.high])
    const volumes = klines.map((item) => item.volume ?? 0)
    return {
      times,
      values,
      volumes,
      ma7: calcMA(7, klines),
      ma25: calcMA(25, klines),
      ma99: calcMA(99, klines),
      volMA7: calcVolMA(7, klines),
      volMA14: calcVolMA(14, klines),
    }
}

function getDisplayMetrics(klines: KLine[], chartData: ReturnType<typeof buildChartData>, index: number) {
  const current = klines[index] ?? klines[klines.length - 1]
  const open = current?.open ?? 0
  const close = current?.close ?? 0
  const high = current?.high ?? 0
  const low = current?.low ?? 0
  const volume = current?.volume ?? 0
  const quoteVolume = current?.quoteAssetVolume ?? 0
  const change = open ? ((close - open) / open) * 100 : 0
  const amplitude = open ? ((high - low) / open) * 100 : 0
  const changeColor = close >= open ? upColor : downColor
  const priceDigits = close < 1 ? 6 : close < 100 ? 4 : 2
  const timeText = chartData.times[index]?.replace(/-/g, '/') ?? ''
  const ma7 = chartData.ma7[index]
  const ma25 = chartData.ma25[index]
  const ma99 = chartData.ma99[index]
  const volMA7 = chartData.volMA7[index]
  const volMA14 = chartData.volMA14[index]

  return {
    timeText,
    open,
    close,
    high,
    low,
    volume,
    quoteVolume,
    change,
    amplitude,
    changeColor,
    priceDigits,
    ma7,
    ma25,
    ma99,
    volMA7,
    volMA14,
  }
}


function overlayRowStyle(top: number, left: number, gap: number): CSSProperties {
  return {
    position: 'absolute',
    top,
    left,
    right: 72,
    display: 'flex',
    gap,
    alignItems: 'center',
    pointerEvents: 'none',
    zIndex: 2,
  }
}

function metricCellStyle(width: number, color: string, fontWeight: number = 500): CSSProperties {
  return {
    width,
    minWidth: width,
    maxWidth: width,
    color,
    display: 'inline-flex',
    alignItems: 'center',
    fontSize: 11,
    fontWeight,
    lineHeight: '16px',
    whiteSpace: 'nowrap',
    overflow: 'hidden',
    textOverflow: 'clip',
    fontVariantNumeric: 'tabular-nums',
    letterSpacing: '-0.1px',
  }
}

const metricLabelStyle: CSSProperties = {
  color: '#848e9c',
  fontWeight: 400,
}

function metricValueStyle(color: string, fontWeight: number = 500): CSSProperties {
  return {
    color,
    fontWeight,
  }
}

function resolveDataIndex(event: any, times: string[]): number | null {
  const tooltipParams = Array.isArray(event) ? event : null
  const firstSeries = tooltipParams?.find((item: any) => typeof item?.dataIndex === 'number')
  if (typeof firstSeries?.dataIndex === 'number') return firstSeries.dataIndex

  const seriesData = Array.isArray(event?.seriesData) ? event.seriesData : null
  const firstSeriesData = seriesData?.find((item: any) => typeof item?.dataIndex === 'number')
  if (typeof firstSeriesData?.dataIndex === 'number') return firstSeriesData.dataIndex

  const axisInfo = event?.axesInfo?.[0]
  if (!axisInfo) return null
  if (typeof axisInfo.value === 'number') return axisInfo.value
  const value = String(axisInfo.value ?? '')
  const index = times.indexOf(value)
  return index >= 0 ? index : null
}
