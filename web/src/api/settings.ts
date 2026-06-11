import api from './client'
import type { UserView, SystemInfo } from '../types'
import { useAuthStore } from '../stores/auth'
import { sha256 } from '../utils/crypto'

// System settings (admin)
export function getSettings() {
  return api.get<{ settings: Record<string, any> }>('/admin/settings')
}

export function updateSettings(settings: Record<string, string>) {
  return api.put('/admin/settings', settings)
}

// User management (admin)
export function getUsers() {
  return api.get<{ items: UserView[] }>('/admin/users')
}

export function deleteUser(id: string) {
  return api.delete(`/admin/users/${id}`)
}

export function toggleAdmin(id: string) {
  return api.put(`/admin/users/${id}/admin`)
}

export async function resetUserPassword(id: string, newPassword: string, targetUsername: string) {
  const hashed = await sha256(`${targetUsername}:${newPassword}`)
  return api.put(`/admin/users/${id}/password`, { new_password: hashed })
}

// Account (any user) - pre-hash passwords like login/register
export async function changePassword(oldPassword: string, newPassword: string) {
  const auth = useAuthStore()
  const oldHashed = await sha256(`${auth.username}:${oldPassword}`)
  const newHashed = await sha256(`${auth.username}:${newPassword}`)
  return api.put('/account/password', { old_password: oldHashed, new_password: newHashed })
}

// System info (admin only)
export function getSystemInfo() {
  return api.get<SystemInfo>('/admin/system')
}
