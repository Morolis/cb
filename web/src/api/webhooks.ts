import api from './client'
import type { Webhook, WebhookLog } from '../types'

export function getWebhooks() {
  return api.get<{ items: Webhook[] }>('/webhooks')
}

export function createWebhook(data: { name: string; url: string; events: string[]; body_template?: string }) {
  return api.post<Webhook>('/webhooks', data)
}

export function deleteWebhook(id: string) {
  return api.delete(`/webhooks/${id}`)
}

export function toggleWebhook(id: string) {
  return api.put<{ active: boolean }>(`/webhooks/${id}/toggle`)
}

export function getWebhookLogs(id: string) {
  return api.get<{ items: WebhookLog[] }>(`/webhooks/${id}/logs`)
}
