import { ROUTES } from '@/router/settings'
import { ExperimentOutlined, InfoCircleOutlined, LineChartOutlined, SettingOutlined } from '@ant-design/icons'

export const MENUS = [
  {
    label: '因子分析',
    value: ROUTES.FACTOR,
    icon: <LineChartOutlined className="text-[20px]" />,
  },
  {
    label: '批量测试',
    value: ROUTES.BATCH_TEST,
    icon: <ExperimentOutlined className="text-[20px]" />,
  },
  {
    label: '配置',
    value: ROUTES.SETTINGS,
    icon: <SettingOutlined className="text-[20px]" />,
  },
  {
    label: '关于',
    value: ROUTES.ABOUT,
    icon: <InfoCircleOutlined className="text-[20px]" />,
  },
]
