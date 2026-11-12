import { useEffect, useRef, useState, type FormEvent } from 'react'
import { APIError } from '../services/api'
import { searchReservationAvailability, updateReservation, type ReceptionReservation } from '../services/hotel'
import { addCalendarDays, formatBRL, formatReservationDate, type AvailabilityOption } from '../services/reservations'
import { ReservationRoomOptions } from './ReservationRoomOptions'

type ChangeKind = 'dates' | 'occupants' | 'room'
type Step = 1 | 2 | 3
const choices: { kind: ChangeKind, title: string, description: string }[] = [
  { kind: 'dates', title: 'Datas da estadia', description: 'Mudar a entrada ou a saída, mantendo as pessoas e o quarto.' },
  { kind: 'occupants', title: 'Quantidade de pessoas', description: 'Ajustar os ocupantes, mantendo o período e o quarto.' },
  { kind: 'room', title: 'Trocar de quarto', description: 'Escolher outra acomodação para o mesmo período e pessoas.' },
]
const control = 'mt-2 w-full rounded-xl border border-forest/20 bg-white px-4 py-3 text-sm focus-visible:outline-orange'
const primary = 'rounded-xl bg-forest px-5 py-3 text-sm font-semibold text-cream hover:bg-moss focus-visible:outline-orange disabled:opacity-50'
const secondary = 'rounded-xl border border-forest/20 px-5 py-3 text-sm font-semibold text-forest hover:bg-forest/10 focus-visible:outline-orange disabled:opacity-50'

function message(cause: unknown) {
  const fields = cause instanceof APIError ? Object.values(cause.fields).join(' ') : ''
  return fields || (cause instanceof Error ? cause.message : 'Não foi possível concluir a solicitação.')
}

