import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as apiLogin, register as apiRegister } from '../api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('cb_token') || '')
  const userId = ref(localStorage.getItem('cb_user_id') || '')
  const username = ref(localStorage.getItem('cb_username') || '')
  const isAdmin = ref(localStorage.getItem('cb_is_admin') === 'true')

  const isAuthenticated = computed(() => !!token.value)

  function saveToStorage(t: string, uid: string, user: string, admin: boolean) {
    token.value = t
    userId.value = uid
    username.value = user
    isAdmin.value = admin
    localStorage.setItem('cb_token', t)
    localStorage.setItem('cb_user_id', uid)
    localStorage.setItem('cb_username', user)
    localStorage.setItem('cb_is_admin', String(admin))
  }

  async function login(user: string, password: string) {
    const { data } = await apiLogin(user, password)
    saveToStorage(data.token, data.user_id, data.username, data.is_admin)
  }

  async function register(user: string, password: string) {
    const { data } = await apiRegister(user, password)
    saveToStorage(data.token, data.user_id, data.username, data.is_admin)
  }

  function logout() {
    token.value = ''
    userId.value = ''
    username.value = ''
    isAdmin.value = false
    localStorage.removeItem('cb_token')
    localStorage.removeItem('cb_user_id')
    localStorage.removeItem('cb_username')
    localStorage.removeItem('cb_is_admin')
  }

  return { token, userId, username, isAdmin, isAuthenticated, login, register, logout }
})
