import { useSnippetsStore } from '../stores/snippets'
import { useAuthStore } from '../stores/auth'
import type { WSMessage } from '../types'

export function useWebSocket() {
  let ws: WebSocket | null = null
  let reconnectTimer: number | null = null
  let reconnectDelay = 1000

  function connect() {
    const auth = useAuthStore()
    if (!auth.token) return

    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
    ws = new WebSocket(`${protocol}//${location.host}/v1/ws?token=${auth.token}`)

    ws.onopen = () => {
      reconnectDelay = 1000
    }

    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        const store = useSnippetsStore()

        switch (msg.type) {
          case 'snippet_created':
            store.addSnippet(msg.payload)
            break
          case 'snippet_deleted':
            store.removeSnippet(msg.payload.id)
            break
          case 'snippet_updated':
            store.updateSnippet(msg.payload)
            break
        }
      } catch {
        // Ignore parse errors
      }
    }

    ws.onclose = () => {
      reconnectTimer = window.setTimeout(() => {
        reconnectDelay = Math.min(reconnectDelay * 2, 30000)
        connect()
      }, reconnectDelay)
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  function disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (ws) {
      ws.close()
      ws = null
    }
  }

  return { connect, disconnect }
}
