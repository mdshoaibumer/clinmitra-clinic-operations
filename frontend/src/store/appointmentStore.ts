import { create } from 'zustand'
import type { Appointment } from '@/types/models'
import type { CreateAppointmentInput } from '@/types/api'
import { parseError } from '@/lib/api'

interface AppointmentState {
  todayAppointments: Appointment[]
  appointments: Appointment[]
  selectedDate: string
  isLoading: boolean
  fetchToday: () => Promise<void>
  fetchByDate: (date: string) => Promise<void>
  fetchWeek: (startDate: string, endDate: string) => Promise<void>
  createAppointment: (input: CreateAppointmentInput) => Promise<Appointment>
  cancelAppointment: (id: string, reason: string) => Promise<void>
  completeAppointment: (id: string) => Promise<void>
  setSelectedDate: (date: string) => void
}

export const useAppointmentStore = create<AppointmentState>((set, get) => ({
  todayAppointments: [],
  appointments: [],
  selectedDate: new Date().toISOString().split('T')[0],
  isLoading: false,

  fetchToday: async () => {
    set({ isLoading: true })
    try {
      const appointments = await window.go.handler.AppointmentHandler.GetTodayAppointments()
      set({ todayAppointments: appointments || [], isLoading: false })
    } catch (error) {
      set({ isLoading: false })
      throw parseError(error)
    }
  },

  fetchByDate: async (date: string) => {
    set({ isLoading: true })
    try {
      const appointments = await window.go.handler.AppointmentHandler.GetAppointmentsByDate(date)
      set({ appointments: appointments || [], isLoading: false })
    } catch (error) {
      set({ isLoading: false })
      throw parseError(error)
    }
  },

  fetchWeek: async (startDate: string, endDate: string) => {
    set({ isLoading: true })
    try {
      const appointments = await window.go.handler.AppointmentHandler.GetWeekAppointments(startDate, endDate)
      set({ appointments: appointments || [], isLoading: false })
    } catch (error) {
      set({ isLoading: false })
      throw parseError(error)
    }
  },

  createAppointment: async (input: CreateAppointmentInput) => {
    try {
      const appointment = await window.go.handler.AppointmentHandler.CreateAppointment(input)
      const { fetchToday, selectedDate, fetchByDate } = get()
      await fetchToday()
      await fetchByDate(selectedDate)
      return appointment
    } catch (error) {
      throw parseError(error)
    }
  },

  cancelAppointment: async (id: string, reason: string) => {
    try {
      await window.go.handler.AppointmentHandler.CancelAppointment(id, reason)
      const { fetchToday, selectedDate, fetchByDate } = get()
      await fetchToday()
      await fetchByDate(selectedDate)
    } catch (error) {
      throw parseError(error)
    }
  },

  completeAppointment: async (id: string) => {
    try {
      await window.go.handler.AppointmentHandler.CompleteAppointment(id)
      const { fetchToday, selectedDate, fetchByDate } = get()
      await fetchToday()
      await fetchByDate(selectedDate)
    } catch (error) {
      throw parseError(error)
    }
  },

  setSelectedDate: (date: string) => set({ selectedDate: date }),
}))
