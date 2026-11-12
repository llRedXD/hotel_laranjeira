import { useEffect, useState } from 'react'
import { Link, useLocation, useParams } from 'react-router-dom'
import { ReservationEditWizard } from '../components/ReservationEditWizard'
import { ReservationStatusBadge } from '../components/ReservationStatusBadge'
import { APIError } from '../services/api'
import { getReservation, type ReceptionReservation } from '../services/hotel'
import { formatBRL, formatReservationDate } from '../services/reservations'

function errorMessage(cause: unknown) {
  const fields = cause instanceof APIError ? Object.values(cause.fields).join(' ') : ''
  return fields || (cause instanceof Error ? cause.message : 'Não foi possível concluir a solicitação.')
}

export function ReservationEditPage() {
  const { id = '' } = useParams()
  const location = useLocation()
  const returnTo = typeof location.state?.returnTo === 'string' && location.state.returnTo.startsWith('/reservas') ? location.state.returnTo : '/reservas'
  const [reservation, setReservation] = useState<ReceptionReservation | null>(null)
  const [hotelDate, setHotelDate] = useState('')
  const [loading, setLoading] = useState(true)
  const [revision, setRevision] = useState(0)
  const [error, setError] = useState('')
  const [saved, setSaved] = useState(false)
  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true); setError(''); setSaved(false)
      try {
        const result = await getReservation(id)
        if (active) { setReservation(result.reservation); setHotelDate(result.hotelDate) }
      } catch (cause) { if (active) { setReservation(null); setError(errorMessage(cause)) } }
      finally { if (active) setLoading(false) }
    }
    void load()
    return () => { active = false }
  }, [id, revision])

  return <div className="mx-auto max-w-5xl px-5 py-8 sm:px-8 sm:py-10">
    <header className="flex flex-wrap items-end justify-between gap-4">
      <div><p className="text-sm font-semibold tracking-[0.18em] text-orange uppercase">Gestão do hotel</p><h1 className="mt-2 font-serif text-4xl font-semibold text-forest">Alterar reserva</h1><p className="mt-2 text-ink/65">Escolha o que deseja modificar e siga as etapas até a confirmação.</p></div>
      <Link to={returnTo} className="rounded-xl border border-forest/20 bg-cream px-5 py-3 text-sm font-semibold text-forest focus-visible:outline-orange">Voltar às reservas</Link>
    </header>
    {loading ? <p role="status" className="mt-8">Carregando reserva…</p> : <>
      {error && <div role="alert" className="mt-6 rounded-xl bg-red-50 p-4 text-red-700"><p>{error}</p><button className="mt-3 rounded-lg px-3 py-2 font-semibold underline focus-visible:outline-orange" onClick={() => setRevision(value => value + 1)}>Recarregar reserva</button></div>}
      {reservation && <section aria-label="Dados atuais da reserva" className="mt-7 rounded-2xl border border-forest/10 bg-cream p-5">
        <p className="font-semibold text-forest">{reservation.guestName}</p><p className="mt-2 text-sm text-ink/65">Quarto {reservation.roomNumber} · {formatReservationDate(reservation.checkIn)} → {formatReservationDate(reservation.checkOut)} · {reservation.occupants} pessoa(s)</p>
        <details className="mt-3"><summary className="cursor-pointer rounded-lg py-1 text-sm font-semibold text-forest focus-visible:outline-orange">Ver dados atuais da reserva</summary><div className="mt-3"><ReservationStatusBadge status={reservation.status} /><p className="mt-3 text-sm">Diária contratada: {formatBRL(reservation.dailyRate)} · Total: {formatBRL(reservation.totalAmount)}</p></div></details>
      </section>}
      {reservation && !reservation.canEdit && <p role="status" className="mt-6 rounded-xl bg-orange/10 p-4">Esta reserva não permite alterações. Apenas reservas não canceladas com entrada futura podem ser alteradas.</p>}
      {saved ? <section className="mt-7 rounded-2xl border border-forest/15 bg-forest/5 p-6"><h2 className="font-serif text-2xl font-semibold text-forest">Alteração concluída</h2><p role="status" className="mt-3 text-forest">Reserva atualizada. A alteração e o responsável foram registrados na auditoria.</p><button type="button" onClick={() => setRevision(value => value + 1)} className="mt-5 rounded-xl border border-forest/20 px-5 py-3 text-sm font-semibold text-forest focus-visible:outline-orange">Fazer outra alteração</button></section> : reservation?.canEdit && <ReservationEditWizard key={`${reservation.id}:${reservation.version}`} reservation={reservation} hotelDate={hotelDate} onReload={() => setRevision(value => value + 1)} onSaved={updated => { setReservation(updated); setSaved(true) }} />}
    </>}
  </div>
}
