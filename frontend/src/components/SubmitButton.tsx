import type { ButtonHTMLAttributes } from 'react'

export function SubmitButton({ children, ...props }: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      className="w-full rounded-xl bg-forest px-4 py-3 font-semibold text-cream transition hover:bg-moss focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-orange disabled:cursor-not-allowed disabled:opacity-60"
      type="submit"
      {...props}
    >
      {children}
    </button>
  )
}
