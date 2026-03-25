<template>
  <div class="min-h-screen bg-gray-100">
    <!-- Impersonation Banner -->
    <div v-if="authStore.isImpersonating" class="bg-yellow-500 text-yellow-900 text-center py-2 text-sm font-medium">
      Viewing as <strong>{{ authStore.user?.username }}</strong>
      <button
        @click="stopImpersonating"
        class="ml-4 bg-yellow-700 hover:bg-yellow-800 text-white text-xs px-3 py-1 rounded"
      >
        Stop Impersonating
      </button>
    </div>

    <nav v-if="authStore.isLoggedIn" class="bg-gray-900 text-white shadow-lg">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-16">
          <span class="text-xl font-bold tracking-tight hidden sm:block">Asterisk Extention Manager</span>
          <span class="text-lg font-bold tracking-tight sm:hidden">AEM</span>
          <div class="flex items-center space-x-4">
            <!-- User dropdown for admins -->
            <div v-if="authStore.isRealAdmin && !authStore.isImpersonating" class="relative">
              <button
                @click="showUserMenu = !showUserMenu"
                class="text-sm text-gray-400 hover:text-white flex items-center gap-1"
              >
                {{ authStore.user?.username }}
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
              </button>
              <div
                v-if="showUserMenu"
                class="absolute right-0 mt-2 w-56 bg-white rounded-md shadow-lg ring-1 ring-black ring-opacity-5 z-50 max-h-80 overflow-y-auto"
              >
                <div class="py-1">
                  <div class="px-4 py-2 text-xs text-gray-500 uppercase font-medium">Impersonate User</div>
                  <button
                    v-for="u in userList"
                    :key="u.id"
                    @click="handleImpersonate(u.id)"
                    class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  >
                    {{ u.username }}
                    <span class="text-gray-400 text-xs ml-1">{{ u.role }}</span>
                  </button>
                </div>
              </div>
            </div>
            <span v-else class="text-sm text-gray-400">{{ authStore.user?.username }}</span>
            <button @click="logout" class="bg-red-600 hover:bg-red-700 text-white text-sm px-3 py-1.5 rounded">
              Logout
            </button>
          </div>
        </div>
        <div class="flex overflow-x-auto -mb-px space-x-1 pb-2 sm:pb-0 sm:space-x-4">
          <router-link to="/extensions" class="text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium whitespace-nowrap">
            Extensions
          </router-link>
          <router-link to="/directory" class="text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium whitespace-nowrap">
            Directory
          </router-link>
          <router-link to="/fax" class="text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium whitespace-nowrap">
            Fax
          </router-link>
          <router-link v-if="authStore.hasCallLogAccess" to="/call-log" class="text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium whitespace-nowrap">
            Call Log
          </router-link>
          <router-link to="/settings" class="text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium whitespace-nowrap">
            Settings
          </router-link>
          <router-link v-if="authStore.isAdmin" to="/admin" class="text-gray-300 hover:text-white px-3 py-2 rounded-md text-sm font-medium whitespace-nowrap">
            Admin
          </router-link>
        </div>
      </div>
    </nav>
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from './stores/auth'
import { useRouter } from 'vue-router'
import api from './api/client'

const authStore = useAuthStore()
const router = useRouter()

interface SimpleUser {
  id: number
  username: string
  role: string
}

const showUserMenu = ref(false)
const userList = ref<SimpleUser[]>([])

async function fetchUsers() {
  if (!authStore.isRealAdmin || authStore.isImpersonating) return
  try {
    const res = await api.get('/admin/users')
    userList.value = res.data.filter((u: SimpleUser) => u.id !== authStore.user?.id)
  } catch {
    // Not admin or token expired
  }
}

watch(() => authStore.isAdmin, (isAdmin) => {
  if (isAdmin) fetchUsers()
})

onMounted(() => {
  if (authStore.isAdmin) fetchUsers()
})

async function handleImpersonate(userId: number) {
  showUserMenu.value = false
  await authStore.impersonate(userId)
  router.push('/extensions')
}

function stopImpersonating() {
  authStore.stopImpersonating()
  router.push('/admin')
}

function logout() {
  authStore.logout()
  router.push('/login')
}

// Close dropdown when clicking outside
function handleClickOutside(e: MouseEvent) {
  if (showUserMenu.value && !(e.target as HTMLElement).closest('.relative')) {
    showUserMenu.value = false
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onUnmounted(() => document.removeEventListener('click', handleClickOutside))
</script>
