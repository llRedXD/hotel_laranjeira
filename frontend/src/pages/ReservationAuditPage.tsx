import { useEffect, useState, type FormEvent } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { ReservationStatusBadge } from '../components/ReservationStatusBadge'
import { APIError } from '../services/api'
import { listReservationEvents, pageSize, type ReservationAuditState, type ReservationEvent } from '../services/hotel'
import { formatBRL, formatReservationDate } from '../services/reservations'

const actions = { imported: 'Registro inicial', created: 'Criação', cancelled: 'Cancelamento', updated: 'Alteração' }
const sources = { import: 'Migração da base', public: 'Site público', staff: 'Equipe do hotel', database: 'Banco de dados' }
const timestamp = new Intl.DateTimeFormat('pt-BR', { dateStyle: 'short', timeStyle: 'short', timeZone: 'America/Sao_Paulo' })
const button = 'rounded-lg px-3 py-2 text-sm font-semibold text-forest hover:bg-forest/10 focus-visible:outline-orange disabled:opacity-45'

function StateDetails({ state }: { state: ReservationAuditState | null }) {
  if (!state) return <p className="text-sm text-ink/60">Sem registro anterior.</p>
  return <dl className="space-y-2 text-sm">
    <div><dt className="text-ink/60">Status</dt><dd><ReservationStatusBadge status={state.status} /></dd></div>
    <div><dt className="text-ink/60">Quarto</dt><dd>{state.roomNumber}</dd></div>
    <div><dt className="text-ink/60">Período previsto</dt><dd>{formatReservationDate(state.checkIn)} → {formatReservationDate(state.checkOut)}</dd></div>
    <div><dt className="text-ink/60">Ocupantes</dt><dd>{state.occupants}</dd></div>
    <div><dt className="text-ink/60">Diária · total</dt><dd>{formatBRL(state.dailyRate)} · {formatBRL(state.totalAmount)}</dd></div>
  </dl>
}

