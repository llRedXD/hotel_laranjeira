import { Link } from 'react-router-dom'

const highlights = [
  ['Reservas organizadas', 'Tenha a operação do hotel sempre à mão.'],
  ['Acolhimento em cada detalhe', 'Uma experiência feita para quem recebe bem.'],
  ['Informações seguras', 'Acesse dados essenciais com tranquilidade.'],
]

export function HomePage() {
  return (
    <>
      <section className="relative overflow-hidden bg-forest px-5 py-20 text-cream sm:px-8 sm:py-28">
        <div className="absolute -top-24 -right-20 size-80 rounded-full bg-orange/30 blur-3xl" />
        <div className="absolute bottom-0 left-1/4 size-56 rounded-full bg-peach/15 blur-3xl" />
        <div className="relative mx-auto max-w-6xl">
          <p className="mb-5 text-sm font-semibold tracking-[0.22em] text-peach uppercase">Gestão hoteleira</p>
          <h1 className="max-w-3xl font-serif text-5xl leading-[1.05] font-semibold sm:text-7xl">O cuidado do seu hotel, em um só lugar.</h1>
          <p className="mt-6 max-w-xl text-lg leading-8 text-cream/80">Uma plataforma pensada para tornar a rotina da recepção mais leve, organizada e humana.</p>
          <div className="mt-9 flex flex-wrap gap-4">
            <Link to="/login" className="rounded-full bg-orange px-6 py-3 font-semibold text-white transition hover:bg-orange/85">Acessar sistema</Link>
            <Link to="/cadastro" className="rounded-full border border-cream/40 px-6 py-3 font-semibold text-cream transition hover:bg-cream/10">Criar acesso</Link>
          </div>
        </div>
      </section>
      <section className="mx-auto w-full max-w-6xl px-5 py-16 sm:px-8">
        <p className="text-sm font-semibold tracking-[0.18em] text-orange uppercase">Feito para receber</p>
        <h2 className="mt-3 max-w-xl font-serif text-3xl font-semibold text-forest sm:text-4xl">Tudo começa com uma boa experiência.</h2>
        <div className="mt-10 grid gap-5 md:grid-cols-3">
          {highlights.map(([title, description], index) => (
            <article key={title} className="rounded-2xl border border-forest/10 bg-cream p-6 shadow-sm">
              <span className="grid size-10 place-items-center rounded-full bg-orange/15 font-serif text-lg font-semibold text-orange">0{index + 1}</span>
              <h3 className="mt-5 font-serif text-xl font-semibold text-forest">{title}</h3>
              <p className="mt-2 leading-6 text-ink/65">{description}</p>
            </article>
          ))}
        </div>
      </section>
    </>
  )
}
