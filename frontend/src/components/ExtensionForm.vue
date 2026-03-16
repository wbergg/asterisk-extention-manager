<template>
  <div class="bg-white rounded-lg shadow p-6 mb-6">
    <h3 class="text-lg font-medium text-gray-900 mb-4">Register New Extension</h3>
    <form @submit.prevent="handleSubmit" class="flex items-end gap-4">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Extension Number</label>
        <input
          v-model.number="extension"
          type="number"
          required
          :min="authStore.user?.min_ext"
          :max="authStore.user?.max_ext"
          class="w-40 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          :placeholder="`${authStore.user?.min_ext}-${authStore.user?.max_ext}`"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Caller ID Name</label>
        <input
          v-model="callerid"
          type="text"
          class="px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          placeholder="e.g. John Doe"
        />
      </div>
      <button
        type="submit"
        :disabled="submitting"
        class="bg-green-600 hover:bg-green-700 text-white font-medium py-2 px-6 rounded-md disabled:opacity-50"
      >
        {{ submitting ? 'Registering...' : 'Register' }}
      </button>
    </form>
    <p v-if="success" class="mt-3 text-green-600 text-sm">Extension registered successfully!</p>
    <p v-if="error" class="mt-3 text-red-600 text-sm">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useExtensionsStore } from '../stores/extensions'

const emit = defineEmits<{ created: [] }>()

const authStore = useAuthStore()
const extStore = useExtensionsStore()

const extension = ref<number | undefined>()
const callerid = ref('')
const submitting = ref(false)
const success = ref(false)
const error = ref<string | null>(null)

async function handleSubmit() {
  if (extension.value === undefined) return
  submitting.value = true
  success.value = false
  error.value = null
  try {
    await extStore.createExtension(extension.value, callerid.value)
    success.value = true
    extension.value = undefined
    callerid.value = ''
    emit('created')
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to register extension'
  } finally {
    submitting.value = false
  }
}
</script>
