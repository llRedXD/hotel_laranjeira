import { useState, type FormEvent } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { AuthCard } from '../components/AuthCard'
import { FormField } from '../components/FormField'
import { SubmitButton } from '../components/SubmitButton'
import { resetPassword } from '../services/api'

type ResetErrors = { password?: string; confirmation?: string; token?: string }

export function PasswordResetPage() {
  const [errors, setErrors] = useState<ResetErrors>({})
  const [reset, setReset] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [searchParams] = useSearchParams()

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const data = new FormData(event.currentTarget)
    const password = String(data.get('password') ?? '')
    const confirmation = String(data.get('confirmation') ?? '')
    const nextErrors: ResetErrors = {}

    if (password.length < 12) nextErrors.password = 'Use uma senha com pelo menos 12 caracteres.'
    if (confirmation !== password) nextErrors.confirmation = 'As senhas não coincidem.'

    setErrors(nextErrors)
    if (Object.keys(nextErrors).length === 0) {
      const token = searchParams.get('token') ?? ''
      if (!token) {
        setErrors({ token: 'O link de redefinição é inválido ou está incompleto.' })
        return
      }
      setIsSubmitting(true)
      try {
        await resetPassword(token, password)
        setReset(true)
      } catch (error) {
        setErrors({ token: error instanceof Error ? error.message : 'Não foi possível redefinir a senha.' })
      } finally {
        setIsSubmitting(false)
      }
    }
  }

  if (reset) {
    return (
      <AuthCard eyebrow="Senha atualizada" title="Tudo pronto!" description="Sua nova senha foi validada nesta demonstração.">
        <div className="space-y-5">
          <p className="rounded-xl bg-forest/10 p-4 text-sm leading-6 text-forest">Quando a API estiver conectada, sua senha será atualizada e os links de recuperação utilizados deixarão de ser válidos.</p>
          <Link className="block w-full rounded-xl bg-forest px-4 py-3 text-center font-semibold text-cream transition hover:bg-moss" to="/login">Entrar com a nova senha</Link>
        </div>
      </AuthCard>
    )
  }

  return (
    <AuthCard eyebrow="Nova senha" title="Redefina sua senha" description="Escolha uma senha segura para voltar a acessar sua conta.">
      <form className="space-y-5" noValidate onSubmit={handleSubmit}>
        <FormField error={errors.password} label="Nova senha" name="password" type="password" autoComplete="new-password" placeholder="Mínimo de 12 caracteres" />
        <FormField error={errors.confirmation} label="Confirmar nova senha" name="confirmation" type="password" autoComplete="new-password" placeholder="Repita sua nova senha" />
        {errors.token && <p className="text-sm text-red-600">{errors.token}</p>}
        <SubmitButton disabled={isSubmitting}>{isSubmitting ? 'Redefinindo...' : 'Redefinir senha'}</SubmitButton>
      </form>
    </AuthCard>
  )
}
