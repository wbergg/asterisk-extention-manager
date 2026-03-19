import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../api/client'

export interface Extension {
  id: number
  extension: number
  user_id: number
  sip_username: string
  sip_password: string
  callerid: string
  context: string
  directory_only: boolean
  created_at: string
}

export const useExtensionsStore = defineStore('extensions', () => {
  const extensions = ref<Extension[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchExtensions() {
    loading.value = true
    error.value = null
    try {
      const res = await api.get('/extensions')
      extensions.value = res.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch extensions'
    } finally {
      loading.value = false
    }
  }

  async function createExtension(ext: number, callerid: string, sip_password?: string) {
    error.value = null
    try {
      const res = await api.post('/extensions', { extension: ext, callerid, sip_password: sip_password || '' })
      extensions.value.push(res.data)
      return res.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to create extension'
      throw e
    }
  }

  async function updateExtension(ext: number, callerid: string, sip_password?: string) {
    error.value = null
    try {
      const res = await api.put(`/extensions/${ext}`, { callerid, sip_password: sip_password || '' })
      const idx = extensions.value.findIndex((e) => e.extension === ext)
      if (idx !== -1) extensions.value[idx] = res.data
      return res.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to update extension'
      throw e
    }
  }

  async function deleteExtension(ext: number) {
    error.value = null
    try {
      await api.delete(`/extensions/${ext}`)
      extensions.value = extensions.value.filter((e) => e.extension !== ext)
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to delete extension'
      throw e
    }
  }

  return { extensions, loading, error, fetchExtensions, createExtension, updateExtension, deleteExtension }
})
