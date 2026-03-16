<template>
  <div class="bg-white rounded-lg shadow p-6 mb-6 border border-blue-200">
    <h3 class="text-lg font-medium text-gray-900 mb-4">
      {{ user ? 'Edit User' : 'Create User' }}
    </h3>
    <form @submit.prevent="handleSubmit" class="space-y-4">
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Username</label>
          <input
            v-model="form.username"
            type="text"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Password{{ user ? ' (leave blank to keep)' : '' }}
          </label>
          <div class="flex gap-2">
            <input
              v-model="form.password"
              type="text"
              :required="!user"
              class="flex-1 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono"
            />
            <button
              type="button"
              @click="generatePassword"
              class="px-3 py-2 bg-gray-200 hover:bg-gray-300 text-gray-700 text-sm rounded-md whitespace-nowrap"
            >
              Generate
            </button>
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Role</label>
          <select
            v-model="form.role"
            class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="user">User</option>
            <option value="admin">Admin</option>
          </select>
        </div>
        <div class="flex items-center pt-6">
          <label class="flex items-center gap-2 cursor-pointer">
            <input
              v-model="form.call_log_access"
              type="checkbox"
              class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
            />
            <span class="text-sm font-medium text-gray-700">Call Log Access</span>
          </label>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Min Extension</label>
          <input
            v-model.number="form.min_ext"
            type="number"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Max Extension</label>
          <input
            v-model.number="form.max_ext"
            type="number"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
      </div>
      <div v-if="error" class="text-red-600 text-sm">{{ error }}</div>
      <div class="flex items-center gap-3">
        <button
          type="submit"
          :disabled="saving"
          class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-md disabled:opacity-50"
        >
          {{ saving ? 'Saving...' : 'Save' }}
        </button>
        <button
          type="button"
          @click="$emit('cancel')"
          class="text-gray-600 hover:text-gray-800 font-medium py-2 px-4"
        >
          Cancel
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watchEffect } from 'vue'
import api from '../api/client'

interface User {
  id: number
  username: string
  role: string
  min_ext: number
  max_ext: number
  call_log_access: boolean
}

const props = defineProps<{ user: User | null }>()
const emit = defineEmits<{
  saved: []
  cancel: []
}>()

const form = reactive({
  username: '',
  password: '',
  role: 'user',
  min_ext: 0,
  max_ext: 0,
  call_log_access: true,
})

const saving = ref(false)
const error = ref<string | null>(null)

watchEffect(() => {
  if (props.user) {
    form.username = props.user.username
    form.password = ''
    form.role = props.user.role
    form.min_ext = props.user.min_ext
    form.max_ext = props.user.max_ext
    form.call_log_access = props.user.call_log_access
  } else {
    form.username = ''
    form.password = ''
    form.role = 'user'
    form.min_ext = 0
    form.max_ext = 0
    form.call_log_access = true
  }
})

function generatePassword() {
  const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
  const arr = new Uint8Array(10)
  crypto.getRandomValues(arr)
  form.password = Array.from(arr, (b) => chars[b % chars.length]).join('')
}

async function handleSubmit() {
  saving.value = true
  error.value = null
  try {
    if (props.user) {
      await api.put(`/admin/users/${props.user.id}`, form)
    } else {
      await api.post('/admin/users', form)
    }
    emit('saved')
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to save user'
  } finally {
    saving.value = false
  }
}
</script>
