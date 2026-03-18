<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="flex items-center justify-between mb-8">
      <h1 class="text-2xl font-bold text-gray-900">Call Log</h1>
      <button
        @click="refresh"
        class="bg-gray-600 hover:bg-gray-700 text-white text-sm px-4 py-2 rounded"
      >
        Refresh
      </button>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8" v-if="stats">
      <div class="bg-white rounded-lg shadow p-6">
        <p class="text-sm font-medium text-gray-500 uppercase">Total Calls</p>
        <p class="text-3xl font-bold text-gray-900 mt-1">{{ stats.total_calls }}</p>
      </div>
      <div class="bg-white rounded-lg shadow p-6">
        <p class="text-sm font-medium text-gray-500 uppercase">Answered</p>
        <p class="text-3xl font-bold text-green-600 mt-1">{{ stats.answered }}</p>
      </div>
      <div class="bg-white rounded-lg shadow p-6">
        <p class="text-sm font-medium text-gray-500 uppercase">Avg Duration</p>
        <p class="text-3xl font-bold text-blue-600 mt-1">{{ formatDuration(stats.avg_duration) }}</p>
      </div>
      <div class="bg-white rounded-lg shadow p-6">
        <p class="text-sm font-medium text-gray-500 uppercase">Answer Rate</p>
        <p class="text-3xl font-bold text-gray-900 mt-1">{{ stats.answer_rate }}%</p>
      </div>
    </div>

    <!-- Top 5 Busiest Days -->
    <div class="bg-white rounded-lg shadow p-6 mb-8" v-if="stats && topDays.length > 0">
      <h3 class="text-sm font-medium text-gray-500 uppercase mb-4">Top 5 Busiest Days</h3>
      <div class="grid grid-cols-1 sm:grid-cols-5 gap-4">
        <div v-for="(day, idx) in topDays" :key="day.date" class="flex items-center gap-3">
          <span class="text-lg font-bold text-gray-400 w-6 text-right">#{{ idx + 1 }}</span>
          <div>
            <p class="text-sm font-semibold text-gray-900">{{ day.date }}</p>
            <p class="text-sm text-gray-500">{{ day.count }} calls</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Charts -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8" v-if="stats">
      <div class="bg-white rounded-lg shadow p-6">
        <h3 class="text-sm font-medium text-gray-500 uppercase mb-4">Calls per Day</h3>
        <Bar :data="callsPerDayData" :options="barOptions" />
      </div>
      <div class="bg-white rounded-lg shadow p-6">
        <h3 class="text-sm font-medium text-gray-500 uppercase mb-4">Call Status Breakdown</h3>
        <Doughnut :data="dispositionData" :options="doughnutOptions" />
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white rounded-lg shadow overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Time</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">From</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">To</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Caller ID</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Duration</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="(call, idx) in cdrRecords" :key="idx">
            <td class="px-6 py-4 text-sm text-gray-500">{{ call.start }}</td>
            <td class="px-6 py-4 text-sm font-mono font-semibold">{{ call.source }}</td>
            <td class="px-6 py-4 text-sm font-mono font-semibold">{{ call.destination }}</td>
            <td class="px-6 py-4 text-sm">{{ call.callerid }}</td>
            <td class="px-6 py-4 text-sm">{{ formatDuration(call.duration) }}</td>
            <td class="px-6 py-4 text-sm">
              <span
                class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
                :class="dispositionCssClass(call.disposition)"
              >
                {{ call.disposition }}
              </span>
            </td>
          </tr>
          <tr v-if="cdrRecords.length === 0 && !loading">
            <td colspan="6" class="px-6 py-8 text-center text-gray-500">No call records found</td>
          </tr>
        </tbody>
      </table>
      <div v-if="loading" class="px-6 py-4 text-center text-gray-500 text-sm">Loading...</div>
      <div v-else-if="!hasMore && cdrRecords.length > 0" class="px-6 py-4 text-center text-gray-400 text-sm">All {{ totalRecords }} records loaded</div>
    </div>
    <div ref="scrollSentinel" class="h-1"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Bar, Doughnut } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  ArcElement,
  Tooltip,
  Legend,
} from 'chart.js'
import api from '../api/client'

