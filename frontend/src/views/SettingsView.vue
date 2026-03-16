<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-8">Settings</h1>

    <section>
      <h2 class="text-xl font-semibold text-gray-800 mb-4">Change Password</h2>
      <div class="bg-white rounded-lg shadow p-6 max-w-md">
        <form @submit.prevent="handleChangePassword" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Current Password</label>
            <input
              v-model="currentPassword"
              type="password"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">New Password</label>
            <div class="flex gap-2">
              <input
                v-model="newPassword"
                :type="showNewPassword ? 'text' : 'password'"
                required
                class="flex-1 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                :class="showNewPassword ? 'font-mono' : ''"
              />
              <button
                type="button"
                @click="generatePassword"
                class="px-3 py-2 bg-gray-200 hover:bg-gray-300 text-gray-700 text-sm rounded-md whitespace-nowrap"
              >
                Generate
              </button>
              <button
                type="button"
                @click="showNewPassword = !showNewPassword"
                class="px-3 py-2 bg-gray-200 hover:bg-gray-300 text-gray-700 text-sm rounded-md whitespace-nowrap"
              >
                {{ showNewPassword ? 'Hide' : 'Show' }}
              </button>
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Confirm New Password</label>
            <input
              v-model="confirmPassword"
              type="password"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <ul class="text-xs space-y-1">
            <li :class="newPassword.length >= 10 ? 'text-green-600' : 'text-gray-400'">
              {{ newPassword.length >= 10 ? '\u2713' : '\u2022' }} At least 10 characters
            </li>
            <li :class="hasUppercase ? 'text-green-600' : 'text-gray-400'">
              {{ hasUppercase ? '\u2713' : '\u2022' }} At least one uppercase letter
            </li>
            <li :class="hasNumber ? 'text-green-600' : 'text-gray-400'">
              {{ hasNumber ? '\u2713' : '\u2022' }} At least one number
            </li>
            <li v-if="newPassword && confirmPassword" :class="passwordsMatch ? 'text-green-600' : 'text-red-500'">
              {{ passwordsMatch ? '\u2713' : '\u2717' }} Passwords match
            </li>
          </ul>

          <div v-if="pwMessage" class="px-4 py-3 rounded text-sm"
            :class="pwError ? 'bg-red-50 border border-red-200 text-red-700' : 'bg-green-50 border border-green-200 text-green-700'">
            {{ pwMessage }}
          </div>

          <button
            type="submit"
            :disabled="!canSubmit"
            class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-md disabled:opacity-50"
          >
            Change Password
          </button>
        </form>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import api from '../api/client'

const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const pwMessage = ref('')
const pwError = ref(false)
const showNewPassword = ref(false)

function generatePassword() {
  const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
  const arr = new Uint8Array(16)
  crypto.getRandomValues(arr)
  const pw = Array.from(arr, (b) => chars[b % chars.length]).join('')
  newPassword.value = pw
  confirmPassword.value = pw
  showNewPassword.value = true
}

const hasUppercase = computed(() => /[A-Z]/.test(newPassword.value))
const hasNumber = computed(() => /[0-9]/.test(newPassword.value))
const passwordsMatch = computed(() => newPassword.value === confirmPassword.value)
const canSubmit = computed(() =>
  currentPassword.value &&
  newPassword.value.length >= 10 &&
  hasUppercase.value &&
  hasNumber.value &&
  passwordsMatch.value
)

async function handleChangePassword() {
  if (!canSubmit.value) return
  pwMessage.value = ''
  try {
    await api.put('/me/password', {
      current_password: currentPassword.value,
      new_password: newPassword.value,
    })
    pwMessage.value = 'Password changed successfully'
    pwError.value = false
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (e: any) {
    pwMessage.value = e.response?.data?.error || 'Failed to change password'
    pwError.value = true
  }
}
</script>
