<template>
  <div class="space-y-4">
    <div class="flex items-start space-x-4">
      <div class="w-16 h-20 bg-gray-200 dark:bg-gray-700 rounded-lg flex items-center justify-center">
        <img 
          v-if="gameImageUrl"
          :src="gameImageUrl" 
          :alt="campaign.game.name"
          class="w-16 h-20 rounded-lg object-cover"
          @error="onImageError"
        >
        <div v-else class="text-gray-400 text-xs text-center p-2">
          {{ campaign.game.name }}
        </div>
      </div>
      <div class="flex-1">
        <h4 class="text-lg font-medium text-gray-900 dark:text-white">{{ campaign.name }}</h4>
        <p class="text-sm text-gray-600 dark:text-gray-400">{{ campaign.game.name }}</p>
        <p class="text-sm text-gray-500 dark:text-gray-500 mt-1">{{ campaign.description }}</p>
        <div class="flex items-center space-x-4 mt-2">
          <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
            {{ campaign.status }}
          </span>
          <span class="text-xs text-gray-500 dark:text-gray-400">
            Ends {{ formatDate(campaign.ends_at) }}
          </span>
        </div>
      </div>
    </div>
    
    <!-- Campaign Drops -->
    <div v-if="campaign.time_based_drops.length > 0" class="border-t border-gray-200 dark:border-gray-700 pt-4">
      <h5 class="text-sm font-medium text-gray-900 dark:text-white mb-3">Campaign Drops</h5>
      <div class="space-y-2">
        <div 
          v-for="drop in campaign.time_based_drops" 
          :key="drop.id"
          class="flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-700 rounded-lg"
        >
          <div class="flex items-center space-x-3">
            <div v-if="drop.benefit_edges?.length > 0" class="flex-shrink-0">
              <img 
                :src="drop.benefit_edges[0].benefit.image_url" 
                :alt="drop.name"
                class="w-8 h-8 rounded object-cover"
              >
            </div>
            <div>
              <p class="text-sm font-medium text-gray-900 dark:text-white">{{ drop.name }}</p>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                {{ getCurrentMinutes(drop) }} / {{ drop.required_minutes_watched }} minutes
              </p>
            </div>
          </div>
          <div class="flex items-center space-x-2">
            <div class="w-16 bg-gray-200 dark:bg-gray-600 rounded-full h-2">
              <div 
                class="bg-twitch-purple h-2 rounded-full transition-all duration-300"
                :style="{ width: `${getProgressPercentage(drop)}%` }"
              ></div>
            </div>
            <span v-if="drop.self?.is_claimed" class="text-xs text-green-600 dark:text-green-400 font-medium">
              CLAIMED
            </span>
            <span v-else-if="getProgressPercentage(drop) >= 100" class="text-xs text-yellow-600 dark:text-yellow-400 font-medium">
              READY
            </span>
            <span v-else class="text-xs text-blue-600 dark:text-blue-400 font-medium">
              {{ Math.round(getProgressPercentage(drop)) }}%
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Campaign, TimeBased } from '@/types'
import { useMinerStore } from '@/stores/miner'

interface Props {
  campaign: Campaign
}

const props = defineProps<Props>()
const minerStore = useMinerStore()

const imageError = ref(false)

const gameImageUrl = computed(() => {
  if (imageError.value || !props.campaign.game.box_art_url) {
    return null
  }
  
  let url = props.campaign.game.box_art_url
  
  // Twitch box art URLs have placeholders that need to be replaced
  if (url.includes('{width}') && url.includes('{height}')) {
    url = url.replace('{width}', '128').replace('{height}', '170')
  }
  
  // Handle empty or invalid URLs
  if (!url || url === 'null' || url === 'undefined') {
    return null
  }
  
  return url
})

function onImageError() {
  imageError.value = true
  console.warn(`Failed to load game image for ${props.campaign.game.name}:`, props.campaign.game.box_art_url)
}

function getCurrentMinutes(drop: TimeBased): number {
  // Check if this drop is from the current active campaign and get real-time progress
  if (minerStore.isRunning && minerStore.currentCampaign?.id === props.campaign.id) {
    // Look for this drop in active drops for real-time progress
    const activeDrop = minerStore.activeDrops.find(ad => ad.id === drop.id)
    if (activeDrop) {
      return activeDrop.current_minutes
    }
  }
  
  // Fall back to the drop's self-reported progress
  return drop.self?.current_minutes_watched || 0
}

function getProgressPercentage(drop: TimeBased): number {
  const currentMinutes = getCurrentMinutes(drop)
  const requiredMinutes = drop.required_minutes_watched
  
  if (requiredMinutes <= 0) return 0
  
  const percentage = (currentMinutes / requiredMinutes) * 100
  return Math.min(100, Math.max(0, percentage))
}

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>