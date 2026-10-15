import { useEffect, useState, type FormEvent } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { formatBRL, formatReservationDate } from '../services/reservations'
import { listReservations, pageSize, type ReceptionReservation } from '../services/hotel'

const control = 'w-full rounded-xl border border-forest/20 bg-white px-4 py-3 text-sm focus-visible:border-orange focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-orange/15'

function reservationStatus(reservation: ReceptionReservation, today: string) {
  if (reservation.checkOut < today) return { label: 'Finalizada', style: 'bg-sand text-ink/65' }
  if (reservation.checkOut === today) return { label: 'Saída hoje', style: 'bg-peach/35 text-orange' }
  if (reservation.checkIn <= today) return { label: 'No hotel', style: 'bg-moss/15 text-forest' }
  return { label: 'Confirmada', style: 'bg-orange/15 text-orange' }
}

function localToday() {
  const date = new Date()
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

export function ReservationsListPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const queryKey = searchParams.toString()
  const currentPage = Math.max(0, Number(searchParams.get('page') || 0))
  const search = searchParams.get('search') ?? ''
  const from = searchParams.get('from') ?? ''
  const to = searchParams.get('to') ?? ''
  const [reservations, setReservations] = useState<ReceptionReservation[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [revision, setRevision] = useState(0)

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setError('')
      const query = new URLSearchParams({ limit: String(pageSize), offset: String(currentPage * pageSize) })
      if (search) query.set('search', search)
      if (from) query.set('from', from)
      if (to) query.set('to', to)
      try {
        const result = await listReservations(query)
        if (active) {
          setReservations(result.reservations)
          setTotal(result.total)
        }
      } catch (cause) {
        if (active) setError(cause instanceof Error ? cause.message : 'Não foi possível carregar as reservas.')
      } finally {
        if (active) setLoading(false)
      }
    }
    void load()
    return () => { active = false }
  }, [queryKey, currentPage, search, from, to, revision])

  function applyFilters(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const data = new FormData(event.currentTarget)
    const next = new URLSearchParams()
    const nextSearch = String(data.get('search') ?? '').trim()
    const nextFrom = String(data.get('from') ?? '')
    const nextTo = String(data.get('to') ?? '')
    if (nextSearch) next.set('search', nextSearch)
    if (nextFrom) next.set('from', nextFrom)
    if (nextTo) next.set('to', nextTo)
    setSearchParams(next)
  }

  function setPage(page: number) {
    const next = new URLSearchParams(searchParams)
    if (page > 0) next.set('page', String(page))
    else next.delete('page')
    setSearchParams(next)
  }

  const today = localToday()
  const firstRecord = total === 0 ? 0 : currentPage * pageSize + 1
  const lastRecord = Math.min((currentPage + 1) * pageSize, total)

  return <div className="mx-auto max-w-7xl px-5 py-8 sm:px-8 sm:py-10">
    <header className="flex flex-wrap items-end justify-between gap-4">
      <div>
        <p className="text-sm font-semibold tracking-[0.18em] text-orange uppercase">Gestão do hotel</p>
        <h1 className="mt-2 font-serif text-4xl font-semibold text-forest">Reservas</h1>
        <p className="mt-2 text-ink/65">Consulte hóspedes, períodos e valores em um só lugar.</p>
      </div>
      <Link to="/reservas/nova" className="rounded-xl bg-forest px-5 py-3 text-sm font-semibold text-cream transition-colors hover:bg-moss focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-orange">+ Nova reserva</Link>
    </header>

    <section aria-label="Filtros de reservas" className="mt-8 rounded-2xl border border-forest/10 bg-cream p-5 shadow-sm">
      <form key={queryKey} onSubmit={applyFilters} className="grid gap-4 md:grid-cols-[minmax(16rem,1fr)_12rem_12rem_auto] md:items-end">
        <div>
          <label htmlFor="reservation-search" className="mb-1.5 block text-sm font-medium">Hóspede, quarto ou documento</label>
          <input id="reservation-search" name="search" type="search" autoComplete="off" maxLength={255} placeholder="Ex.: Ana Souza ou quarto 101…" defaultValue={search} className={control} />
        </div>
        <div>
          <label htmlFor="reservation-from" className="mb-1.5 block text-sm font-medium">Entrada a partir de</label>
          <input id="reservation-from" name="from" type="date" autoComplete="off" defaultValue={from} className={control} />
        </div>
        <div>
          <label htmlFor="reservation-to" className="mb-1.5 block text-sm font-medium">Entrada até</label>
          <input id="reservation-to" name="to" type="date" autoComplete="off" defaultValue={to} className={control} />
        </div>
        <button type="submit" className="rounded-xl border border-forest px-5 py-3 text-sm font-semibold text-forest transition-colors hover:bg-forest/10 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-orange">Filtrar</button>
      </form>
    </section>

    <section aria-label="Lista de reservas" className="mt-6 overflow-hidden rounded-2xl border border-forest/10 bg-cream shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-forest/10 px-5 py-4">
        <div>
          <h2 className="font-serif text-xl font-semibold text-forest">Todas as reservas</h2>
          <p className="mt-1 text-sm text-ink/60">{loading ? 'Consultando reservas…' : `${total} ${total === 1 ? 'reserva encontrada' : 'reservas encontradas'}`}</p>
        </div>
      </div>
      {loading ? <p role="status" className="p-8 text-center text-ink/60">Carregando reservas…</p> : error ? <div role="alert" className="p-6 text-red-700"><p>{error}</p><button type="button" onClick={() => setRevision(value => value + 1)} className="mt-2 rounded-lg px-3 py-2 text-sm font-semibold underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-orange">Tentar novamente</button></div> : reservations.length === 0 ? <div className="p-12 text-center"><h3 className="font-serif text-2xl text-forest">Nenhuma reserva encontrada</h3><p className="mt-2 text-ink/60">{search || from || to ? 'Ajuste os filtros para ampliar sua busca.' : 'As reservas confirmadas aparecerão aqui.'}</p></div> : <div className="overflow-x-auto">
        <table className="w-full min-w-[760px] text-left text-sm">
          <caption className="sr-only">Reservas do Hotel Laranjeira</caption>
          <thead className="bg-sand/60 text-xs tracking-wide text-ink/60 uppercase"><tr><th className="px-5 py-4">Hóspede</th><th className="px-4 py-4">Quarto</th><th className="px-4 py-4">Entrada · saída</th><th className="px-4 py-4">Pessoas</th><th className="px-4 py-4">Status</th><th className="px-5 py-4 text-right">Total</th></tr></thead>
          <tbody className="divide-y divide-forest/10">{reservations.map(reservation => {
            const status = reservationStatus(reservation, today)
            return <tr key={reservation.id} className="hover:bg-sand/30">
              <td className="max-w-64 break-words px-5 py-4"><span className="font-semibold text-forest">{reservation.guestName}</span><span className="mt-1 block text-xs text-ink/55">#{reservation.id.slice(0, 8)}</span></td>
              <td className="px-4 py-4"><span className="font-medium">{reservation.roomNumber}</span><span className="mt-1 block text-xs text-ink/55">{reservation.categoryName}</span></td>
              <td className="whitespace-nowrap px-4 py-4">{formatReservationDate(reservation.checkIn)} <span className="text-ink/40">→</span> {formatReservationDate(reservation.checkOut)}</td>
              <td className="px-4 py-4">{reservation.occupants}</td>
              <td className="px-4 py-4"><span className={`rounded-full px-3 py-1 text-xs font-semibold ${status.style}`}>{status.label}</span></td>
              <td className="whitespace-nowrap px-5 py-4 text-right font-semibold tabular-nums">{formatBRL(reservation.totalAmount)}</td>
            </tr>
          })}</tbody>
        </table>
      </div>}
      <footer className="flex flex-wrap items-center justify-between gap-3 border-t border-forest/10 px-5 py-4 text-sm text-ink/65">
        <p>{loading ? ' ' : total ? `Exibindo ${firstRecord}–${lastRecord} de ${total}` : ' '}</p>
        <div className="flex gap-2">
          <button type="button" className="rounded-lg px-3 py-2 font-medium text-forest hover:bg-forest/10 focus-visible:outline focus-visible:outline-2 focus-visible:outline-orange disabled:opacity-45" disabled={loading || currentPage === 0} onClick={() => setPage(currentPage - 1)}>Anterior</button>
          <button type="button" className="rounded-lg px-3 py-2 font-medium text-forest hover:bg-forest/10 focus-visible:outline focus-visible:outline-2 focus-visible:outline-orange disabled:opacity-45" disabled={loading || (currentPage + 1) * pageSize >= total} onClick={() => setPage(currentPage + 1)}>Próxima</button>
        </div>
      </footer>
    </section>
  </div>
}
