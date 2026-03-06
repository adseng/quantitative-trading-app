import { Table, Tag } from 'antd'

import type { CozeResult, CozeScenario } from '../types'

const scenarioOrder = ['LONG', 'SHORT', 'SIDEWAYS'] as const

function sortScenarios(list: CozeScenario[] | undefined) {
  if (!list?.length) return []
  return [...list].sort((left, right) => scenarioOrder.indexOf(left.direction as any) - scenarioOrder.indexOf(right.direction as any))
}

const scenarioColumns = [
  { title: '方向', dataIndex: 'direction', key: 'direction', width: 80 },
  { title: '概率', dataIndex: 'probability', key: 'probability', width: 60, render: (value: number) => `${value}%` },
  { title: '入场', dataIndex: 'entry_price', key: 'entry_price', width: 90, render: (value: number | null) => (value != null ? value.toFixed(2) : '-') },
  { title: '止损', dataIndex: 'stop_loss', key: 'stop_loss', width: 90, render: (value: number | null) => (value != null ? value.toFixed(2) : '-') },
  { title: '止盈1', dataIndex: 'take_profit_1', key: 'take_profit_1', width: 90, render: (value: number | null) => (value != null ? value.toFixed(2) : '-') },
  { title: '止盈2', dataIndex: 'take_profit_2', key: 'take_profit_2', width: 90, render: (value: number | null) => (value != null ? value.toFixed(2) : '-') },
  { title: '风报比', dataIndex: 'risk_reward_ratio', key: 'risk_reward_ratio', width: 70, render: (value: number | null) => (value != null ? value.toFixed(1) : '-') },
  {
    title: '动作',
    dataIndex: 'action',
    key: 'action',
    width: 80,
    render: (action: string) => {
      const config: Record<string, { color: string }> = {
        EXECUTE: { color: 'green' },
        WAIT: { color: 'orange' },
        SKIP: { color: 'default' },
      }
      return <Tag color={(config[action] ?? { color: 'default' }).color}>{action || '-'}</Tag>
    },
  },
  { title: '触发条件', dataIndex: 'trigger_condition', key: 'trigger_condition', ellipsis: true },
]

interface PredictionResultsProps {
  results: CozeResult[]
}

export function PredictionResults({ results }: PredictionResultsProps) {
  if (results.length === 0) {
    return <div className="text-gray-500 py-8 text-center">暂无预测结果</div>
  }

  return (
    <div className="space-y-4">
      {results.map((result, index) => (
        <div key={`${result.timestamp}-${index}`} className="border rounded p-3 bg-gray-50/50">
          <div className="flex gap-4 mb-2 text-sm">
            <span>{result.timestamp}</span>
            <span>{result.symbol}</span>
            <span>当前价: {result.current_price?.toFixed(2)}</span>
            <span className="text-gray-600">{result.market_structure || result.resultType}</span>
          </div>
          {result.rawAnswer && !result.scenarios?.length ? (
            <pre className="text-xs overflow-auto max-h-32 whitespace-pre-wrap">{result.rawAnswer}</pre>
          ) : (
            <Table
              dataSource={sortScenarios(result.scenarios)}
              columns={scenarioColumns}
              rowKey="direction"
              size="small"
              pagination={false}
            />
          )}
        </div>
      ))}
    </div>
  )
}
