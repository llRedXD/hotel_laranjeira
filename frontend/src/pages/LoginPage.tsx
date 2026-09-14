import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { AuthCard } from '../components/AuthCard'
import { FormField } from '../components/FormField'
import { SubmitButton } from '../components/SubmitButton'

type LoginErrors = { email?: string; password?: string }

export function LoginPage() {
  const [errors, setErrors] = useState<LoginErrors>({})
  const [submitted, setSubmitted] = useState(false)

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const data = new FormData(event.currentTarget)
    const email = String(data.get('email') ?? '').trim()
    const password = String(data.get('password') ?? '')
    const nextErrors: LoginErrors = {}

    if (!email) nextErrors.email = 'Informe seu e-mail.'
    else if (!/^\S+@\S+\.\S+$/.test(email)) nextErrors.email = 'Informe um e-mail válido.'
    if (!password) nextErrors.password = 'Informe sua senha.'

    setErrors(nextErrors)
    setSubmitted(Object.keys(nextErrors).length === 0)
  }

  return (
    <AuthCard eyebrow="Bem-vindo de volta" title="Entre na sua conta" description="Acesse o ambiente de gestão do Hotel Laranjeira.">
      <form className="space-y-5" noValidate onSubmit={handleSubmit}>
        <FormField error={errors.email} label="E-mail" name="email" type="email" autoComplete="email" placeholder="voce@hotel.com" />
        <FormField error={errors.password} label="Senha" name="password" type="password" autoComplete="current-password" placeholder="Sua senha" />
        {submitted && <p className="rounded-xl bg-forest/10 px-4 py-3 text-sm text-forest">Dados validados. A autenticação será conectada em breve.</p>}
        <SubmitButton>Entrar</SubmitButton>
      </form>
      <p className="mt-6 text-center text-sm text-ink/65">Ainda não tem acesso? <Link className="font-semibold text-orange hover:underline" to="/cadastro">Cadastre-se</Link>.</p>
    </AuthCard>
  )
}
