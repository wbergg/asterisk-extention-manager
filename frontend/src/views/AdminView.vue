<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-8">Administration</h1>

    <!-- Users Section -->
    <section class="mb-12">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-xl font-semibold text-gray-800">Users</h2>
        <button
          @click="showUserForm = true; editingUser = null"
          class="bg-blue-600 hover:bg-blue-700 text-white text-sm px-4 py-2 rounded"
        >
          Add User
        </button>
      </div>

      <UserForm
        v-if="showUserForm"
        :user="editingUser"
        @saved="handleUserSaved"
        @cancel="showUserForm = false"
      />

      <UserTable
        :users="users"
        @edit="handleEditUser"
        @delete="handleDeleteUser"
      />
    </section>

    <!-- Extensions Section -->
    <section class="mb-12">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-xl font-semibold text-gray-800">All Extensions</h2>
        <div class="flex gap-2">
          <button
            @click="showDirectoryForm = !showDirectoryForm"
            class="bg-green-600 hover:bg-green-700 text-white text-sm px-4 py-2 rounded"
          >
            Add Directory Entry
          </button>
          <button
            @click="handleForceSync"
            :disabled="syncing"
            class="bg-yellow-500 hover:bg-yellow-600 text-white text-sm px-4 py-2 rounded disabled:opacity-50"
          >
            {{ syncing ? 'Syncing...' : 'Force Sync' }}
          </button>
        </div>
      </div>

      <!-- Add Directory Entry Form -->
      <div v-if="showDirectoryForm" class="bg-white rounded-lg shadow p-6 mb-6 border border-green-200">
        <h3 class="text-lg font-medium text-gray-900 mb-1">Add Directory / Phonebook Entry</h3>
        <p class="text-sm text-gray-500 mb-4">For numbers already routed in extensions.conf (IVR, ring groups, etc.). No SIP credentials are generated and no pjsip config is written.</p>
        <form @submit.prevent="handleDirectoryCreate" class="space-y-4">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Extension Number</label>
              <input
                v-model.number="dirExt"
                type="number"
                required
                class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                placeholder="e.g. 200"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Display Name (Caller ID)</label>
              <input
                v-model="dirCallerID"
                type="text"
                required
                class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                placeholder="e.g. Reception"
              />
            </div>
          </div>
          <p v-if="dirError" class="text-red-600 text-sm">{{ dirError }}</p>
          <div class="flex gap-4">
            <button type="submit" class="bg-green-600 hover:bg-green-700 text-white font-medium py-2 px-6 rounded-md">Add Entry</button>
            <button type="button" @click="showDirectoryForm = false" class="text-gray-600 hover:text-gray-800 font-medium py-2 px-4">Cancel</button>
          </div>
        </form>
      </div>

      <div v-if="syncMessage" class="mb-4 px-4 py-3 rounded text-sm"
        :class="syncError ? 'bg-red-50 border border-red-200 text-red-700' : 'bg-green-50 border border-green-200 text-green-700'">
        {{ syncMessage }}
      </div>

      <!-- Edit Extension Form -->
      <div v-if="editingExt" class="bg-white rounded-lg shadow p-6 mb-6 border border-blue-200">
        <h3 class="text-lg font-medium text-gray-900 mb-4">Edit Extension {{ editingExt.extension }}</h3>
        <form @submit.prevent="handleExtSave" class="space-y-4">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Caller ID</label>
              <input
                v-model="editingExt.callerid"
                type="text"
                class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div v-if="!editingExt.directory_only">
              <label class="block text-sm font-medium text-gray-700 mb-1">SIP Password (leave blank to keep)</label>
              <div class="flex gap-2">
                <input
                  v-model="editingExtPassword"
                  type="text"
                  class="w-full min-w-0 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono"
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
          <div class="flex gap-4">
            <button type="submit" class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-md">Save</button>
            <button type="button" @click="editingExt = null" class="text-gray-600 hover:text-gray-800 font-medium py-2 px-4">Cancel</button>
          </div>
        </form>
      </div>

      <div class="bg-white rounded-lg shadow overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Extension</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">SIP Username</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">SIP Password</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Caller ID</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Created By</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-for="ext in allExtensions" :key="ext.id" :class="ext.directory_only ? 'bg-green-50' : ''">
              <td class="px-6 py-4 text-sm font-mono font-semibold">
                {{ ext.extension }}
                <span v-if="ext.directory_only" class="ml-2 inline-block px-1.5 py-0.5 text-xs font-medium bg-green-100 text-green-700 rounded">Directory</span>
              </td>
              <td class="px-6 py-4 text-sm font-mono text-gray-400">{{ ext.directory_only ? '—' : ext.sip_username }}</td>
              <td class="px-6 py-4 text-sm font-mono">
                <span v-if="ext.directory_only" class="text-gray-400">—</span>
                <span v-else-if="!visiblePasswords.has(ext.id)">
                  <button @click="visiblePasswords.add(ext.id)" class="text-blue-600 hover:text-blue-800 text-xs">Show</button>
                </span>
                <span v-else>
                  {{ ext.sip_password }}
                  <button @click="visiblePasswords.delete(ext.id)" class="text-gray-400 hover:text-gray-600 text-xs ml-2">Hide</button>
                </span>
              </td>
              <td class="px-6 py-4 text-sm">{{ ext.callerid || '-' }}</td>
              <td class="px-6 py-4 text-sm">{{ userMap[ext.user_id] || ext.user_id }}</td>
              <td class="px-6 py-4 text-sm text-gray-500">{{ new Date(ext.created_at).toLocaleString() }}</td>
              <td class="px-6 py-4 text-sm space-x-3">
                <button @click="handleEditExt(ext)" class="text-blue-600 hover:text-blue-800 font-medium">Edit</button>
                <button @click="handleDeleteExt(ext.extension)" class="text-red-600 hover:text-red-800 font-medium">Delete</button>
              </td>
            </tr>
            <tr v-if="allExtensions.length === 0">
              <td colspan="7" class="px-6 py-8 text-center text-gray-500">No extensions registered</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- Blocked Extensions Section -->
    <section class="mb-12">
      <h2 class="text-xl font-semibold text-gray-800 mb-4">Blocked Extensions</h2>

      <div class="bg-white rounded-lg shadow p-6 mb-4">
        <form @submit.prevent="handleBlock" class="grid grid-cols-1 sm:grid-cols-2 lg:flex lg:items-end gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Extension Number</label>
            <input
              v-model.number="blockExt"
              type="number"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="e.g. 1009"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Reason (optional)</label>
            <input
              v-model="blockReason"
              type="text"
              class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="e.g. Reserved for lobby phone"
            />
          </div>
          <button type="submit" class="w-full sm:w-auto bg-red-600 hover:bg-red-700 text-white font-medium py-2 px-6 rounded-md">Block</button>
        </form>
        <p v-if="blockError" class="mt-3 text-red-600 text-sm">{{ blockError }}</p>
      </div>

      <div class="bg-white rounded-lg shadow overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Extension</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Reason</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Blocked At</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-for="b in blockedExtensions" :key="b.id">
              <td class="px-6 py-4 text-sm font-mono font-semibold">{{ b.extension }}</td>
              <td class="px-6 py-4 text-sm text-gray-600">{{ b.reason || '-' }}</td>
              <td class="px-6 py-4 text-sm text-gray-500">{{ new Date(b.created_at).toLocaleString() }}</td>
              <td class="px-6 py-4 text-sm">
                <button @click="handleUnblock(b.extension)" class="text-green-600 hover:text-green-800 font-medium">Unblock</button>
              </td>
            </tr>
            <tr v-if="blockedExtensions.length === 0">
              <td colspan="4" class="px-6 py-8 text-center text-gray-500">No blocked extensions</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import api from '../api/client'
