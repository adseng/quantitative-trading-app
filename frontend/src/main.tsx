import { createRoot } from 'react-dom/client'
import { App as AntdApp, ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN.js'
import { ErrorBoundary } from 'react-error-boundary'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import Router from '@/router'
import './style.css'

dayjs.locale('zh-cn')

const theme = {
  token: {
    colorText: '#242f57',
    controlHeight: 40,
  },
}

const container = document.getElementById('root')
const root = createRoot(container!)

root.render(
  <ConfigProvider theme={theme} locale={zhCN}>
    <AntdApp>
      <ErrorBoundary fallback={<div className="p-4">加载出错</div>}>
        <Router />
      </ErrorBoundary>
    </AntdApp>
  </ConfigProvider>
)
