import api from './client'
import type { AuthResponse } from '../types'

function sha256(str: string): Promise<string> {
  const buf = new TextEncoder().encode(str)
  return crypto.subtle.digest('SHA-256', buf).then((hash) => {
    return Array.from(new Uint8Array(hash)).map((b) => b.toString(16).padStart(2, '0')).join('')
  })
}

async function preHashPassword(username: string, password: string): Promise<string> {
  return sha256(`${username}:${password}`)
}

export async function login(username: string, password: string) {
  const hashed = await preHashPassword(username, password)
  return api.post<AuthResponse>('/auth/login', { username, password: hashed })
}

export async function register(username: string, password: string) {
  const hashed = await preHashPassword(username, password)
  return api.post<AuthResponse>('/auth/register', { username, password: hashed })
}
