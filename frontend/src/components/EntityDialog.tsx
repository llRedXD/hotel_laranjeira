import { useEffect, useRef, useState, type ChangeEvent, type FormEvent } from 'react'
import { APIError } from '../services/api'
import { saveEntity, type Entity, type Resource } from '../services/hotel'
import { FormField } from './FormField'
import { SubmitButton } from './SubmitButton'

export type FieldSpec = { key: string, label: string, required?: boolean, type?: string, max?: number }
const fields: Record<Resource, FieldSpec[]> = {
  guests: [
    { key: 'name', label: 'Nome completo', required: true, max: 120 },
    { key: 'documentType', label: 'Tipo de documento' },
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

function formatDocument(value: string) {
  const clean = value.replace(/[^\dA-Za-z]/g, '').toUpperCase().slice(0, 30)
  if (!/^\d+$/.test(clean) || clean.length > 11) return clean
  if (clean.length <= 3) return clean
  if (clean.length <= 6) return `${clean.slice(0, 3)}.${clean.slice(3)}`
  if (clean.length <= 9) return `${clean.slice(0, 3)}.${clean.slice(3, 6)}.${clean.slice(6)}`
  return `${clean.slice(0, 3)}.${clean.slice(3, 6)}.${clean.slice(6, 9)}-${clean.slice(9)}`
}

function formatPhone(value: string) {
  const digits = value.replace(/\D/g, '').slice(0, 11)
  if (digits.length <= 2) return digits.length ? `(${digits}` : ''
  if (digits.length <= 6) return `(${digits.slice(0, 2)}) ${digits.slice(2)}`
  const split = digits.length > 10 ? 7 : 6
  return `(${digits.slice(0, 2)}) ${digits.slice(2, split)}-${digits.slice(split)}`
}
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
    if (resource === 'guests') {
      values.document = values.document?.replace(/[^\dA-Za-z]/g, '').toUpperCase() || ''
      values.phone = values.phone?.replace(/\D/g, '') || ''
      if (values.phone && !/^\d{10,11}$/.test(values.phone)) {
        setErrors({ phone: 'Informe um telefone com DDD (10 ou 11 dígitos).' }); setMessage('Revise os campos destacados.'); return
      }
      if (values.email && !/^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/i.test(values.email.trim())) {
        setErrors({ email: 'Informe um e-mail válido.' }); setMessage('Revise os campos destacados.'); return
      }
    }
    setBusy(true); setErrors({}); setMessage('')
    try { await saveEntity(resource, entity?.id, values); onSaved() }
    catch (error) {
      setMessage(error instanceof Error ? error.message : 'Não foi possível salvar.')
      if (error instanceof APIError) setErrors(error.fields)
    } finally { setBusy(false) }
  }
  function value(key: string): string {
    const raw = entity?.[key as keyof Entity]
    if (raw == null) return ''
    const text = String(raw)
    if (resource === 'guests' && key === 'document') return formatDocument(text)
    if (resource === 'guests' && key === 'phone') return formatPhone(text)
    return text
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
          {fields[resource].map(field => field.key === 'documentType' ? <div key={field.key}><label htmlFor={field.key} className="mb-1.5 block text-sm font-medium">{field.label}</label><select className={selectClass} id={field.key} name={field.key} defaultValue={value(field.key)} aria-invalid={!!errors[field.key]}><option value="">Não classificado</option><option value="RG">RG</option><option value="CPF">CPF</option></select>{errors[field.key] && <p role="alert" className="text-sm text-red-700">{errors[field.key]}</p>}</div> : <FormField key={field.key} id={field.key} name={field.key} label={field.label + (field.required ? ' *' : '')}
            type={field.type ?? 'text'} defaultValue={value(field.key)} required={field.required} maxLength={field.max}
            autoComplete={resource === 'guests' ? ({ name: 'name', email: 'email', phone: 'tel', address: 'street-address' }[field.key] ?? 'off') : 'off'}
            inputMode={resource === 'guests' && (field.key === 'document' || field.key === 'phone') ? 'tel' : undefined}
            pattern={resource === 'guests' && field.key === 'email' ? '^[^\\s@]+@[^\\s@]+\\.[^\\s@]{2,}$' : undefined}
            onChange={resource === 'guests' && (field.key === 'document' || field.key === 'phone') ? (event: ChangeEvent<HTMLInputElement>) => {
              event.currentTarget.value = field.key === 'document' ? formatDocument(event.currentTarget.value) : formatPhone(event.currentTarget.value)
            } : undefined}
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
