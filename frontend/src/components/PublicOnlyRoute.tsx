import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'

export function PublicOnlyRoute() {
  const { isLoading, user } = useAuth()
  if (isLoading) return null
  return user ? <Navigate to="/painel" replace /> : <Outlet />
}
