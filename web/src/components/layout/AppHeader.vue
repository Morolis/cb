<script setup lang="ts">
import { useDevicesStore } from '../../stores/devices'
import { useI18n } from '../../i18n'

const devicesStore = useDevicesStore()
const { t, lang, setLang } = useI18n()
</script>

<template>
  <nav class="bg-white/80 backdrop-blur-md shadow-sm border-b border-gray-100 sticky top-0 z-50">
    <div class="max-w-5xl mx-auto px-4 py-3 flex items-center justify-between">
      <div class="flex items-center gap-6">
        <router-link to="/" class="text-lg font-bold text-gray-900 tracking-tight">cb</router-link>
        <div class="flex gap-1">
          <router-link to="/" class="px-3 py-1.5 rounded-lg text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-100 transition-colors">
            {{ t('nav.home') }}
          </router-link>
          <router-link to="/snippets" class="px-3 py-1.5 rounded-lg text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-100 transition-colors">
            {{ t('nav.snippets') }}
          </router-link>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <div
          v-if="devicesStore.devices.length > 0"
          class="flex items-center gap-1.5 text-xs text-green-600 bg-green-50 px-2.5 py-1 rounded-full"
        >
          <span class="w-1.5 h-1.5 bg-green-500 rounded-full animate-pulse" />
          {{ devicesStore.devices.length }} {{ t('nav.online') }}
        </div>

        <!-- Language switcher -->
        <div class="flex items-center bg-gray-100 rounded-lg p-0.5">
          <button
            @click="setLang('zh')"
            :class="[
              'px-2.5 py-1 text-xs font-medium rounded-md transition-all duration-150',
              lang === 'zh' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-500 hover:text-gray-700'
            ]"
          >
            中文
          </button>
          <button
            @click="setLang('en')"
            :class="[
              'px-2.5 py-1 text-xs font-medium rounded-md transition-all duration-150',
              lang === 'en' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-500 hover:text-gray-700'
            ]"
          >
            EN
          </button>
        </div>

        <router-link to="/settings" class="p-2 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z" clip-rule="evenodd" />
          </svg>
        </router-link>
      </div>
    </div>
  </nav>
</template>
