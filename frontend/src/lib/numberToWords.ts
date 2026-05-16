/**
 * Converts a number to words in the Indian Numbering System (Lakhs, Crores)
 */
export function numberToWords(num: number): string {
  if (num === 0) return 'Zero'

  const units = ['', 'One', 'Two', 'Three', 'Four', 'Five', 'Six', 'Seven', 'Eight', 'Nine', 'Ten', 'Eleven', 'Twelve', 'Thirteen', 'Fourteen', 'Fifteen', 'Sixteen', 'Seventeen', 'Eighteen', 'Nineteen']
  const tens = ['', '', 'Twenty', 'Thirty', 'Forty', 'Fifty', 'Sixty', 'Seventy', 'Eighty', 'Ninety']

  function convert(n: number): string {
    if (n < 20) return units[n]
    if (n < 100) return tens[Math.floor(n / 10)] + (n % 10 !== 0 ? ' ' + units[n % 10] : '')
    if (n < 1000) return units[Math.floor(n / 100)] + ' Hundred' + (n % 100 !== 0 ? ' and ' + convert(n % 100) : '')
    if (n < 100000) return convert(Math.floor(n / 1000)) + ' Thousand' + (n % 1000 !== 0 ? ' ' + convert(n % 1000) : '')
    if (n < 10000000) return convert(Math.floor(n / 100000)) + ' Lakh' + (n % 100000 !== 0 ? ' ' + convert(n % 100000) : '')
    return convert(Math.floor(n / 10000000)) + ' Crore' + (n % 10000000 !== 0 ? ' ' + convert(n % 10000000) : '')
  }

  return convert(num).trim()
}

/**
 * Specifically for Indian Rupees (handling paise as well)
 */
export function amountToWords(amountPaise: number): string {
  const rupees = Math.floor(amountPaise / 100)
  const paise = Math.round(amountPaise % 100)

  let result = numberToWords(rupees) + ' Rupees'
  if (paise > 0) {
    result += ' and ' + numberToWords(paise) + ' Paise'
  }
  return result + ' Only'
}
