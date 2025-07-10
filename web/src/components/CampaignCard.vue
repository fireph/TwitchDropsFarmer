<template>
  <div class="space-y-4">
    <div class="flex items-start space-x-4">
      <img 
        :src="campaign.game.box_art_url" 
        :alt="campaign.game.name"
        class="w-16 h-20 rounded-lg object-cover"
      >
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
                {{ drop.self.current_minutes_watched }} / {{ drop.required_minutes_watched }} minutes
              </p>
            </div>
          </div>
          <div class="flex items-center space-x-2">
            <div class="w-16 bg-gray-200 dark:bg-gray-600 rounded-full h-2">
              <div 
                class="bg-twitch-purple h-2 rounded-full transition-all duration-300"
                :style="{ width: `${Math.min(100, (drop.self.current_minutes_watched / drop.required_minutes_watched) * 100)}%` }"
              ></div>
            </div>
            <span v-if="drop.self.is_claimed" class="text-xs text-green-600 dark:text-green-400 font-medium">
              CLAIMED
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Campaign } from '@/types'

interface Props {
  campaign: Campaign
}

defineProps<Props>()

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