export function ReservationAuditPage() {
  const [params, setParams] = useSearchParams()
  const reservationId = params.get('reservationId') ?? ''
  const search = params.get('search') ?? ''
  const requestedPage = Number(params.get('page') || 0)
  const page = Number.isInteger(requestedPage) && requestedPage >= 0 && requestedPage <= 5000 ? requestedPage : 0
  const [events, setEvents] = useState<ReservationEvent[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [revision, setRevision] = useState(0)
  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true); setError('')
      const query = new URLSearchParams({ limit: String(pageSize), offset: String(page * pageSize) })
      if (reservationId) query.set('reservationId', reservationId)
      if (search) query.set('search', search)
      try {
        const result = await listReservationEvents(query)
        if (active) { setEvents(result.events); setTotal(result.total) }
      } catch (cause) {
        if (active) {
          setEvents([]); setTotal(0)
          const fields = cause instanceof APIError ? Object.values(cause.fields).join(' ') : ''
          setError(fields || (cause instanceof Error ? cause.message : 'Não foi possível consultar a auditoria.'))
        }
      } finally { if (active) setLoading(false) }
    }
    void load()
    return () => { active = false }
  }, [reservationId, search, page, revision])
  function filter(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const next = new URLSearchParams()
    if (reservationId) next.set('reservationId', reservationId)
    const value = String(new FormData(event.currentTarget).get('search') ?? '').trim()
    if (value) next.set('search', value)
    setParams(next)
  }
  function changePage(nextPage: number) {
    const next = new URLSearchParams(params)
    if (nextPage) next.set('page', String(nextPage)); else next.delete('page')
    setParams(next)
  }
  return <div className="mx-auto max-w-7xl px-5 py-8 sm:px-8 sm:py-10">
    <header>
      <p className="text-sm font-semibold tracking-[0.18em] text-orange uppercase">Gestão do hotel</p>
      <h1 className="mt-2 font-serif text-4xl font-semibold text-forest">Auditoria de reservas</h1>
      <p className="mt-2 text-ink/65">Consulte ações realizadas, responsáveis e o estado anterior e posterior de cada reserva.</p>
      <p className="mt-2 text-sm text-ink/60">Reservas existentes começam com um registro inicial. Mudanças previstas pela passagem das datas não geram ações de auditoria.</p>
    </header>
    {reservationId && <div className="mt-6 rounded-xl bg-cream p-4"><p className="break-all text-sm">Reserva #{reservationId}</p><Link to="/auditoria" className="mt-2 inline-block font-semibold text-orange underline">Ver toda a auditoria</Link></div>}
    <form key={`${reservationId}:${search}`} onSubmit={filter} className="mt-6 flex flex-wrap items-end gap-3 rounded-2xl border border-forest/10 bg-cream p-5">
      <div className="min-w-48 flex-1"><label htmlFor="audit-search" className="mb-1.5 block text-sm font-medium">Hóspede, quarto ou referência da reserva</label><input id="audit-search" name="search" type="search" maxLength={255} defaultValue={search} className="w-full rounded-xl border border-forest/20 bg-white px-4 py-3 text-sm focus-visible:outline-orange" /></div>
      <button type="submit" className={button}>Buscar</button>
      {search && <Link className={button} to={reservationId ? `/auditoria?reservationId=${encodeURIComponent(reservationId)}` : '/auditoria'}>Limpar busca</Link>}
    </form>
    <section aria-label="Eventos de auditoria" className="mt-6 overflow-hidden rounded-2xl border border-forest/10 bg-cream shadow-sm">
      <div className="border-b border-forest/10 p-5"><h2 className="font-serif text-xl font-semibold text-forest">Ações registradas</h2><p className="mt-1 text-sm text-ink/60">{loading ? 'Consultando…' : error ? 'Consulta indisponível' : `${total} ${total === 1 ? 'registro encontrado' : 'registros encontrados'}`}</p></div>
      {loading ? <p role="status" className="p-8 text-center">Carregando auditoria…</p> : error ? <div role="alert" className="p-6 text-red-700"><p>{error}</p><button className={button} onClick={() => setRevision(value => value + 1)}>Tentar novamente</button></div> : events.length === 0 ? <p className="p-8 text-center text-ink/65">Nenhum registro encontrado.</p> : <ul className="divide-y divide-forest/10">{events.map(event => <li key={event.id} className="p-5 sm:p-6">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div><h3 className="font-semibold text-forest">{actions[event.action]} · {event.guestName}</h3><p className="mt-1 text-sm text-ink/65">Quarto {event.roomNumber} · {event.actorName ?? sources[event.source]}</p><Link to={`/auditoria?reservationId=${encodeURIComponent(event.reservationId)}`} className="mt-1 inline-block text-xs font-semibold text-orange underline">Reserva #{event.reservationId.slice(-8)}</Link></div>
          <time dateTime={event.occurredAt} className="text-sm text-ink/65">{timestamp.format(new Date(event.occurredAt))} · São Paulo</time>
        </div>
        <div className="mt-4 flex flex-wrap items-center gap-2">{event.beforeState ? <ReservationStatusBadge status={event.beforeState.status} /> : <span className="text-sm text-ink/60">Sem registro anterior</span>}<span aria-label="passou para">→</span><ReservationStatusBadge status={event.afterState.status} /></div>
        {event.action === 'imported' && <p className="mt-3 text-sm text-ink/60">Estado encontrado ao iniciar a auditoria; ações anteriores não foram reconstruídas.</p>}
        <details className="mt-4"><summary className="cursor-pointer rounded-lg py-2 text-sm font-semibold text-forest focus-visible:outline-orange">Ver antes e depois</summary><div className="mt-3 grid gap-5 rounded-xl bg-sand/50 p-4 sm:grid-cols-2"><div><h4 className="mb-3 font-semibold">Antes</h4><StateDetails state={event.beforeState} /></div><div><h4 className="mb-3 font-semibold">Depois</h4><StateDetails state={event.afterState} /></div></div></details>
      </li>)}</ul>}
      <footer className="flex flex-wrap items-center justify-between gap-3 border-t border-forest/10 p-5"><p className="text-sm text-ink/65">Página {page + 1}</p><div className="flex gap-2"><button className={button} disabled={loading || page === 0} onClick={() => changePage(page - 1)}>Anterior</button><button className={button} disabled={loading || !!error || (page + 1) * pageSize >= total} onClick={() => changePage(page + 1)}>Próxima</button></div></footer>
    </section>
  </div>
}
