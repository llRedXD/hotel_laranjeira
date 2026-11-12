import { Navigate, Route, Routes } from 'react-router-dom'
import { DashboardLayout } from './components/DashboardLayout'
import { Layout } from './components/Layout'
import { ProtectedRoute } from './components/ProtectedRoute'
import { PublicOnlyRoute } from './components/PublicOnlyRoute'
import { DashboardPage } from './pages/DashboardPage'
import { LoginPage } from './pages/LoginPage'
import { PasswordRecoveryPage } from './pages/PasswordRecoveryPage'
import { PasswordResetPage } from './pages/PasswordResetPage'
import { RecordsPage } from './pages/RecordsPage'
import { ReservationPage } from './pages/ReservationPage'
import { ReceptionReservationsPage } from './pages/ReceptionReservationsPage'
import { ReservationsListPage } from './pages/ReservationsListPage'
import { ReservationAuditPage } from './pages/ReservationAuditPage'
import { ReservationEditPage } from './pages/ReservationEditPage'

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Navigate to="/reservar" replace />} />
        <Route path="/reservar" element={<ReservationPage />} />
        <Route element={<PublicOnlyRoute />}>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/recuperar-senha" element={<PasswordRecoveryPage />} />
          <Route path="/redefinir-senha" element={<PasswordResetPage />} />
          <Route path="/cadastro" element={<Navigate to="/login" replace />} />
        </Route>
      </Route>
      <Route element={<ProtectedRoute />}>
        <Route element={<DashboardLayout />}>
          <Route path="/painel" element={<DashboardPage />} />
          <Route path="/reservas/nova" element={<ReceptionReservationsPage />} />
          <Route path="/hospedes" element={<RecordsPage key="guests" resource="guests" />} />
          <Route path="/categorias" element={<RecordsPage key="categories" resource="categories" />} />
          <Route path="/quartos" element={<RecordsPage key="rooms" resource="rooms" />} />
          <Route path="/reservas" element={<ReservationsListPage />} />
          <Route path="/auditoria" element={<ReservationAuditPage />} />
          <Route path="/reservas/:id/editar" element={<ReservationEditPage />} />
        </Route>
      </Route>
    </Routes>
  )
}
