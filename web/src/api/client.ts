import axios from 'axios'

const api = axios.create({
  baseURL: '/v1',
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('cb_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status
    const url = error.config?.url || ''

    // Only redirect to login on 401 if it's NOT an auth/password endpoint
    const noRedirectPaths = ['/auth/login', '/auth/register', '/account/password']
    if (status === 401 && !noRedirectPaths.some(p => url.includes(p))) {
      localStorage.removeItem('cb_token')
      localStorage.removeItem('cb_user_id')
      localStorage.removeItem('cb_username')
      localStorage.removeItem('cb_is_admin')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default api
