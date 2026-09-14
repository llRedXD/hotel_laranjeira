import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { AuthCard } from '../components/AuthCard'
import { FormField } from '../components/FormField'
import { SubmitButton } from '../components/SubmitButton'
import { requestPasswordRecovery } from '../services/api'

type RecoveryErrors = { email?: string }

export function PasswordRecoveryPage() {
  const [errors, setErrors] = useState<RecoveryErrors>({})
  const [requested, setRequested] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const email = String(new FormData(event.currentTarget).get('email') ?? '').trim()
    const nextErrors: RecoveryErrors = {}

    if (!email) nextErrors.email = 'Informe seu e-mail.'
    else if (!/^\S+@\S+\.\S+$/.test(email)) nextErrors.email = 'Informe um e-mail válido.'

    setErrors(nextErrors)
    if (Object.keys(nextErrors).length === 0) {
      setIsSubmitting(true)
      try {
        await requestPasswordRecovery(email)
        setRequested(true)
      } catch (error) {
        setErrors({ email: error instanceof Error ? error.message : 'Não foi possível enviar as instruções.' })
      } finally {
        setIsSubmitting(false)
      }
    }
  }

  if (requested) {
    return (
      <AuthCard eyebrow="Pedido enviado" title="Verifique seu e-mail" description="Se houver uma conta para este endereço, você receberá um link seguro para redefinir sua senha.">
        <div className="space-y-5">
          <div className="rounded-xl bg-forest/10 p-4 text-sm leading-6 text-forest">
            Por segurança, não informamos se o e-mail está cadastrado. O link terá validade limitada quando a autenticação estiver conectada.
          </div>
          <Link className="block text-center text-sm font-medium text-orange hover:underline" to="/login">Voltar para entrar</Link>
        </div>
      </AuthCard>
    )
  }

  return (
    <AuthCard eyebrow="Recuperar acesso" title="Esqueceu sua senha?" description="Informe o e-mail associado à sua conta para receber as instruções de redefinição.">
      <form className="space-y-5" noValidate onSubmit={handleSubmit}>
        <FormField error={errors.email} label="E-mail" name="email" type="email" autoComplete="email" placeholder="voce@hotel.com" />
        <SubmitButton disabled={isSubmitting}>{isSubmitting ? 'Enviando...' : 'Enviar instruções'}</SubmitButton>
      </form>
      <p className="mt-6 text-center text-sm text-ink/65">Lembrou sua senha? <Link className="font-semibold text-orange hover:underline" to="/login">Volte para entrar</Link>.</p>
    </AuthCard>
  )
}
