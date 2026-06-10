import api from './client'
import type { AuthResponse } from '../types'
import { sha256 } from '../utils/crypto'

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
