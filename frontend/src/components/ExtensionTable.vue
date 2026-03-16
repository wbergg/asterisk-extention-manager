<template>
  <div class="bg-white rounded-lg shadow overflow-hidden">
    <!-- Inline Edit Form -->
    <div v-if="editingExt" class="p-6 border-b border-blue-200 bg-blue-50">
      <h3 class="text-lg font-medium text-gray-900 mb-4">Edit Extension {{ editingExt.extension }}</h3>
      <form @submit.prevent="handleSave" class="space-y-4">
        <div class="flex items-end gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Caller ID</label>
            <input
              v-model="editCallerid"
              type="text"
              class="px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">New SIP Password (leave blank to keep)</label>
            <div class="flex gap-2">
              <input
                v-model="editSipPassword"
                type="text"
                class="px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono"
                placeholder="unchanged"
              />
              <button
                type="button"
                @click="generateSipPassword"
                class="px-3 py-2 bg-gray-200 hover:bg-gray-300 text-gray-700 text-sm rounded-md whitespace-nowrap"
              >
                Generate
              </button>
            </div>
          </div>
        </div>

        <ul v-if="editSipPassword" class="text-xs space-y-1">
          <li :class="editSipPassword.length >= 8 ? 'text-green-600' : 'text-gray-400'">
            {{ editSipPassword.length >= 8 ? '\u2713' : '\u2022' }} At least 8 characters
          </li>
          <li :class="sipHasUppercase ? 'text-green-600' : 'text-gray-400'">
            {{ sipHasUppercase ? '\u2713' : '\u2022' }} At least one uppercase letter
          </li>
          <li :class="sipHasNumber ? 'text-green-600' : 'text-gray-400'">
            {{ sipHasNumber ? '\u2713' : '\u2022' }} At least one number
          </li>
        </ul>

        <div v-if="editError" class="px-4 py-3 rounded text-sm bg-red-50 border border-red-200 text-red-700">
          {{ editError }}
        </div>

        <div class="flex gap-4">
          <button
            type="submit"
            :disabled="!canSave"
            class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-md disabled:opacity-50"
          >
            Save
          </button>
          <button type="button" @click="editingExt = null" class="text-gray-600 hover:text-gray-800 font-medium py-2 px-4">Cancel</button>
        </div>
      </form>
    </div>

    <table class="min-w-full divide-y divide-gray-200">
      <thead class="bg-gray-50">
        <tr>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Extension</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">SIP Username</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">SIP Password</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Caller ID</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-200">
        <tr v-if="loading">
          <td colspan="6" class="px-6 py-8 text-center text-gray-500">Loading...</td>
        </tr>
        <tr v-else-if="extensions.length === 0">
          <td colspan="6" class="px-6 py-8 text-center text-gray-500">No extensions registered yet</td>
        </tr>
        <tr v-for="ext in extensions" :key="ext.id">
          <td class="px-6 py-4 text-sm font-mono font-semibold">{{ ext.extension }}</td>
          <td class="px-6 py-4 text-sm font-mono">{{ ext.sip_username }}</td>
          <td class="px-6 py-4 text-sm font-mono">
            <span v-if="!visiblePasswords.has(ext.id)">
              <button @click="visiblePasswords.add(ext.id)" class="text-blue-600 hover:text-blue-800 text-xs">Show</button>
            </span>
            <span v-else>
              {{ ext.sip_password }}
              <button @click="visiblePasswords.delete(ext.id)" class="text-gray-400 hover:text-gray-600 text-xs ml-2">Hide</button>
            </span>
          </td>
          <td class="px-6 py-4 text-sm">{{ ext.callerid || '-' }}</td>
          <td class="px-6 py-4 text-sm text-gray-500">{{ new Date(ext.created_at).toLocaleString() }}</td>
          <td class="px-6 py-4 text-sm space-x-3">
            <button
              @click="startEdit(ext)"
              class="text-blue-600 hover:text-blue-800 font-medium"
            >
              Edit
            </button>
            <button
              @click="$emit('delete', ext.extension)"
              class="text-red-600 hover:text-red-800 font-medium"
            >
              Delete
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import type { Extension } from '../stores/extensions'

defineProps<{
  extensions: Extension[]
  loading: boolean
}>()

const emit = defineEmits<{
  delete: [ext: number]
  edit: [ext: number, callerid: string, sipPassword: string]
}>()

const visiblePasswords = reactive(new Set<number>())
const editingExt = ref<Extension | null>(null)
const editCallerid = ref('')
const editSipPassword = ref('')
const editError = ref('')

const sipHasUppercase = computed(() => /[A-Z]/.test(editSipPassword.value))
const sipHasNumber = computed(() => /[0-9]/.test(editSipPassword.value))
const sipPasswordValid = computed(() => {
  if (!editSipPassword.value) return true // blank = keep current
  return editSipPassword.value.length >= 8 && sipHasUppercase.value && sipHasNumber.value
})
const canSave = computed(() => sipPasswordValid.value)

function generateSipPassword() {
  const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
  const arr = new Uint8Array(16)
  crypto.getRandomValues(arr)
  editSipPassword.value = Array.from(arr, (b) => chars[b % chars.length]).join('')
}

function startEdit(ext: Extension) {
  editingExt.value = ext
  editCallerid.value = ext.callerid
  editSipPassword.value = ''
  editError.value = ''
}

function handleSave() {
  if (!editingExt.value || !canSave.value) return
  editError.value = ''
  emit('edit', editingExt.value.extension, editCallerid.value, editSipPassword.value)
  editingExt.value = null
}
</script>
