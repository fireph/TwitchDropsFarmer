import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { MinerStatus, Config } from '@/types'
import { apiService } from '@/services/api'

export const useMinerStore = defineStore('miner', () => {
  const status = ref<MinerStatus>({
    is_running: false,
    current_progress: 0,
    total_campaigns: 0,
    claimed_drops: 0,
    last_update: new Date().toISOString(),
    next_switch: new Date().toISOString(),
    error_message: '',
    active_drops: []
  })
  
  const config = ref<Config>({
    server_address: ':8080',
    twitch_client_id: '',
    priority_games: [],
    claim_drops: true,
    webhook_url: '',
    check_interval: 60,
    switch_threshold: 5,
    minimum_points: 50,
    maximum_streams: 3,
    theme: 'dark',
    language: 'en',
    show_tray: true,
    start_minimized: false
  })

  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const isRunning = computed(() => status.value.is_running)
  const currentStream = computed(() => status.value.current_stream)
  const currentCampaign = computed(() => status.value.current_campaign)
  const activeDrops = computed(() => status.value.active_drops || [])
  const claimedDrops = computed(() => status.value.claimed_drops)

  async function startMiner() {
    try {
      isLoading.value = true
      error.value = null
      
      await apiService.post('/api/miner/start')
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to start miner'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function stopMiner() {
    try {
      isLoading.value = true
      error.value = null
      
      await apiService.post('/api/miner/stop')
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to stop miner'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function fetchStatus() {
    try {
      const response = await apiService.get<MinerStatus>('/api/miner/status')
      status.value = response
    } catch (err) {
      console.error('Failed to fetch miner status:', err)
    }
  }


  async function fetchConfig() {
    try {
      const response = await apiService.get<Config>('/api/config')
      config.value = response
    } catch (err) {
      console.error('Failed to fetch config:', err)
    }
  }

  async function saveConfig() {
    try {
      isLoading.value = true
      error.value = null
      
      await apiService.post('/api/config', config.value)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to save config'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function addGame(gameName: string) {
    try {
      isLoading.value = true
      error.value = null
      
      await apiService.post('/api/config/game', {
        game_name: gameName
      })
      
      // Refresh config after adding game
      await fetchConfig()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to add game'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  function updateStatus(newStatus: MinerStatus) {
    status.value = newStatus
  }

  function updateConfig(newConfig: Config) {
    config.value = newConfig
  }

  function clearError() {
    error.value = null
  }

  return {
    status,
    config,
    isLoading,
    error,
    isRunning,
    currentStream,
    currentCampaign,
    activeDrops,
    claimedDrops,
    startMiner,
    stopMiner,
    fetchStatus,
    fetchConfig,
    saveConfig,
    addGame,
    updateStatus,
    updateConfig,
    clearError
  }
})