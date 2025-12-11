import { create } from 'zustand'
import { api } from '@/shared/lib/api'
import { getUser } from '@/shared/lib/telegram'

interface User {
  id: number
  first_name?: string
  last_name?: string
  username?: string
}

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null
  login: () => Promise<void>
  logout: () => void
  setUser: (user: User) => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  isAuthenticated: false,
  isLoading: false,
  error: null,

  login: async () => {
    set({ isLoading: true, error: null })
    try {
      const tgUser = getUser()
      if (!tgUser) {
        throw new Error('Telegram user not available')
      }

      const response = await api.auth()
      const user: User = {
        id: tgUser.id,
        first_name: tgUser.first_name,
        last_name: tgUser.last_name,
        username: tgUser.username,
      }

      set({
        user,
        token: response.token,
        isAuthenticated: true,
        isLoading: false,
      })
    } catch (error: any) {
      set({
        error: error.message || 'Authentication failed',
        isLoading: false,
        isAuthenticated: false,
      })
    }
  },

  logout: () => {
    api.setToken('')
    set({
      user: null,
      token: null,
      isAuthenticated: false,
    })
  },

  setUser: (user: User) => {
    set({ user })
  },
}))

