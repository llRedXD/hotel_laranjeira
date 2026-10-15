import { publicRequest } from './api'

export type AvailabilityOption = {
  roomId: string
  number: string
  categoryName: string
  capacity: number
  dailyRate: string
  nights: number
  estimatedTotal: string
}

export type ReservationGuestInput = {
  name: string
  documentType: 'RG' | 'CPF'
  document: string
  email: string
}

export type PublicReservationInput = {
  roomId: string
  checkIn: string
  checkOut: string
  occupants: number
  expectedDailyRate: string
  guest: ReservationGuestInput
}

export type ReservationConfirmation = {
  id: string
  roomId: string
  checkIn: string
  checkOut: string
  occupants: number
  dailyRate: string
  totalAmount: string
}

export type StaffReservationInput = Omit<PublicReservationInput, 'guest'> & { guestId: string }

export function formatBRL(value: string) {
  const [whole, fraction = '00'] = value.split('.')
  return `R$ ${BigInt(whole).toLocaleString('pt-BR')},${fraction.padEnd(2, '0')}`
}

export function localDateString(date: Date) {
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${date.getFullYear()}-${month}-${day}`
}

export function addCalendarDays(value: string, days: number) {
  if (!value) return ''
  const [year, month, day] = value.split('-').map(Number)
  const date = new Date(year, month - 1, day)
  date.setDate(date.getDate() + days)
  return localDateString(date)
}

export function formatReservationDate(value: string) {
  const [year, month, day] = value.split('-').map(Number)
  return new Intl.DateTimeFormat('pt-BR').format(new Date(year, month - 1, day))
}

export async function searchAvailability(input: { checkIn: string, checkOut: string, occupants: number }) {
  const query = new URLSearchParams({ checkIn: input.checkIn, checkOut: input.checkOut, occupants: String(input.occupants) })
  const result = await publicRequest<{ rooms: AvailabilityOption[] }>(`/public/availability?${query}`)
  return result.rooms
}

export async function createPublicReservation(input: PublicReservationInput) {
  const result = await publicRequest<{ reservation: ReservationConfirmation }>('/public/reservations', {
    method: 'POST', body: JSON.stringify(input),
  })
  return result.reservation
}
