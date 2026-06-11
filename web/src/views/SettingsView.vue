<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useI18n } from '../i18n'
import { getSettings, updateSettings, getUsers, deleteUser, toggleAdmin, resetUserPassword, changePassword, getSystemInfo } from '../api/settings'
import api from '../api/client'
import { getWebhooks, createWebhook, deleteWebhook, toggleWebhook, getWebhookLogs } from '../api/webhooks'
import type { UserView, SystemInfo, Webhook, WebhookLog } from '../types'
import AppHeader from '../components/layout/AppHeader.vue'

const auth = useAuthStore()
const router = useRouter()
const { t, lang } = useI18n()

const activeTab = ref('account')
const tabs = computed(() => {
  const list = [
    { key: 'account', label: t('settings.tab.account') },
    { key: 'webhooks', label: t('settings.tab.webhooks') },
  ]
  if (auth.isAdmin) {
    list.push({ key: 'system', label: t('settings.tab.system') })
    list.push({ key: 'server', label: t('settings.tab.server') })
    list.push({ key: 'users', label: t('settings.tab.users') })
  }
  return list
})

const sysInfo = ref<SystemInfo | null>(null)
const settings = ref<Record<string, any>>({})
const settingsLoading = ref(false)
const settingsMsg = ref('')
const corsOrigin = ref('')
const tlsLoading = ref(false)
const tlsMsg = ref('')
const tlsDisableConfirm = ref(false)
const tlsCountdown = ref(3)
let tlsDisableTimer: number | null = null
const users = ref<UserView[]>([])
const newUsername = ref('')
const newUserPass = ref('')
const addUserError = ref('')
const addUserLoading = ref(false)
const deleteConfirm = ref('')
const resetTarget = ref('')
const resetNewPass = ref('')
const toggleConfirm = ref('')
const oldPass = ref('')
const newPass = ref('')
const passMsg = ref('')
const passError = ref('')
const passLoading = ref(false)
let passTimer: number | null = null

