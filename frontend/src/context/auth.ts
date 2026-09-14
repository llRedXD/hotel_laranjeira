import { createContext } from 'react'

export type AuthUser = {
  id: string
  email: string
  name: string
  role: 'admin' | 'reception'
}

export type AuthContextValue = {
  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  isLoading: boolean
  user: AuthUser | null
}

export const AuthContext = createContext<AuthContextValue | undefined>(undefined)
