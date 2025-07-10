<template>
  <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm">
    <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
      <h3 class="text-lg font-medium text-gray-900 dark:text-white">Drop Farming Settings</h3>
    </div>
    <div class="p-6">
      <div class="space-y-6">
        <!-- Priority Games -->
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Games to Farm Drops For
          </label>
          <div class="flex space-x-2 mb-2">
            <input
              v-model="newGameName"
              type="text"
              placeholder="Enter game name (e.g., Fortnite, League of Legends)"
              class="flex-1 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white rounded-lg px-3 py-2 focus:ring-2 focus:ring-twitch-purple focus:border-transparent"
              @keydown.enter="addGame"
            >
            <button
              @click="addGame"
              :disabled="!newGameName.trim() || minerStore.isLoading"
              class="bg-twitch-purple hover:bg-twitch-purple-dark disabled:bg-gray-400 text-white px-4 py-2 rounded-lg font-medium transition-colors"
            >
              Add Game
            </button>
          </div>
          <div class="space-y-2">
            <draggable 
              v-model="localConfig.priority_games" 
              group="games"
              @end="onGameReorder"
              item-key="id"
              class="space-y-2"
            >
              <template #item="{ element: game }">
                <div class="flex items-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg border border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500 transition-colors">
                  <svg class="w-5 h-5 text-gray-400 cursor-grab mr-3 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M7 2a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM7 8a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM7 14a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM13 2a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM13 8a2 2 0 1 0 0 4 2 2 0 0 0 0-4zM13 14a2 2 0 1 0 0 4 2 2 0 0 0 0-4z"></path>
                  </svg>
                  <div class="flex-1">
                    <div class="mb-1">
                      <span class="text-sm font-medium text-gray-900 dark:text-white">{{ game.name }}</span>
                    </div>
                    <div class="text-xs text-gray-500 dark:text-gray-400 space-y-1">
                      <div><span class="font-medium">Slug:</span> {{ game.slug || 'N/A' }}</div>
                      <div><span class="font-medium">ID:</span> {{ game.id || 'N/A' }}</div>
                    </div>
                  </div>
                  <button
                    @click="removeGame(game)"
                    class="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300 text-sm font-medium ml-4 px-2 py-1 rounded hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors flex-shrink-0"
                  >
                    Remove
                  </button>
                </div>
              </template>
            </draggable>
          </div>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">
            Add games you want to farm drops for. The miner will automatically find streams for these games. Drag items to reorder priority.
          </p>
        </div>

        <!-- Auto-claim drops -->
        <div class="flex items-center justify-between">
          <div>
            <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
              Auto-claim completed drops
            </label>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              Automatically claim drops when they're completed
            </p>
          </div>
          <label class="relative inline-flex items-center cursor-pointer">
            <input 
              type="checkbox" 
              v-model="localConfig.claim_drops"
              @change="updateSetting('claim_drops', $event.target.checked)"
              class="sr-only peer"
            >
            <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-purple-300 dark:peer-focus:ring-purple-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-purple-600"></div>
          </label>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useMinerStore } from '@/stores/miner'
import type { GameConfig } from '@/types'
import draggable from 'vuedraggable'

const minerStore = useMinerStore()
const newGameName = ref('')
const localConfig = reactive({ ...minerStore.config })

// Watch for config changes from the store
watch(() => minerStore.config, (newConfig) => {
  Object.assign(localConfig, newConfig)
}, { deep: true })

async function addGame() {
  if (!newGameName.value.trim()) return
  
  try {
    await minerStore.addGame(newGameName.value.trim())
    newGameName.value = ''
  } catch (error) {
    console.error('Failed to add game:', error)
  }
}

async function removeGame(game: GameConfig) {
  try {
    // Remove from local config
    const index = localConfig.priority_games.findIndex(g => g.id === game.id)
    if (index !== -1) {
      localConfig.priority_games.splice(index, 1)
    }
    
    // Update the store config
    minerStore.config.priority_games = localConfig.priority_games
    await minerStore.saveConfig()
  } catch (error) {
    console.error('Failed to remove game:', error)
  }
}

async function updateSetting(key: string, value: any) {
  // Update the store config
  ;(minerStore.config as any)[key] = value
  await minerStore.saveConfig()
}

async function onGameReorder() {
  try {
    // Update the store config with the new order
    minerStore.config.priority_games = localConfig.priority_games
    await minerStore.saveConfig()
  } catch (error) {
    console.error('Failed to save game order:', error)
  }
}
</script>