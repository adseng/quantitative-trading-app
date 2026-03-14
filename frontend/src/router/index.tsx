import { ROUTES } from '@/router/settings'
import Layout from '@/structure/Layout'
import CozePredict from '@/views/cozePredict/CozePredict'
import BoxRangeReversal from '@/views/boxRangeReversal/BoxRangeReversal'
import EMATrendPullback from '@/views/emaTrendPullback/EMATrendPullback'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'

export default function Router() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path={ROUTES.ROOT} element={<Layout />}>
          <Route index element={<Navigate to={ROUTES.EMA_TREND_PULLBACK} replace />} />
          <Route path={ROUTES.EMA_TREND_PULLBACK.slice(1)} element={<EMATrendPullback />} />
          <Route path={ROUTES.BOX_RANGE_REVERSAL.slice(1)} element={<BoxRangeReversal />} />
          <Route path={ROUTES.BOX_PULLBACK.slice(1)} element={<CozePredict />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
