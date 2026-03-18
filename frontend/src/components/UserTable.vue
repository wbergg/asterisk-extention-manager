<template>
  <div class="bg-white rounded-lg shadow overflow-x-auto">
    <table class="min-w-full divide-y divide-gray-200">
      <thead class="bg-gray-50">
        <tr>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">ID</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Username</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Role</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Extension Range</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Call Log</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-200">
        <tr v-for="user in users" :key="user.id">
          <td class="px-6 py-4 text-sm">{{ user.id }}</td>
          <td class="px-6 py-4 text-sm font-medium">{{ user.username }}</td>
          <td class="px-6 py-4 text-sm">
            <span
              class="px-2 py-1 rounded text-xs font-medium"
              :class="user.role === 'admin' ? 'bg-purple-100 text-purple-800' : 'bg-blue-100 text-blue-800'"
            >
              {{ user.role }}
            </span>
          </td>
          <td class="px-6 py-4 text-sm font-mono">{{ user.min_ext }} - {{ user.max_ext }}</td>
          <td class="px-6 py-4 text-sm">
            <span
              class="px-2 py-1 rounded text-xs font-medium"
              :class="user.call_log_access ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-500'"
            >
              {{ user.call_log_access ? 'Yes' : 'No' }}
            </span>
          </td>
          <td class="px-6 py-4 text-sm space-x-3">
            <button @click="$emit('edit', user)" class="text-blue-600 hover:text-blue-800 font-medium">Edit</button>
            <button @click="$emit('delete', user.id)" class="text-red-600 hover:text-red-800 font-medium">Delete</button>
          </td>
        </tr>
        <tr v-if="users.length === 0">
          <td colspan="6" class="px-6 py-8 text-center text-gray-500">No users</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
interface User {
  id: number
  username: string
  role: string
  min_ext: number
  max_ext: number
  call_log_access: boolean
}

defineProps<{ users: User[] }>()
defineEmits<{
  edit: [user: User]
  delete: [id: number]
}>()
</script>
