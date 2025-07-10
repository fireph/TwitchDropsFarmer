import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, AuthStatus, DeviceCodeResponse } from '@/types'
import { apiService } from '@/services/api'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!user.value)

  async function checkAuthStatus() {
    try {
      isLoading.value = true
      error.value = null
      
      const response = await apiService.get<AuthStatus>('/api/auth/status')
      
      if (response.is_logged_in && response.user) {
        user.value = response.user
      } else {
        user.value = null
      }
    } catch (err) {
      console.error('Auth status check failed:', err)
      user.value = null
    } finally {
      isLoading.value = false
    }
  }

  async function initiateDeviceFlow(): Promise<DeviceCodeResponse> {
    try {
      isLoading.value = true
      error.value = null
      
      const response = await apiService.get<DeviceCodeResponse>('/api/auth/url')
      return response
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to initiate login'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function startTokenPolling(deviceCode: string): Promise<void> {
    try {
      await apiService.post('/api/auth/callback', { device_code: deviceCode })
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to start token polling'
      throw err
    }
  }

  async function logout() {
    try {
      isLoading.value = true
      error.value = null
      
      await apiService.post('/api/auth/logout')
      user.value = null
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to logout'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  function clearError() {
    error.value = null
  }

  return {
    user,
    isLoading,
    error,
    isAuthenticated,
    checkAuthStatus,
    initiateDeviceFlow,
    startTokenPolling,
    logout,
    clearError
  }
})