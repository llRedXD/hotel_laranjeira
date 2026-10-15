import { useEffect, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { EntityDialog } from '../components/EntityDialog'
import { allCategories, changeAvailability, getEntity, listEntities, pageSize, type Entity, type Resource } from '../services/hotel'

const content = {
  guests: { title: 'Hóspedes', singular: 'hóspede', description: 'Informações de quem passa pelo Hotel Laranjeira.', search: 'CPF ou RG; nome como alternativa' },
  categories: { title: 'Categorias', singular: 'categoria', description: 'Organize os tipos de acomodação do hotel.', search: '' },
  rooms: { title: 'Quartos', singular: 'quarto', description: 'Capacidade, diárias e disponibilidade em um só lugar.', search: 'Número do quarto' },
}
const control = 'rounded-xl border border-forest/20 bg-white px-4 py-3 text-sm'
const action = 'rounded-lg px-3 py-2 text-sm font-medium text-forest hover:bg-forest/10 focus-visible:outline-orange disabled:opacity-50'
export function RecordsPage({ resource }: { resource: Resource }) {
  const info = content[resource]
  const [records, setRecords] = useState<Entity[]>([])
  const [categories, setCategories] = useState<Entity[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState('')
  const [available, setAvailable] = useState('')
  const [page, setPage] = useState(0)
  const [revision, setRevision] = useState(0)
  const [busy, setBusy] = useState<string | null>(null)
  const [dialog, setDialog] = useState<{ entity?: Entity, readOnly: boolean } | null>(null)
  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true); setError('')
      try {
        const query = new URLSearchParams({ limit: String(pageSize), offset: String(page * pageSize), search })
        if (category) query.set('categoryId', category)
        if (available) query.set('isAvailable', available)
        const [items, options] = await Promise.all([listEntities(resource, query), resource === 'rooms' ? allCategories() : Promise.resolve([])])
        if (active) { setRecords(items); setCategories(options) }
      } catch (err) { if (active) setError(err instanceof Error ? err.message : 'Não foi possível carregar os registros.') }
      finally { if (active) setLoading(false) }
    }
    void load()
    return () => { active = false }
  }, [resource, search, category, available, page, revision])
  async function open(entity: Entity, readOnly: boolean) {
    setBusy(entity.id); setError('')
    try { setDialog({ entity: resource === 'categories' ? entity : await getEntity(resource, entity.id), readOnly }) }
    catch (err) { setError(err instanceof Error ? err.message : 'Falha ao consultar registro.') }
    finally { setBusy(null) }
  }
  async function toggle(entity: Entity) {
    setBusy(entity.id); setError(''); setSuccess('')
    try { await changeAvailability(entity.id, !entity.isAvailable); setSuccess('Disponibilidade atualizada.'); setRevision(value => value + 1) }
    catch (err) { setError(err instanceof Error ? err.message : 'Não foi possível atualizar.') }
    finally { setBusy(null) }
  }
  function submitSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSearch(String(new FormData(event.currentTarget).get('search') ?? '').trim()); setPage(0)
  }
  return <div className="mx-auto max-w-7xl px-5 py-8 sm:px-8 sm:py-10">
    <header className="flex flex-wrap items-end justify-between gap-4">
      <div><p className="text-sm font-semibold tracking-[0.18em] text-orange uppercase">Gestão do hotel</p><h1 className="mt-2 font-serif text-4xl font-semibold text-forest">{info.title}</h1><p className="mt-2 text-ink/65">{info.description}</p></div>
      <button type="button" onClick={() => { setSuccess(''); setDialog({ readOnly: false }) }} disabled={loading || !!error || (resource === 'rooms' && categories.length === 0)}
        className="rounded-full bg-forest px-6 py-3 text-sm font-semibold text-cream transition hover:bg-moss disabled:opacity-50">Cadastrar {info.singular}</button>
    </header>
    {resource === 'rooms' && <p className="mt-4 text-sm text-ink/65">Precisa de outro tipo de acomodação? <Link to="/categorias" className="font-semibold text-orange underline">Gerenciar categorias</Link></p>}
    {success && <p role="status" className="mt-6 rounded-xl bg-forest/10 p-4 text-forest">{success}</p>}
    <section aria-label="Consulta de registros" className="mt-8 overflow-hidden rounded-2xl border border-forest/10 bg-cream shadow-sm">
      <div className="flex flex-wrap items-end gap-3 border-b border-forest/10 p-5">
        {resource !== 'categories' && <form onSubmit={submitSearch} className="flex flex-1 flex-wrap items-end gap-2">
          <div className="min-w-48 flex-1"><label htmlFor="search" className="mb-1 block text-sm text-ink/70">{resource === 'guests' ? 'Buscar hóspede pelo documento' : `Buscar ${info.title.toLowerCase()}`}</label><input id="search" name="search" type={resource === 'guests' ? 'search' : 'text'} autoComplete="off" maxLength={255} placeholder={info.search} className={control + ' w-full'} />{resource === 'guests' && <p className="mt-1.5 text-xs text-ink/60">CPF e RG são priorizados nos resultados. Também é possível buscar pelo nome.</p>}</div>
          <button className={control + ' font-semibold text-forest'} type="submit">Buscar</button>
        </form>}
        {resource === 'rooms' && <>
          <div><label htmlFor="category-filter" className="mb-1 block text-sm text-ink/70">Categoria</label><select id="category-filter" className={control} value={category} onChange={e => { setCategory(e.target.value); setPage(0) }}><option value="">Todas as categorias</option>{categories.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}</select></div>
          <div><label htmlFor="availability-filter" className="mb-1 block text-sm text-ink/70">Disponibilidade</label><select id="availability-filter" className={control} value={available} onChange={e => { setAvailable(e.target.value); setPage(0) }}><option value="">Todos os status</option><option value="true">Disponível</option><option value="false">Indisponível</option></select></div>
        </>}
        {resource === 'categories' && <p className="text-sm text-ink/60">Categorias cadastradas no hotel</p>}
      </div>
      {loading ? <p role="status" className="p-8 text-center text-ink/60">Carregando registros...</p> : error ? <div role="alert" className="p-6 text-red-700"><p>{error}</p><button type="button" onClick={() => setRevision(v => v + 1)} className={action}>Tentar novamente</button></div> : records.length === 0 ? <div className="p-12 text-center"><h2 className="font-serif text-2xl text-forest">Nenhum registro encontrado</h2><p className="mt-2 text-ink/60">{search || category || available ? 'Revise os filtros para ampliar sua busca.' : `Comece cadastrando ${info.singular === 'categoria' ? 'uma categoria' : 'um '+info.singular}.`}</p></div> :
        <div className="overflow-x-auto"><table className="w-full text-left text-sm">
          <caption className="sr-only">Lista de {info.title.toLowerCase()}</caption>
          <thead className="bg-sand/60 text-xs tracking-wide text-ink/60 uppercase"><tr>{resource === 'guests' ? <><th className="px-5 py-4">Documento</th><th className="px-5 py-4">Nome</th></> : <th className="px-5 py-4">{resource === 'rooms' ? 'Quarto' : 'Nome'}</th>}
            {resource === 'guests' ? <th className="px-5 py-4">Contato</th> : resource === 'categories' ? <th className="px-5 py-4">Descrição</th> : <><th className="px-5 py-4">Categoria</th><th className="px-5 py-4">Capacidade</th><th className="px-5 py-4">Diária</th><th className="px-5 py-4">Status</th></>}
            <th className="px-5 py-4 text-right">Ações</th></tr></thead>
          <tbody className="divide-y divide-forest/10">{records.map(entity => <tr key={entity.id} className="hover:bg-sand/30">
            {resource === 'guests' ? <><td className="px-5 py-5">{entity.document ? <>{entity.documentType ? `${entity.documentType}: ` : ''}{entity.document}{!entity.documentType && <span className="ml-2 text-xs text-orange">não classificado</span>}</> : '—'}</td><td className="max-w-64 break-words px-5 py-5 font-semibold text-forest">{entity.name}</td><td className="px-5 py-5"><p>{entity.email || '—'}</p><p className="mt-1 text-ink/60">{entity.phone}</p></td></> : <><td className="max-w-64 break-words px-5 py-5 font-semibold text-forest">{entity.name ?? entity.number}</td>{resource === 'categories' ? <td className="max-w-xs break-words px-5 py-5 text-ink/65">{entity.description || '—'}</td> : <>
              <td className="px-5 py-5">{entity.categoryName}</td><td className="px-5 py-5">{entity.capacity} pessoas</td><td className="whitespace-nowrap px-5 py-5">{Number(entity.dailyRate).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}</td><td className="px-5 py-5"><span className={`rounded-full px-3 py-1 text-xs font-semibold ${entity.isAvailable ? 'bg-forest/10 text-forest' : 'bg-orange/15 text-orange'}`}>{entity.isAvailable ? 'Disponível' : 'Indisponível'}</span></td>
            </>}</>}
            <td className="px-3 py-4 text-right"><div className="flex flex-wrap justify-end gap-1">
              <button type="button" className={action} disabled={busy !== null} onClick={() => void open(entity, true)}>Detalhes</button>
              <button type="button" className={action} disabled={busy !== null} onClick={() => void open(entity, false)}>Editar</button>
              {resource === 'rooms' && <button type="button" className={action} disabled={busy !== null} onClick={() => void toggle(entity)}>{entity.isAvailable ? 'Indisponibilizar' : 'Disponibilizar'}</button>}
            </div></td>
          </tr>)}</tbody>
        </table></div>}
      <footer className="flex items-center justify-between border-t border-forest/10 px-5 py-4 text-sm text-ink/65">
        <p>Página {page + 1}</p><div className="flex gap-2"><button className={action} disabled={loading || page === 0} onClick={() => setPage(p => p - 1)}>Anterior</button><button className={action} disabled={loading || !!error || records.length < pageSize} onClick={() => setPage(p => p + 1)}>Próxima</button></div>
      </footer>
    </section>
    {!loading && !error && resource === 'rooms' && categories.length === 0 && <p className="mt-5 rounded-xl bg-orange/10 p-4 text-sm text-ink">Cadastre uma categoria antes de adicionar quartos. <Link to="/categorias" className="font-semibold underline">Cadastrar categoria</Link></p>}
    {dialog && <EntityDialog resource={resource} entity={dialog.entity} readOnly={dialog.readOnly} categories={categories} onClose={() => setDialog(null)} onSaved={() => { setDialog(null); setSuccess('Registro salvo com sucesso.'); setRevision(v => v + 1) }} />}
  </div>
}
