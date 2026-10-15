export function formatGuestDocument(type: 'RG' | 'CPF', value: string) {
  if (type !== 'CPF') return value.slice(0, 30)

  const digits = value.replace(/\D/g, '').slice(0, 11)
  if (digits.length <= 3) return digits
  if (digits.length <= 6) return `${digits.slice(0, 3)}.${digits.slice(3)}`
  if (digits.length <= 9) return `${digits.slice(0, 3)}.${digits.slice(3, 6)}.${digits.slice(6)}`
  return `${digits.slice(0, 3)}.${digits.slice(3, 6)}.${digits.slice(6, 9)}-${digits.slice(9)}`
}

export function isValidCPF(value: string) {
  const digits = value.replace(/\D/g, '')
  if (digits.length !== 11 || /^(\d)\1{10}$/.test(digits)) return false

  const checkDigit = (length: number) => {
    let sum = 0
    for (let index = 0; index < length; index += 1) {
      sum += Number(digits[index]) * (length + 1 - index)
    }
    const remainder = (sum * 10) % 11
    return remainder === 10 ? 0 : remainder
  }

  return checkDigit(9) === Number(digits[9]) && checkDigit(10) === Number(digits[10])
}

export function guestDocumentError(type: 'RG' | 'CPF', value: string) {
  if (!value.trim()) return 'Informe o número do documento.'
  if (type === 'CPF') {
    const digits = value.replace(/\D/g, '')
    if (digits.length !== 11) return 'O CPF precisa ter 11 números.'
    if (!isValidCPF(value)) return 'CPF inválido. Confira os números digitados.'
  }
  if (type === 'RG' && value.trim().length > 30) return 'O RG deve ter até 30 caracteres.'
  return ''
}
