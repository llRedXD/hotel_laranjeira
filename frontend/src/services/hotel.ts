import { authenticatedRequest } from './api'
import type { AvailabilityOption, StaffReservationInput } from './reservations'

export type Resource = 'guests' | 'categories' | 'rooms'
export type Entity = {
  id: string
  name?: string
  documentType?: 'RG' | 'CPF' | null
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
export type ReservationStatus = 'confirmed' | 'in_progress' | 'departure_today' | 'ended' | 'cancelled'
export type ReceptionReservation = {
  roomId: string
  guestId: string
  status: ReservationStatus
  version: number
  canEdit: boolean
  canCancel: boolean
  id: string
  checkIn: string
  checkOut: string
  occupants: number
  dailyRate: string
  totalAmount: string
  guestName: string
  roomNumber: string
  categoryName: string
}
export type ReservationSummary = { date: string, guestsInHouse: number, checkInsToday: number, availableRooms: number }
export type ReservationAuditState = {
  status: ReservationStatus, roomId: string, roomNumber: string, checkIn: string, checkOut: string,
  occupants: number, dailyRate: string, totalAmount: string, version: number,
}
export type ReservationEvent = {
  id: string, reservationId: string, action: 'imported' | 'created' | 'cancelled' | 'updated',
  source: 'import' | 'public' | 'staff' | 'database', actorId: string | null, actorName: string | null,
  guestName: string, roomNumber: string, occurredAt: string,
  beforeState: ReservationAuditState | null, afterState: ReservationAuditState,
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
export async function searchReservationAvailability(input: { checkIn: string, checkOut: string, occupants: number, reservationId?: string }) {
  const query = new URLSearchParams({ checkIn: input.checkIn, checkOut: input.checkOut, occupants: String(input.occupants) })
  if (input.reservationId) query.set('reservationId', input.reservationId)
  const result = await authenticatedRequest<{ rooms: AvailabilityOption[] }>(`/reservations/availability?${query}`)
  return result.rooms
}
export async function createStaffReservation(input: StaffReservationInput) {
  return authenticatedRequest<{ reservation: {
    id: string, guestId: string, roomId: string, checkIn: string, checkOut: string,
    occupants: number, dailyRate: string, totalAmount: string, guestName: string,
  } }>('/reservations', { method: 'POST', body: JSON.stringify(input) })
}
export async function listReservations(query: URLSearchParams) {
  return authenticatedRequest<{ reservations: ReceptionReservation[], total: number }>(`/reservations?${query}`)
}
export async function getReservationSummary(date?: string) {
  const query = new URLSearchParams(date ? { date } : {})
  const result = await authenticatedRequest<{ summary: ReservationSummary }>(`/reservations/summary?${query}`)
  return result.summary
}
export async function cancelReservation(id: string, version: number) {
  return authenticatedRequest<{ reservation: ReceptionReservation }>(`/reservations/${encodeURIComponent(id)}/cancel`, {
    method: 'POST', body: JSON.stringify({ version }),
  })
}
export async function getReservation(id: string) {
  return authenticatedRequest<{ reservation: ReceptionReservation, hotelDate: string }>(`/reservations/${encodeURIComponent(id)}`)
}
export async function updateReservation(id: string, input: {
  roomId: string, checkIn: string, checkOut: string, occupants: number, expectedDailyRate: string, version: number,
}) {
  return authenticatedRequest<{ reservation: ReceptionReservation }>(`/reservations/${encodeURIComponent(id)}`, {
    method: 'PATCH', body: JSON.stringify(input),
  })
}
export async function listReservationEvents(query: URLSearchParams) {
  return authenticatedRequest<{ events: ReservationEvent[], total: number }>(`/reservation-events?${query}`)
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
