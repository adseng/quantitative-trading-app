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
  if (value >= 1e6) return `${(value / 1e6).toFixed(2)}M`
  if (value >= 1e3) return `${(value / 1e3).toFixed(3)}K`
  return value.toFixed(2)
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
  const updateAxisPointerHandlerRef = useRef<((event: any) => void) | null>(null)
  const dataZoomHandlerRef = useRef<((event: any) => void) | null>(null)
  const globalOutHandlerRef = useRef<(() => void) | null>(null)
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null)
  const [chartUpdateTrigger, setChartUpdateTrigger] = useState(0)

  const chartData = useMemo(() => buildChartData(klines), [klines])
  chartDataRef.current = chartData
  klinesRef.current = klines
  const displayIndex =
    hoveredIndex != null && hoveredIndex >= 0 && hoveredIndex < klines.length ? hoveredIndex : Math.max(klines.length - 1, 0)
  const displayMetrics = useMemo(() => getDisplayMetrics(klines, chartData, displayIndex), [chartData, displayIndex, klines])

  useEffect(() => {
    dataZoomRef.current = null
  }, [visibleBars])

  const flushPendingRefresh = () => {
    if (!pendingRefreshRef.current) return
    pendingRefreshRef.current = false
    fromThrottleRef.current = true
    setChartUpdateTrigger((value) => value + 1)
  }

  useEffect(() => {
    if (klines.length === 0 || !chartRef.current) return
    if (!chartInstance.current) chartInstance.current = echarts.init(chartRef.current)

    const chart = chartInstance.current
    if (!updateAxisPointerHandlerRef.current) {
      updateAxisPointerHandlerRef.current = (event: any) => {
        const nextIndex = resolveDataIndex(event, chartDataRef.current?.times ?? [])
        if (nextIndex == null) return
        hoveredDataIndexRef.current = nextIndex
        isHoveringRef.current = true
        setHoveredIndex((current) => (current === nextIndex ? current : nextIndex))
      }
      chart.on('updateAxisPointer', updateAxisPointerHandlerRef.current)
    }
    if (!dataZoomHandlerRef.current) {
      dataZoomHandlerRef.current = (event: any) => {
        const payload = Array.isArray(event?.batch) ? event.batch[0] : event
        const start = Number(payload?.start)
        const end = Number(payload?.end)
        if (Number.isFinite(start) && Number.isFinite(end)) {
          dataZoomRef.current = { start, end }
        }
      }
      chart.on('datazoom', dataZoomHandlerRef.current)
    }
    if (!globalOutHandlerRef.current) {
      globalOutHandlerRef.current = () => {
        hoveredDataIndexRef.current = null
        isHoveringRef.current = false
        chart.dispatchAction({ type: 'hideTip' })
        setHoveredIndex(null)
        flushPendingRefresh()
      }
      chart.getZr().on('globalout', globalOutHandlerRef.current)
    }

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
        setChartUpdateTrigger((value) => value + 1)
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
      const { times, values, volumes, ma7, ma25, ma99, volMA5, volMA10 } = chartData
      const latestMetrics = getDisplayMetrics(klines, chartData, Math.max(klines.length - 1, 0))
      const priceDigits = latestMetrics.priceDigits

      const bars = Math.max(1, visibleBars)
      const defaultZoomStart = klines.length <= bars ? 0 : 100 - (bars / klines.length) * 100
      const defaultZoomEnd = 100
      const useDefaultZoom = lastChartUpdateRef.current === 0 || dataZoomRef.current == null
      let start = defaultZoomStart
      let end = defaultZoomEnd

      if (!useDefaultZoom) {
        const prevOpt = chartInstance.current?.getOption() as any
        const zoom = Array.isArray(prevOpt?.dataZoom) ? prevOpt.dataZoom.find((item: any) => item.xAxisIndex != null) : null
        if (zoom?.start != null && zoom?.end != null) {
          start = zoom.start
          end = zoom.end
        } else if (dataZoomRef.current) {
          start = dataZoomRef.current.start
          end = dataZoomRef.current.end
        }
      }

      dataZoomRef.current = { start, end }
      const isFirstRender = lastChartUpdateRef.current === 0

      const formatTimeLabel = (index: number) => {
        return formatXAxisLabel(klines, index)
      }

      chartInstance.current?.setOption(
        {
          backgroundColor: '#0b0e11',
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
              const nextIndex = resolveDataIndex(params, chartData.times)
              if (nextIndex != null) {
                hoveredDataIndexRef.current = nextIndex
                isHoveringRef.current = true
                setHoveredIndex((current) => (current === nextIndex ? current : nextIndex))
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
            { left: 54, right: 72, top: 38, height: 348 },
            { left: 54, right: 72, top: 402, height: 82 },
          ],
          xAxis: [
            {
              type: 'category',
              data: times,
              boundaryGap: true,
              axisLine: { lineStyle: { color: '#2b3139' } },
              axisLabel: {
                show: false,
              },
              splitNumber: 7,
              axisPointer: {
                label: {
                  formatter: (params: any) => {
                    const index = resolveDataIndex({ axesInfo: [params] }, times)
                    if (index == null) return ''
                    return formatDateTime(klines[index]?.openTime ?? 0)
                  },
                },
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
              axisLabel: {
                color: '#848e9c',
                fontSize: 11,
                hideOverlap: true,
                margin: 14,
                formatter: (_value: string, index: number) => formatTimeLabel(index),
              },
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
              axisLabel: { color: '#848e9c', fontSize: 11, formatter: (value: number) => Number(value).toFixed(priceDigits) },
              axisPointer: {
                label: {
                  formatter: ({ value }: { value: number }) => formatPrice(Number(value), priceDigits),
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
                formatter: (value: number) => (value >= 1e6 ? `${(value / 1e6).toFixed(1)}M` : value >= 1e3 ? `${(value / 1e3).toFixed(1)}K` : `${Math.round(value)}`),
              },
              splitLine: { show: false },
            },
          ],
          dataZoom: [
            { type: 'inside', xAxisIndex: [0, 1], start, end },
            {
              type: 'slider',
              xAxisIndex: [0, 1],
              bottom: 8,
              height: 18,
              start,
              end,
              brushSelect: false,
              fillerColor: 'rgba(132, 142, 156, 0.12)',
              borderColor: '#2b3139',
              handleStyle: { color: '#6b7280', borderColor: '#6b7280' },
              moveHandleStyle: { color: '#6b7280' },
              textStyle: { color: '#6b7280', fontSize: 10 },
            },
          ],
          graphic: [
            buildWatermarkGraphic(),
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
                  formatter: () => `${(klines[klines.length - 1]?.close ?? 0).toFixed(priceDigits)}`,
                  color: '#ffffff',
                  backgroundColor: (klines[klines.length - 1]?.close ?? 0) >= (klines[klines.length - 1]?.open ?? 0) ? upColor : downColor,
                  padding: [3, 6],
                  borderRadius: 3,
                  distance: 6,
                },
                lineStyle: { color: '#848e9c', type: 'dashed', width: 1, opacity: 0.5 },
                data: [{ yAxis: klines[klines.length - 1]?.close ?? 0 }],
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
              data: volumes,
              barWidth: '60%',
              itemStyle: {
                color: (params: any) => {
                  const index = params?.dataIndex ?? 0
                  const current = klines[index]
                  return current && (current.close ?? 0) >= (current.open ?? 0) ? upColor : downColor
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
        false,
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
  }, [chartData, chartUpdateTrigger, klines, streaming, visibleBars])

  useEffect(() => {
    const chart = chartInstance.current
    const onResize = () => chart?.resize()
    window.addEventListener('resize', onResize)
    return () => {
      if (renderRaf.current) cancelAnimationFrame(renderRaf.current)
      if (chart && updateAxisPointerHandlerRef.current) {
        chart.off('updateAxisPointer', updateAxisPointerHandlerRef.current)
      }
      if (chart && dataZoomHandlerRef.current) {
        chart.off('datazoom', dataZoomHandlerRef.current)
      }
      if (chart && globalOutHandlerRef.current) {
        chart.getZr().off('globalout', globalOutHandlerRef.current)
      }
      window.removeEventListener('resize', onResize)
      chart?.dispose()
      chartInstance.current = null
      updateAxisPointerHandlerRef.current = null
      dataZoomHandlerRef.current = null
      globalOutHandlerRef.current = null
    }
  }, [])

  return (
    <div
      style={{
        position: 'relative',
        width: '100%',
        minHeight: 520,
        height: 520,
        background: '#0b0e11',
        border: '1px solid #1e2329',
        borderRadius: 6,
        overflow: 'hidden',
      }}
    >
      <div style={overlayRowStyle(8, 54)}>
        <span style={textStyle('#848e9c', 400)}>{displayMetrics.timeText}</span>
        <span style={textStyle('#eaecef')}>{`开 ${formatPrice(displayMetrics.open, displayMetrics.priceDigits)}`}</span>
        <span style={textStyle('#eaecef')}>{`高 ${formatPrice(displayMetrics.high, displayMetrics.priceDigits)}`}</span>
        <span style={textStyle('#eaecef')}>{`低 ${formatPrice(displayMetrics.low, displayMetrics.priceDigits)}`}</span>
        <span style={textStyle(displayMetrics.changeColor, 600)}>{`收 ${formatPrice(displayMetrics.close, displayMetrics.priceDigits)}`}</span>
        <span style={textStyle(displayMetrics.changeColor, 600)}>{`涨跌 ${displayMetrics.change >= 0 ? '+' : ''}${displayMetrics.change.toFixed(2)}%`}</span>
        <span style={textStyle('#848e9c')}>{`振幅 ${displayMetrics.amplitude.toFixed(2)}%`}</span>
      </div>
      <div style={overlayRowStyle(24, 54)}>
        <span style={textStyle('#f0b90b')}>{`MA(7) ${displayMetrics.ma7 != null ? formatPrice(displayMetrics.ma7, displayMetrics.priceDigits) : '-'}`}</span>
        <span style={textStyle('#d946ef')}>{`MA(25) ${displayMetrics.ma25 != null ? formatPrice(displayMetrics.ma25, displayMetrics.priceDigits) : '-'}`}</span>
        <span style={textStyle('#7c3aed')}>{`MA(99) ${displayMetrics.ma99 != null ? formatPrice(displayMetrics.ma99, displayMetrics.priceDigits) : '-'}`}</span>
      </div>
      <div style={overlayRowStyle(398, 54)}>
        <span style={textStyle('#848e9c')}>{`Vol(BTC) ${formatVolume(displayMetrics.volume)}`}</span>
        <span style={textStyle('#848e9c')}>{`Vol(USDT) ${formatVolume(displayMetrics.volume * displayMetrics.close)}`}</span>
        <span style={textStyle('#38bdf8')}>{`MA5 ${displayMetrics.volMA5 != null ? formatVolume(displayMetrics.volMA5) : '-'}`}</span>
        <span style={textStyle('#fb7185')}>{`MA10 ${displayMetrics.volMA10 != null ? formatVolume(displayMetrics.volMA10) : '-'}`}</span>
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
      volMA5: calcVolMA(5, klines),
      volMA10: calcVolMA(10, klines),
    }
}

function getDisplayMetrics(klines: KLine[], chartData: ReturnType<typeof buildChartData>, index: number) {
  const current = klines[index] ?? klines[klines.length - 1]
  const open = current?.open ?? 0
  const close = current?.close ?? 0
  const high = current?.high ?? 0
  const low = current?.low ?? 0
  const volume = current?.volume ?? 0
  const change = open ? ((close - open) / open) * 100 : 0
  const amplitude = open ? ((high - low) / open) * 100 : 0
  const changeColor = close >= open ? upColor : downColor
  const priceDigits = close < 1 ? 6 : close < 100 ? 4 : 2
  const timeText = chartData.times[index]?.replace(/-/g, '/') ?? ''
  const ma7 = chartData.ma7[index]
  const ma25 = chartData.ma25[index]
  const ma99 = chartData.ma99[index]
  const volMA5 = chartData.volMA5[index]
  const volMA10 = chartData.volMA10[index]

  return {
    timeText,
    open,
    close,
    high,
    low,
    volume,
    change,
    amplitude,
    changeColor,
    priceDigits,
    ma7,
    ma25,
    ma99,
    volMA5,
    volMA10,
  }
}

function buildWatermarkGraphic() {
  return {
    id: 'watermark',
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
  }
}

function overlayRowStyle(top: number, left: number): CSSProperties {
  return {
    position: 'absolute',
    top,
    left,
    right: 72,
    display: 'flex',
    gap: 16,
    alignItems: 'center',
    flexWrap: 'wrap',
    pointerEvents: 'none',
    zIndex: 2,
  }
}

function textStyle(color: string, fontWeight: number = 500): CSSProperties {
  return {
    color,
    fontSize: 11,
    fontWeight,
    lineHeight: '16px',
    whiteSpace: 'nowrap',
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
