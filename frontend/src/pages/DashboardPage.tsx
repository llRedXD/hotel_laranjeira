import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { getReservationSummary, listReservations, type ReceptionReservation, type ReservationSummary } from '../services/hotel'
import { formatBRL, formatReservationDate, localDateString } from '../services/reservations'

const shortcuts = [
  { title: 'Hóspedes', description: 'Consulte e atualize cadastros.', to: '/hospedes' },
  { title: 'Quartos', description: 'Veja diárias e disponibilidade.', to: '/quartos' },
  { title: 'Categorias', description: 'Organize os tipos de acomodação.', to: '/categorias' },
]

function localToday() {
  return localDateString(new Date())
}

function greeting() {
  const hour = new Date().getHours()
  if (hour < 12) return 'Bom dia'
  if (hour < 18) return 'Boa tarde'
  return 'Boa noite'
}

export function DashboardPage() {
  const { user } = useAuth()
  const [summary, setSummary] = useState<ReservationSummary | null>(null)
  const [reservations, setReservations] = useState<ReceptionReservation[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [revision, setRevision] = useState(0)

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setError('')
      const query = new URLSearchParams({ limit: '5', offset: '0' })
      try {
        const [dailySummary, reservationPage] = await Promise.all([
          getReservationSummary(localToday()),
          listReservations(query),
        ])
        if (active) {
          setSummary(dailySummary)
          setReservations(reservationPage.reservations.filter(reservation => reservation.checkOut > localToday()).slice(0, 5))
        }
      } catch (cause) {
        if (active) setError(cause instanceof Error ? cause.message : 'Não foi possível carregar o painel.')
      } finally {
        if (active) setLoading(false)
      }
    }
    void load()
    return () => { active = false }
  }, [revision])

  const metrics = [
    { label: 'Hóspedes hospedados esta noite', value: summary?.guestsInHouse ?? 0, detail: 'reservas que incluem esta noite', accent: 'bg-orange/15 text-orange' },
    { label: 'Check-ins hoje', value: summary?.checkInsToday ?? 0, detail: 'chegadas previstas para hoje', accent: 'bg-forest/10 text-forest' },
    { label: 'Quartos livres esta noite', value: summary?.availableRooms ?? 0, detail: 'considerando as reservas confirmadas', accent: 'bg-moss/15 text-moss' },
  ]

  return <div className="mx-auto max-w-7xl px-5 py-8 sm:px-8 sm:py-10">
    <header className="flex flex-wrap items-end justify-between gap-4">
      <div>
        <p className="text-sm font-semibold tracking-[0.18em] text-orange uppercase">Painel da recepção</p>
        <h1 className="mt-2 font-serif text-3xl font-semibold text-forest sm:text-4xl">{greeting()}, {user?.name}.</h1>
        <p className="mt-2 text-ink/65">A operação de hoje, em um só lugar.</p>
      </div>
      <div className="flex flex-wrap items-center gap-3">
        <p className="rounded-full bg-cream px-4 py-2 text-sm font-medium text-ink/65 shadow-sm">{new Intl.DateTimeFormat('pt-BR', { dateStyle: 'long' }).format(new Date())}</p>
        <Link to="/reservas/nova" className="rounded-xl bg-forest px-5 py-3 text-sm font-semibold text-cream transition-colors hover:bg-moss focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-orange">+ Nova reserva</Link>
      </div>
    </header>

    {error && <div role="alert" className="mt-6 rounded-xl bg-red-50 p-4 text-red-700"><p>{error}</p><button type="button" onClick={() => setRevision(value => value + 1)} className="mt-2 font-semibold underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-orange">Tentar novamente</button></div>}

    <section className="mt-8 grid gap-4 sm:grid-cols-3" aria-label="Resumo operacional de hoje">
      {metrics.map(metric => <article key={metric.label} className="rounded-2xl border border-forest/10 bg-cream p-5 shadow-sm">
        <span className={`inline-flex rounded-lg px-2.5 py-1 text-xs font-semibold ${metric.accent}`}>{metric.label}</span>
        <p className="mt-4 font-serif text-4xl font-semibold tabular-nums text-forest">{loading ? '—' : metric.value}</p>
        <p className="mt-1 text-sm text-ink/60">{metric.detail}</p>
      </article>)}
    </section>

    <section className="mt-8 overflow-hidden rounded-2xl border border-forest/10 bg-cream shadow-sm" aria-label="Próximas reservas">
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-forest/10 px-5 py-5 sm:px-6">
        <div><h2 className="font-serif text-2xl font-semibold text-forest">Próximas reservas</h2><p className="mt-1 text-sm text-ink/60">Estadias em andamento e chegadas futuras.</p></div>
        <Link to="/reservas" className="text-sm font-semibold text-orange underline underline-offset-4 hover:text-forest focus-visible:outline focus-visible:outline-2 focus-visible:outline-orange">Ver todas</Link>
      </div>
      {loading ? <p role="status" className="p-8 text-center text-ink/60">Carregando reservas…</p> : reservations.length === 0 ? <div className="p-10 text-center"><h3 className="font-serif text-xl font-semibold text-forest">Nenhuma estadia próxima</h3><p className="mt-2 text-sm text-ink/60">Quando uma reserva for confirmada, ela aparecerá aqui.</p><Link to="/reservas/nova" className="mt-4 inline-block text-sm font-semibold text-orange underline">Criar reserva</Link></div> : <div className="overflow-x-auto">
        <table className="w-full min-w-[620px] text-left text-sm">
          <caption className="sr-only">Próximas reservas do hotel</caption>
          <thead className="bg-sand/60 text-xs tracking-wide text-ink/60 uppercase"><tr><th className="px-5 py-4">Hóspede</th><th className="px-4 py-4">Quarto</th><th className="px-4 py-4">Entrada · saída</th><th className="px-5 py-4 text-right">Total</th></tr></thead>
          <tbody className="divide-y divide-forest/10">{reservations.map(reservation => <tr key={reservation.id} className="hover:bg-sand/30">
            <td className="max-w-64 break-words px-5 py-4 font-semibold text-forest">{reservation.guestName}</td>
            <td className="px-4 py-4">{reservation.roomNumber}<span className="mt-1 block text-xs text-ink/55">{reservation.categoryName}</span></td>
            <td className="whitespace-nowrap px-4 py-4">{formatReservationDate(reservation.checkIn)} <span className="text-ink/40">→</span> {formatReservationDate(reservation.checkOut)}</td>
            <td className="whitespace-nowrap px-5 py-4 text-right font-semibold tabular-nums">{formatBRL(reservation.totalAmount)}</td>
          </tr>)}</tbody>
        </table>
      </div>}
    </section>

    <section className="mt-8" aria-label="Acesso rápido">
      <div className="flex flex-wrap items-end justify-between gap-2"><div><h2 className="font-serif text-2xl font-semibold text-forest">Cadastros</h2><p className="mt-1 text-sm text-ink/60">Acesse os dados de apoio da recepção.</p></div></div>
      <div className="mt-4 grid gap-3 sm:grid-cols-3">{shortcuts.map(shortcut => <Link key={shortcut.to} to={shortcut.to} className="rounded-xl border border-forest/10 bg-cream p-4 transition-colors hover:border-forest/25 hover:bg-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-orange">
        <span className="font-semibold text-forest">{shortcut.title}</span><span className="mt-1 block text-sm text-ink/60">{shortcut.description}</span>
      </Link>)}</div>
    </section>
  </div>
}
