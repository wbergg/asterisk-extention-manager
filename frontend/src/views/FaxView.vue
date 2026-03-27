<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-8">Send Fax</h1>

    <!-- Send Form -->
    <div class="bg-white rounded-lg shadow p-6 mb-8">
      <form @submit.prevent="handleSend" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Subject</label>
          <input
            v-model="subject"
            type="text"
            placeholder="Fax subject line"
            class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-blue-500 focus:border-blue-500"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Destination</label>
          <select
            v-model="destination"
            class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="" disabled>Select fax destination</option>
            <option
              v-for="dest in faxStore.destinations"
              :key="dest.extension"
              :value="dest.extension"
            >
              {{ dest.name }} ({{ dest.extension }})
            </option>
          </select>
          <p v-if="faxStore.destinations.length === 0" class="mt-1 text-sm text-gray-500">
            No fax destinations found. Extensions with "fax" in their Caller ID will appear here.
          </p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Message (optional)</label>
          <textarea
            v-model="message"
            rows="4"
            placeholder="Text to append as an extra page"
            class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-blue-500 focus:border-blue-500"
          ></textarea>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">File</label>
          <div
            @dragover.prevent="dragOver = true"
            @dragleave.prevent="dragOver = false"
            @drop.prevent="handleDrop"
            @click="fileInput?.click()"
            class="border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors"
            :class="dragOver ? 'border-blue-500 bg-blue-50' : 'border-gray-300 hover:border-gray-400'"
          >
            <div v-if="!selectedFile">
              <p class="text-sm text-gray-600">Drag and drop a file here, or click to select</p>
              <p class="text-xs text-gray-400 mt-1">PDF, PNG, or JPG (max 10MB)</p>
            </div>
            <div v-else>
              <p class="text-sm font-medium text-gray-900">{{ selectedFile.name }}</p>
              <p class="text-xs text-gray-500 mt-1">{{ formatSize(selectedFile.size) }}</p>
              <button
                type="button"
                @click.stop="selectedFile = null"
                class="mt-2 text-xs text-red-600 hover:text-red-800"
              >
                Remove
              </button>
            </div>
          </div>
          <input
            ref="fileInput"
            type="file"
            accept=".pdf,.png,.jpg,.jpeg"
            class="hidden"
            @change="handleFileSelect"
          />
        </div>

        <div v-if="faxStore.error" class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded text-sm">
          {{ faxStore.error }}
        </div>

        <div v-if="successMsg" class="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded text-sm">
          {{ successMsg }}
        </div>

        <button
          type="submit"
          :disabled="faxStore.sending || !selectedFile || !destination"
          class="bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white text-sm px-6 py-2 rounded"
        >
          {{ faxStore.sending ? 'Sending...' : 'Send Fax' }}
        </button>
      </form>
    </div>

    <!-- Fax History -->
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold text-gray-900">Fax History</h2>
      <button
        @click="faxStore.fetchJobs()"
        class="bg-gray-600 hover:bg-gray-700 text-white text-sm px-4 py-2 rounded"
      >
        Refresh
      </button>
    </div>
    <div class="bg-white rounded-lg shadow overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Subject</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Destination</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="job in faxStore.jobs" :key="job.id">
            <td class="px-6 py-4 text-sm text-gray-500">{{ formatDate(job.created_at) }}</td>
            <td class="px-6 py-4 text-sm text-gray-900">{{ job.subject || '(no subject)' }}</td>
            <td class="px-6 py-4 text-sm font-mono font-semibold">{{ job.destination_ext }}</td>
            <td class="px-6 py-4 text-sm">
              <span
                class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
                :class="statusClass(job.status)"
              >
                {{ job.status }}
              </span>
              <span v-if="job.error_message" class="ml-2 text-xs text-red-500">{{ job.error_message }}</span>
            </td>
          </tr>
          <tr v-if="faxStore.jobs.length === 0 && !faxStore.loading">
            <td colspan="4" class="px-6 py-8 text-center text-gray-500">No faxes sent yet</td>
          </tr>
        </tbody>
      </table>
      <div v-if="faxStore.loading" class="px-6 py-4 text-center text-gray-500 text-sm">Loading...</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useFaxStore } from '../stores/fax'

const faxStore = useFaxStore()

const subject = ref('')
const destination = ref<number | ''>('')
const message = ref('')
const selectedFile = ref<File | null>(null)
const dragOver = ref(false)
const successMsg = ref('')
const fileInput = ref<HTMLInputElement | null>(null)

onMounted(() => {
  faxStore.fetchDestinations()
  faxStore.fetchJobs()
})

function handleFileSelect(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files && input.files.length > 0) {
    selectedFile.value = input.files[0]
  }
}

function handleDrop(e: DragEvent) {
  dragOver.value = false
  if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
    const file = e.dataTransfer.files[0]
    const ext = file.name.toLowerCase().split('.').pop()
    if (['pdf', 'png', 'jpg', 'jpeg'].includes(ext || '')) {
      selectedFile.value = file
    }
  }
}

async function handleSend() {
  if (!selectedFile.value || !destination.value) return
  successMsg.value = ''
  faxStore.error = null
  try {
    await faxStore.sendFax(subject.value, destination.value as number, selectedFile.value, message.value)
    successMsg.value = 'Fax queued successfully'
    subject.value = ''
    destination.value = ''
    message.value = ''
    selectedFile.value = null
    if (fileInput.value) fileInput.value.value = ''
  } catch {
    // error is set in store
  }
}

function statusClass(status: string): string {
  switch (status) {
    case 'queued': return 'bg-yellow-100 text-yellow-800'
    case 'attempted': return 'bg-orange-100 text-orange-800'
    case 'sent': return 'bg-green-100 text-green-800'
    case 'failed': return 'bg-red-100 text-red-800'
    case 'converting': return 'bg-blue-100 text-blue-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString()
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}
</script>