import UserTable from '../components/UserTable.vue'
import UserForm from '../components/UserForm.vue'
import type { Extension } from '../stores/extensions'

const visiblePasswords = reactive(new Set<number>())

interface User {
  id: number
  username: string
  role: string
  min_ext: number
  max_ext: number
  call_log_access: boolean
}

const users = ref<User[]>([])
const allExtensions = ref<Extension[]>([])
const showUserForm = ref(false)
const editingUser = ref<User | null>(null)
const editingExt = ref<Extension | null>(null)
const editingExtPassword = ref('')
const userMap = computed(() => {
  const m: Record<number, string> = {}
  for (const u of users.value) m[u.id] = u.username
  return m
})
const syncing = ref(false)
const syncMessage = ref('')
const syncError = ref(false)

interface BlockedExt {
  id: number
  extension: number
  reason: string
  created_at: string
}

const blockedExtensions = ref<BlockedExt[]>([])
const blockExt = ref<number | undefined>()
const blockReason = ref('')
const blockError = ref('')

const showDirectoryForm = ref(false)
const dirExt = ref<number | undefined>()
const dirCallerID = ref('')
const dirError = ref('')

async function fetchUsers() {
  const res = await api.get('/admin/users')
  users.value = res.data
}

async function fetchAllExtensions() {
  const res = await api.get('/admin/extensions')
  allExtensions.value = res.data
}

