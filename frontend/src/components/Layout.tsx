import { NavLink, Outlet } from 'react-router-dom'

const linkClassName = ({ isActive }: { isActive: boolean }) =>
  `transition-colors hover:text-orange ${isActive ? 'text-forest' : 'text-ink/70'}`

export function Layout() {
  return (
    <div className="flex min-h-screen flex-col">
      <header className="border-b border-forest/10 bg-cream/90 backdrop-blur">
        <nav className="mx-auto flex max-w-6xl items-center justify-between px-5 py-4 sm:px-8" aria-label="Navegação principal">
          <NavLink to="/reservar" className="flex items-center gap-2 text-forest">
            <span className="grid size-9 place-items-center rounded-full bg-orange text-lg font-bold text-white">L</span>
            <span className="font-serif text-xl font-semibold tracking-tight">Hotel Laranjeira</span>
          </NavLink>
          <div className="flex items-center gap-4 text-sm font-medium sm:gap-6">
            <NavLink to="/reservar" className={linkClassName}>Reservar</NavLink>
            <NavLink to="/login" className={linkClassName}>Entrar</NavLink>
          </div>
        </nav>
      </header>
      <main className="flex flex-1 flex-col"><Outlet /></main>
      <footer className="border-t border-forest/10 bg-cream px-5 py-5 text-center text-sm text-ink/60">
        Hotel Laranjeira · Hospitalidade que começa no detalhe.
      </footer>
    </div>
  )
}
