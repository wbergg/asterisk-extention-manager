<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-gray-900">My Extensions</h1>
      <p class="text-gray-600 mt-1">
        Your permitted range: <span class="font-mono font-semibold">{{ authStore.user?.min_ext }}</span> -
        <span class="font-mono font-semibold">{{ authStore.user?.max_ext }}</span>
      </p>
    </div>

    <div v-if="extStore.error" class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-6">
      {{ extStore.error }}
    </div>

    <ExtensionForm @created="extStore.fetchExtensions()" />

    <ExtensionTable
      :extensions="extStore.extensions"
      :loading="extStore.loading"
      @edit="handleEdit"
      @delete="handleDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useExtensionsStore } from '../stores/extensions'
import ExtensionForm from '../components/ExtensionForm.vue'
import ExtensionTable from '../components/ExtensionTable.vue'

const authStore = useAuthStore()
const extStore = useExtensionsStore()

onMounted(() => {
  extStore.fetchExtensions()
})

async function handleEdit(ext: number, callerid: string, sipPassword: string) {
  await extStore.updateExtension(ext, callerid, sipPassword)
}

async function handleDelete(ext: number) {
  if (confirm(`Delete extension ${ext}?`)) {
    await extStore.deleteExtension(ext)
  }
}
</script>
