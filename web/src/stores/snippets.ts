import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { SnippetPreview } from '../types'
import { listSnippets, deleteSnippet } from '../api/snippets'

export const useSnippetsStore = defineStore('snippets', () => {
  const items = ref<SnippetPreview[]>([])
  const total = ref(0)
  const loading = ref(false)

  async function fetch(limit = 20, offset = 0, category?: string, tag?: string) {
    loading.value = true
    try {
      const { data } = await listSnippets(limit, offset, category, tag)
      items.value = data.items
      total.value = data.total ?? data.items.length
    } finally {
      loading.value = false
    }
  }

  function addSnippet(snippet: SnippetPreview) {
    const existing = items.value.findIndex((s) => s.id === snippet.id)
    if (existing >= 0) {
      items.value[existing] = snippet
    } else {
      items.value.unshift(snippet)
    }
  }

  function removeSnippet(id: string) {
    items.value = items.value.filter((s) => s.id !== id)
  }

  function updateSnippet(snippet: SnippetPreview) {
    const idx = items.value.findIndex((s) => s.id === snippet.id)
    if (idx >= 0) {
      items.value[idx] = snippet
    }
  }

  async function remove(id: string) {
    await deleteSnippet(id)
    removeSnippet(id)
  }

  return { items, total, loading, fetch, addSnippet, removeSnippet, updateSnippet, remove }
})
