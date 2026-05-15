import { create } from 'zustand'
import type { AuthResponse } from '@/types/api'
import { parseError } from '@/lib/api'

interface AuthState {
  user: AuthResponse['user'] | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
  checkSession: () => Promise<void>
  changePassword: (oldPassword: string, newPassword: string) => Promise<void>
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  isLoading: false,

  login: async (username: string, password: string) => {
    set({ isLoading: true })
    try {
      const response = await window.go.handler.AuthHandler.Login(username, password)
      set({
        user: response.user,
        isAuthenticated: response.loggedIn,
        isLoading: false,
      })
    } catch (error) {
      set({ isLoading: false })
      throw parseError(error)
    }
  },

  logout: async () => {
    try {
      await window.go.handler.AuthHandler.Logout()
    } finally {
      set({ user: null, isAuthenticated: false })
    }
  },

  checkSession: async () => {
    try {
      const response = await window.go.handler.AuthHandler.GetCurrentUser()
      set({
        user: response.loggedIn ? response.user : null,
        isAuthenticated: response.loggedIn,
      })
    } catch {
      set({ user: null, isAuthenticated: false })
    }
  },

  changePassword: async (oldPassword: string, newPassword: string) => {
    try {
      await window.go.handler.AuthHandler.ChangePassword(oldPassword, newPassword)
    } catch (error) {
      throw parseError(error)
    }
  },
}))
