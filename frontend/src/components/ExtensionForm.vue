<template>
  <div class="bg-white rounded-lg shadow p-6 mb-6">
    <h3 class="text-lg font-medium text-gray-900 mb-4">Register New Extension</h3>
    <form @submit.prevent="handleSubmit" class="space-y-4">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:flex lg:items-end gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Extension Number</label>
          <input
            v-model.number="extension"
            type="number"
            required
            :min="authStore.user?.min_ext"
            :max="authStore.user?.max_ext"
            class="w-full lg:w-40 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            :placeholder="`${authStore.user?.min_ext}-${authStore.user?.max_ext}`"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Caller ID Name</label>
          <input
            v-model="callerid"
            type="text"
            class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            placeholder="e.g. John Doe"
          />
        </div>
        <div class="sm:col-span-2 lg:col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">SIP Password (blank = auto)</label>
          <div class="flex gap-2">
            <input
              v-model="sipPassword"
              type="text"
              class="w-full min-w-0 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono"
              placeholder="auto-generated"
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
        <button
          type="submit"
          :disabled="submitting || !passwordValid"
          class="w-full sm:w-auto bg-green-600 hover:bg-green-700 text-white font-medium py-2 px-6 rounded-md disabled:opacity-50"
        >
          {{ submitting ? 'Registering...' : 'Register' }}
        </button>
      </div>

      <ul v-if="sipPassword" class="text-xs space-y-1">
        <li :class="sipPassword.length >= 8 ? 'text-green-600' : 'text-gray-400'">
          {{ sipPassword.length >= 8 ? '\u2713' : '\u2022' }} At least 8 characters
        </li>
        <li :class="hasUppercase ? 'text-green-600' : 'text-gray-400'">
          {{ hasUppercase ? '\u2713' : '\u2022' }} At least one uppercase letter
        </li>
        <li :class="hasNumber ? 'text-green-600' : 'text-gray-400'">
          {{ hasNumber ? '\u2713' : '\u2022' }} At least one number
        </li>
      </ul>
    </form>
    <p v-if="success" class="mt-3 text-green-600 text-sm">Extension registered successfully!</p>
    <p v-if="error" class="mt-3 text-red-600 text-sm">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useExtensionsStore } from '../stores/extensions'

const emit = defineEmits<{ created: [] }>()

const authStore = useAuthStore()
const extStore = useExtensionsStore()

const extension = ref<number | undefined>()
const callerid = ref('')
const sipPassword = ref('')
const submitting = ref(false)
const success = ref(false)
const error = ref<string | null>(null)

const hasUppercase = computed(() => /[A-Z]/.test(sipPassword.value))
const hasNumber = computed(() => /[0-9]/.test(sipPassword.value))
const passwordValid = computed(() => {
  if (!sipPassword.value) return true // blank = auto-generate
  return sipPassword.value.length >= 8 && hasUppercase.value && hasNumber.value
})

function generatePassword() {
  const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
  const arr = new Uint8Array(10)
  crypto.getRandomValues(arr)
  sipPassword.value = Array.from(arr, (b) => chars[b % chars.length]).join('')
}

async function handleSubmit() {
  if (extension.value === undefined) return
  submitting.value = true
  success.value = false
  error.value = null
  try {
    await extStore.createExtension(extension.value, callerid.value, sipPassword.value)
    success.value = true
    extension.value = undefined
    callerid.value = ''
    sipPassword.value = ''
    emit('created')
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to register extension'
  } finally {
    submitting.value = false
  }
}
</script>
