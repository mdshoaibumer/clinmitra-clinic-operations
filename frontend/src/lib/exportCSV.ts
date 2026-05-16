/**
 * Export data to CSV and trigger download
 */
export function exportToCSV(data: Record<string, unknown>[], filename: string, columns?: { key: string; header: string }[]) {
  if (!data || data.length === 0) return

  const cols = columns || Object.keys(data[0]).map(key => ({ key, header: key }))

  // Build CSV content
  const headers = cols.map(c => `"${c.header}"`).join(',')
  const rows = data.map(row =>
    cols.map(col => {
      const value = row[col.key]
      if (value === null || value === undefined) return '""'
      const str = String(value).replace(/"/g, '""')
      return `"${str}"`
    }).join(',')
  )

  const csvContent = [headers, ...rows].join('\n')

  // Add BOM for Excel compatibility with special characters (₹, etc.)
  const bom = '\uFEFF'
  const blob = new Blob([bom + csvContent], { type: 'text/csv;charset=utf-8;' })

  // Trigger download
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `${filename}.csv`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(link.href)
}

/**
 * Format paise as rupee string for CSV export (no HTML)
 */
export function formatCurrencyForExport(paise: number): string {
  const rupees = paise / 100
  return `₹${rupees.toFixed(2)}`
}