export function ReservationEditWizard({ reservation, hotelDate, onSaved, onReload }: {
  reservation: ReceptionReservation, hotelDate: string, onSaved: (reservation: ReceptionReservation) => void, onReload: () => void,
}) {
  const [step, setStep] = useState<Step>(1)
  const [kind, setKind] = useState<ChangeKind | null>(null)
  const [dates, setDates] = useState({ checkIn: reservation.checkIn, checkOut: reservation.checkOut, occupants: reservation.occupants })
  const [choosingRoom, setChoosingRoom] = useState(false)
  const [rooms, setRooms] = useState<AvailabilityOption[]>([])
  const [searched, setSearched] = useState(false)
  const [blockedRoom, setBlockedRoom] = useState(false)
  const [quote, setQuote] = useState<AvailabilityOption | null>(null)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [needsRefresh, setNeedsRefresh] = useState(false)
  const heading = useRef<HTMLHeadingElement>(null)
  useEffect(() => { heading.current?.focus() }, [step, kind, choosingRoom])

  function clearQuote() {
    setRooms([]); setSearched(false); setBlockedRoom(false); setQuote(null); setError('')
  }
  function choose(next: ChangeKind) {
    clearQuote(); setKind(next); setChoosingRoom(next === 'room')
    setDates({ checkIn: reservation.checkIn, checkOut: reservation.checkOut, occupants: reservation.occupants })
    setStep(2)
  }
  function back() {
    clearQuote()
    if (step === 3) setStep(2)
    else if (choosingRoom && kind !== 'room') setChoosingRoom(false)
    else { setStep(1); setKind(null); setChoosingRoom(false) }
  }
  async function checkAvailability(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (busy || needsRefresh || (!choosingRoom && !draftChanged)) return
    clearQuote(); setBusy(true)
    try {
      const options = await searchReservationAvailability({ ...dates, reservationId: reservation.id })
      if (choosingRoom) {
        setRooms(options.filter(room => room.roomId !== reservation.roomId)); setSearched(true)
      } else {
        const current = options.find(room => room.roomId === reservation.roomId)
        if (current) { setQuote(current); setStep(3) }
        else { setRooms(options.filter(room => room.roomId !== reservation.roomId)); setBlockedRoom(true) }
      }
    } catch (cause) {
      setError(message(cause))
      if (cause instanceof APIError && (cause.status === 409 || cause.status === 404)) setNeedsRefresh(true)
    } finally { setBusy(false) }
  }
  async function save() {
    if (!quote || busy || needsRefresh) return
    setBusy(true); setError('')
    try {
      const result = await updateReservation(reservation.id, { ...dates, roomId: quote.roomId, expectedDailyRate: quote.dailyRate, version: reservation.version })
      onSaved(result.reservation)
    } catch (cause) {
      setError(message(cause))
      if (cause instanceof APIError && (cause.status === 409 || cause.status === 404)) setNeedsRefresh(true)
      else if (cause instanceof APIError && cause.status === 422) { setQuote(null); setRooms([]); setSearched(false); setStep(2) }
    } finally { setBusy(false) }
  }

  const draftChanged = kind === 'dates' ? dates.checkIn !== reservation.checkIn || dates.checkOut !== reservation.checkOut : dates.occupants !== reservation.occupants
  const title = step === 1 ? 'O que deseja alterar?' : step === 3 ? 'Revise a alteração' : choosingRoom ? 'Escolha outro quarto' : kind === 'dates' ? 'Ajuste as datas' : 'Ajuste a quantidade de pessoas'
  const comparisons = quote ? [
    { label: 'Período', before: `${formatReservationDate(reservation.checkIn)} → ${formatReservationDate(reservation.checkOut)}`, after: `${formatReservationDate(dates.checkIn)} → ${formatReservationDate(dates.checkOut)}` },
    { label: 'Pessoas', before: String(reservation.occupants), after: String(dates.occupants) },
    { label: 'Quarto', before: reservation.roomNumber, after: quote.number },
    { label: 'Diária', before: formatBRL(reservation.dailyRate), after: formatBRL(quote.dailyRate) },
    { label: 'Total', before: formatBRL(reservation.totalAmount), after: formatBRL(quote.estimatedTotal) },
  ] : []
  return <section className="mt-7" aria-label="Fluxo de alteração da reserva">
    <ol aria-label="Etapas da alteração" className="mb-6 grid grid-cols-3 gap-2">
      {['Escolher', 'Ajustar', 'Revisar'].map((label, index) => <li key={label} aria-current={step === index + 1 ? 'step' : undefined} className={`rounded-xl border p-3 text-sm font-semibold ${step === index + 1 ? 'border-forest bg-forest text-cream' : index + 1 < step ? 'border-forest/20 bg-forest/10 text-forest' : 'border-forest/10 bg-cream text-ink/55'}`}><span className="block text-xs sm:inline">{index + 1}{index + 1 < step ? ' ✓' : '.'}</span> {label}</li>)}
    </ol>
    {error && <div role="alert" className="mb-5 rounded-xl bg-red-50 p-4 text-red-700"><p>{error}</p>{needsRefresh && <><p className="mt-2 text-sm">Recarregue a reserva e refaça a alteração com os dados atuais.</p><button type="button" className="mt-3 rounded-lg px-3 py-2 font-semibold underline focus-visible:outline-orange" onClick={onReload}>Recarregar reserva</button></>}</div>}
    <div className="rounded-2xl border border-forest/10 bg-cream p-5 shadow-sm sm:p-7">
      <p className="text-xs font-semibold tracking-widest text-orange uppercase">Etapa {step} de 3</p>
      <h2 ref={heading} tabIndex={-1} className="mt-2 font-serif text-2xl font-semibold text-forest outline-none">{title}</h2>
      <fieldset disabled={busy || needsRefresh} className="mt-5 min-w-0 space-y-5 disabled:opacity-60">
        {step === 1 && <div className="grid gap-3 sm:grid-cols-3">{choices.map(choice => <button key={choice.kind} type="button" onClick={() => choose(choice.kind)} className="rounded-xl border border-forest/20 bg-white p-5 text-left transition hover:border-orange hover:bg-orange/5 focus-visible:outline-orange"><span className="block text-base font-semibold text-forest">{choice.title}</span><span className="mt-2 block text-sm text-ink/65">{choice.description}</span></button>)}</div>}
        {step === 2 && <>
          <form onSubmit={checkAvailability} className="space-y-5">
            {choosingRoom ? <p className="rounded-xl bg-sand/60 p-4 text-sm">Período: {formatReservationDate(dates.checkIn)} → {formatReservationDate(dates.checkOut)} · {dates.occupants} pessoa(s). {kind === 'room' ? 'As datas e a quantidade de pessoas serão mantidas.' : 'A revisão também incluirá a troca de quarto.'}</p> : <>
              {kind === 'dates' ? <div className="grid gap-4 sm:grid-cols-2">
                <label className="text-sm font-medium" htmlFor="edit-check-in">Nova entrada<input id="edit-check-in" type="date" required min={addCalendarDays(hotelDate, 1)} value={dates.checkIn} className={control} onChange={event => { clearQuote(); const checkIn = event.target.value; setDates({ ...dates, checkIn, checkOut: checkIn && dates.checkOut <= checkIn ? addCalendarDays(checkIn, 1) : dates.checkOut }) }} /></label>
                <label className="text-sm font-medium" htmlFor="edit-check-out">Nova saída<input id="edit-check-out" type="date" required min={dates.checkIn ? addCalendarDays(dates.checkIn, 1) : undefined} value={dates.checkOut} className={control} onChange={event => { clearQuote(); setDates({ ...dates, checkOut: event.target.value }) }} /></label>
              </div> : <label className="block max-w-sm text-sm font-medium" htmlFor="edit-occupants">Nova quantidade de pessoas<input id="edit-occupants" type="number" required min={1} step={1} value={dates.occupants} className={control} onChange={event => { clearQuote(); setDates({ ...dates, occupants: Number(event.target.value) }) }} /></label>}
              <p className="text-sm text-ink/65">{kind === 'dates' ? `Quarto ${reservation.roomNumber} e ${reservation.occupants} pessoa(s) serão mantidos.` : `Quarto ${reservation.roomNumber} e período de ${formatReservationDate(reservation.checkIn)} a ${formatReservationDate(reservation.checkOut)} serão mantidos.`}</p>
            </>}
            <div className="flex flex-wrap gap-3"><button type="button" onClick={back} className={secondary}>{choosingRoom && kind !== 'room' ? 'Voltar ao ajuste' : 'Escolher outra alteração'}</button><button type="submit" disabled={!choosingRoom && !draftChanged} className={primary}>{busy ? 'Verificando…' : choosingRoom ? 'Consultar quartos disponíveis' : 'Revisar alteração'}</button></div>
          </form>
          {blockedRoom && <div role="status" className="rounded-xl border border-orange/25 bg-orange/10 p-4"><p>O quarto atual não está disponível para essa alteração.</p>{rooms.length > 0 ? <><p className="mt-2 text-sm">Ajuste os dados ou inclua uma troca de quarto.</p><button type="button" className={`${secondary} mt-3`} onClick={() => { setChoosingRoom(true); setBlockedRoom(false); setSearched(true) }}>Escolher outro quarto</button></> : <p className="mt-2 text-sm">Nenhum quarto atende à alteração. Revise os dados ou escolha outro tipo de alteração.</p>}</div>}
          {choosingRoom && searched && (rooms.length > 0 ? <ReservationRoomOptions rooms={rooms} onSelect={room => { setQuote(room); setStep(3); setError('') }} /> : <p role="status" className="rounded-xl bg-sand/60 p-5 text-sm">Nenhum outro quarto está disponível para esse período e quantidade de pessoas.</p>)}
        </>}
        {step === 3 && quote && <>
          <p className="text-sm text-ink/65">Confira o antes e depois. A reserva só será atualizada ao confirmar.</p>
          <div className="overflow-x-auto"><table className="w-full table-fixed text-left text-xs sm:text-sm"><caption className="sr-only">Comparação da reserva antes e depois da alteração</caption><thead><tr className="border-b border-forest/15"><th className="w-1/4 py-3 pr-2 sm:pr-4">Informação</th><th className="px-2 py-3 sm:px-3">Antes</th><th className="px-2 py-3 sm:px-3">Depois</th></tr></thead><tbody>{comparisons.map(row => <tr key={row.label} className="border-b border-forest/10"><th scope="row" className="py-4 pr-2 font-medium sm:pr-4">{row.label}</th><td className="break-words px-2 py-4 text-ink/65 sm:px-3">{row.before}</td><td className={`break-words px-2 py-4 sm:px-3 ${row.before !== row.after ? 'font-semibold text-forest' : 'text-ink/65'}`}>{row.after}</td></tr>)}</tbody></table></div>
          <p className="text-sm text-ink/65">{quote.roomId === reservation.roomId ? 'A diária contratada será preservada no mesmo quarto.' : 'A troca de quarto aplica a diária atual da nova acomodação.'} O hóspede responsável permanece o mesmo.</p>
          <div className="flex flex-wrap gap-3"><button type="button" onClick={back} className={secondary}>Voltar ao ajuste</button><button type="button" onClick={() => void save()} className={primary}>{busy ? 'Salvando…' : 'Confirmar alteração'}</button></div>
        </>}
      </fieldset>
    </div>
  </section>
}
