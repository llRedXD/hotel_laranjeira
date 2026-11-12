import type { ReservationStatus } from '../services/hotel'

const statuses: Record<ReservationStatus, { label: string, style: string }> = {
  confirmed: { label: 'Confirmada', style: 'bg-orange/15 text-orange' },
  in_progress: { label: 'Estadia prevista em andamento', style: 'bg-moss/15 text-forest' },
  departure_today: { label: 'Saída prevista hoje', style: 'bg-peach/35 text-orange' },
  ended: { label: 'Encerrada pelo período', style: 'bg-sand text-ink/65' },
  cancelled: { label: 'Cancelada', style: 'bg-red-50 text-red-700' },
}

export function ReservationStatusBadge({ status }: { status: ReservationStatus }) {
  const display = statuses[status] ?? { label: 'Status indisponível', style: 'bg-sand text-ink/65' }
  return <span className={`inline-block rounded-full px-3 py-1 text-xs font-semibold ${display.style}`}>{display.label}</span>
}
