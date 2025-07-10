<template>
  <div class="h-full flex flex-col">
    <!-- Header -->
    <header class="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center h-16">
          <div class="flex items-center">
            <div class="flex-shrink-0">
              <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
                Twitch Drops Farmer
              </h1>
            </div>
          </div>
          
          <div class="flex items-center space-x-4">
            <!-- Theme Toggle -->
            <button 
              @click="themeStore.toggleTheme()"
              class="p-2 rounded-lg bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
            >
              <svg class="w-5 h-5 text-gray-600 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"></path>
              </svg>
            </button>
            
            <!-- User Profile -->
            <div class="flex items-center space-x-3">
              <div v-if="authStore.user" class="flex items-center space-x-3">
                <img 
                  :src="authStore.user.profile_image_url" 
                  :alt="authStore.user.display_name"
                  class="h-8 w-8 rounded-full"
                >
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ authStore.user.display_name }}
                </span>
                <button 
                  @click="logout"
                  class="text-sm text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300"
                >
                  Logout
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="flex-1 overflow-y-auto">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <!-- Status Cards -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
          <StatusCard 
            title="Status"
            :value="minerStore.isRunning ? 'Running' : 'Stopped'"
            :color="minerStore.isRunning ? 'green' : 'red'"
            icon="status"
          />
          
          <StatusCard 
            title="Active Drops"
            :value="minerStore.activeDrops.length"
            color="blue"
            icon="clock"
          />
          
          <StatusCard 
            title="Watching"
            :value="minerStore.currentStream?.user_name || 'None'"
            color="orange"
            icon="video"
          />
        </div>

        <!-- Controls -->
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6 mb-8">
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-lg font-medium text-gray-900 dark:text-white">Drop Miner Controls</h3>
              <p class="text-sm text-gray-600 dark:text-gray-400">Start or stop the automatic drop farming</p>
            </div>
            <div class="flex space-x-4">
              <button 
                @click="startMiner"
                :disabled="minerStore.isRunning || minerStore.isLoading"
                class="bg-green-600 hover:bg-green-700 disabled:bg-gray-400 text-white px-4 py-2 rounded-lg font-medium transition-colors"
              >
                Start Mining
              </button>
              <button 
                @click="stopMiner"
                :disabled="!minerStore.isRunning || minerStore.isLoading"
                class="bg-red-600 hover:bg-red-700 disabled:bg-gray-400 text-white px-4 py-2 rounded-lg font-medium transition-colors"
              >
                Stop Mining
              </button>
            </div>
          </div>
        </div>

        <!-- Current Campaign -->
        <div v-if="minerStore.currentCampaign" class="bg-white dark:bg-gray-800 rounded-lg shadow-sm mb-8">
          <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-medium text-gray-900 dark:text-white">Current Campaign</h3>
              <div class="flex items-center space-x-2">
                <div class="pulse-dot w-2 h-2 bg-green-500 rounded-full"></div>
                <span class="text-sm text-green-600 dark:text-green-400 font-medium">FARMING</span>
              </div>
            </div>
          </div>
          <div class="p-6">
            <CampaignCard :campaign="minerStore.currentCampaign" />
          </div>
        </div>

        <!-- Active Drops -->
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm mb-8">
          <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-medium text-gray-900 dark:text-white">Active Drops</h3>
              <span class="text-sm text-gray-500 dark:text-gray-400">
                {{ minerStore.activeDrops.length }} active
              </span>
            </div>
          </div>
          <div class="p-6">
            <div v-if="minerStore.activeDrops.length === 0" class="text-center text-gray-500 dark:text-gray-400">
              <p>No active drops found. Start the miner to begin farming!</p>
            </div>
            <div v-else class="space-y-4">
              <DropCard 
                v-for="drop in minerStore.activeDrops" 
                :key="drop.id"
                :drop="drop"
              />
            </div>
          </div>
        </div>

        <!-- Current Stream -->
        <div v-if="minerStore.currentStream" class="bg-white dark:bg-gray-800 rounded-lg shadow-sm mb-8">
          <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <h3 class="text-lg font-medium text-gray-900 dark:text-white">Current Stream</h3>
          </div>
          <div class="p-6">
            <StreamCard :stream="minerStore.currentStream" />
          </div>
        </div>

        <!-- Settings -->
        <SettingsCard />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'
import { useMinerStore } from '@/stores/miner'
import { webSocketService } from '@/services/websocket'
import StatusCard from '@/components/StatusCard.vue'
import CampaignCard from '@/components/CampaignCard.vue'
import DropCard from '@/components/DropCard.vue'
import SettingsCard from '@/components/SettingsCard.vue'
import StreamCard from '@/components/StreamCard.vue'

const router = useRouter()
const authStore = useAuthStore()
const themeStore = useThemeStore()
const minerStore = useMinerStore()

async function logout() {
  try {
    await authStore.logout()
    router.push('/login')
  } catch (error) {
    console.error('Logout failed:', error)
  }
}

async function startMiner() {
  try {
    await minerStore.startMiner()
  } catch (error) {
    console.error('Failed to start miner:', error)
  }
}

async function stopMiner() {
  try {
    await minerStore.stopMiner()
  } catch (error) {
    console.error('Failed to stop miner:', error)
  }
}

onMounted(async () => {
  // Initialize WebSocket connection - this will now provide real-time progress updates
  webSocketService.connect()
  
  // Fetch initial data
  await Promise.all([
    minerStore.fetchStatus(),
    minerStore.fetchConfig()
  ])
})

onUnmounted(() => {
  webSocketService.disconnect()
})
</script>