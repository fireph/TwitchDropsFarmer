<template>
  <div class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-sm">
    <div class="flex items-center">
      <div class="flex-shrink-0">
        <div :class="[
          'h-8 w-8 rounded-full flex items-center justify-center',
          colorClasses[color]
        ]">
          <svg class="h-5 w-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path v-if="icon === 'status'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
            <path v-else-if="icon === 'clock'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            <path v-else-if="icon === 'user'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path>
            <path v-else-if="icon === 'video'" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"></path>
          </svg>
        </div>
      </div>
      <div class="ml-4">
        <p class="text-sm font-medium text-gray-600 dark:text-gray-400">{{ title }}</p>
        <p class="text-lg font-semibold text-gray-900 dark:text-white">{{ value }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  title: string
  value: string | number
  color: 'green' | 'red' | 'blue' | 'purple' | 'orange'
  icon: 'status' | 'clock' | 'user' | 'video'
}

defineProps<Props>()

const colorClasses = {
  green: 'bg-green-500',
  red: 'bg-red-500',
  blue: 'bg-blue-500',
  purple: 'bg-purple-500',
  orange: 'bg-orange-500'
}
</script>