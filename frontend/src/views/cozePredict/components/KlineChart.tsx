import { useEffect, useMemo, useRef, useState } from 'react'
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
  const [chartUpdateTrigger, setChartUpdateTrigger] = useState(0)

  const chartData = useMemo(() => {
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
  }, [klines])

  useEffect(() => {
    dataZoomRef.current = null
  }, [visibleBars])

  useEffect(() => {
    if (klines.length === 0 || !chartRef.current) return
    if (!chartInstance.current) chartInstance.current = echarts.init(chartRef.current)

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
      const last = klines[klines.length - 1]
      const lastPrice = last?.close ?? 0
      const priceDigits = lastPrice < 1 ? 6 : lastPrice < 100 ? 4 : 2
      const lastOpen = last?.open ?? 0
      const lastChange = lastOpen ? ((lastPrice - lastOpen) / lastOpen) * 100 : 0
      const lastChangeColor = lastPrice >= lastOpen ? upColor : downColor
      const lastMA7 = typeof ma7[ma7.length - 1] === 'number' ? (ma7[ma7.length - 1] as number) : null
      const lastMA25 = typeof ma25[ma25.length - 1] === 'number' ? (ma25[ma25.length - 1] as number) : null
      const lastMA99 = typeof ma99[ma99.length - 1] === 'number' ? (ma99[ma99.length - 1] as number) : null

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
        const timestamp = klines[index]?.openTime
        if (timestamp == null) return ''
        const date = new Date(timestamp)
        const hour = date.getUTCHours()
        const minute = date.getUTCMinutes()
        return `${hour.toString().padStart(2, '0')}:${minute.toString().padStart(2, '0')}`
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
              const point = params?.find((item) => item?.seriesType === 'candlestick') ?? params?.[0]
              const index = point?.dataIndex
              if (index == null || !klines[index]) return ''
              const current = klines[index]
              const open = current.open ?? 0
              const close = current.close ?? 0
              const high = current.high ?? 0
              const low = current.low ?? 0
              const volume = current.volume ?? 0
              const change = open ? ((close - open) / open) * 100 : 0
              const amplitude = open ? ((high - low) / open) * 100 : 0
              const color = close >= open ? upColor : downColor
              const timeText = times[index].replace(/-/g, '/')
              const lines: string[] = [
                `<div style="font-weight:600;margin-bottom:6px;">${timeText}</div>`,
                `开 ${formatPrice(open)}  高 ${formatPrice(high)}`,
                `低 ${formatPrice(low)}  收 <span style="color:${color};font-weight:600">${formatPrice(close)}</span>`,
                `涨跌幅 <span style="color:${color};font-weight:600">${change >= 0 ? '+' : ''}${change.toFixed(2)}%</span>  振幅 ${amplitude.toFixed(2)}%`,
                `MA(7) <span style="color:#f0b90b">${ma7[index] != null ? formatPrice(ma7[index] as number) : '-'}</span>  MA(25) <span style="color:#c994ff">${ma25[index] != null ? formatPrice(ma25[index] as number) : '-'}</span>  MA(99) <span style="color:#a78bfa">${ma99[index] != null ? formatPrice(ma99[index] as number) : '-'}</span>`,
                `Vol(BTC) ${formatVolume(volume)}  Vol(USDT) ${formatVolume(volume * close)}`,
              ]
              if (volMA5[index] != null || volMA10[index] != null) {
                lines.push(`Vol MA5 ${formatVolume((volMA5[index] as number) ?? 0)}  Vol MA10 ${formatVolume((volMA10[index] as number) ?? 0)}`)
              }
              return lines.join('<br/>')
            },
          },
          grid: [
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
                formatter: (_value: string, index: number) => formatTimeLabel(index),
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
              axisLabel: { color: '#848e9c', formatter: (value: number) => Number(value).toFixed(priceDigits) },
              splitLine: { lineStyle: { color: '#1e2329' } },
            },
            {
              gridIndex: 1,
              scale: true,
              position: 'right',
              axisLine: { lineStyle: { color: '#2b3139' } },
              axisLabel: {
                color: '#848e9c',
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
                    text: `涨跌幅 ${lastChange >= 0 ? '+' : ''}${lastChange.toFixed(2)}%  振幅 ${lastOpen ? ((((last?.high ?? 0) - (last?.low ?? 0)) / lastOpen) * 100).toFixed(2) : '0'}%`,
                    fill: lastChangeColor,
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
                    text: `最新 ${formatPrice(lastPrice)}  (${lastChange >= 0 ? '+' : ''}${lastChange.toFixed(2)}%)`,
                    fill: lastChangeColor,
                    fontSize: 12,
                    fontWeight: 600,
                  },
                },
              ],
            },
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
                    text: `Vol(BTC) ${formatVolume(last?.volume ?? 0)}  Vol(USDT) ${formatVolume((last?.volume ?? 0) * (last?.close ?? 0))}`,
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
        !isFirstRender,
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
      window.removeEventListener('resize', onResize)
      chart?.dispose()
      chartInstance.current = null
    }
  }, [])

  return (
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
  )
}
