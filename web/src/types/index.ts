export interface Snippet {
  id: string
  user_id?: string
  alias?: string
  description?: string
  content: string
  encrypted: boolean
  category?: string
  language?: string
  tags?: string[]
  expires_at?: string
  created_at: string
  updated_at: string
}

export interface SnippetPreview {
  id: string
  alias?: string
  description?: string
  preview: string
  encrypted: boolean
  category?: string
  language?: string
  tags?: string[]
  expires_at?: string
  created_at: string
}

export interface AuthResponse {
  token: string
  user_id: string
  username: string
  is_admin: boolean
}

export interface Device {
  id: string
  name: string
  type: string
  last_seen: string
  created_at: string
}

export interface WSMessage {
  type: 'snippet_created' | 'snippet_deleted' | 'snippet_updated' | 'device_online' | 'device_offline' | 'connected'
  payload: any
}

export interface PaginatedResponse<T> {
  items: T[]
  total?: number
}

export interface UserView {
  id: string
  username: string
  is_admin: boolean
  created_at: string
}

export interface SystemInfo {
  user_count: number
  snippet_count: number
  device_count: number
  db_size_bytes: number
  uptime_seconds: number
  started_at: string
}

export interface Webhook {
  id: string
  name: string
  url: string
  events: string[]
  body_template?: string
  active: boolean
  created_at: string
}

export interface WebhookLog {
  id: number
  webhook_id: string
  event_type: string
  status_code: number
  error?: string
  created_at: string
}
