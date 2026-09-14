import type { ReactNode } from 'react'

type AuthCardProps = {
  children: ReactNode
  eyebrow: string
  title: string
  description: string
}

export function AuthCard({ children, eyebrow, title, description }: AuthCardProps) {
  return (
    <section className="mx-auto my-10 w-full max-w-md px-5 sm:my-16">
      <div className="rounded-3xl border border-forest/10 bg-cream p-7 shadow-xl shadow-forest/5 sm:p-9">
        <p className="mb-2 text-sm font-semibold tracking-[0.18em] text-orange uppercase">{eyebrow}</p>
        <h1 className="font-serif text-3xl font-semibold text-forest">{title}</h1>
        <p className="mt-3 text-sm leading-6 text-ink/65">{description}</p>
        <div className="mt-7">{children}</div>
      </div>
    </section>
  )
}
