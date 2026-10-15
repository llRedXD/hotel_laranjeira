import { useState, type FormEvent } from 'react'
import { EntityDialog } from '../components/EntityDialog'
import { ReservationRoomOptions } from '../components/ReservationRoomOptions'
import { ReservationSearchForm } from '../components/ReservationSearchForm'
import { APIError } from '../services/api'
import { createStaffReservation, getEntity, listEntities, searchReservationAvailability, type Entity } from '../services/hotel'
import { formatBRL, localDateString, type AvailabilityOption } from '../services/reservations'

const todayDate = new Date()
const today = localDateString(todayDate)
const tomorrowDate = new Date(todayDate)
tomorrowDate.setDate(tomorrowDate.getDate() + 1)
const tomorrow = localDateString(tomorrowDate)
const control = 'w-full rounded-xl border border-forest/20 bg-white px-4 py-3 text-sm focus:border-orange focus:outline-none focus:ring-4 focus:ring-orange/15'

export function ReceptionReservationsPage() {
  const [guestSearch, setGuestSearch] = useState('')
  const [guests, setGuests] = useState<Entity[]>([])
  const [selectedGuest, setSelectedGuest] = useState<Entity | null>(null)
  const [dates, setDates] = useState({ checkIn: today, checkOut: tomorrow, occupants: 1 })
  const [rooms, setRooms] = useState<AvailabilityOption[]>([])
  const [selectedRoom, setSelectedRoom] = useState<AvailabilityOption | null>(null)
  const [editGuest, setEditGuest] = useState(false)
  const [loadingGuests, setLoadingGuests] = useState(false)
  const [busy, setBusy] = useState(false)
  const [searched, setSearched] = useState(false)
  const [success, setSuccess] = useState('')
  const [error, setError] = useState('')

  async function handleGuestSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setLoadingGuests(true)
    setError('')
    try {
      const query = new URLSearchParams({ limit: '100', offset: '0' })
      if (guestSearch.trim()) query.set('search', guestSearch.trim())
      setGuests(await listEntities('guests', query))
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Não foi possível buscar hóspedes.')
    } finally { setLoadingGuests(false) }
  }

  async function handleAvailability(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    setSuccess('')
    setSelectedRoom(null)
    if (!selectedGuest) {
      setError('Selecione um hóspede antes de consultar os quartos.')
      return
    }
    if (!dates.checkIn || !dates.checkOut || dates.checkOut <= dates.checkIn || dates.occupants < 1) {
      setError('Informe um período válido e a quantidade total de pessoas.')
      return
    }
    if (!selectedGuest.document || !selectedGuest.documentType || !selectedGuest.email) {
      setError('Complete o tipo e número do documento e o e-mail do hóspede antes de reservar.')
      return
    }
    setBusy(true)
    try {
      setRooms(await searchReservationAvailability(dates))
      setSearched(true)
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Não foi possível consultar os quartos.')
    } finally { setBusy(false) }
  }

  async function handleConfirm() {
    if (!selectedGuest || !selectedRoom) return
    setBusy(true)
    setError('')
    setSuccess('')
    try {
      const { reservation } = await createStaffReservation({
        guestId: selectedGuest.id, roomId: selectedRoom.roomId,
        checkIn: dates.checkIn, checkOut: dates.checkOut, occupants: dates.occupants, expectedDailyRate: selectedRoom.dailyRate,
      })
      setSuccess(`Reserva ${reservation.id} confirmada para ${selectedGuest.name}, no quarto ${selectedRoom.number}, de ${reservation.checkIn} a ${reservation.checkOut}.`)
      setSelectedRoom(null)
      setSearched(false)
    } catch (cause) {
      if (cause instanceof APIError && cause.status === 409) setError('A disponibilidade ou os dados mudaram antes da confirmação. Consulte novamente para continuar.')
      else setError(cause instanceof Error ? cause.message : 'Não foi possível confirmar a reserva.')
    } finally { setBusy(false) }
  }

  async function refreshSelectedGuest() {
    if (!selectedGuest) return
    try {
      setSelectedGuest(await getEntity('guests', selectedGuest.id))
      setEditGuest(false)
      setError('')
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Não foi possível atualizar o hóspede.') }
  }

  return <div className="mx-auto max-w-6xl px-5 py-8 sm:px-8 sm:py-10">
    <header className="mb-8">
      <p className="text-sm font-semibold tracking-[0.18em] text-orange uppercase">Atendimento</p>
      <h1 className="mt-2 font-serif text-4xl font-semibold text-forest">Nova reserva</h1>
      <p className="mt-2 text-ink/65">Selecione o hóspede, consulte o período e confirme a acomodação.</p>
    </header>

    <section className="rounded-2xl border border-forest/10 bg-cream p-5 shadow-sm sm:p-7">
      <h2 className="font-serif text-2xl font-semibold text-forest">1. Selecionar hóspede</h2>
      <form onSubmit={handleGuestSearch} className="mt-4 flex flex-col gap-3 sm:flex-row sm:items-end">
        <div className="flex-1"><label htmlFor="guest-search" className="mb-1.5 block text-sm font-medium">Documento do hóspede (CPF ou RG)</label><input id="guest-search" type="search" autoComplete="off" className={control} placeholder="Digite o CPF ou RG; também é possível buscar pelo nome" value={guestSearch} onChange={event => setGuestSearch(event.target.value)} /><p className="mt-1.5 text-xs text-ink/60">Os documentos correspondentes aparecem primeiro. Se não souber o documento, busque pelo nome.</p></div>
        <button disabled={loadingGuests} className="rounded-xl border border-forest px-5 py-3 text-sm font-semibold text-forest disabled:opacity-60" type="submit">{loadingGuests ? 'Buscando...' : 'Buscar hóspedes'}</button>
      </form>
      {guests.length > 0 && <div className="mt-4"><label htmlFor="guest-select" className="mb-1.5 block text-sm font-medium">Resultados</label><select id="guest-select" className={control} value={selectedGuest?.id ?? ''} onChange={event => setSelectedGuest(guests.find(guest => guest.id === event.target.value) ?? null)}><option value="">Selecione um hóspede</option>{guests.map(guest => <option key={guest.id} value={guest.id}>{guest.documentType ?? 'Documento'} {guest.document ?? 'ausente'} · {guest.name} · {guest.email ?? 'sem e-mail'}</option>)}</select></div>}
      {selectedGuest && <div className="mt-4 flex flex-wrap items-center justify-between gap-3 rounded-xl bg-white p-4">
        <div><p className="font-semibold text-forest">{selectedGuest.documentType ?? 'Documento'} · {selectedGuest.document ?? 'ausente'}</p><p className="mt-1 text-sm text-ink/65">{selectedGuest.name} · {selectedGuest.email ?? 'E-mail ausente'}</p></div>
        <button type="button" className="rounded-lg px-3 py-2 text-sm font-semibold text-orange underline" onClick={() => setEditGuest(true)}>Completar cadastro</button>
      </div>}
      {guests.length === 0 && !loadingGuests && <p className="mt-4 text-sm text-ink/60">Busque para selecionar um hóspede cadastrado.</p>}
    </section>

    <section className="mt-7 rounded-2xl border border-forest/10 bg-cream p-5 shadow-sm sm:p-7">
      <h2 className="mb-4 font-serif text-2xl font-semibold text-forest">2. Período e acomodação</h2>
      <ReservationSearchForm value={dates} busy={busy} onChange={setDates} onSubmit={handleAvailability} />
      {searched && <div className="mt-7"><ReservationRoomOptions rooms={rooms} selectedRoomId={selectedRoom?.roomId} onSelect={setSelectedRoom} /></div>}
      {selectedRoom && selectedGuest && <div className="mt-6 flex flex-col gap-4 rounded-xl border border-forest/10 bg-white p-5 sm:flex-row sm:items-center sm:justify-between">
        <div><p className="font-semibold text-forest">Quarto {selectedRoom.number} · {selectedRoom.categoryName}</p><p className="mt-1 text-sm text-ink/65">{selectedRoom.nights} diárias × {formatBRL(selectedRoom.dailyRate)} · {dates.occupants} pessoa(s)</p><p className="mt-1 text-sm font-semibold text-forest">Total previsto: {formatBRL(selectedRoom.estimatedTotal)}</p><p className="mt-1 text-xs text-ink/60">Hóspede: {selectedGuest.name}</p></div>
        <button type="button" disabled={busy} onClick={() => void handleConfirm()} className="rounded-xl bg-forest px-6 py-3 text-sm font-semibold text-cream hover:bg-moss disabled:opacity-60">{busy ? 'Confirmando...' : 'Confirmar reserva'}</button>
      </div>}
    </section>
    {error && <p role="alert" className="mt-5 rounded-xl bg-red-50 p-4 text-sm text-red-700">{error}</p>}
    {success && <p role="status" className="mt-5 rounded-xl bg-forest/10 p-4 text-sm text-forest">{success}</p>}
    {editGuest && selectedGuest && <EntityDialog resource="guests" entity={selectedGuest} readOnly={false} categories={[]} onClose={() => setEditGuest(false)} onSaved={() => void refreshSelectedGuest()} />}
  </div>
}
