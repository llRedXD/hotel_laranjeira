import { useAuth } from '../hooks/useAuth'
import { Link } from 'react-router-dom'

const metrics = [
  { label: 'Reservas ativas', value: '24', detail: '5 novas desde ontem', accent: 'bg-orange/15 text-orange' },
  { label: 'Hóspedes hospedados', value: '38', detail: 'Em 21 quartos ocupados', accent: 'bg-forest/10 text-forest' },
  { label: 'Quartos disponíveis', value: '12', detail: 'de 35 quartos', accent: 'bg-moss/15 text-moss' },
  { label: 'Check-ins previstos', value: '7', detail: 'para hoje', accent: 'bg-peach/35 text-orange' },
]

const movement = [
  { time: '11:00', guest: 'Marina Azevedo', room: 'Suíte 203', type: 'Check-out', color: 'bg-orange' },
  { time: '13:30', guest: 'Rafael Martins', room: 'Quarto 108', type: 'Check-in', color: 'bg-forest' },
  { time: '14:00', guest: 'Helena Costa', room: 'Suíte 301', type: 'Check-in', color: 'bg-forest' },
  { time: '15:30', guest: 'Lucas Almeida', room: 'Quarto 112', type: 'Check-in', color: 'bg-forest' },
]

const shortcuts = [
  { title: 'Hóspedes', description: 'Cadastre e consulte visitantes.', to: '/hospedes' },
  { title: 'Quartos', description: 'Veja disponibilidade e diárias.', to: '/quartos' },
  { title: 'Categorias', description: 'Organize os tipos de acomodação.', to: '/categorias' },
]

export function DashboardPage() {
  const { user } = useAuth()

  return (
    <div className="mx-auto max-w-7xl px-5 py-8 sm:px-8 sm:py-10">
      <header className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-sm font-semibold tracking-[0.18em] text-orange uppercase">Visão geral</p>
          <h1 className="mt-2 font-serif text-3xl font-semibold text-forest sm:text-4xl">Olá, {user?.name}.</h1>
          <p className="mt-2 text-ink/65">Acesse os cadastros do Hotel Laranjeira. Os indicadores abaixo são demonstrativos.</p>
        </div>
        <p className="rounded-full bg-cream px-4 py-2 text-sm font-medium text-ink/65 shadow-sm">{new Date().toLocaleDateString('pt-BR', { dateStyle: 'long' })}</p>
      </header>

      <section className="mt-8 grid gap-4 sm:grid-cols-2 xl:grid-cols-4" aria-label="Resumo operacional">
        {metrics.map((metric) => (
          <article key={metric.label} className="rounded-2xl border border-forest/10 bg-cream p-5 shadow-sm">
            <span className={`inline-flex rounded-lg px-2.5 py-1 text-xs font-semibold ${metric.accent}`}>{metric.label}</span>
            <p className="mt-4 font-serif text-4xl font-semibold text-forest">{metric.value}</p>
            <p className="mt-1 text-sm text-ink/60">{metric.detail}</p>
          </article>
        ))}
      </section>

      <div className="mt-8 grid gap-8 xl:grid-cols-[1.45fr_1fr]">
        <section className="rounded-2xl border border-forest/10 bg-cream p-6 shadow-sm sm:p-7">
          <div className="flex items-center justify-between gap-4">
            <div>
              <h2 className="font-serif text-2xl font-semibold text-forest">Movimento de hoje</h2>
              <p className="mt-1 text-sm text-ink/60">Check-ins e check-outs programados.</p>
            </div>
            <span className="rounded-full bg-sand px-3 py-1 text-sm font-medium text-forest">{movement.length} atividades</span>
          </div>
          <ol className="mt-6 divide-y divide-forest/10">
            {movement.map((item) => (
              <li key={`${item.time}-${item.guest}`} className="flex items-center gap-4 py-4 first:pt-0 last:pb-0">
                <span className={`size-2 rounded-full ${item.color}`} aria-hidden="true" />
                <time className="w-11 text-sm font-semibold text-forest">{item.time}</time>
                <div className="min-w-0 flex-1">
                  <p className="truncate font-medium text-ink">{item.guest}</p>
                  <p className="text-sm text-ink/60">{item.room}</p>
                </div>
                <span className="rounded-full bg-sand px-3 py-1 text-xs font-semibold text-ink/70">{item.type}</span>
              </li>
            ))}
          </ol>
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
