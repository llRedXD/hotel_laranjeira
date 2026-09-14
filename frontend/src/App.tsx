import { Navigate, Route, Routes } from 'react-router-dom'
import { DashboardLayout } from './components/DashboardLayout'
import { Layout } from './components/Layout'
import { ProtectedRoute } from './components/ProtectedRoute'
import { PublicOnlyRoute } from './components/PublicOnlyRoute'
import { DashboardPage } from './pages/DashboardPage'
import { HomePage } from './pages/HomePage'
import { LoginPage } from './pages/LoginPage'
import { PasswordRecoveryPage } from './pages/PasswordRecoveryPage'
import { PasswordResetPage } from './pages/PasswordResetPage'

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<HomePage />} />
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
        </Route>
      </Route>
    </Routes>
  )
}
