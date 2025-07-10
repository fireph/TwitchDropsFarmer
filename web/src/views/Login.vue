<template>
  <div class="min-h-full flex flex-col justify-center py-12 sm:px-6 lg:px-8">
    <div class="sm:mx-auto sm:w-full sm:max-w-md">
      <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900 dark:text-white">
        Twitch Drops Farmer
      </h2>
      <p class="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
        Login with your Twitch account to start farming drops
      </p>
    </div>

    <div class="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
      <div class="bg-white dark:bg-gray-800 py-8 px-4 shadow sm:rounded-lg sm:px-10">
        <div class="space-y-6">
          <!-- Login Status -->
          <div v-if="loginSuccess" class="rounded-md bg-green-50 dark:bg-green-900 p-4">
            <div class="flex">
              <div class="flex-shrink-0">
                <svg class="h-5 w-5 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
                </svg>
              </div>
              <div class="ml-3">
                <p class="text-sm font-medium text-green-800 dark:text-green-200">
                  Successfully logged in! Redirecting to dashboard...
                </p>
              </div>
            </div>
          </div>

          <!-- Error Message -->
          <div v-if="authStore.error" class="rounded-md bg-red-50 dark:bg-red-900 p-4">
            <div class="flex">
              <div class="flex-shrink-0">
                <svg class="h-5 w-5 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 15.5c-.77.833.192 2.5 1.732 2.5z"></path>
                </svg>
              </div>
              <div class="ml-3">
                <p class="text-sm font-medium text-red-800 dark:text-red-200">
                  {{ authStore.error }}
                </p>
              </div>
            </div>
          </div>

          <!-- Login Form -->
          <div v-if="!showDeviceCode && !authStore.isLoading">
            <div class="text-center">
              <div class="mb-6">
                <svg class="mx-auto h-16 w-16 text-twitch-purple" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M11.571 4.714h1.715v5.143H11.57zm4.715 0H18v5.143h-1.714zM6 0L1.714 4.286v15.428C1.714 21.068 2.647 22 3.571 22h16.286c.924 0 1.857-.932 1.857-2.286V3.429C21.714 2.575 20.781 1.714 19.857 1.714H7.714zm14.571 4.286v14.857c0 .924-.932 1.857-2.285 1.857H3.429C2.575 20 1.714 19.068 1.714 18.143V4.286c0-.854.861-1.715 1.715-1.715h14.857c.924 0 1.857.861 1.857 1.715z"/>
                </svg>
              </div>
              <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">
                Connect your Twitch account
              </h3>
              <p class="text-sm text-gray-600 dark:text-gray-400 mb-6">
                We need access to your Twitch account to monitor drop campaigns and claim rewards automatically.
              </p>
              <button 
                @click="initiateLogin"
                class="w-full flex justify-center py-3 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-twitch-purple hover:bg-twitch-purple-dark focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-twitch-purple transition-colors"
              >
                <svg class="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M11.571 4.714h1.715v5.143H11.57zm4.715 0H18v5.143h-1.714zM6 0L1.714 4.286v15.428C1.714 21.068 2.647 22 3.571 22h16.286c.924 0 1.857-.932 1.857-2.286V3.429C21.714 2.575 20.781 1.714 19.857 1.714H7.714zm14.571 4.286v14.857c0 .924-.932 1.857-2.285 1.857H3.429C2.575 20 1.714 19.068 1.714 18.143V4.286c0-.854.861-1.715 1.715-1.715h14.857c.924 0 1.857.861 1.857 1.715z"/>
                </svg>
                Login with Twitch
              </button>
            </div>
          </div>

          <!-- Device Code Display -->
          <div v-if="showDeviceCode && deviceCodeData">
            <div class="text-center">
              <div class="mb-6">
                <svg class="mx-auto h-16 w-16 text-twitch-purple" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M11.571 4.714h1.715v5.143H11.57zm4.715 0H18v5.143h-1.714zM6 0L1.714 4.286v15.428C1.714 21.068 2.647 22 3.571 22h16.286c.924 0 1.857-.932 1.857-2.286V3.429C21.714 2.575 20.781 1.714 19.857 1.714H7.714zm14.571 4.286v14.857c0 .924-.932 1.857-2.285 1.857H3.429C2.575 20 1.714 19.068 1.714 18.143V4.286c0-.854.861-1.715 1.715-1.715h14.857c.924 0 1.857.861 1.857 1.715z"/>
                </svg>
              </div>
              <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">
                Almost there!
              </h3>
              <div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-4 mb-4">
                <p class="text-sm text-gray-600 dark:text-gray-400 mb-2">
                  1. A new tab will open to Twitch activation page
                </p>
                <p class="text-sm text-gray-600 dark:text-gray-400 mb-2">
                  2. Enter this code when prompted:
                </p>
                <div class="bg-white dark:bg-gray-800 border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg p-3 mb-2">
                  <code class="text-2xl font-bold text-twitch-purple">{{ deviceCodeData.user_code }}</code>
                </div>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  Code expires in {{ Math.floor(deviceCodeData.expires_in / 60) }} minutes
                </p>
              </div>
              <button 
                @click="openActivationPage"
                class="w-full flex justify-center py-3 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-twitch-purple hover:bg-twitch-purple-dark focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-twitch-purple transition-colors"
              >
                Open Twitch Activation Page
              </button>
            </div>
          </div>

          <!-- Loading State -->
          <div v-if="authStore.isLoading" class="text-center">
            <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-twitch-purple"></div>
            <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
              Connecting to Twitch...
            </p>
          </div>
        </div>

        <div class="mt-8 border-t border-gray-200 dark:border-gray-700 pt-6">
          <div class="text-center">
            <p class="text-xs text-gray-500 dark:text-gray-400">
              By logging in, you agree to allow this application to access your Twitch account for drop farming purposes.
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import type { DeviceCodeResponse } from '@/types'

