import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { AuthCard } from '../components/AuthCard'
import { FormField } from '../components/FormField'
import { SubmitButton } from '../components/SubmitButton'

type RegisterErrors = { name?: string; email?: string; password?: string; confirmation?: string }

export function RegisterPage() {
  const [errors, setErrors] = useState<RegisterErrors>({})
  const [submitted, setSubmitted] = useState(false)

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const data = new FormData(event.currentTarget)
    const name = String(data.get('name') ?? '').trim()
    const email = String(data.get('email') ?? '').trim()
    const password = String(data.get('password') ?? '')
    const confirmation = String(data.get('confirmation') ?? '')
    const nextErrors: RegisterErrors = {}

    if (!name) nextErrors.name = 'Informe seu nome.'
    if (!email) nextErrors.email = 'Informe seu e-mail.'
    else if (!/^\S+@\S+\.\S+$/.test(email)) nextErrors.email = 'Informe um e-mail válido.'
    if (password.length < 6) nextErrors.password = 'Use uma senha com pelo menos 6 caracteres.'
    if (confirmation !== password) nextErrors.confirmation = 'As senhas não coincidem.'

    setErrors(nextErrors)
    setSubmitted(Object.keys(nextErrors).length === 0)
  }

  return (
    <AuthCard eyebrow="Primeiro acesso" title="Crie sua conta" description="Preencha seus dados para começar a usar a plataforma.">
      <form className="space-y-5" noValidate onSubmit={handleSubmit}>
        <FormField error={errors.name} label="Nome completo" name="name" type="text" autoComplete="name" placeholder="Como devemos chamar você?" />
        <FormField error={errors.email} label="E-mail" name="email" type="email" autoComplete="email" placeholder="voce@hotel.com" />
        <FormField error={errors.password} label="Senha" name="password" type="password" autoComplete="new-password" placeholder="Mínimo de 6 caracteres" />
        <FormField error={errors.confirmation} label="Confirmar senha" name="confirmation" type="password" autoComplete="new-password" placeholder="Repita sua senha" />
        {submitted && <p className="rounded-xl bg-forest/10 px-4 py-3 text-sm text-forest">Cadastro validado. A criação de conta será conectada em breve.</p>}
        <SubmitButton>Criar acesso</SubmitButton>
      </form>
      <p className="mt-6 text-center text-sm text-ink/65">Já possui acesso? <Link className="font-semibold text-orange hover:underline" to="/login">Entre</Link>.</p>
    </AuthCard>
  )
}
