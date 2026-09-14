import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useState } from 'react'
import { useAuth } from '../hooks/useAuth'

const menuItems = [
  { label: 'Visão geral', to: '/painel', icon: '⌂' },
  { label: 'Reservas', to: '/painel', icon: '▣', disabled: true },
  { label: 'Hóspedes', to: '/painel', icon: '♙', disabled: true },
  { label: 'Quartos', to: '/painel', icon: '▤', disabled: true },
  { label: 'Pagamentos', to: '/painel', icon: '◈', disabled: true },
]

export function DashboardLayout() {
  const { logout, user } = useAuth()
  const navigate = useNavigate()
  const [logoutError, setLogoutError] = useState('')

  async function handleLogout() {
    try {
      await logout()
      navigate('/login', { replace: true })
    } catch {
      setLogoutError('Não foi possível encerrar a sessão. Tente novamente.')
    }
  }

  return (
    <div className="min-h-screen bg-sand lg:flex">
      <aside className="bg-forest px-5 py-5 text-cream lg:flex lg:w-72 lg:flex-col lg:px-6 lg:py-8">
        {logoutError && <p role="alert">{logoutError}</p>}
        <div className="flex items-center justify-between lg:block">
          <div className="flex items-center gap-3">
            <span className="grid size-10 place-items-center rounded-full bg-orange text-lg font-bold text-white">L</span>
            <div>
              <p className="font-serif text-xl font-semibold">Hotel Laranjeira</p>
              <p className="text-xs text-cream/60">Gestão hoteleira</p>
            </div>
          </div>
          <button className="rounded-lg px-3 py-2 text-sm text-cream/70 hover:bg-cream/10 lg:hidden" type="button" onClick={handleLogout}>Sair</button>
        </div>

        <nav className="mt-6 flex gap-2 overflow-x-auto pb-1 lg:mt-12 lg:flex-col" aria-label="Módulos do sistema">
          {menuItems.map((item) => (
            <NavLink
              key={item.label}
              to={item.to}
              aria-disabled={item.disabled}
              onClick={(event) => {
                if (item.disabled) event.preventDefault()
              }}
              className={({ isActive }) => `flex shrink-0 items-center gap-3 rounded-xl px-4 py-3 text-sm font-medium transition ${isActive && !item.disabled ? 'bg-cream text-forest shadow-sm' : 'text-cream/70 hover:bg-cream/10 hover:text-cream'} ${item.disabled ? 'cursor-not-allowed opacity-55' : ''}`}
            >
              <span aria-hidden="true" className="text-base">{item.icon}</span>
              {item.label}
              {item.disabled && <span className="ml-auto text-[10px] tracking-wide uppercase lg:inline">em breve</span>}
            </NavLink>
          ))}
        </nav>

        <div className="mt-8 hidden border-t border-cream/15 pt-6 lg:block">
          <p className="truncate text-sm font-semibold">{user?.name}</p>
          <p className="mt-1 truncate text-xs text-cream/60">{user?.email}</p>
          <button className="mt-5 w-full rounded-xl border border-cream/25 px-4 py-2.5 text-sm font-semibold transition hover:bg-cream/10" type="button" onClick={handleLogout}>Sair da conta</button>
        </div>
      </aside>
      <main className="min-w-0 flex-1"><Outlet /></main>
    </div>
  )
}
