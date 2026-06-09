<script setup lang="ts">
import { ref } from 'vue'
import { createSnippet } from '../../api/snippets'
import { useI18n } from '../../i18n'

const emit = defineEmits<{ created: [] }>()
const { t } = useI18n()

const content = ref('')
const alias = ref('')
const description = ref('')
const category = ref('')
const language = ref('')
const tags = ref('')
const ttl = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  if (!content.value.trim()) { error.value = t('editor.content_required'); return }
  error.value = ''
  loading.value = true
  try {
    await createSnippet({
      content: content.value,
      alias: alias.value || undefined,
      description: description.value || undefined,
      category: category.value || undefined,
      language: language.value || undefined,
      tags: tags.value ? tags.value.split(',').map((t) => t.trim()) : undefined,
      ttl: ttl.value || undefined,
    })
    content.value = ''; alias.value = ''; description.value = ''
    category.value = ''; language.value = ''; tags.value = ''; ttl.value = ''
    emit('created')
  } catch (e: any) {
    error.value = e.response?.data?.error || t('editor.create_failed')
  } finally { loading.value = false }
}
</script>

<template>
  <div class="card p-5">
    <h3 class="font-medium text-gray-900 mb-4">{{ t('editor.title') }}</h3>
    <form @submit.prevent="submit" class="space-y-3">
      <textarea v-model="content" rows="4" :placeholder="t('editor.content')" class="input font-mono" />

      <div class="grid grid-cols-2 gap-3">
        <input v-model="alias" :placeholder="t('editor.alias')" class="input" />
        <input v-model="description" :placeholder="t('editor.desc')" class="input" />
        <input v-model="category" :placeholder="t('editor.category')" class="input" />
        <input v-model="language" :placeholder="t('editor.language')" class="input" />
        <input v-model="tags" :placeholder="t('editor.tags')" class="input" />
        <input v-model="ttl" :placeholder="t('editor.ttl')" class="input" />
      </div>

      <div v-if="error" class="bg-red-50 border border-red-100 text-red-700 px-4 py-3 rounded-xl text-sm">{{ error }}</div>

      <button type="submit" :disabled="loading" class="btn-primary">
        {{ loading ? t('editor.creating') : t('editor.create') }}
      </button>
    </form>
  </div>
</template>