const router = useRouter()
const authStore = useAuthStore()

const showDeviceCode = ref(false)
const deviceCodeData = ref<DeviceCodeResponse | null>(null)
const loginSuccess = ref(false)
const pollInterval = ref<number | null>(null)

async function initiateLogin() {
  try {
    authStore.clearError()
    const deviceCode = await authStore.initiateDeviceFlow()
    
    deviceCodeData.value = deviceCode
    showDeviceCode.value = true
    
    // Auto-open the activation page
    window.open(deviceCode.verification_uri, '_blank')
    
    // Start polling for token
    await authStore.startTokenPolling(deviceCode.device_code)
    startPolling()
    
  } catch (error) {
    console.error('Login initiation failed:', error)
  }
}

function openActivationPage() {
  if (deviceCodeData.value) {
    window.open(deviceCodeData.value.verification_uri, '_blank')
  }
}

function startPolling() {
  if (pollInterval.value) {
    clearInterval(pollInterval.value)
  }
  
  pollInterval.value = setInterval(async () => {
    try {
      await authStore.checkAuthStatus()
      
      if (authStore.isAuthenticated) {
        loginSuccess.value = true
        clearInterval(pollInterval.value!)
        
        // Redirect after a short delay
        setTimeout(() => {
          router.push('/')
        }, 1500)
      }
    } catch (error) {
      console.error('Polling error:', error)
    }
  }, (deviceCodeData.value?.interval || 5) * 1000)
  
  // Stop polling after expiry time
  setTimeout(() => {
    if (pollInterval.value) {
      clearInterval(pollInterval.value)
      if (!authStore.isAuthenticated) {
        authStore.error = 'Authentication timed out. Please try again.'
        showDeviceCode.value = false
      }
    }
  }, (deviceCodeData.value?.expires_in || 600) * 1000)
}

onMounted(() => {
  // Check if already authenticated
  if (authStore.isAuthenticated) {
    router.push('/')
  }
})

onUnmounted(() => {
  if (pollInterval.value) {
    clearInterval(pollInterval.value)
  }
})
</script>