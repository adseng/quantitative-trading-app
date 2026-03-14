import { ROUTES } from '@/router/settings'
import { LineChartOutlined } from '@ant-design/icons'

export const MENUS = [
  {
    label: 'EMA 趋势回踩',
    value: ROUTES.EMA_TREND_PULLBACK,
    icon: <LineChartOutlined className="text-[20px]" />,
  },
  {
    label: '箱体突破回踩',
    value: ROUTES.BOX_PULLBACK,
    icon: <LineChartOutlined className="text-[20px]" />,
  },
]
