import { ROUTES } from '@/router/settings'
import Layout from '@/structure/Layout'
import About from '@/views/about/About'
import BatchTest from '@/views/batchTest/BatchTest'
import FactorAnalysis from '@/views/emotion/Emotion'
import Settings from '@/views/settings/Settings'
import { BrowserRouter, Route, Routes } from 'react-router-dom'

export default function Router() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path={ROUTES.ROOT} element={<Layout />}>
          <Route index element={<FactorAnalysis />} />
          <Route path={ROUTES.BATCH_TEST} element={<BatchTest />} />
          <Route path={ROUTES.SETTINGS} element={<Settings />} />
          <Route path={ROUTES.ABOUT} element={<About />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
