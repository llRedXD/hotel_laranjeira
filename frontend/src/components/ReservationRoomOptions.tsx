import { formatBRL, type AvailabilityOption } from '../services/reservations'

export function ReservationRoomOptions({ rooms, onSelect, selectedRoomId }: {
  rooms: AvailabilityOption[]
  selectedRoomId?: string
  onSelect: (room: AvailabilityOption) => void
}) {
  if (rooms.length === 0) return <div className="rounded-2xl border border-forest/10 bg-cream p-8 text-center"><h2 className="font-serif text-2xl text-forest">Nenhum quarto disponível</h2><p className="mt-2 text-sm text-ink/65">Tente outras datas ou uma quantidade menor de pessoas.</p></div>
  return <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
    {rooms.map(room => <article key={room.roomId} className={`rounded-2xl border bg-cream p-5 shadow-sm ${selectedRoomId === room.roomId ? 'border-orange ring-2 ring-orange/25' : 'border-forest/10'}`}>
      <p className="text-xs font-semibold tracking-[0.16em] text-orange uppercase">{room.categoryName}</p>
      <h3 className="mt-2 font-serif text-2xl font-semibold text-forest">Quarto {room.number}</h3>
      <p className="mt-2 text-sm text-ink/65">Até {room.capacity} pessoas</p>
      <dl className="mt-5 space-y-2 border-t border-forest/10 pt-4 text-sm">
        <div className="flex justify-between gap-3"><dt className="text-ink/65">Diária</dt><dd className="font-medium">{formatBRL(room.dailyRate)}</dd></div>
        <div className="flex justify-between gap-3"><dt className="text-ink/65">Diárias</dt><dd className="font-medium">{room.nights}</dd></div>
        <div className="flex justify-between gap-3 text-base"><dt className="font-semibold text-forest">Total previsto</dt><dd className="font-semibold text-forest">{formatBRL(room.estimatedTotal)}</dd></div>
      </dl>
      <button type="button" className="mt-5 w-full rounded-xl border border-forest px-4 py-3 text-sm font-semibold text-forest hover:bg-forest hover:text-cream" onClick={() => onSelect(room)}>{selectedRoomId === room.roomId ? 'Quarto selecionado' : 'Escolher este quarto'}</button>
    </article>)}
  </div>
}
