import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Device } from '../types'
import { listDevices, heartbeat } from '../api/devices'

export const useDevicesStore = defineStore('devices', () => {
  const devices = ref<Device[]>([])
  let heartbeatTimer: number | null = null

  async function fetch() {
    try {
      const { data } = await listDevices()
      devices.value = data.items
    } catch {
      // Ignore errors
    }
  }

  function startHeartbeat() {
    heartbeat('web-browser', 'web')
    heartbeatTimer = window.setInterval(() => {
      heartbeat('web-browser', 'web')
    }, 60000)
  }

  function stopHeartbeat() {
    if (heartbeatTimer) {
      clearInterval(heartbeatTimer)
      heartbeatTimer = null
    }
  }

  return { devices, fetch, startHeartbeat, stopHeartbeat }
})
