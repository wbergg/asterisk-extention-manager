import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../api/client'

interface User {
  id: number
  username: string
  role: string
  min_ext: number
  max_ext: number
  call_log_access: boolean
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<User | null>(JSON.parse(localStorage.getItem('user') || 'null'))
  const adminToken = ref<string | null>(localStorage.getItem('adminToken'))
  const adminUser = ref<User | null>(JSON.parse(localStorage.getItem('adminUser') || 'null'))

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isRealAdmin = computed(() => !!adminToken.value || user.value?.role === 'admin')
  const hasCallLogAccess = computed(() => user.value?.role === 'admin' || user.value?.call_log_access === true)
  const isImpersonating = computed(() => !!adminToken.value)

  async function login(username: string, password: string) {
    const res = await api.post('/login', { username, password })
    token.value = res.data.token
    user.value = res.data.user
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('user', JSON.stringify(res.data.user))
  }

  async function impersonate(userId: number) {
    // Save admin session
    adminToken.value = token.value
    adminUser.value = user.value
    localStorage.setItem('adminToken', token.value!)
    localStorage.setItem('adminUser', JSON.stringify(user.value))

    // Fetch impersonated user token using admin token
    const res = await api.post(`/admin/impersonate/${userId}`)
    token.value = res.data.token
    user.value = res.data.user
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('user', JSON.stringify(res.data.user))
  }

  function stopImpersonating() {
    token.value = adminToken.value
    user.value = adminUser.value
    localStorage.setItem('token', adminToken.value!)
    localStorage.setItem('user', JSON.stringify(adminUser.value))
    adminToken.value = null
    adminUser.value = null
    localStorage.removeItem('adminToken')
    localStorage.removeItem('adminUser')
  }

  function logout() {
    token.value = null
    user.value = null
    adminToken.value = null
    adminUser.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    localStorage.removeItem('adminToken')
    localStorage.removeItem('adminUser')
  }

  return { token, user, adminUser, isLoggedIn, isAdmin, isRealAdmin, hasCallLogAccess, isImpersonating, login, impersonate, stopImpersonating, logout }
})
