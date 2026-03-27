import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../api/client'

export interface FaxDestination {
  extension: number
  name: string
}

export interface FaxJob {
  id: number
  user_id: number
  subject: string
  destination_ext: number
  original_file: string
  tiff_file: string
  call_file: string
  status: string
  error_message: string
  created_at: string
}

export const useFaxStore = defineStore('fax', () => {
  const destinations = ref<FaxDestination[]>([])
  const jobs = ref<FaxJob[]>([])
  const loading = ref(false)
  const sending = ref(false)
  const error = ref<string | null>(null)

  async function fetchDestinations() {
    try {
      const res = await api.get('/fax/destinations')
      destinations.value = res.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch fax destinations'
    }
  }

  async function fetchJobs() {
    loading.value = true
    error.value = null
    try {
      const res = await api.get('/fax/jobs')
      jobs.value = res.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to fetch fax jobs'
    } finally {
      loading.value = false
    }
  }

  async function sendFax(subject: string, destination: number, file: File, message: string) {
    sending.value = true
    error.value = null
    try {
      const formData = new FormData()
      formData.append('subject', subject)
      formData.append('destination', String(destination))
      formData.append('file', file)
      if (message) formData.append('message', message)
      const res = await api.post('/fax/send', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      jobs.value.unshift(res.data)
      return res.data
    } catch (e: any) {
      error.value = e.response?.data?.error || 'Failed to send fax'
      throw e
    } finally {
      sending.value = false
    }
  }

  return { destinations, jobs, loading, sending, error, fetchDestinations, fetchJobs, sendFax }
})
