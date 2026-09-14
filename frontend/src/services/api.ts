const apiBaseURL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api/v1'

export type AuthUser = {
  id: string
  name: string
  email: string
  role: 'admin' | 'reception'
}

type APIErrorBody = {
  error?: {
    code?: string
    message?: string
    fields?: Record<string, string>
  }
}

export class APIError extends Error {
  fields: Record<string, string>

  constructor(message: string, fields: Record<string, string> = {}) {
    super(message)
    this.fields = fields
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${apiBaseURL}${path}`, {
    ...options,
    credentials: 'include',
    headers: {
      ...(options.body ? { 'Content-Type': 'application/json' } : {}),
      ...options.headers,
    },
  })
  if (!response.ok) {
    const body = await response.json().catch(() => ({})) as APIErrorBody
    throw new APIError(body.error?.message ?? 'Não foi possível concluir sua solicitação.', body.error?.fields)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

export function login(email: string, password: string) {
  return request<{ user: AuthUser, accessToken: string }>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
}

export function refreshSession() {
  return request<{ user: AuthUser, accessToken: string }>('/auth/refresh', { method: 'POST' })
}

export function logout() {
  return request<void>('/auth/logout', { method: 'POST' })
}

export function currentUser(accessToken: string) {
  return request<{ user: AuthUser }>('/auth/me', {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

export function requestPasswordRecovery(email: string) {
  return request<{ message: string }>('/auth/password-recovery', {
    method: 'POST',
    body: JSON.stringify({ email }),
  })
}

export function resetPassword(token: string, password: string) {
  return request<void>('/auth/password-reset', {
    method: 'POST',
    body: JSON.stringify({ token, password }),
  })
}
