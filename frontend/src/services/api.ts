const apiBaseURL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api/v1'
let accessToken: string | null = null
type Session = { user: AuthUser, accessToken: string }
let refreshPending: Promise<Session> | null = null

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
  status: number
  fields: Record<string, string>

  constructor(message: string, fields: Record<string, string> = {}, status = 0) {
    super(message)
    this.fields = fields
    this.status = status
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
    throw new APIError(body.error?.message ?? 'Não foi possível concluir sua solicitação.', body.error?.fields, response.status)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

export async function login(email: string, password: string) {
  const session = await request<Session>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
  accessToken = session.accessToken
  return session
}

export function refreshSession() {
  if (!refreshPending) {
    refreshPending = request<Session>('/auth/refresh', { method: 'POST' })
      .then(session => { accessToken = session.accessToken; return session })
      .catch((error: unknown) => { accessToken = null; throw error })
      .finally(() => { refreshPending = null })
  }
  return refreshPending
}

export async function logout() {
  if (refreshPending) await refreshPending.catch(() => undefined)
  await request<void>('/auth/logout', { method: 'POST' })
  accessToken = null
}

export async function authenticatedRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
  if (!accessToken) await refreshSession()
  const send = () => request<T>(path, { ...options, headers: { ...options.headers, Authorization: `Bearer ${accessToken}` } })
  try { return await send() } catch (error) {
    if (!(error instanceof APIError) || error.status !== 401) throw error
    await refreshSession()
    return send()
  }
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
