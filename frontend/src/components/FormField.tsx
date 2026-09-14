import type { InputHTMLAttributes } from 'react'

type FormFieldProps = InputHTMLAttributes<HTMLInputElement> & {
  error?: string
  label: string
}

export function FormField({ error, id: explicitId, label, ...inputProps }: FormFieldProps) {
  const id = explicitId ?? inputProps.name
  return (
    <div>
      <label className="mb-1.5 block text-sm font-medium text-ink" htmlFor={id}>{label}</label>
      <input
        id={id}
        className={`w-full rounded-xl border bg-white px-4 py-3 text-ink outline-none transition placeholder:text-ink/35 focus:ring-4 ${error ? 'border-red-400 focus:border-red-500 focus:ring-red-100' : 'border-forest/15 focus:border-orange focus:ring-orange/15'}`}
        aria-invalid={Boolean(error)}
        aria-describedby={error ? `${id}-error` : undefined}
        {...inputProps}
      />
      {error && <p id={`${id}-error`} className="mt-1.5 text-sm text-red-600">{error}</p>}
    </div>
  )
}
