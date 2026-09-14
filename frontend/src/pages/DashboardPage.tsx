import { useEffect, useState } from 'react'
import { useAuth } from '../hooks/useAuth'
import { Link } from 'react-router-dom'
import { listAll, type Entity } from '../services/hotel'

const shortcuts = [
  { title: 'Hóspedes', description: 'Cadastre e consulte visitantes.', to: '/hospedes' },
  { title: 'Quartos', description: 'Veja disponibilidade e diárias.', to: '/quartos' },
  { title: 'Categorias', description: 'Organize os tipos de acomodação.', to: '/categorias' },
]

export function DashboardPage() {
  const { user } = useAuth()
  const [guests, setGuests] = useState<Entity[]>([])
  const [rooms, setRooms] = useState<Entity[]>([])
  const [categories, setCategories] = useState<Entity[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  useEffect(() => {
    let active = true
    Promise.all([listAll('guests'), listAll('rooms'), listAll('categories')])
      .then(([guestData, roomData, categoryData]) => {
        if (!active) return
        setGuests(guestData); setRooms(roomData); setCategories(categoryData)
      })
      .catch(err => { if (active) setError(err instanceof Error ? err.message : 'Não foi possível carregar os indicadores.') })
      .finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [])
  const availableRooms = rooms.filter(room => room.isAvailable).length
  const metrics = [
    { label: 'Hóspedes cadastrados', value: guests.length, detail: 'Registros ativos no sistema', accent: 'bg-orange/15 text-orange' },
    { label: 'Quartos disponíveis', value: availableRooms, detail: `de ${rooms.length} quartos cadastrados`, accent: 'bg-moss/15 text-moss' },
    { label: 'Quartos cadastrados', value: rooms.length, detail: 'Acomodações do hotel', accent: 'bg-forest/10 text-forest' },
    { label: 'Categorias de quarto', value: categories.length, detail: 'Tipos de acomodação', accent: 'bg-peach/35 text-orange' },
  ]

  return (
    <div className="mx-auto max-w-7xl px-5 py-8 sm:px-8 sm:py-10">
      <header className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-sm font-semibold tracking-[0.18em] text-orange uppercase">Visão geral</p>
          <h1 className="mt-2 font-serif text-3xl font-semibold text-forest sm:text-4xl">Olá, {user?.name}.</h1>
          <p className="mt-2 text-ink/65">Acompanhe os cadastros e a disponibilidade atual do Hotel Laranjeira.</p>
        </div>
        <p className="rounded-full bg-cream px-4 py-2 text-sm font-medium text-ink/65 shadow-sm">{new Date().toLocaleDateString('pt-BR', { dateStyle: 'long' })}</p>
      </header>

      {error && <div role="alert" className="mt-6 rounded-xl bg-red-50 p-4 text-red-700"><p>{error}</p><button type="button" onClick={() => window.location.reload()} className="mt-2 font-semibold underline">Tentar novamente</button></div>}

      <section className="mt-8 grid gap-4 sm:grid-cols-2 xl:grid-cols-4" aria-label="Resumo operacional">
        {metrics.map((metric) => (
          <article key={metric.label} className="rounded-2xl border border-forest/10 bg-cream p-5 shadow-sm">
            <span className={`inline-flex rounded-lg px-2.5 py-1 text-xs font-semibold ${metric.accent}`}>{metric.label}</span>
            <p className="mt-4 font-serif text-4xl font-semibold text-forest">{loading ? '—' : metric.value}</p>
            <p className="mt-1 text-sm text-ink/60">{loading ? 'Consultando a API...' : metric.detail}</p>
          </article>
        ))}
      </section>

      <div className="mt-8 grid gap-8 xl:grid-cols-[1.45fr_1fr]">
        <section className="rounded-2xl border border-forest/10 bg-cream p-6 shadow-sm sm:p-7">
          <div className="flex items-center justify-between gap-4">
            <div>
              <h2 className="font-serif text-2xl font-semibold text-forest">Status das acomodações</h2>
              <p className="mt-1 text-sm text-ink/60">Disponibilidade atual dos quartos cadastrados.</p>
            </div>
            <span className="rounded-full bg-sand px-3 py-1 text-sm font-medium text-forest">{loading ? '—' : `${availableRooms}/${rooms.length}`}</span>
          </div>
          <div className="mt-6 rounded-xl bg-sand/60 p-5"><p className="text-4xl font-serif font-semibold text-forest">{loading ? '—' : `${availableRooms} disponíveis`}</p><p className="mt-2 text-sm text-ink/60">{rooms.length - availableRooms} indisponíveis no momento.</p><Link to="/quartos" className="mt-4 inline-block text-sm font-semibold text-orange underline">Gerenciar quartos</Link></div>
        </section>

        <section className="rounded-2xl bg-forest p-6 text-cream shadow-sm sm:p-7">
          <p className="text-sm font-semibold tracking-[0.18em] text-peach uppercase">Acesso rápido</p>
          <h2 className="mt-2 font-serif text-2xl font-semibold">Próximos passos</h2>
          <div className="mt-6 space-y-3">
            {shortcuts.map((shortcut) => (
              <Link key={shortcut.title} to={shortcut.to} className="block w-full rounded-xl border border-cream/15 p-4 text-left transition hover:bg-cream/10">
                <span className="block font-semibold">{shortcut.title}</span>
                <span className="mt-1 block text-sm text-cream/65">{shortcut.description}</span>
              </Link>
            ))}
          </div>
          <p className="mt-5 text-xs text-cream/55">Reservas e pagamentos estarão disponíveis em uma próxima etapa.</p>
        </section>
      </div>
    </div>
  )
}
