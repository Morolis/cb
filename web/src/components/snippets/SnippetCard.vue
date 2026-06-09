<script setup lang="ts">
import { useClipboard } from '../../composables/useClipboard'
import { useI18n } from '../../i18n'
import type { SnippetPreview } from '../../types'
import { useRouter } from 'vue-router'

const props = defineProps<{ snippet: SnippetPreview }>()
const emit = defineEmits<{ deleted: [] }>()

const router = useRouter()
const { copied, copy } = useClipboard()
const { t } = useI18n()

function goToDetail() {
  router.push(`/snippets/${props.snippet.id}`)
}
</script>

<template>
  <div
    class="card-hover p-4 cursor-pointer"
    @click="goToDetail"
  >
    <div class="flex items-start justify-between">
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 mb-1.5 flex-wrap">
          <span v-if="snippet.alias" class="font-semibold text-gray-900 text-sm">
            {{ snippet.alias }}
          </span>
          <span v-else class="font-mono text-xs text-gray-400 bg-gray-50 px-2 py-0.5 rounded-lg">
            {{ snippet.id.slice(0, 8) }}
          </span>
          <span
            v-if="snippet.encrypted"
            class="px-2 py-0.5 bg-amber-50 text-amber-700 rounded-lg text-xs font-medium"
          >
            {{ t('snippets.encrypted') }}
          </span>
          <span
            v-if="snippet.category"
            class="px-2 py-0.5 bg-blue-50 text-blue-600 rounded-lg text-xs font-medium"
          >
            {{ snippet.category }}
          </span>
        </div>

        <p v-if="snippet.description" class="text-gray-500 text-xs mb-2">
          {{ snippet.description }}
        </p>

        <pre class="text-sm text-gray-600 bg-gray-50 rounded-xl px-3.5 py-2.5 overflow-x-auto max-h-20 overflow-hidden font-mono">{{ snippet.preview }}</pre>

        <div class="flex items-center gap-3 mt-2.5 text-xs text-gray-400">
          <span>{{ new Date(snippet.created_at).toLocaleString() }}</span>
          <span v-if="snippet.expires_at" class="text-amber-500">
            {{ new Date(snippet.expires_at).toLocaleString() }}
          </span>
          <div v-if="snippet.tags?.length" class="flex gap-1">
            <span v-for="tag in snippet.tags" :key="tag" class="px-1.5 py-0.5 bg-gray-100 rounded-lg text-gray-500">
              {{ tag }}
            </span>
          </div>
        </div>
      </div>

      <div class="flex gap-1 ml-3" @click.stop>
        <button
          @click="copy(snippet.preview)"
          class="p-2 rounded-xl text-gray-400 hover:text-blue-600 hover:bg-blue-50 transition-all"
          :title="copied ? t('detail.copied') : t('detail.copy')"
        >
          <svg v-if="!copied" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
            <path d="M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z" />
            <path d="M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z" />
          </svg>
          <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-green-500" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>