ChartJS.register(CategoryScale, LinearScale, BarElement, ArcElement, Tooltip, Legend)

interface CDRRecord {
  source: string
  destination: string
  callerid: string
  start: string
  answer: string
  end: string
  duration: number
  billsec: number
  disposition: string
  channel: string
  dst_channel: string
}

interface CDRStatsData {
  total_calls: number
  answered: number
  avg_duration: number
  answer_rate: number
  calls_per_day: Record<string, number>
  dispositions: Record<string, number>
}

const PAGE_SIZE = 100
const cdrRecords = ref<CDRRecord[]>([])
const stats = ref<CDRStatsData | null>(null)
const loading = ref(false)
const hasMore = ref(true)
const totalRecords = ref(0)
const scrollSentinel = ref<HTMLElement | null>(null)
let observer: IntersectionObserver | null = null

async function fetchPage() {
  if (loading.value || !hasMore.value) return
  loading.value = true
  const offset = cdrRecords.value.length
  const res = await api.get(`/cdr?offset=${offset}&limit=${PAGE_SIZE}`)
  const page = res.data
  cdrRecords.value.push(...page.records)
  hasMore.value = page.has_more
  totalRecords.value = page.total
  loading.value = false
}

async function fetchStats() {
  const res = await api.get('/cdr/stats')
  stats.value = res.data
}

async function refresh() {
  cdrRecords.value = []
  hasMore.value = true
  stats.value = null
  fetchPage()
  fetchStats()
}

const topDays = computed(() => {
  if (!stats.value) return []
  const cpd = stats.value.calls_per_day
  return Object.entries(cpd)
    .map(([date, count]) => ({ date, count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 5)
})

const callsPerDayData = computed(() => {
  if (!stats.value) return { labels: [], datasets: [] }
  const cpd = stats.value.calls_per_day
  const days = Object.keys(cpd).sort()
  return {
    labels: days,
    datasets: [
      {
        label: 'Calls',
        data: days.map((d) => cpd[d]),
        backgroundColor: 'rgba(59, 130, 246, 0.6)',
        borderColor: 'rgb(59, 130, 246)',
        borderWidth: 1,
      },
    ],
  }
})

const dispositionData = computed(() => {
  if (!stats.value) return { labels: [], datasets: [] }
  const disp = stats.value.dispositions
  const labels = Object.keys(disp)
  const colorMap: Record<string, string> = {
    ANSWERED: 'rgba(34, 197, 94, 0.7)',
    'NO ANSWER': 'rgba(234, 179, 8, 0.7)',
    BUSY: 'rgba(249, 115, 22, 0.7)',
    FAILED: 'rgba(239, 68, 68, 0.7)',
  }
  return {
    labels,
    datasets: [
      {
        data: labels.map((l) => disp[l]),
        backgroundColor: labels.map((l) => colorMap[l] || 'rgba(156, 163, 175, 0.7)'),
      },
    ],
  }
})

const barOptions = {
  responsive: true,
  plugins: { legend: { display: false } },
  scales: {
    y: { beginAtZero: true, ticks: { precision: 0 } },
  },
}

const doughnutOptions = {
  responsive: true,
  plugins: { legend: { position: 'bottom' as const } },
}

function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return m > 0 ? `${m}m ${s}s` : `${s}s`
}

function dispositionCssClass(disposition: string): string {
  switch (disposition) {
    case 'ANSWERED':
      return 'bg-green-100 text-green-800'
    case 'NO ANSWER':
      return 'bg-yellow-100 text-yellow-800'
    case 'BUSY':
      return 'bg-orange-100 text-orange-800'
    case 'FAILED':
      return 'bg-red-100 text-red-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

onMounted(() => {
  fetchPage()
  fetchStats()

  observer = new IntersectionObserver(
    (entries) => {
      if (entries[0].isIntersecting) {
        fetchPage()
      }
    },
    { rootMargin: '200px' }
  )
  if (scrollSentinel.value) {
    observer.observe(scrollSentinel.value)
  }
})

onUnmounted(() => {
  observer?.disconnect()
})
</script>
