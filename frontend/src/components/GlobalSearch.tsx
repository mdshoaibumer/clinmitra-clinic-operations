import { useEffect, useState, useRef, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { formatCurrency } from '@/lib/utils'
import { Input } from '@/components/ui/input'
import { Kbd } from '@/components/ui/kbd'
import {
  Search,
  Users,
  Receipt,
  Calendar,
  X,
} from 'lucide-react'
import type { Patient } from '@/types/models'
import type { Invoice } from '@/types/models'

interface SearchResult {
  type: 'patient' | 'invoice' | 'appointment' | 'action'
  id: string
  title: string
  subtitle?: string
  path: string
  icon: typeof Users
}

export default function GlobalSearch() {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<SearchResult[]>([])
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [isSearching, setIsSearching] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const navigate = useNavigate()

  // Open with Ctrl+K
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault()
        setOpen(true)
      }
      if (e.key === 'Escape' && open) {
        e.preventDefault()
        setOpen(false)
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [open])

  // Focus input when opened
  useEffect(() => {
    if (open) {
      setTimeout(() => inputRef.current?.focus(), 50)
    } else {
      setQuery('')
      setResults([])
      setSelectedIndex(0)
    }
  }, [open])

  // Search function
  const performSearch = useCallback(async (q: string) => {
    if (!q.trim()) {
      setResults([])
      return
    }

    setIsSearching(true)
    const searchResults: SearchResult[] = []

    try {
      // Search patients
      const patientResponse = await window.go.handler.PatientHandler.ListPatients(1, 5, q)
      if (patientResponse?.patients) {
        patientResponse.patients.forEach((p: Patient) => {
          searchResults.push({
            type: 'patient',
            id: p.id,
            title: p.name,
            subtitle: p.phone,
            path: `/patients/${p.id}`,
            icon: Users,
          })
        })
      }
    } catch { /* ignore */ }

    try {
      // Search invoices by number
      const invoiceResponse = await window.go.handler.InvoiceHandler.ListInvoices(1, 5, '', q, '', '', '')
      if (invoiceResponse?.invoices) {
        invoiceResponse.invoices.forEach((inv: Invoice) => {
          searchResults.push({
            type: 'invoice',
            id: inv.id,
            title: `Invoice ${inv.invoiceNumber}`,
            subtitle: `${inv.patient?.name || 'Patient'} • ${formatCurrency(inv.totalAmount)}`,
            path: `/billing/${inv.id}`,
            icon: Receipt,
          })
        })
      }
    } catch { /* ignore */ }

    // Quick actions
    const actions: SearchResult[] = [
      { type: 'action', id: 'new-patient', title: 'New Patient', subtitle: 'Register a new patient', path: '/patients?action=new', icon: Users },
      { type: 'action', id: 'new-invoice', title: 'New Invoice', subtitle: 'Create a new invoice', path: '/billing?action=new', icon: Receipt },
      { type: 'action', id: 'new-appointment', title: 'New Appointment', subtitle: 'Book an appointment', path: '/appointments?action=new', icon: Calendar },
    ]

    const matchingActions = actions.filter(a =>
      a.title.toLowerCase().includes(q.toLowerCase())
    )
    searchResults.push(...matchingActions)

    setResults(searchResults)
    setSelectedIndex(0)
    setIsSearching(false)
  }, [])

  // Debounced search
  useEffect(() => {
    const timer = setTimeout(() => {
      performSearch(query)
    }, 200)
    return () => clearTimeout(timer)
  }, [query, performSearch])

  const handleSelect = (result: SearchResult) => {
    navigate(result.path)
    setOpen(false)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setSelectedIndex((i) => Math.min(i + 1, results.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setSelectedIndex((i) => Math.max(i - 1, 0))
    } else if (e.key === 'Enter' && results[selectedIndex]) {
      e.preventDefault()
      handleSelect(results[selectedIndex])
    }
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[20vh]">
      {/* Backdrop */}
      <div className="fixed inset-0 bg-black/50" onClick={() => setOpen(false)} />

      {/* Search Dialog */}
      <div className="relative w-full max-w-lg bg-white rounded-lg shadow-2xl border overflow-hidden">
        {/* Search Input */}
        <div className="flex items-center border-b px-4">
          <Search className="h-4 w-4 text-muted-foreground mr-2" />
          <Input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Search patients, invoices, or type a command..."
            className="border-0 focus-visible:ring-0 focus-visible:ring-offset-0 h-12"
          />
          <button onClick={() => setOpen(false)} className="p-1 hover:bg-muted rounded">
            <X className="h-4 w-4" />
          </button>
        </div>

        {/* Results */}
        <div className="max-h-[300px] overflow-y-auto">
          {isSearching && (
            <div className="p-4 text-center text-sm text-muted-foreground">Searching...</div>
          )}

          {!isSearching && query && results.length === 0 && (
            <div className="p-4 text-center text-sm text-muted-foreground">No results found.</div>
          )}

          {!isSearching && results.length > 0 && (
            <div className="py-2">
              {results.map((result, index) => (
                <button
                  key={`${result.type}-${result.id}`}
                  className={`w-full flex items-center gap-3 px-4 py-2.5 text-left text-sm hover:bg-muted/50 ${
                    index === selectedIndex ? 'bg-muted' : ''
                  }`}
                  onClick={() => handleSelect(result)}
                  onMouseEnter={() => setSelectedIndex(index)}
                >
                  <result.icon className="h-4 w-4 text-muted-foreground flex-shrink-0" />
                  <div className="flex-1 min-w-0">
                    <p className="font-medium truncate">{result.title}</p>
                    {result.subtitle && (
                      <p className="text-xs text-muted-foreground truncate">{result.subtitle}</p>
                    )}
                  </div>
                  <span className="text-xs text-muted-foreground capitalize">{result.type}</span>
                </button>
              ))}
            </div>
          )}

          {!query && (
            <div className="p-4 space-y-2">
              <p className="text-xs text-muted-foreground font-medium">Quick Actions</p>
              <button
                className="w-full flex items-center gap-3 px-3 py-2 text-left text-sm hover:bg-muted rounded-md"
                onClick={() => { navigate('/patients?action=new'); setOpen(false) }}
              >
                <Users className="h-4 w-4 text-muted-foreground" />
                <span>New Patient</span>
                <Kbd className="ml-auto">Ctrl+N</Kbd>
              </button>
              <button
                className="w-full flex items-center gap-3 px-3 py-2 text-left text-sm hover:bg-muted rounded-md"
                onClick={() => { navigate('/billing?action=new'); setOpen(false) }}
              >
                <Receipt className="h-4 w-4 text-muted-foreground" />
                <span>New Invoice</span>
                <Kbd className="ml-auto">Ctrl+B</Kbd>
              </button>
              <button
                className="w-full flex items-center gap-3 px-3 py-2 text-left text-sm hover:bg-muted rounded-md"
                onClick={() => { navigate('/appointments?action=new'); setOpen(false) }}
              >
                <Calendar className="h-4 w-4 text-muted-foreground" />
                <span>New Appointment</span>
              </button>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="border-t px-4 py-2 flex items-center gap-4 text-xs text-muted-foreground">
          <span><Kbd>↑↓</Kbd> Navigate</span>
          <span><Kbd>↵</Kbd> Select</span>
          <span><Kbd>Esc</Kbd> Close</span>
        </div>
      </div>
    </div>
  )
}
