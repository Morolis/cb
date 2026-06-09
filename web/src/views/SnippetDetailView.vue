<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getSnippet, updateSnippet, listVersions } from '../api/snippets'
import { useClipboard } from '../composables/useClipboard'
import { useI18n } from '../i18n'
import type { Snippet } from '../types'
import AppHeader from '../components/layout/AppHeader.vue'
import hljs from 'highlight.js/lib/core'
import 'highlight.js/styles/github-dark.css'
import go from 'highlight.js/lib/languages/go'
import python from 'highlight.js/lib/languages/python'
import bash from 'highlight.js/lib/languages/bash'
import javascript from 'highlight.js/lib/languages/javascript'
import typescript from 'highlight.js/lib/languages/typescript'
import sql from 'highlight.js/lib/languages/sql'
import json from 'highlight.js/lib/languages/json'
import yaml from 'highlight.js/lib/languages/yaml'
import xml from 'highlight.js/lib/languages/xml'
import css from 'highlight.js/lib/languages/css'
import java from 'highlight.js/lib/languages/java'
import rust from 'highlight.js/lib/languages/rust'
import cpp from 'highlight.js/lib/languages/cpp'
import dockerfile from 'highlight.js/lib/languages/dockerfile'

hljs.registerLanguage('go', go)
hljs.registerLanguage('python', python)
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('shell', bash)
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('typescript', typescript)
hljs.registerLanguage('sql', sql)
hljs.registerLanguage('json', json)
hljs.registerLanguage('yaml', yaml)
hljs.registerLanguage('xml', xml)
hljs.registerLanguage('html', xml)
hljs.registerLanguage('css', css)
hljs.registerLanguage('java', java)
hljs.registerLanguage('rust', rust)
hljs.registerLanguage('cpp', cpp)
hljs.registerLanguage('dockerfile', dockerfile)

const route = useRoute()
const { copied, copy } = useClipboard()
const { t } = useI18n()

const snippet = ref<Snippet | null>(null)
const versions = ref<any[]>([])
const editing = ref(false)
const editContent = ref('')
const loading = ref(true)

onMounted(async () => {
  const id = route.params.id as string
  try {
    const { data } = await getSnippet(id)
    snippet.value = data
    editContent.value = data.content
    const vRes = await listVersions(id)
    versions.value = vRes.data.items || []
  } finally {
    loading.value = false
  }
})

function highlight(code: string, lang?: string) {
  if (lang && hljs.getLanguage(lang)) {
    return hljs.highlight(code, { language: lang }).value
  }
  return hljs.highlightAuto(code).value
}

async function saveEdit() {
  if (!snippet.value) return
  await updateSnippet(snippet.value.id, { content: editContent.value })
  snippet.value.content = editContent.value
  editing.value = false
  const vRes = await listVersions(snippet.value.id)
  versions.value = vRes.data.items || []
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 to-blue-50">
    <AppHeader />

    <main class="max-w-5xl mx-auto px-4 py-8">
      <div v-if="loading" class="card p-12 text-center text-gray-400">{{ t('detail.loading') }}</div>

      <div v-else-if="!snippet" class="card p-12 text-center text-red-500">{{ t('detail.not_found') }}</div>

      <div v-else>
        <div class="flex items-center justify-between mb-4">
          <div>
            <h2 class="text-xl font-semibold text-gray-900">
              {{ snippet.alias || snippet.id }}
            </h2>
            <p v-if="snippet.description" class="text-gray-500 text-sm mt-1">
              {{ snippet.description }}
            </p>
          </div>
          <div class="flex gap-2">
            <button @click="copy(snippet.content)" class="btn-secondary">
              {{ copied ? t('detail.copied') : t('detail.copy') }}
            </button>
            <button @click="editing = !editing" class="btn-ghost text-blue-600">
              {{ editing ? t('detail.cancel') : t('detail.edit') }}
            </button>
          </div>
        </div>

        <div v-if="snippet.tags?.length" class="flex gap-2 mb-4">
          <span
            v-for="tag in snippet.tags"
            :key="tag"
            class="px-2 py-0.5 bg-gray-100 rounded text-xs text-gray-600"
          >
            {{ tag }}
          </span>
        </div>

        <div v-if="editing" class="mb-6">
          <textarea
            v-model="editContent"
            rows="12"
            class="w-full px-4 py-3 border rounded-md font-mono text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button @click="saveEdit" class="btn-primary mt-2">
            {{ t('detail.save') }}
          </button>
        </div>

        <div v-else class="card overflow-hidden mb-6">
          <pre class="p-4 overflow-x-auto text-sm"><code v-html="highlight(snippet.content, snippet.language)" /></pre>
        </div>

        <div v-if="snippet.category" class="text-sm text-gray-500 mb-2">
          {{ t('detail.category') }}: <span class="font-medium">{{ snippet.category }}</span>
        </div>

        <div v-if="versions.length > 0" class="mt-8">
          <h3 class="text-lg font-semibold mb-3">{{ t('detail.versions') }}</h3>
          <div class="space-y-2">
            <div
              v-for="v in versions"
              :key="v.id"
              class="card p-4 text-sm"
            >
              <div class="text-gray-500 mb-2 text-xs">
                {{ new Date(v.created_at).toLocaleString() }}
              </div>
              <pre class="text-xs text-gray-700 overflow-x-auto max-h-32 overflow-y-auto bg-gray-50 rounded-xl p-3">{{ v.content.slice(0, 200) }}{{ v.content.length > 200 ? '...' : '' }}</pre>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>
