import { authenticatedRequest } from './api'

export type Resource = 'guests' | 'categories' | 'rooms'
export type Entity = {
  id: string
  name?: string
  document?: string | null
  email?: string | null
  phone?: string | null
  address?: string | null
  description?: string | null
  number?: string
  categoryId?: string
  categoryName?: string
  capacity?: number
  dailyRate?: string
  isAvailable?: boolean
}
const paths: Record<Resource, string> = { guests: '/guests', categories: '/room-categories', rooms: '/rooms' }
export const pageSize = 20
export async function listEntities(resource: Resource, query: URLSearchParams): Promise<Entity[]> {
  const result = await authenticatedRequest<{ guests?: Entity[], categories?: Entity[], rooms?: Entity[] }>(`${paths[resource]}?${query}`)
  return result[resource] ?? []
}
export async function getEntity(resource: Resource, id: string): Promise<Entity> {
  const result = await authenticatedRequest<{ guest?: Entity, room?: Entity }>(`${paths[resource]}/${id}`)
  const entity = result.guest ?? result.room
  if (!entity) throw new Error('Registro não encontrado.')
  return entity
}
export async function saveEntity(resource: Resource, id: string | undefined, values: Record<string, string>) {
  const payload: Record<string, string | number | boolean | null> = { ...values }
  if (resource === 'rooms') {
    payload.capacity = Number(values.capacity)
    payload.dailyRate = values.dailyRate.replace(',', '.')
    payload.isAvailable = values.isAvailable === 'true'
  } else {
    for (const key of Object.keys(payload)) if (key !== 'name' && payload[key] === '') payload[key] = null
  }
  await authenticatedRequest(`${paths[resource]}${id ? '/'+id : ''}`, { method: id ? 'PATCH' : 'POST', body: JSON.stringify(payload) })
}
export async function changeAvailability(id: string, available: boolean) {
  await authenticatedRequest(`/rooms/${id}/availability`, { method: 'PATCH', body: JSON.stringify({ isAvailable: available }) })
}
export async function allCategories(): Promise<Entity[]> {
  return listAll('categories')
}
export async function listAll(resource: Resource): Promise<Entity[]> {
  const result: Entity[] = []
  for (let offset = 0; offset <= 100000; offset += 100) {
    const page = await listEntities(resource, new URLSearchParams({ limit: '100', offset: String(offset) }))
    result.push(...page)
    if (page.length < 100) break
  }
  return result
}
