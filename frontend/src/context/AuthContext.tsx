import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { currentUser, login as loginRequest, logout as logoutRequest, refreshSession } from '../services/api'
import { AuthContext, type AuthContextValue, type AuthUser } from './auth'

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    async function restoreSession() {
      try {
        const session = await refreshSession()
        const profile = await currentUser(session.accessToken)
        setUser(profile.user)
      } catch {
        setUser(null)
      } finally {
        setIsLoading(false)
      }
    }
    void restoreSession()
  }, [])

  const value = useMemo<AuthContextValue>(() => ({
    user,
    isLoading,
    login: async (email, password) => {
      const session = await loginRequest(email, password)
      setUser(session.user)
    },
    logout: async () => {
      try {
        await logoutRequest()
      } finally {
        setUser(null)
      }
    },
  }), [isLoading, user])

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
