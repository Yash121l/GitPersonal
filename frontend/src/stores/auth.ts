import { defineStore } from 'pinia'

import { ApiError, api, type User } from '@/lib/api'

interface AuthState {
  currentUser: User | null
  initialized: boolean
  loading: boolean
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    currentUser: null,
    initialized: false,
    loading: false,
  }),
  getters: {
    isAuthenticated: (state) => state.currentUser !== null,
  },
  actions: {
    reset() {
      this.currentUser = null
      this.initialized = true
      this.loading = false
    },
    async ensureLoaded(force = false) {
      if (!force && (this.initialized || this.loading)) {
        return this.currentUser
      }

      this.loading = true
      try {
        const response = await api.me()
        this.currentUser = response.user
      } catch (error) {
        if (error instanceof ApiError && error.status === 401) {
          this.currentUser = null
        } else {
          throw error
        }
      } finally {
        this.initialized = true
        this.loading = false
      }

      return this.currentUser
    },
    async login(payload: { username: string; password: string }) {
      const response = await api.login(payload)
      this.currentUser = response.user
      this.initialized = true
      return response.user
    },
    async register(payload: { username: string; password: string }) {
      const response = await api.register(payload)
      this.currentUser = response.user
      this.initialized = true
      return response.user
    },
    async logout() {
      try {
        await api.logout()
      } finally {
        this.reset()
      }
    },
  },
})
