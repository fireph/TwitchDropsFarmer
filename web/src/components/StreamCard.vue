<template>
  <div class="flex items-start space-x-4">
    <div class="flex-shrink-0">
      <img 
        :src="stream.preview_image_url.replace('{width}', '320').replace('{height}', '180')" 
        :alt="stream.title"
        class="w-32 h-18 rounded-lg object-cover"
      >
    </div>
    <div class="flex-1">
      <h4 class="text-lg font-medium text-gray-900 dark:text-white">{{ stream.user_name }}</h4>
      <p class="text-sm text-gray-600 dark:text-gray-400">{{ stream.game_name }}</p>
      <p class="text-sm text-gray-500 dark:text-gray-500 mt-1 line-clamp-2">{{ stream.title }}</p>
      <div class="flex items-center space-x-4 mt-2">
        <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200">
          🔴 LIVE
        </span>
        <span class="text-xs text-gray-500 dark:text-gray-400">
          {{ formatViewerCount(stream.viewer_count) }} viewers
        </span>
        <!-- <span class="text-xs text-gray-500 dark:text-gray-400">
          Started {{ formatStartTime(stream.started_at) }}
        </span> -->
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Stream } from '@/types'

interface Props {
  stream: Stream
}

defineProps<Props>()

function formatViewerCount(count: number): string {
  if (count < 1000) return count.toString()
  if (count < 1000000) return (count / 1000).toFixed(1) + 'K'
  return (count / 1000000).toFixed(1) + 'M'
}

function formatStartTime(startTime: string): string {
  const start = new Date(startTime)
  const now = new Date()
  const diffMs = now.getTime() - start.getTime()
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
  const diffMinutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60))
  
  if (diffHours > 0) {
    return `${diffHours}h ${diffMinutes}m ago`
  } else {
    return `${diffMinutes}m ago`
  }
}
</script>