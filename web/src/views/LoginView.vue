<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useI18n } from '../i18n'

const router = useRouter()
const auth = useAuthStore()
const { t, lang, setLang } = useI18n()

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const isRegister = ref(false)

async function handleSubmit() {
  error.value = ''
  loading.value = true
  try {
    if (isRegister.value) {
      await auth.register(username.value, password.value)
    } else {
      await auth.login(username.value, password.value)
    }
    router.push('/')
  } catch (e: any) {
    error.value = e.response?.data?.error || e.message || 'Request failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-blue-50">
    <div class="w-full max-w-md mx-4">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-gray-900 tracking-tight">cb</h1>
        <p class="text-sm text-gray-500 mt-1">{{ t('login.title') }}</p>
      </div>

      <div class="card p-8">
        <form @submit.prevent="handleSubmit" class="space-y-5">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('login.username') }}</label>
            <input
              v-model="username"
              type="text"
              required
              minlength="2"
              maxlength="32"
              class="input"
              :placeholder="t('login.placeholder.user')"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1.5">{{ t('login.password') }}</label>
            <input
              v-model="password"
              type="password"
              required
              minlength="6"
              class="input"
              :placeholder="t('login.placeholder.pass')"
            />
          </div>

          <div v-if="error" class="bg-red-50 border border-red-100 text-red-700 px-4 py-3 rounded-xl text-sm">
            {{ error }}
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="btn-primary w-full py-2.5"
          >
            {{ loading ? t('login.submitting') : (isRegister ? t('login.register_btn') : t('login.submit')) }}
          </button>
        </form>

        <div class="text-center mt-5 pt-5 border-t border-gray-100">
          <button
            @click="isRegister = !isRegister; error = ''"
            class="text-sm text-blue-600 hover:text-blue-700 transition-colors"
          >
            {{ isRegister ? t('login.has_account') : t('login.no_account') }}
          </button>
        </div>
      </div>

      <div class="flex justify-center mt-4">
        <div class="flex items-center bg-white/60 rounded-lg p-0.5 shadow-sm">
          <button
            @click="setLang('zh')"
            :class="[
              'px-3 py-1 text-xs font-medium rounded-md transition-all duration-150',
              lang === 'zh' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-500 hover:text-gray-700'
            ]"
          >
            中文
          </button>
          <button
            @click="setLang('en')"
            :class="[
              'px-3 py-1 text-xs font-medium rounded-md transition-all duration-150',
              lang === 'en' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-500 hover:text-gray-700'
            ]"
          >
            EN
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
