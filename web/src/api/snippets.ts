import api from './client'
import type { Snippet, SnippetPreview, PaginatedResponse } from '../types'

export interface CreateSnippetPayload {
  content: string
  alias?: string
  description?: string
  ttl?: string
  encrypted?: boolean
  category?: string
  language?: string
  tags?: string[]
}

export interface UpdateSnippetPayload {
  content?: string
  alias?: string
  description?: string
  category?: string
  language?: string
  tags?: string[]
}

export function createSnippet(data: CreateSnippetPayload) {
  return api.post<Snippet>('/snippets', data)
}

export function listSnippets(limit = 20, offset = 0, category?: string, tag?: string) {
  const params: any = { limit, offset }
  if (category) params.category = category
  if (tag) params.tag = tag
  return api.get<PaginatedResponse<SnippetPreview>>('/snippets', { params })
}

export function getSnippet(id: string) {
  return api.get<Snippet>(`/snippets/${id}`)
}

export function getSnippetByAlias(alias: string) {
  return api.get<Snippet>(`/snippets/alias/${alias}`)
}

export function updateSnippet(id: string, data: UpdateSnippetPayload) {
  return api.put<Snippet>(`/snippets/${id}`, data)
}

export function deleteSnippet(id: string) {
  return api.delete(`/snippets/${id}`)
}

export function listVersions(id: string) {
  return api.get(`/snippets/${id}/versions`)
}
