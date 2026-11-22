import { useEffect, useRef, useState } from 'react'
import { APIError } from '../services/api'
import {
  checkInReservation,
  checkOutReservation,
  quoteCheckOutReservation,
  type ReceptionReservation,
  type ReservationCheckoutQuote,
} from '../services/hotel'
import { formatBRL } from '../services/reservations'

type StayAction = 'check-in' | 'check-out'

function formatHotelDateTime(value: string) {
  return new Intl.DateTimeFormat('pt-BR', {
    dateStyle: 'medium', timeStyle: 'short', timeZone: 'America/Sao_Paulo',
  }).format(new Date(value))
}

export function ReservationStayActions({ reservation, onUpdated, onReload }: {
  reservation: ReceptionReservation
  onUpdated: (reservation: ReceptionReservation) => void
  onReload: () => void
}) {
  const dialog = useRef<HTMLDialogElement>(null)
  const [action, setAction] = useState<StayAction | null>(null)
  const [quote, setQuote] = useState<ReservationCheckoutQuote | null>(null)
  const [loadingQuote, setLoadingQuote] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [stale, setStale] = useState(false)

  useEffect(() => {
    const element = dialog.current
    if (action) element?.showModal()
    else element?.close()
    return () => element?.close()
  }, [action])

  async function beginCheckOut() {
    setError('')
    setStale(false)
    setLoadingQuote(true)
    try {
      const result = await quoteCheckOutReservation(reservation.id, reservation.version)
      setQuote(result.quote)
      setAction('check-out')
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Não foi possível calcular o valor final.')
      setStale(cause instanceof APIError && (cause.status === 409 || cause.status === 404))
    } finally {
      setLoadingQuote(false)
    }
  }

  async function confirm() {
    if (!action || busy || stale) return
    setBusy(true)
    setError('')
    try {
      const result = action === 'check-in'
        ? await checkInReservation(reservation.id, reservation.version)
        : await checkOutReservation(reservation.id, reservation.version)
      setAction(null)
      onUpdated(result.reservation)
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Não foi possível concluir a operação.')
      setStale(cause instanceof APIError && (cause.status === 409 || cause.status === 404))
    } finally {
      setBusy(false)
    }
  }

  return <>
    <div className="flex flex-wrap gap-2">
      {reservation.canCheckIn && <button type="button" onClick={() => { setError(''); setStale(false); setAction('check-in') }} className="rounded-lg bg-forest px-3 py-2 text-sm font-semibold text-cream hover:bg-moss focus-visible:outline focus-visible:outline-2 focus-visible:outline-orange">Realizar check-in</button>}
      {reservation.canCheckOut && <button type="button" disabled={loadingQuote} onClick={() => void beginCheckOut()} className="rounded-lg bg-orange px-3 py-2 text-sm font-semibold text-white hover:bg-orange/90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-forest disabled:opacity-50">{loadingQuote ? 'Calculando…' : 'Realizar check-out'}</button>}
      {error && !action && <span role="alert" className="text-sm text-red-700">{error}</span>}
      {stale && !action && <button type="button" onClick={() => { setError(''); setStale(false); onReload() }} className="text-sm font-semibold text-forest underline">Atualizar reserva</button>}
    </div>
    <dialog ref={dialog} aria-labelledby="stay-action-title" onCancel={event => { event.preventDefault(); if (!busy) setAction(null) }} className="m-auto max-h-[90dvh] w-[calc(100%-2rem)] max-w-lg overflow-y-auto rounded-2xl bg-cream p-6 text-ink shadow-xl backdrop:bg-forest/60">
      <h2 id="stay-action-title" className="font-serif text-2xl font-semibold text-forest">{action === 'check-in' ? 'Confirmar check-in?' : 'Confirmar check-out?'}</h2>
      <p className="mt-3 font-semibold">{reservation.guestName}</p>
      <p className="mt-1 text-sm">Quarto {reservation.roomNumber} · reserva {reservation.checkIn} → {reservation.checkOut}</p>
      {action === 'check-in' && <p className="mt-4 text-sm text-ink/70">O check-in será registrado agora e o quarto passará ao estado “Ocupado”.</p>}
      {action === 'check-out' && quote && <div className="mt-4 rounded-xl bg-white p-4 text-sm">
        <p>Saída prevista para agora: {formatHotelDateTime(quote.checkoutAt)}</p>
        <p className="mt-2">{quote.actualNights} {quote.actualNights === 1 ? 'diária' : 'diárias'} × {formatBRL(quote.dailyRate)}</p>
        <p className="mt-2 text-base font-semibold text-forest">Valor final previsto: {formatBRL(quote.finalAmount)}</p>
        <p className="mt-2 text-xs text-ink/60">O sistema recalculará o valor ao confirmar e exibirá o total registrado.</p>
      </div>}
      {error && <p role="alert" className="mt-4 rounded-lg bg-red-50 p-3 text-sm text-red-700">{error}</p>}
      <div className="mt-6 flex flex-wrap justify-end gap-3">
        <button type="button" disabled={busy} onClick={() => setAction(null)} className="rounded-xl border border-forest/20 px-4 py-3 text-sm font-semibold disabled:opacity-50">Voltar</button>
        {stale ? <button type="button" onClick={() => { setAction(null); setError(''); setStale(false); onReload() }} className="rounded-xl bg-forest px-4 py-3 text-sm font-semibold text-cream">Atualizar reserva</button> : <button type="button" disabled={busy || !action} onClick={() => void confirm()} className="rounded-xl bg-forest px-4 py-3 text-sm font-semibold text-cream disabled:opacity-50">{busy ? 'Salvando…' : 'Confirmar'}</button>}
      </div>
    </dialog>
  </>
}
