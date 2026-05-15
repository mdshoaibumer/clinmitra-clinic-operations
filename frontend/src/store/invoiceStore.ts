import { create } from 'zustand'
import type { Invoice, Payment } from '@/types/models'
import type { CreateInvoiceInput, RecordPaymentInput } from '@/types/api'
import { PAGE_SIZE } from '@/lib/constants'
import { parseError } from '@/lib/api'

interface InvoiceState {
  invoices: Invoice[]
  totalCount: number
  currentInvoice: Invoice | null
  page: number
  statusFilter: string
  isLoading: boolean
  fetchInvoices: () => Promise<void>
  fetchInvoice: (id: string) => Promise<void>
  createInvoice: (input: CreateInvoiceInput) => Promise<Invoice>
  recordPayment: (input: RecordPaymentInput) => Promise<Payment>
  voidInvoice: (id: string, reason: string) => Promise<void>
  setPage: (page: number) => void
  setStatusFilter: (status: string) => void
}

export const useInvoiceStore = create<InvoiceState>((set, get) => ({
  invoices: [],
  totalCount: 0,
  currentInvoice: null,
  page: 1,
  statusFilter: '',
  isLoading: false,

  fetchInvoices: async () => {
    set({ isLoading: true })
    try {
      const { page, statusFilter } = get()
      const response = await window.go.handler.InvoiceHandler.ListInvoices(
        page, PAGE_SIZE, statusFilter, '', '', '', ''
      )
      set({
        invoices: response.invoices || [],
        totalCount: response.total,
        isLoading: false,
      })
    } catch (error) {
      set({ isLoading: false })
      throw parseError(error)
    }
  },

  fetchInvoice: async (id: string) => {
    set({ isLoading: true })
    try {
      const invoice = await window.go.handler.InvoiceHandler.GetInvoice(id)
      set({ currentInvoice: invoice, isLoading: false })
    } catch (error) {
      set({ isLoading: false })
      throw parseError(error)
    }
  },

  createInvoice: async (input: CreateInvoiceInput) => {
    try {
      const invoice = await window.go.handler.InvoiceHandler.CreateInvoice(input)
      const { fetchInvoices } = get()
      await fetchInvoices()
      return invoice
    } catch (error) {
      throw parseError(error)
    }
  },

  recordPayment: async (input: RecordPaymentInput) => {
    try {
      const payment = await window.go.handler.InvoiceHandler.RecordPayment(input)
      // Refresh current invoice
      const { currentInvoice, fetchInvoice } = get()
      if (currentInvoice) {
        await fetchInvoice(currentInvoice.id)
      }
      return payment
    } catch (error) {
      throw parseError(error)
    }
  },

  voidInvoice: async (id: string, reason: string) => {
    try {
      await window.go.handler.InvoiceHandler.VoidInvoice(id, reason)
      const { fetchInvoice } = get()
      await fetchInvoice(id)
    } catch (error) {
      throw parseError(error)
    }
  },

  setPage: (page: number) => set({ page }),
  setStatusFilter: (status: string) => set({ statusFilter: status, page: 1 }),
}))
