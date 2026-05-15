import { create } from 'zustand'
import type { Patient, PatientTreatment } from '@/types/models'
import type { CreatePatientInput } from '@/types/api'
import { PAGE_SIZE } from '@/lib/constants'
import { parseError } from '@/lib/api'

interface PatientState {
  patients: Patient[]
  totalCount: number
  currentPatient: Patient | null
  patientHistory: PatientTreatment[]
  searchQuery: string
  page: number
  isLoading: boolean
  fetchPatients: () => Promise<void>
  fetchPatient: (id: string) => Promise<void>
  fetchPatientHistory: (patientId: string) => Promise<void>
  createPatient: (input: CreatePatientInput) => Promise<Patient>
  updatePatient: (id: string, input: CreatePatientInput) => Promise<Patient>
  deletePatient: (id: string) => Promise<void>
  setSearch: (query: string) => void
  setPage: (page: number) => void
}

export const usePatientStore = create<PatientState>((set, get) => ({
  patients: [],
  totalCount: 0,
  currentPatient: null,
  patientHistory: [],
  searchQuery: '',
  page: 1,
  isLoading: false,

  fetchPatients: async () => {
    set({ isLoading: true })
    try {
      const { page, searchQuery } = get()
      const response = await window.go.handler.PatientHandler.ListPatients(page, PAGE_SIZE, searchQuery)
      set({
        patients: response.patients || [],
        totalCount: response.total,
        isLoading: false,
      })
    } catch (error) {
      set({ isLoading: false })
      throw parseError(error)
    }
  },

  fetchPatient: async (id: string) => {
    set({ isLoading: true })
    try {
      const patient = await window.go.handler.PatientHandler.GetPatient(id)
      set({ currentPatient: patient, isLoading: false })
    } catch (error) {
      set({ isLoading: false })
      throw parseError(error)
    }
  },

  fetchPatientHistory: async (patientId: string) => {
    try {
      const history = await window.go.handler.PatientHandler.GetPatientHistory(patientId)
      set({ patientHistory: history || [] })
    } catch {
      set({ patientHistory: [] })
    }
  },

  createPatient: async (input: CreatePatientInput) => {
    try {
      const patient = await window.go.handler.PatientHandler.CreatePatient(input)
      const { fetchPatients } = get()
      await fetchPatients()
      return patient
    } catch (error) {
      throw parseError(error)
    }
  },

  updatePatient: async (id: string, input: CreatePatientInput) => {
    try {
      const patient = await window.go.handler.PatientHandler.UpdatePatient(id, input)
      set({ currentPatient: patient })
      return patient
    } catch (error) {
      throw parseError(error)
    }
  },

  deletePatient: async (id: string) => {
    try {
      await window.go.handler.PatientHandler.DeletePatient(id)
      const { fetchPatients } = get()
      await fetchPatients()
    } catch (error) {
      throw parseError(error)
    }
  },

  setSearch: (query: string) => {
    set({ searchQuery: query, page: 1 })
  },

  setPage: (page: number) => {
    set({ page })
  },
}))