// Webhooks
const webhooks = ref<Webhook[]>([])
const whName = ref('')
const whUrl = ref('')
const whEvents = ref(['created', 'updated', 'deleted'])
const whTemplate = ref('')
const whError = ref('')
const whLoading = ref(false)
const whLogs = ref<WebhookLog[]>([])
const whLogsTarget = ref('')
const whDeleteConfirm = ref('')
const templateVars = computed(() => [
  { name: '{{.Event}}', desc: t('webhooks.var.event') + ' (created/updated/deleted)' },
  { name: '{{.DateTime}}', desc: t('webhooks.var.datetime') + ' (ISO8601)' },
  { name: '{{.Snippet.ID}}', desc: t('webhooks.var.id') + ' (UUID)' },
  { name: '{{.Snippet.UserID}}', desc: t('webhooks.var.userid') },
  { name: '{{.Snippet.Alias}}', desc: t('webhooks.var.alias') },
  { name: '{{.Snippet.Description}}', desc: t('webhooks.var.desc') },
  { name: '{{.Snippet.Content}}', desc: t('webhooks.var.content') },
  { name: '{{.Snippet.Encrypted}}', desc: t('webhooks.var.encrypted') + ' (true/false)' },
  { name: '{{.Snippet.Category}}', desc: t('webhooks.var.category') },
  { name: '{{.Snippet.Language}}', desc: t('webhooks.var.lang') },
  { name: '{{.Snippet.ExpiresAt}}', desc: t('webhooks.var.expires') },
  { name: '{{.Snippet.CreatedAt}}', desc: t('webhooks.var.created') },
  { name: '{{.Snippet.UpdatedAt}}', desc: t('webhooks.var.updated') },
])
const dingtalkExample = '{\n' +
  '  "msgtype": "text",\n' +
  '  "text": {\n' +
  '    "content": {{json (printf "[cb %s] %s: %s" .Event .Snippet.Alias .Snippet.Content)}}\n' +
  '  },\n' +
  '  "at": {\n' +
  '    "isAtAll": false\n' +
  '  }\n' +
  '}'

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function formatUptime(seconds: number): string {
  const d = Math.floor(seconds / 86400)
  const h = Math.floor((seconds % 86400) / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (d > 0) return `${d}d ${h}h ${m}m`
  if (h > 0) return `${h}h ${m}m`
  return `${m}m`
}

async function loadSystemInfo() {
  try { const { data } = await getSystemInfo(); sysInfo.value = data } catch {}
}

async function loadSettings() {
  if (!auth.isAdmin) return
  settingsLoading.value = true
  try {
    const { data } = await getSettings()
    settings.value = data.settings || {}
    corsOrigin.value = data.settings?.cors_origin || ''
  } finally { settingsLoading.value = false }
}

async function saveCORS() {
  settingsMsg.value = ''
  try {
    await updateSettings({ cors_origin: corsOrigin.value })
    settingsMsg.value = t('server.saved')
    setTimeout(() => { settingsMsg.value = '' }, 3000)
  } catch (e: any) { settingsMsg.value = e.response?.data?.error || 'Failed' }
}

async function handleEnableTLS() {
  tlsMsg.value = ''
  tlsLoading.value = true
  try {
    const { data } = await api.post('/admin/settings/tls/enable')
    tlsMsg.value = data.message || 'HTTPS enabled'
    settings.value.tls_enabled = true
    settings.value.http_redirect = true
    setTimeout(() => {
      const host = window.location.hostname
      window.location.href = `https://${host}`
    }, 2000)
  } catch (e: any) { tlsMsg.value = e.response?.data?.error || 'Failed' }
  finally { tlsLoading.value = false }
}

function handleDisableTLS() {
  // Phase 1: counting down
  if (tlsCountdown.value > 0 && !tlsDisableConfirm.value) {
    tlsDisableConfirm.value = true
    tlsCountdown.value = 3
    if (tlsDisableTimer) clearInterval(tlsDisableTimer)
    tlsDisableTimer = window.setInterval(() => {
      tlsCountdown.value--
      if (tlsCountdown.value <= 0) {
        clearInterval(tlsDisableTimer!)
        tlsDisableTimer = null
        // Start 10s window for confirmation
        tlsDisableTimer = window.setTimeout(() => {
          tlsDisableConfirm.value = false
          tlsCountdown.value = 3
          tlsDisableTimer = null
        }, 10000)
      }
    }, 1000)
    return
  }
  // Phase 2: confirmed
  if (tlsDisableTimer) { clearInterval(tlsDisableTimer); clearTimeout(tlsDisableTimer); tlsDisableTimer = null }
  tlsDisableConfirm.value = false
  tlsCountdown.value = 3
  tlsLoading.value = true
  api.post('/admin/settings/tls/disable').then(() => {
    settings.value.tls_enabled = false
    setTimeout(() => {
      window.location.href = `http://${window.location.hostname}`
    }, 1000)
  }).catch((e: any) => { alert(e.response?.data?.error || 'Failed') })
    .finally(() => { tlsLoading.value = false })
}

async function handleAddUser() {
  addUserError.value = ''
  if (!newUsername.value || !newUserPass.value) { addUserError.value = 'Username and password required'; return }
  if (newUserPass.value.length < 6) { addUserError.value = 'Password must be at least 6 characters'; return }
  addUserLoading.value = true
  try {
    await api.post('/admin/users', { username: newUsername.value, password: newUserPass.value })
    newUsername.value = ''; newUserPass.value = ''
    await loadUsers()
  } catch (e: any) { addUserError.value = e.response?.data?.error || 'Failed' }
  finally { addUserLoading.value = false }
}

async function loadUsers() {
  if (!auth.isAdmin) return
  try { const { data } = await getUsers(); users.value = data.items || [] } catch {}
}

async function handleDeleteUser(id: string) {
  if (deleteConfirm.value !== id) { deleteConfirm.value = id; setTimeout(() => { deleteConfirm.value = '' }, 3000); return }
  try { await deleteUser(id); users.value = users.value.filter((u) => u.id !== id); deleteConfirm.value = '' }
  catch (e: any) { alert(e.response?.data?.error || 'Failed') }
}

async function handleToggleAdmin(id: string) {
  if (toggleConfirm.value !== id) { toggleConfirm.value = id; setTimeout(() => { toggleConfirm.value = '' }, 3000); return }
  try {
    await toggleAdmin(id)
    const user = users.value.find((u) => u.id === id)
    if (user) user.is_admin = !user.is_admin
    toggleConfirm.value = ''
  } catch (e: any) { alert(e.response?.data?.error || 'Failed') }
}

async function handleResetPassword(id: string) {
  if (resetTarget.value !== id) { resetTarget.value = id; resetNewPass.value = ''; return }
  if (resetNewPass.value.length < 6) { alert(t('account.pass_min')); return }
  const targetUser = users.value.find(u => u.id === id)
  try { await resetUserPassword(id, resetNewPass.value, targetUser?.username || ''); alert(t('users.pass_reset_ok')); resetTarget.value = ''; resetNewPass.value = '' }
  catch (e: any) { alert(e.response?.data?.error || 'Failed') }
}

async function handleChangePassword() {
  passMsg.value = ''
  passError.value = ''
  if (passTimer) { clearTimeout(passTimer); passTimer = null }
  passLoading.value = true

  let success = false
  try {
    const { data } = await changePassword(oldPass.value, newPass.value)
    passMsg.value = data.message || t('account.pass_success')
    oldPass.value = ''
    newPass.value = ''
    success = true
  } catch (e: any) {
    passError.value = e.response?.data?.error || t('account.pass_error')
  } finally {
    passLoading.value = false
  }

  // Only redirect on success
  if (success) {
    passTimer = window.setTimeout(() => {
      auth.logout()
      router.push('/login')
    }, 2000)
  }
}

onMounted(() => {
  if (auth.isAdmin) { loadSystemInfo(); loadSettings(); loadUsers() }
  loadWebhooks()
})

// Webhooks
async function loadWebhooks() {
  try { const { data } = await getWebhooks(); webhooks.value = data.items || [] } catch {}
}

async function handleCreateWebhook() {
  whError.value = ''
  if (!whName.value || !whUrl.value) { whError.value = 'Name and URL are required'; return }
  if (whEvents.value.length === 0) { whError.value = 'Select at least one event'; return }
  whLoading.value = true
  try {
    await createWebhook({ name: whName.value, url: whUrl.value, events: whEvents.value, body_template: whTemplate.value || undefined })
    whName.value = ''; whUrl.value = ''; whTemplate.value = ''; whEvents.value = ['created', 'updated', 'deleted']
    await loadWebhooks()
  } catch (e: any) { whError.value = e.response?.data?.error || 'Failed' }
  finally { whLoading.value = false }
}

async function handleToggleWebhook(id: string) {
  try { await toggleWebhook(id); await loadWebhooks() } catch (e: any) { alert(e.response?.data?.error || 'Failed') }
}

async function handleDeleteWebhook(id: string) {
  if (whDeleteConfirm.value !== id) { whDeleteConfirm.value = id; setTimeout(() => { whDeleteConfirm.value = '' }, 3000); return }
  try { await deleteWebhook(id); await loadWebhooks(); whDeleteConfirm.value = '' }
  catch (e: any) { alert(e.response?.data?.error || 'Failed') }
}

async function handleViewLogs(id: string) {
  if (whLogsTarget.value === id) { whLogsTarget.value = ''; whLogs.value = []; return }
  whLogsTarget.value = id
  try { const { data } = await getWebhookLogs(id); whLogs.value = data.items || [] } catch { whLogs.value = [] }
}

function toggleEvent(ev: string) {
  const idx = whEvents.value.indexOf(ev)
  if (idx >= 0) whEvents.value.splice(idx, 1)
  else whEvents.value.push(ev)
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 to-blue-50">
    <AppHeader />

    <main class="max-w-5xl mx-auto px-4 py-8">
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-xl font-semibold text-gray-900">{{ t('settings.title') }}</h2>
        <span v-if="auth.isAdmin" class="px-2.5 py-1 bg-amber-50 text-amber-700 rounded-full text-xs font-medium">
          {{ t('settings.admin') }}
        </span>
      </div>

      <div class="flex gap-1 bg-white rounded-xl p-1 shadow-card mb-6">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          @click="activeTab = tab.key"
          :class="[
            'flex-1 px-4 py-2 text-sm font-medium rounded-lg transition-all duration-150',
            activeTab === tab.key
              ? 'bg-blue-600 text-white shadow-sm'
              : 'text-gray-500 hover:text-gray-700 hover:bg-gray-50'
          ]"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- System Info -->
      <div v-if="activeTab === 'system'">
        <div v-if="sysInfo" class="grid grid-cols-2 gap-4">
          <div class="card p-5"><div class="text-sm text-gray-500 mb-1">{{ t('system.users') }}</div><div class="text-2xl font-bold">{{ sysInfo.user_count }}</div></div>
          <div class="card p-5"><div class="text-sm text-gray-500 mb-1">{{ t('system.snippets') }}</div><div class="text-2xl font-bold">{{ sysInfo.snippet_count }}</div></div>
          <div class="card p-5"><div class="text-sm text-gray-500 mb-1">{{ t('system.devices') }}</div><div class="text-2xl font-bold">{{ sysInfo.device_count }}</div></div>
          <div class="card p-5"><div class="text-sm text-gray-500 mb-1">{{ t('system.db_size') }}</div><div class="text-2xl font-bold">{{ formatBytes(sysInfo.db_size_bytes) }}</div></div>
          <div class="card p-5 col-span-2">
            <div class="text-sm text-gray-500 mb-1">{{ t('system.uptime') }}</div>
            <div class="text-2xl font-bold">{{ formatUptime(sysInfo.uptime_seconds) }}</div>
            <div class="text-xs text-gray-400 mt-1">{{ t('system.started') }}: {{ new Date(sysInfo.started_at).toLocaleString() }}</div>
          </div>
        </div>
        <div v-else class="card p-12 text-center text-gray-400">{{ t('common.loading') }}</div>
      </div>

      <!-- Account -->
      <div v-if="activeTab === 'account'" class="space-y-4">
        <div class="card p-6">
          <h3 class="font-medium text-gray-900 mb-4">{{ t('account.profile') }}</h3>
          <div class="space-y-2.5 text-sm">
            <div class="flex justify-between"><span class="text-gray-500">{{ t('account.username') }}</span><span class="font-medium">{{ auth.username }}</span></div>
            <div class="flex justify-between"><span class="text-gray-500">{{ t('account.user_id') }}</span><code class="text-xs bg-gray-50 px-2 py-0.5 rounded-lg">{{ auth.userId }}</code></div>
            <div class="flex justify-between"><span class="text-gray-500">{{ t('account.role') }}</span><span :class="auth.isAdmin ? 'text-amber-600' : 'text-gray-600'" class="font-medium">{{ auth.isAdmin ? t('account.role_admin') : t('account.role_user') }}</span></div>
          </div>
        </div>

        <div class="card p-6">
          <h3 class="font-medium text-gray-900 mb-4">{{ t('account.change_pass') }}</h3>
          <form @submit.prevent="handleChangePassword" class="space-y-3 max-w-sm">
            <input v-model="oldPass" type="password" :placeholder="t('account.current_pass')" required class="input" />
            <input v-model="newPass" type="password" :placeholder="t('account.new_pass')" required minlength="6" class="input" />
            <div v-if="passError" class="bg-red-50 border border-red-100 text-red-700 px-4 py-3 rounded-xl text-sm">{{ passError }}</div>
            <div v-if="passMsg" class="bg-green-50 border border-green-100 text-green-700 px-4 py-3 rounded-xl text-sm">{{ passMsg }}</div>
            <button type="submit" :disabled="passLoading" class="btn-primary">{{ passLoading ? '...' : t('account.submit_pass') }}</button>
          </form>
        </div>

        <button @click="auth.logout(); router.push('/login')" class="btn-danger">{{ t('account.logout') }}</button>
      </div>

      <!-- Webhooks -->
      <div v-if="activeTab === 'webhooks'" class="space-y-4">
        <div class="card p-6">
          <h3 class="font-medium text-gray-900 mb-4">{{ t('webhooks.add') }}</h3>
          <form @submit.prevent="handleCreateWebhook" class="space-y-3">
            <div class="grid grid-cols-2 gap-3">
              <input v-model="whName" :placeholder="t('webhooks.name')" required class="input" />
              <input v-model="whUrl" :placeholder="t('webhooks.url')" required type="url" class="input" />
            </div>
            <div class="flex gap-2">
              <label v-for="ev in ['created', 'updated', 'deleted']" :key="ev" class="flex items-center gap-1.5 text-sm cursor-pointer">
                <input type="checkbox" :checked="whEvents.includes(ev)" @change="toggleEvent(ev)" class="rounded" />
                <span class="px-2 py-0.5 bg-gray-100 rounded text-xs font-mono">{{ ev }}</span>
              </label>
            </div>
            <div>
              <textarea v-model="whTemplate" :placeholder="t('webhooks.template_placeholder')" rows="10" class="input font-mono text-xs" />
              <p class="text-xs text-gray-400 mt-1">{{ t('webhooks.template_hint') }}</p>
              <details class="mt-1.5">
                <summary class="text-xs text-blue-500 cursor-pointer hover:text-blue-700">{{ t('webhooks.template_vars') }}</summary>
                <div class="mt-1.5 p-3 bg-gray-50 rounded-lg text-xs font-mono text-gray-600 space-y-1">
                  <div v-for="v in templateVars" :key="v.name">
                    <span class="text-blue-600" v-text="v.name"></span> — {{ v.desc }}
                  </div>
                </div>
              </details>
              <details class="mt-1.5">
                <summary class="text-xs text-blue-500 cursor-pointer hover:text-blue-700">{{ t('webhooks.example') }}</summary>
                <div class="mt-1.5 p-3 bg-gray-50 rounded-lg text-xs space-y-3">
                  <div>
                    <div class="font-medium text-gray-700 mb-1">{{ t('webhooks.example_dingtalk') }}</div>
                    <p class="text-gray-500 mb-1.5" v-if="lang === 'zh'">URL 填钉钉机器人的 webhook 地址，模板填写：</p>
                    <p class="text-gray-500 mb-1.5" v-else>Use DingTalk robot webhook URL, template:</p>
                    <pre class="bg-white p-2 rounded border border-gray-200 text-xs overflow-x-auto whitespace-pre-wrap break-all" v-text="dingtalkExample"></pre>
                    <button type="button" class="mt-1.5 text-xs text-blue-500 hover:text-blue-700" @click="whTemplate = dingtalkExample">
                      {{ lang === 'zh' ? '填入此示例' : 'Use this example' }}
                    </button>
                  </div>
                </div>
              </details>
            </div>
            <div v-if="whError" class="bg-red-50 border border-red-100 text-red-700 px-4 py-3 rounded-xl text-sm">{{ whError }}</div>
            <button type="submit" :disabled="whLoading" class="btn-primary">{{ whLoading ? '...' : t('webhooks.add') }}</button>
          </form>
        </div>

        <div v-if="webhooks.length === 0" class="card p-12 text-center text-gray-400">{{ t('webhooks.empty') }}</div>

        <div v-for="wh in webhooks" :key="wh.id" class="card p-5">
          <div class="flex items-start justify-between mb-2">
            <div>
              <span class="font-medium text-gray-900">{{ wh.name }}</span>
              <span :class="wh.active ? 'bg-green-50 text-green-700' : 'bg-gray-100 text-gray-500'" class="ml-2 px-2 py-0.5 rounded-lg text-xs font-medium">
                {{ wh.active ? t('webhooks.active') : t('webhooks.inactive') }}
              </span>
            </div>
            <div class="flex gap-1.5">
              <button @click="handleToggleWebhook(wh.id)" class="btn-ghost text-xs px-2.5 py-1">{{ t('webhooks.toggle') }}</button>
              <button @click="handleViewLogs(wh.id)" class="btn-ghost text-xs px-2.5 py-1 text-blue-600">{{ t('webhooks.logs') }}</button>
              <button @click="handleDeleteWebhook(wh.id)" :class="whDeleteConfirm === wh.id ? 'btn-danger text-xs px-2.5 py-1' : 'btn-ghost text-xs px-2.5 py-1 text-red-600'">
                {{ whDeleteConfirm === wh.id ? t('webhooks.confirm_delete') : t('webhooks.delete') }}
              </button>
            </div>
          </div>
          <div class="text-xs text-gray-500 space-y-1">
            <div><span class="text-gray-400 w-12 inline-block">URL:</span> <code class="bg-gray-50 px-1.5 py-0.5 rounded">{{ wh.url }}</code></div>
            <div><span class="text-gray-400 w-12 inline-block">Events:</span> {{ wh.events.join(', ') }}</div>
            <div v-if="wh.body_template"><span class="text-gray-400 w-12 inline-block">Template:</span> <code class="bg-gray-50 px-1.5 py-0.5 rounded text-xs">{{ wh.body_template.length > 100 ? wh.body_template.slice(0, 100) + '...' : wh.body_template }}</code></div>
          </div>

          <!-- Logs -->
          <div v-if="whLogsTarget === wh.id" class="mt-3 pt-3 border-t border-gray-100">
            <h4 class="text-xs font-medium text-gray-500 mb-2">{{ t('webhooks.logs') }}</h4>
            <div v-if="whLogs.length === 0" class="text-xs text-gray-400">{{ t('webhooks.no_logs') }}</div>
            <div v-else class="space-y-1 max-h-48 overflow-y-auto">
              <div v-for="log in whLogs" :key="log.id" class="flex items-center gap-2 text-xs">
                <span :class="log.status_code >= 200 && log.status_code < 300 ? 'text-green-600' : 'text-red-600'" class="font-mono w-8">
                  {{ log.status_code >= 200 && log.status_code < 300 ? t('webhooks.status_ok') : (log.status_code || t('webhooks.status_fail')) }}
                </span>
                <span class="text-gray-500 font-mono">{{ log.event_type }}</span>
                <span class="text-gray-400">{{ new Date(log.created_at).toLocaleString() }}</span>
                <span v-if="log.error" class="text-red-500 truncate">{{ log.error }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Server Config -->
      <div v-if="activeTab === 'server' && auth.isAdmin" class="space-y-4">
        <div class="card p-6">
          <h3 class="font-medium text-gray-900 mb-4">{{ t('server.status') }}</h3>
          <div v-if="settingsLoading" class="text-gray-400 text-sm">{{ t('common.loading') }}</div>
          <div v-else class="space-y-3">
            <div class="flex items-center gap-3">
              <span class="text-sm text-gray-500 w-32">{{ t('server.cors') }}</span>
              <input v-model="corsOrigin" placeholder="*" class="input flex-1" />
              <button @click="saveCORS" class="btn-secondary">{{ t('server.save') }}</button>
            </div>
            <p class="text-xs text-gray-400">{{ t('server.cors_hint') }}</p>
            <p v-if="settingsMsg" class="text-sm text-green-600">{{ settingsMsg }}</p>
          </div>
        </div>

        <div class="card p-6">
          <h3 class="font-medium text-gray-900 mb-3">{{ t('server.https') }}</h3>
          <p class="text-sm text-gray-500 mb-4">{{ t('server.https_desc') }}</p>
          <div v-if="settings.tls_enabled">
            <div class="flex items-center gap-2 mb-3">
              <span class="w-2 h-2 bg-green-500 rounded-full"></span>
              <span class="text-sm text-green-600 font-medium">{{ t('server.https_active') }}</span>
            </div>
            <p class="text-xs text-gray-400 mb-4">{{ t('server.https_redirect_note') }}</p>
            <div class="border-t border-gray-100 pt-4 flex items-center gap-3">
              <button @click="handleDisableTLS" :disabled="tlsCountdown > 0 && tlsDisableConfirm && !tlsLoading"
                :class="[
                  'px-4 py-2 rounded-xl text-sm font-medium transition-all',
                  tlsLoading ? 'bg-gray-200 text-gray-400' :
                  tlsDisableConfirm && tlsCountdown > 0 ? 'bg-gray-200 text-gray-500' :
                  tlsDisableConfirm && tlsCountdown <= 0 ? 'bg-red-600 text-white' :
                  'border border-red-200 text-red-600 hover:bg-red-50'
                ]">
                {{ tlsLoading ? '...' :
                   tlsDisableConfirm && tlsCountdown > 0 ? t('server.disabling') + ' (' + tlsCountdown + ')' :
                   tlsDisableConfirm && tlsCountdown <= 0 ? t('server.confirm_disable') :
                   t('server.disable_https') }}
              </button>
              <span class="text-xs text-gray-400">{{ t('server.disable_https_hint') }}</span>
            </div>
          </div>
          <div v-else>
            <div class="bg-amber-50 border border-amber-100 rounded-lg p-3 mb-3">
              <p class="text-xs text-amber-700">{{ t('server.https_note') }}</p>
            </div>
            <button @click="handleEnableTLS" :disabled="tlsLoading" class="btn-primary">
              {{ tlsLoading ? '...' : t('server.enable_https') }}
            </button>
            <p v-if="tlsMsg" class="mt-2 text-sm" :class="tlsMsg.includes('enabled') ? 'text-green-600' : 'text-red-600'">{{ tlsMsg }}</p>
          </div>
        </div>

        <div class="card p-6">
          <h3 class="font-medium text-gray-900 mb-3">{{ t('server.reverse_proxy') }}</h3>
          <p class="text-sm text-gray-500 mb-4">{{ t('server.proxy_desc') }}</p>

          <div class="space-y-4">
            <div>
              <h4 class="text-sm font-medium text-gray-700 mb-2">Caddy {{ t('server.recommended') }}</h4>
              <pre class="bg-gray-50 p-3 rounded-lg text-xs font-mono text-gray-700 overflow-x-auto">cb.yourdomain.com {
    reverse_proxy localhost:8080
}</pre>
              <p class="text-xs text-gray-400 mt-1">{{ t('server.caddy_hint') }}</p>
            </div>

            <div class="border-t border-gray-100 pt-4">
              <h4 class="text-sm font-medium text-gray-700 mb-2">Nginx</h4>
              <pre class="bg-gray-50 p-3 rounded-lg text-xs font-mono text-gray-700 overflow-x-auto whitespace-pre-wrap">server {
    listen 443 ssl http2;
    server_name cb.yourdomain.com;

    ssl_certificate     /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /v1/ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_read_timeout 86400s;
    }
}</pre>
              <p class="text-xs text-gray-400 mt-1">{{ t('server.nginx_hint') }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Users -->
      <div v-if="activeTab === 'users' && auth.isAdmin" class="space-y-4">
        <div class="card p-6">
          <h3 class="font-medium text-gray-900 mb-3">{{ t('users.add_user') }}</h3>
          <form @submit.prevent="handleAddUser" class="flex gap-3 items-end">
            <div class="flex-1">
              <input v-model="newUsername" :placeholder="t('users.username')" required class="input" />
            </div>
            <div class="flex-1">
              <input v-model="newUserPass" type="password" :placeholder="t('account.new_pass')" required minlength="6" class="input" />
            </div>
            <button type="submit" :disabled="addUserLoading" class="btn-primary whitespace-nowrap">
              {{ addUserLoading ? '...' : t('users.add_user') }}
            </button>
          </form>
          <p v-if="addUserError" class="mt-2 text-sm text-red-600">{{ addUserError }}</p>
        </div>

        <div class="card overflow-hidden">
          <table class="w-full text-sm">
            <thead class="bg-gray-50 border-b border-gray-100">
              <tr>
                <th class="text-left px-5 py-3 font-medium text-gray-500">{{ t('users.username') }}</th>
                <th class="text-left px-5 py-3 font-medium text-gray-500">{{ t('users.role') }}</th>
                <th class="text-left px-5 py-3 font-medium text-gray-500">{{ t('users.created') }}</th>
                <th class="text-right px-5 py-3 font-medium text-gray-500">{{ t('users.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in users" :key="user.id" class="border-b border-gray-50 last:border-0 hover:bg-gray-50/50">
                <td class="px-5 py-3"><span class="font-medium">{{ user.username }}</span><span v-if="user.id === auth.userId" class="text-xs text-gray-400 ml-1">{{ t('users.you') }}</span></td>
                <td class="px-5 py-3"><span :class="user.is_admin ? 'bg-amber-50 text-amber-700' : 'bg-gray-100 text-gray-500'" class="px-2 py-0.5 rounded-lg text-xs font-medium">{{ user.is_admin ? t('users.admin') : t('users.user') }}</span></td>
                <td class="px-5 py-3 text-gray-500">{{ new Date(user.created_at).toLocaleDateString() }}</td>
                <td class="px-5 py-3 text-right">
                  <template v-if="user.id !== auth.userId">
                    <div class="flex gap-1.5 justify-end">
                      <button @click="handleToggleAdmin(user.id)" :class="toggleConfirm === user.id ? 'btn-danger text-xs px-2.5 py-1' : 'btn-ghost text-xs px-2.5 py-1 text-amber-600'">{{ toggleConfirm === user.id ? t('users.confirm') : (user.is_admin ? t('users.revoke_admin') : t('users.make_admin')) }}</button>
                      <button @click="handleResetPassword(user.id)" class="btn-ghost text-xs px-2.5 py-1 text-blue-600">{{ resetTarget === user.id ? t('snippets.cancel') : t('users.reset_pass') }}</button>
                      <button @click="handleDeleteUser(user.id)" :class="deleteConfirm === user.id ? 'btn-danger text-xs px-2.5 py-1' : 'btn-ghost text-xs px-2.5 py-1 text-red-600'">{{ deleteConfirm === user.id ? t('users.confirm_delete') : t('users.delete') }}</button>
                    </div>
                  </template>
                  <span v-else class="text-xs text-gray-300">-</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-if="resetTarget" class="card p-5 mt-4">
          <h4 class="font-medium text-sm mb-3">{{ t('users.reset_pass') }}: {{ users.find(u => u.id === resetTarget)?.username }}</h4>
          <div class="flex gap-2">
            <input v-model="resetNewPass" type="password" :placeholder="t('account.new_pass')" minlength="6" class="input flex-1" />
            <button @click="handleResetPassword(resetTarget)" class="btn-primary">{{ t('users.reset_pass') }}</button>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>
