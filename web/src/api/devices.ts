import api from './client'
import type { Device } from '../types'

export function heartbeat(name: string, type = 'web') {
  return api.post('/devices/heartbeat', { name, type })
}

export function listDevices() {
  return api.get<{ items: Device[] }>('/devices')
}
