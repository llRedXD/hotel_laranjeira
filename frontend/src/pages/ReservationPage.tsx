import { useState, type FormEvent } from 'react'
import { APIError } from '../services/api'
import { createPublicReservation, formatBRL, formatReservationDate, localDateString, searchAvailability, type AvailabilityOption, type ReservationConfirmation, type ReservationGuestInput } from '../services/reservations'
import { formatGuestDocument, guestDocumentError } from '../services/guestDocument'
import { ReservationRoomOptions } from '../components/ReservationRoomOptions'
import { ReservationSearchForm } from '../components/ReservationSearchForm'

const todayDate = new Date()
const today = localDateString(todayDate)
const tomorrowDate = new Date(todayDate)
tomorrowDate.setDate(tomorrowDate.getDate() + 1)
const tomorrow = localDateString(tomorrowDate)
const control = 'w-full rounded-xl border border-forest/20 bg-white px-4 py-3 text-sm focus:border-orange focus:outline-none focus:ring-4 focus:ring-orange/15'

export function ReservationPage() {
  const [dates, setDates] = useState({ checkIn: today, checkOut: tomorrow, occupants: 1 })
  const [rooms, setRooms] = useState<AvailabilityOption[]>([])
  const [selectedRoom, setSelectedRoom] = useState<AvailabilityOption | null>(null)
  const [guest, setGuest] = useState<ReservationGuestInput>({ name: '', documentType: 'CPF', document: '', email: '' })
  const [reviewing, setReviewing] = useState(false)
  const [confirmation, setConfirmation] = useState<ReservationConfirmation | null>(null)
  const [busy, setBusy] = useState(false)
  const [searched, setSearched] = useState(false)
  const [availabilityStale, setAvailabilityStale] = useState(false)
  const [documentError, setDocumentError] = useState('')
  const [error, setError] = useState('')

  const activeStep = reviewing ? 4 : selectedRoom ? 3 : searched ? 2 : 1

  function handleSearchValuesChange(nextDates: typeof dates) {
    const searchChanged = nextDates.checkIn !== dates.checkIn
      || nextDates.checkOut !== dates.checkOut
      || nextDates.occupants !== dates.occupants

    setDates(nextDates)
    if (!searchChanged) return

    if (searched || selectedRoom) setAvailabilityStale(true)
    setRooms([])
    setSearched(false)
    setSelectedRoom(null)
    setReviewing(false)
    setError('')
    setDocumentError('')
  }

  async function handleSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    setAvailabilityStale(false)
    setSelectedRoom(null)
    setReviewing(false)
    setDocumentError('')
    if (!dates.checkIn || !dates.checkOut || dates.checkOut <= dates.checkIn || dates.occupants < 1) {
      setError('Informe um período válido e a quantidade total de pessoas.')
      return
    }
    setBusy(true)
    try {
      setRooms(await searchAvailability(dates))
      setSearched(true)
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Não foi possível consultar os quartos.')
    } finally { setBusy(false) }
  }

  function handleGuestSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    if (!selectedRoom) return
    const validationError = guestDocumentError(guest.documentType, guest.document)
    if (validationError) {
      setDocumentError(validationError)
      window.requestAnimationFrame(() => document.getElementById('document-number')?.focus())
      return
    }
    setDocumentError('')
    setReviewing(true)
  }

  async function handleConfirm() {
    if (!selectedRoom) return
    setBusy(true)
    setError('')
    try {
      setConfirmation(await createPublicReservation({ ...dates, roomId: selectedRoom.roomId, expectedDailyRate: selectedRoom.dailyRate, guest }))
    } catch (cause) {
      if (cause instanceof APIError && cause.status === 409) setError('A disponibilidade mudou antes da confirmação. Consulte novamente para escolher outro quarto.')
      else setError(cause instanceof Error ? cause.message : 'Não foi possível confirmar a reserva.')
    } finally { setBusy(false) }
  }

  async function recheckAvailability() {
    setSelectedRoom(null)
    setReviewing(false)
    setBusy(true)
    setError('')
    try {
      setRooms(await searchAvailability(dates))
      setSearched(true)
      setAvailabilityStale(false)
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Não foi possível consultar os quartos.')
    } finally { setBusy(false) }
  }

  if (confirmation) return <section className="mx-auto w-full max-w-3xl flex-1 px-5 py-12 sm:px-8">
    <div className="rounded-3xl border border-forest/10 bg-cream p-7 shadow-sm sm:p-10">
      <p className="text-sm font-semibold tracking-[0.18em] text-orange uppercase">Reserva confirmada</p>
      <h1 className="mt-3 font-serif text-4xl font-semibold text-forest">Até breve, {guest.name.split(' ')[0]}.</h1>
      <p className="mt-3 text-ink/70">Sua reserva foi registrada. Guarde o código para apresentar à recepção no dia da chegada.</p>
      <dl className="mt-8 grid gap-4 border-t border-forest/10 pt-6 sm:grid-cols-2">
        <div><dt className="text-sm text-ink/60">Código</dt><dd className="mt-1 font-semibold text-forest">{confirmation.id}</dd></div>
        <div><dt className="text-sm text-ink/60">Quarto</dt><dd className="mt-1 font-semibold text-forest">{selectedRoom?.number}</dd></div>
        <div><dt className="text-sm text-ink/60">Período previsto</dt><dd className="mt-1 font-semibold text-forest">{formatReservationDate(confirmation.checkIn)} a {formatReservationDate(confirmation.checkOut)}</dd></div>
        <div><dt className="text-sm text-ink/60">Total previsto</dt><dd className="mt-1 font-semibold text-forest">{formatBRL(confirmation.totalAmount)}</dd></div>
      </dl>
    </div>
  </section>

  return <section className="mx-auto w-full max-w-6xl flex-1 px-5 py-10 sm:px-8 sm:py-14">
    <header className="mb-8 max-w-3xl">
      <p className="text-sm font-semibold tracking-[0.18em] text-orange uppercase">Hospedagem</p>
      <h1 className="mt-2 font-serif text-4xl font-semibold text-forest sm:text-5xl">Encontre seu quarto</h1>
      <p className="mt-3 text-ink/65">Informe as datas previstas e veja as opções disponíveis antes de reservar. Não é necessário criar uma conta.</p>
    </header>

    <nav aria-label="Etapas da reserva" className="mb-6 rounded-2xl border border-forest/10 bg-cream px-4 py-4 sm:px-6">
      <ol className="grid grid-cols-4 gap-2">
        {['Período', 'Quarto', 'Dados', 'Revisão'].map((label, index) => {
          const step = index + 1
          const active = step === activeStep
          const complete = step < activeStep
          return <li key={label} aria-current={active ? 'step' : undefined} className={`flex min-w-0 flex-col items-center gap-2 text-center text-xs font-medium sm:flex-row sm:justify-center sm:text-sm ${active || complete ? 'text-forest' : 'text-ink/50'}`}>
            <span aria-hidden="true" className={`grid size-7 shrink-0 place-items-center rounded-full border text-xs ${active ? 'border-forest bg-forest text-cream' : complete ? 'border-moss bg-moss text-white' : 'border-forest/20 bg-white'}`}>{complete ? '✓' : step}</span>
            <span className="truncate">{label}</span>
          </li>
        })}
      </ol>
    </nav>

    <ReservationSearchForm value={dates} busy={busy} onChange={handleSearchValuesChange} onSubmit={handleSearch} />
    {availabilityStale && <p className="mt-3 text-sm text-forest" role="status" aria-live="polite">As datas ou a quantidade de pessoas mudaram. Consulte os quartos novamente para atualizar os valores.</p>}
    {error && <p role="alert" className="mt-4 rounded-xl bg-red-50 p-4 text-sm text-red-700">{error}</p>}
    {searched && !selectedRoom && <section className="mt-8" aria-label="Quartos disponíveis">
      <h2 className="mb-4 font-serif text-2xl font-semibold text-forest">Quartos disponíveis</h2>
      <ReservationRoomOptions rooms={rooms} onSelect={room => { setSelectedRoom(room); setDocumentError('') }} />
    </section>}

    {selectedRoom && !reviewing && <section className="mt-8 grid gap-6 lg:grid-cols-[0.8fr_1.2fr]">
      <div className="h-fit rounded-2xl border border-forest/10 bg-cream p-5">
        <p className="text-xs font-semibold tracking-[0.16em] text-orange uppercase">Sua escolha</p>
        <h2 className="mt-2 font-serif text-2xl font-semibold text-forest">Quarto {selectedRoom.number} · {selectedRoom.categoryName}</h2>
        <p className="mt-2 text-sm text-ink/65">{selectedRoom.nights} {selectedRoom.nights === 1 ? 'diária' : 'diárias'} · {formatBRL(selectedRoom.dailyRate)} por diária</p>
        <p className="mt-4 text-lg font-semibold text-forest">Total previsto: {formatBRL(selectedRoom.estimatedTotal)}</p>
        <button type="button" onClick={() => { setSelectedRoom(null); setDocumentError('') }} className="mt-4 text-sm font-semibold text-orange underline">Escolher outro quarto</button>
      </div>
      <form onSubmit={handleGuestSubmit} className="rounded-2xl border border-forest/10 bg-cream p-5 sm:p-7">
        <h2 className="font-serif text-2xl font-semibold text-forest">Dados do responsável</h2>
        <p className="mt-2 text-sm text-ink/65">Usaremos esses dados para identificar a reserva na recepção.</p>
        <div className="mt-5 grid gap-4 sm:grid-cols-2">
          <div className="sm:col-span-2"><label htmlFor="guest-name" className="mb-1.5 block text-sm font-medium">Nome completo</label><input className={control} id="guest-name" name="name" autoComplete="name" required maxLength={120} value={guest.name} onChange={event => setGuest({ ...guest, name: event.target.value })} /></div>
          <div><label htmlFor="document-type" className="mb-1.5 block text-sm font-medium">Tipo de documento</label><select className={control} id="document-type" name="documentType" required value={guest.documentType} onChange={event => {
            const documentType = event.target.value as 'RG' | 'CPF'
            setGuest(current => ({
              ...current,
              documentType,
              document: documentType === 'RG'
                ? current.document.replace(/[.-]/g, '').slice(0, 30)
                : formatGuestDocument(documentType, current.document),
            }))
            setDocumentError('')
          }}><option value="CPF">CPF</option><option value="RG">RG</option></select></div>
          <div><label htmlFor="document-number" className="mb-1.5 block text-sm font-medium">Número do documento</label><input className={control} id="document-number" name="document" autoComplete="off" inputMode={guest.documentType === 'CPF' ? 'numeric' : 'text'} maxLength={guest.documentType === 'CPF' ? 14 : 30} aria-invalid={documentError ? 'true' : undefined} aria-describedby={documentError ? 'document-hint document-error' : 'document-hint'} required value={guest.document} onChange={event => {
            setGuest(current => ({ ...current, document: formatGuestDocument(current.documentType, event.target.value) }))
            setDocumentError('')
          }} /><p id="document-hint" className="mt-1 text-xs text-ink/60">{guest.documentType === 'CPF' ? 'Digite os 11 números do CPF.' : 'Informe o RG como aparece no documento.'}</p>{documentError && <p id="document-error" className="mt-1 text-sm text-red-700" aria-live="polite">{documentError}</p>}</div>
          <div className="sm:col-span-2"><label htmlFor="guest-email" className="mb-1.5 block text-sm font-medium">E-mail</label><input className={control} id="guest-email" name="email" type="email" autoComplete="email" spellCheck={false} required maxLength={255} value={guest.email} onChange={event => setGuest({ ...guest, email: event.target.value })} /></div>
        </div>
        <button className="mt-6 w-full rounded-xl bg-forest px-5 py-3 font-semibold text-cream hover:bg-moss" type="submit">Revisar reserva</button>
      </form>
    </section>}

    {selectedRoom && reviewing && <section className="mt-8 rounded-2xl border border-forest/10 bg-cream p-5 sm:p-8">
      <p className="text-xs font-semibold tracking-[0.16em] text-orange uppercase">Revisão</p>
      <h2 className="mt-2 font-serif text-3xl font-semibold text-forest">Confirme seus dados</h2>
      <div className="mt-6 grid gap-6 md:grid-cols-2">
        <div><h3 className="font-semibold text-forest">Hospedagem</h3><p className="mt-2 text-sm">Quarto {selectedRoom.number} · {selectedRoom.categoryName}</p><p className="mt-1 text-sm">{formatReservationDate(dates.checkIn)} a {formatReservationDate(dates.checkOut)} · {dates.occupants} {dates.occupants === 1 ? 'pessoa' : 'pessoas'}</p><p className="mt-3 text-sm">{selectedRoom.nights} {selectedRoom.nights === 1 ? 'diária' : 'diárias'} × {formatBRL(selectedRoom.dailyRate)}</p><p className="mt-1 text-lg font-semibold text-forest">{formatBRL(selectedRoom.estimatedTotal)}</p></div>
        <div><h3 className="font-semibold text-forest">Responsável</h3><p className="mt-2 text-sm">{guest.name}</p><p className="mt-1 text-sm">{guest.documentType}: {guest.document}</p><p className="mt-1 text-sm">{guest.email}</p></div>
      </div>
      <div className="mt-7 flex flex-col-reverse gap-3 border-t border-forest/10 pt-5 sm:flex-row sm:justify-between">
        <button type="button" disabled={busy} onClick={() => setReviewing(false)} className="rounded-xl border border-forest/20 px-5 py-3 text-sm font-semibold text-forest">Voltar e editar</button>
        <button type="button" disabled={busy} onClick={() => void handleConfirm()} className="rounded-xl bg-forest px-6 py-3 text-sm font-semibold text-cream hover:bg-moss disabled:opacity-60">{busy ? 'Confirmando…' : 'Confirmar reserva'}</button>
      </div>
      {error && <p role="alert" className="mt-4 rounded-xl bg-red-50 p-4 text-sm text-red-700">{error} <button type="button" className="font-semibold underline" onClick={() => void recheckAvailability()}>Consultar disponibilidade novamente</button></p>}
    </section>}
  </section>
}
