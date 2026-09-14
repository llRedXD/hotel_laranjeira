import { useState, type FormEvent } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { AuthCard } from '../components/AuthCard'
import { FormField } from '../components/FormField'
import { SubmitButton } from '../components/SubmitButton'
import { useAuth } from '../hooks/useAuth'

type LoginErrors = { email?: string; password?: string }

export function LoginPage() {
  const [errors, setErrors] = useState<LoginErrors>({})
  const [isSubmitting, setIsSubmitting] = useState(false)
  const { login } = useAuth()
  const location = useLocation()
  const navigate = useNavigate()

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const data = new FormData(event.currentTarget)
    const email = String(data.get('email') ?? '').trim()
    const password = String(data.get('password') ?? '')
    const nextErrors: LoginErrors = {}

    if (!email) nextErrors.email = 'Informe seu e-mail.'
    else if (!/^\S+@\S+\.\S+$/.test(email)) nextErrors.email = 'Informe um e-mail válido.'
    if (!password) nextErrors.password = 'Informe sua senha.'

    setErrors(nextErrors)
    if (Object.keys(nextErrors).length === 0) {
      setIsSubmitting(true)
      try {
        await login(email, password)
        const destination = typeof location.state?.from === 'string' ? location.state.from : '/painel'
        navigate(destination, { replace: true })
      } catch (error) {
        setErrors({ password: error instanceof Error ? error.message : 'Não foi possível entrar.' })
      } finally {
        setIsSubmitting(false)
      }
    }
  }

  return (
    <AuthCard eyebrow="Acesso da equipe" title="Entre na sua conta" description="Acesse o ambiente de gestão do Hotel Laranjeira com suas credenciais autorizadas.">
      <form className="space-y-5" noValidate onSubmit={handleSubmit}>
        <FormField error={errors.email} label="E-mail" name="email" type="email" autoComplete="email" placeholder="voce@hotel.com" />
        <div>
          <FormField error={errors.password} label="Senha" name="password" type="password" autoComplete="current-password" placeholder="Sua senha" />
          <Link className="mt-2 inline-block text-sm font-medium text-orange hover:underline" to="/recuperar-senha">Esqueci minha senha</Link>
        </div>
        <SubmitButton disabled={isSubmitting}>{isSubmitting ? 'Entrando...' : 'Entrar'}</SubmitButton>
      </form>
    </AuthCard>
  )
}
