import { Table, Tag } from 'antd'

import type { BacktestTrade } from '../types'

interface PredictionResultsProps {
  results: BacktestTrade[]
}

const columns = [
  {
    title: '方向',
    dataIndex: 'direction',
    key: 'direction',
    width: 86,
    render: (value: string) => <Tag color={value === 'LONG' ? 'green' : 'volcano'}>{value}</Tag>,
  },
  { title: '开仓时间', dataIndex: 'entryTime', key: 'entryTime', width: 150, render: (value: number) => formatDateTime(value) },
  { title: '平仓时间', dataIndex: 'exitTime', key: 'exitTime', width: 150, render: (value: number) => formatDateTime(value) },
  { title: '开仓价', dataIndex: 'entryPrice', key: 'entryPrice', width: 96, render: formatPrice },
  { title: '平仓价', dataIndex: 'exitPrice', key: 'exitPrice', width: 96, render: formatPrice },
  { title: '止损', dataIndex: 'stopLoss', key: 'stopLoss', width: 96, render: formatPrice },
  { title: '止盈', dataIndex: 'takeProfit', key: 'takeProfit', width: 96, render: formatPrice },
  { title: '盈亏', dataIndex: 'pnl', key: 'pnl', width: 96, render: formatPnl },
  { title: '盈亏%', dataIndex: 'pnlPercent', key: 'pnlPercent', width: 86, render: (value: number) => `${value.toFixed(2)}%` },
  { title: '持仓K线', dataIndex: 'holdBars', key: 'holdBars', width: 90 },
  { title: '原因', dataIndex: 'exitReason', key: 'exitReason', width: 180 },
  { title: '余额', dataIndex: 'balanceAfter', key: 'balanceAfter', width: 108, render: formatPrice },
]

export function PredictionResults({ results }: PredictionResultsProps) {
  if (results.length === 0) {
    return <div className="text-gray-500 py-8 text-center">暂无订单结果</div>
  }

  return <Table rowKey="id" dataSource={results} columns={columns} size="small" scroll={{ x: 1300 }} pagination={{ pageSize: 10 }} />
}

function formatDateTime(value: number): string {
  if (!value) return '-'
  const date = new Date(value)
  const year = date.getFullYear()
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  const hours = `${date.getHours()}`.padStart(2, '0')
  const minutes = `${date.getMinutes()}`.padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}

function formatPrice(value: number): string {
  return Number(value || 0).toFixed(2)
}

function formatPnl(value: number): JSX.Element {
  const color = value >= 0 ? '#16a34a' : '#dc2626'
  return <span style={{ color }}>{value.toFixed(2)}</span>
}
