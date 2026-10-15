type ReservationSearchValues = { checkIn: string, checkOut: string, occupants: number }

const control = 'w-full rounded-xl border border-forest/20 bg-white px-4 py-3 text-sm text-ink focus:border-orange focus:outline-none focus:ring-4 focus:ring-orange/15'

export function ReservationSearchForm({ value, busy, onChange, onSubmit }: {
  value: ReservationSearchValues
  busy: boolean
  onChange: (value: ReservationSearchValues) => void
  onSubmit: (event: React.FormEvent<HTMLFormElement>) => void
}) {
  return <form onSubmit={onSubmit} className="grid gap-4 rounded-2xl border border-forest/10 bg-cream p-5 shadow-sm sm:grid-cols-2 lg:grid-cols-4">
    <div><label htmlFor="checkIn" className="mb-1.5 block text-sm font-medium text-ink">Check-in previsto</label><input className={control} id="checkIn" name="checkIn" type="date" required value={value.checkIn} onChange={event => onChange({ ...value, checkIn: event.target.value })} /></div>
    <div><label htmlFor="checkOut" className="mb-1.5 block text-sm font-medium text-ink">Check-out previsto</label><input className={control} id="checkOut" name="checkOut" type="date" required min={value.checkIn || undefined} value={value.checkOut} onChange={event => onChange({ ...value, checkOut: event.target.value })} /></div>
    <div><label htmlFor="occupants" className="mb-1.5 block text-sm font-medium text-ink">Pessoas</label><input className={control} id="occupants" name="occupants" type="number" min="1" step="1" required value={value.occupants} onChange={event => onChange({ ...value, occupants: Number(event.target.value) })} /></div>
    <div className="flex items-end"><button className="w-full rounded-xl bg-forest px-5 py-3 text-sm font-semibold text-cream transition hover:bg-moss disabled:opacity-60" type="submit" disabled={busy}>{busy ? 'Consultando...' : 'Consultar quartos'}</button></div>
  </form>
}