async function fetchBlocked() {
  const res = await api.get('/admin/blocked')
  blockedExtensions.value = res.data
}

async function handleBlock() {
  if (!blockExt.value) return
  blockError.value = ''
  try {
    await api.post('/admin/blocked', { extension: blockExt.value, reason: blockReason.value })
    blockExt.value = undefined
    blockReason.value = ''
    fetchBlocked()
  } catch (e: any) {
    blockError.value = e.response?.data?.error || 'Failed to block extension'
  }
}

async function handleDirectoryCreate() {
  if (!dirExt.value) return
  dirError.value = ''
  try {
    await api.post('/admin/extensions/directory', { extension: dirExt.value, callerid: dirCallerID.value })
    dirExt.value = undefined
    dirCallerID.value = ''
    showDirectoryForm.value = false
    fetchAllExtensions()
  } catch (e: any) {
    dirError.value = e.response?.data?.error || 'Failed to create directory entry'
  }
}

async function handleUnblock(ext: number) {
  if (confirm(`Unblock extension ${ext}?`)) {
    await api.delete(`/admin/blocked/${ext}`)
    fetchBlocked()
  }
}

onMounted(() => {
  fetchUsers()
  fetchAllExtensions()
  fetchBlocked()
})

function handleEditUser(user: User) {
  editingUser.value = { ...user }
  showUserForm.value = true
}

async function handleDeleteUser(id: number) {
  if (confirm('Delete this user? Their extensions will also be deleted.')) {
    await api.delete(`/admin/users/${id}`)
    fetchUsers()
    fetchAllExtensions()
  }
}

function handleUserSaved() {
  showUserForm.value = false
  editingUser.value = null
  fetchUsers()
}

function generateSipPassword() {
  const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
  const arr = new Uint8Array(16)
  crypto.getRandomValues(arr)
  editingExtPassword.value = Array.from(arr, (b) => chars[b % chars.length]).join('')
}

function handleEditExt(ext: Extension) {
  editingExt.value = { ...ext }
  editingExtPassword.value = ''
}

async function handleExtSave() {
  if (!editingExt.value) return
  await api.put(`/admin/extensions/${editingExt.value.extension}`, {
    callerid: editingExt.value.callerid,
    sip_password: editingExtPassword.value,
  })
  editingExt.value = null
  fetchAllExtensions()
}

async function handleDeleteExt(extNum: number) {
  if (confirm(`Delete extension ${extNum}?`)) {
    await api.delete(`/admin/extensions/${extNum}`)
    fetchAllExtensions()
  }
}

async function handleForceSync() {
  syncing.value = true
  syncMessage.value = ''
  try {
    await api.post('/admin/sync')
    syncMessage.value = 'Asterisk configuration synced successfully'
    syncError.value = false
  } catch (e: any) {
    syncMessage.value = e.response?.data?.error || 'Sync failed'
    syncError.value = true
  } finally {
    syncing.value = false
  }
}
</script>
