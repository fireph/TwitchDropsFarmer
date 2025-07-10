<template>
  <div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
    <div class="flex items-center justify-between">
      <div class="flex items-center space-x-3">
        <div class="w-2 h-2 bg-blue-500 rounded-full"></div>
        <div>
          <h4 class="text-sm font-medium text-gray-900 dark:text-white">{{ drop.name }}</h4>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ drop.game_name }}</p>
        </div>
      </div>
      <div class="text-right">
        <p class="text-sm font-medium text-gray-900 dark:text-white">
          {{ Math.round(drop.progress * 100) }}%
        </p>
        <p class="text-xs text-gray-500 dark:text-gray-400">
          {{ drop.current_minutes }} / {{ drop.required_minutes }} min
        </p>
      </div>
    </div>
    
    <div class="mt-3">
      <div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400 mb-1">
        <span>Progress</span>
        <span v-if="!drop.is_claimed && drop.progress < 1">
          {{ formatTimeRemaining(drop.estimated_time) }}
        </span>
        <span v-else-if="drop.is_claimed" class="text-green-600 dark:text-green-400 font-medium">
          CLAIMED
        </span>
        <span v-else class="text-yellow-600 dark:text-yellow-400 font-medium">
          READY TO CLAIM
        </span>
      </div>
      <div class="w-full bg-gray-200 dark:bg-gray-600 rounded-full h-2">
        <div 
          class="bg-gradient-to-r from-twitch-purple to-twitch-purple-dark h-2 rounded-full transition-all duration-300"
          :style="{ width: `${Math.min(100, drop.progress * 100)}%` }"
        ></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ActiveDrop } from '@/types'

interface Props {
  drop: ActiveDrop
}

defineProps<Props>()

function formatTimeRemaining(estimatedTime: string): string {
  const now = new Date()
  const target = new Date(estimatedTime)
  const diffMs = target.getTime() - now.getTime()
  
  if (diffMs <= 0) return 'Now'
  
  const diffMinutes = Math.ceil(diffMs / (1000 * 60))
  
  if (diffMinutes < 60) {
    return `${diffMinutes}m remaining`
  } else {
    const hours = Math.floor(diffMinutes / 60)
    const minutes = diffMinutes % 60
    return `${hours}h ${minutes}m remaining`
  }
}
</script>