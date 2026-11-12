import { useEffect, useRef, useState } from 'react'
import { APIError } from '../services/api'
import { cancelReservation, type ReceptionReservation } from '../services/hotel'
import { formatReservationDate } from '../services/reservations'

export function CancelReservationDialog({ reservation, onClose, onCancelled, onRefresh }: {
  reservation: ReceptionReservation, onClose: () => void, onCancelled: () => void, onRefresh: () => void,
}) {
  const dialog = useRef<HTMLDialogElement>(null)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [stale, setStale] = useState(false)
  useEffect(() => { const element = dialog.current; element?.showModal(); return () => element?.close() }, [])
  async function confirm() {
    if (busy || stale) return
    setBusy(true); setError('')
    try { await cancelReservation(reservation.id, reservation.version); onCancelled() }
    catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Não foi possível cancelar a reserva.')
      setStale(cause instanceof APIError && (cause.status === 409 || cause.status === 404))
    } finally { setBusy(false) }
  }
  return <dialog ref={dialog} aria-labelledby="cancel-title" onCancel={event => { event.preventDefault(); if (!busy) onClose() }} className="m-auto max-h-[90dvh] w-[calc(100%-2rem)] max-w-lg overflow-y-auto rounded-2xl bg-cream p-6 text-ink shadow-xl backdrop:bg-forest/60">
    <h2 id="cancel-title" className="font-serif text-2xl font-semibold text-forest">Cancelar reserva?</h2>
    <p className="mt-3 font-semibold">{reservation.guestName}</p>
    <p className="mt-1">Quarto {reservation.roomNumber} · {formatReservationDate(reservation.checkIn)} → {formatReservationDate(reservation.checkOut)}</p>
    <p className="mt-4 text-sm text-ink/70">O quarto será liberado para esse período. A reserva e o registro do cancelamento permanecerão no histórico. Esta ação não permite reativação.</p>
    {error && <p role="alert" className="mt-4 rounded-lg bg-red-50 p-3 text-sm text-red-700">{error}</p>}
    <div className="mt-6 flex flex-wrap justify-end gap-3">
      <button type="button" disabled={busy} onClick={onClose} className="rounded-xl border border-forest/20 px-4 py-3 text-sm font-semibold disabled:opacity-50">Voltar</button>
      {stale ? <button type="button" onClick={onRefresh} className="rounded-xl bg-forest px-4 py-3 text-sm font-semibold text-cream">Atualizar consulta</button> : <button type="button" disabled={busy} onClick={() => void confirm()} className="rounded-xl bg-red-700 px-4 py-3 text-sm font-semibold text-white disabled:opacity-50">{busy ? 'Cancelando…' : 'Confirmar cancelamento'}</button>}
    </div>
  </dialog>
}
