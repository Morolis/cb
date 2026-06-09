<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useSnippetsStore } from '../stores/snippets'
import { useDevicesStore } from '../stores/devices'
import { useWebSocket } from '../composables/useWebSocket'
import { useI18n } from '../i18n'
import AppHeader from '../components/layout/AppHeader.vue'
import SnippetCard from '../components/snippets/SnippetCard.vue'

const snippetsStore = useSnippetsStore()
const devicesStore = useDevicesStore()
const { connect, disconnect } = useWebSocket()
const { t } = useI18n()

onMounted(() => {
  snippetsStore.fetch(10)
  devicesStore.fetch()
  devicesStore.startHeartbeat()
  connect()
})

onUnmounted(() => {
  disconnect()
  devicesStore.stopHeartbeat()
})
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 to-blue-50">
    <AppHeader />

    <main class="max-w-5xl mx-auto px-4 py-8">
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-xl font-semibold text-gray-900">{{ t('home.title') }}</h2>
        <router-link
          to="/snippets"
          class="text-sm text-blue-600 hover:text-blue-700 transition-colors"
        >
          {{ t('home.view_all') }}
        </router-link>
      </div>

      <div v-if="snippetsStore.loading" class="text-center py-12 text-gray-400">
        {{ t('common.loading') }}
      </div>

      <div v-else-if="snippetsStore.items.length === 0" class="card p-12 text-center">
        <div class="text-gray-300 text-5xl mb-4">📋</div>
        <p class="text-lg font-medium text-gray-500 mb-2">{{ t('home.empty') }}</p>
        <p class="text-sm text-gray-400">{{ t('home.empty_hint') }}</p>
      </div>

      <div v-else class="space-y-3">
        <SnippetCard
          v-for="snippet in snippetsStore.items"
          :key="snippet.id"
          :snippet="snippet"
        />
      </div>
    </main>
  </div>
</template>
