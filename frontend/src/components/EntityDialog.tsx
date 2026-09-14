import { useEffect, useRef, useState, type FormEvent } from 'react'
import { APIError } from '../services/api'
import { saveEntity, type Entity, type Resource } from '../services/hotel'
import { FormField } from './FormField'
import { SubmitButton } from './SubmitButton'

export type FieldSpec = { key: string, label: string, required?: boolean, type?: string, max?: number }
const fields: Record<Resource, FieldSpec[]> = {
  guests: [
    { key: 'name', label: 'Nome completo', required: true, max: 120 },
    { key: 'document', label: 'Documento', max: 30 },
    { key: 'email', label: 'E-mail', type: 'email', max: 255 },
    { key: 'phone', label: 'Telefone', type: 'tel', max: 30 },
    { key: 'address', label: 'Endereço', max: 2000 },
  ],
  categories: [{ key: 'name', label: 'Nome da categoria', required: true, max: 80 }, { key: 'description', label: 'Descrição', max: 2000 }],
  rooms: [{ key: 'number', label: 'Número do quarto', required: true, max: 20 }, { key: 'capacity', label: 'Capacidade', type: 'number', required: true }, { key: 'dailyRate', label: 'Diária (R$)', type: 'number', required: true }],
}
const names: Record<Resource, string> = { guests: 'hóspede', categories: 'categoria', rooms: 'quarto' }
const selectClass = 'w-full rounded-xl border border-forest/20 bg-white px-4 py-3 text-ink'
export function EntityDialog({ resource, entity, readOnly, categories, onClose, onSaved }: {
  resource: Resource, entity?: Entity, readOnly: boolean, categories: Entity[], onClose: () => void, onSaved: () => void
}) {
  const dialog = useRef<HTMLDialogElement>(null)
  const [busy, setBusy] = useState(false)
  const [message, setMessage] = useState('')
  const [errors, setErrors] = useState<Record<string, string>>({})
  useEffect(() => { const element = dialog.current; element?.showModal(); return () => element?.close() }, [])
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (busy) return
    const values = Object.fromEntries(new FormData(event.currentTarget).entries()) as Record<string, string>
    setBusy(true); setErrors({}); setMessage('')
    try { await saveEntity(resource, entity?.id, values); onSaved() }
    catch (error) {
      setMessage(error instanceof Error ? error.message : 'Não foi possível salvar.')
      if (error instanceof APIError) setErrors(error.fields)
    } finally { setBusy(false) }
  }
  function value(key: string): string {
    const raw = entity?.[key as keyof Entity]
    return raw == null ? '' : String(raw)
  }
  return (
    <dialog ref={dialog} onCancel={event => { event.preventDefault(); if (!busy) onClose() }}
      aria-labelledby="entity-title" className="m-auto max-h-[90dvh] w-[calc(100%-2rem)] max-w-xl overflow-y-auto rounded-2xl bg-cream p-6 text-ink shadow-xl backdrop:bg-forest/60 sm:p-8">
      <div className="mb-6 flex items-start justify-between gap-4">
        <div>
          <p className="text-xs font-semibold tracking-widest text-orange uppercase">{readOnly ? 'Consulta' : 'Cadastro'}</p>
          <h2 id="entity-title" className="mt-1 font-serif text-2xl text-forest">{readOnly ? 'Detalhes do registro' : `${entity ? 'Editar' : 'Cadastrar'} ${names[resource]}`}</h2>
        </div>
        <button type="button" disabled={busy} onClick={onClose} aria-label="Fechar" className="rounded-lg px-3 py-2 hover:bg-sand">✕</button>
      </div>
      {readOnly ? <>
        <dl className="space-y-4">
          {fields[resource].map(field => <div key={field.key}><dt className="text-sm text-ink/60">{field.label}</dt><dd className="mt-1 break-words font-medium">{field.key === 'dailyRate' ? Number(value(field.key)).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' }) : value(field.key) || 'Não informado'}</dd></div>)}
          {resource === 'rooms' && <><div><dt className="text-sm text-ink/60">Categoria</dt><dd>{entity?.categoryName}</dd></div><div><dt className="text-sm text-ink/60">Disponibilidade</dt><dd>{entity?.isAvailable ? 'Disponível' : 'Indisponível'}</dd></div></>}
        </dl>
        <button type="button" onClick={onClose} className="mt-6 rounded-xl bg-forest px-5 py-3 text-cream">Fechar detalhes</button>
      </> : <form onSubmit={submit} className="space-y-4">
        <p className="text-sm text-ink/60">Campos marcados com * são obrigatórios.</p>
        <fieldset disabled={busy} className="space-y-4">
          {fields[resource].map(field => <FormField key={field.key} id={field.key} name={field.key} label={field.label + (field.required ? ' *' : '')}
            type={field.type ?? 'text'} defaultValue={value(field.key)} required={field.required} maxLength={field.max}
            min={field.key === 'capacity' ? 1 : field.key === 'dailyRate' ? 0 : undefined}
            max={field.key === 'capacity' ? 2147483647 : field.key === 'dailyRate' ? '9999999999.99' : undefined}
            step={field.key === 'dailyRate' ? '0.01' : field.key === 'capacity' ? '1' : undefined} error={errors[field.key]} />)}
          {resource === 'rooms' && <>
            <div><label htmlFor="categoryId" className="mb-1.5 block text-sm font-medium">Categoria *</label>
              <select className={selectClass} id="categoryId" name="categoryId" required defaultValue={entity?.categoryId ?? ''} aria-invalid={!!errors.categoryId}>
                <option value="">Selecione uma categoria</option>
                {categories.map(category => <option key={category.id} value={category.id}>{category.name}</option>)}
              </select>{errors.categoryId && <p role="alert" className="text-sm text-red-700">{errors.categoryId}</p>}
            </div>
            <div><label htmlFor="isAvailable" className="mb-1.5 block text-sm font-medium">Disponibilidade *</label>
              <select className={selectClass} id="isAvailable" name="isAvailable" defaultValue={String(entity?.isAvailable ?? true)}><option value="true">Disponível</option><option value="false">Indisponível</option></select>
            </div>
          </>}
        </fieldset>
        {message && <p role="alert" className="rounded-lg bg-red-50 p-3 text-sm text-red-700">{message}</p>}
        <div className="flex gap-3 pt-3">
          <button type="button" disabled={busy} onClick={onClose} className="rounded-xl border border-forest/20 px-5 py-3">Cancelar</button>
          <SubmitButton disabled={busy || (resource === 'rooms' && categories.length === 0)}>{busy ? 'Salvando...' : 'Salvar'}</SubmitButton>
        </div>
      </form>}
    </dialog>
  )
}
