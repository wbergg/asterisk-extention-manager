<template>
  <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-8">Directory</h1>

    <div class="mb-4">
      <input
        v-model="search"
        type="text"
        placeholder="Search by name or extension..."
        class="w-full max-w-sm px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      />
    </div>

    <div class="bg-white rounded-lg shadow overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Extension</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="entry in filtered" :key="entry.extension">
            <td class="px-6 py-4 text-sm font-mono font-semibold">{{ entry.extension }}</td>
            <td class="px-6 py-4 text-sm">{{ entry.name }}</td>
          </tr>
          <tr v-if="filtered.length === 0">
            <td colspan="2" class="px-6 py-8 text-center text-gray-500">
              {{ entries.length === 0 ? 'No extensions registered' : 'No matches' }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import api from '../api/client'

interface DirectoryEntry {
  extension: number
  name: string
}

const entries = ref<DirectoryEntry[]>([])
const search = ref('')

const filtered = computed(() => {
  if (!search.value) return entries.value
  const q = search.value.toLowerCase()
  return entries.value.filter(
    (e) => e.name.toLowerCase().includes(q) || String(e.extension).includes(q)
  )
})

onMounted(async () => {
  const res = await api.get('/directory')
  entries.value = res.data
})
</script>
