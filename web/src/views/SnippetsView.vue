<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useSnippetsStore } from '../stores/snippets'
import { useI18n } from '../i18n'
import AppHeader from '../components/layout/AppHeader.vue'
import SnippetCard from '../components/snippets/SnippetCard.vue'
import SnippetEditor from '../components/snippets/SnippetEditor.vue'

const snippetsStore = useSnippetsStore()
const { t } = useI18n()
const showCreate = ref(false)
const category = ref('')
const searchTag = ref('')

onMounted(() => { snippetsStore.fetch(50) })

function applyFilter() {
  snippetsStore.fetch(50, 0, category.value || undefined, searchTag.value || undefined)
}

function onCreated() {
  showCreate.value = false
  snippetsStore.fetch(50)
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 to-blue-50">
    <AppHeader />

    <main class="max-w-5xl mx-auto px-4 py-8">
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-xl font-semibold text-gray-900">{{ t('snippets.title') }}</h2>
        <button @click="showCreate = !showCreate" class="btn-primary">
          {{ showCreate ? t('snippets.cancel') : t('snippets.new') }}
        </button>
      </div>

      <SnippetEditor v-if="showCreate" @created="onCreated" class="mb-6" />

      <div class="flex gap-3 mb-6">
        <input v-model="category" :placeholder="t('snippets.filter_category')" class="input flex-1" @keyup.enter="applyFilter" />
        <input v-model="searchTag" :placeholder="t('snippets.filter_tag')" class="input flex-1" @keyup.enter="applyFilter" />
        <button @click="applyFilter" class="btn-secondary">{{ t('snippets.filter') }}</button>
      </div>

      <div v-if="snippetsStore.loading" class="card p-12 text-center text-gray-400">
        {{ t('common.loading') }}
      </div>

      <div v-else class="space-y-3">
        <SnippetCard
          v-for="snippet in snippetsStore.items"
          :key="snippet.id"
          :snippet="snippet"
          @deleted="snippetsStore.remove(snippet.id)"
        />
      </div>
    </main>
  </div>
</template>
