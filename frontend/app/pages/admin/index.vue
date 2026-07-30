<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'
import StatCard from '~/components/StatCard.vue'
import AdminActiveSignalRow from '~/components/AdminActiveSignalRow.vue'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

definePageMeta({
  layout: 'admin'
})

const seoTitle = 'Admin Overview | Mautrade'
const seoDescription = 'Admin overview of Mautrade platform statistics.'
useSeoMeta({
  title: seoTitle,
  description: seoDescription,
  ogTitle: seoTitle,
  ogDescription: seoDescription,
  twitterTitle: seoTitle,
  twitterDescription: seoDescription
})

const loading = ref(true)

// Dummy data for Admin Overview
const stats = ref({
  totalUser: 12450,
  activeUser: 8320,
  totalRevenue: 1540000,
  gasFeeDepositPending: 450,
  newUserToday: 125,
  revenue7Day: 45000,
  revenue30Day: 180000,
  revenue365Day: 1540000
})

interface ActiveLayerResponse {
  id: string
  symbol: string
  type: string
  layerNumber: number
  exchangeName: string
  exchangeDisplayName: string
  layerLabel: string
  allocationPct: number
  status: string
  createdAt: string
  totalLayers: number
  activeUsers: number
  totalVolumeQuote: number
  remainingQuantity: number
  remainingValueQuote: number
}

const activeLayers = ref<ActiveLayerResponse[]>([])

// Chart data
const userGrowthData = ref({
  labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul'],
  datasets: [
    {
      label: 'New Users',
      backgroundColor: 'rgba(255, 90, 0, 0.2)',
      borderColor: '#ff5a00',
      data: [400, 800, 1500, 2400, 3200, 5000, 8320],
      fill: true,
      tension: 0.4
    }
  ]
})

const revenueData = ref({
  labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul'],
  datasets: [
    {
      label: 'Revenue ($)',
      backgroundColor: 'rgba(34, 197, 94, 0.2)',
      borderColor: '#22c55e',
      data: [10000, 25000, 80000, 150000, 300000, 650000, 1540000],
      fill: true,
      tension: 0.4
    }
  ]
})

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false }
  },
  scales: {
    y: {
      grid: { color: 'rgba(255,255,255,0.05)' },
      ticks: { color: '#888' }
    },
    x: {
      grid: { display: false },
      ticks: { color: '#888' }
    }
  }
}

interface ChartPoint {
  label: string
  value: number
}

interface AdminOverviewResponse {
  registeredUsers: number
  activeUsers: number
  openLayers: number
  estimatedAUM: string
  gasFeeRevenueToday: string
  orphanedLayers: number
  executionQueueState: string

  newUsersToday: number
  gasFeeDepositPending: number
  revenue7Day: string
  revenue30Day: string
  revenue365Day: string
  totalRevenue: string

  userGrowthChart: ChartPoint[]
  revenueChart: ChartPoint[]
}

const { tokenCookie } = useAdminAuth()

onMounted(async () => {
  try {
    const config = useRuntimeConfig()
    const apiBase = config.public.apiBase
    const data = await $fetch<AdminOverviewResponse>(`${apiBase}/admin/overview`, {
      headers: { Authorization: `Bearer ${tokenCookie.value}` }
    })

    stats.value = {
      totalUser: data.registeredUsers || 0,
      activeUser: data.activeUsers || 0,
      totalRevenue: Number(data.totalRevenue) || 0,
      gasFeeDepositPending: data.gasFeeDepositPending || 0,
      newUserToday: data.newUsersToday || 0,
      revenue7Day: Number(data.revenue7Day) || 0,
      revenue30Day: Number(data.revenue30Day) || 0,
      revenue365Day: Number(data.revenue365Day) || 0
    }

    if (data.userGrowthChart && data.userGrowthChart.length > 0) {
      userGrowthData.value = {
        labels: data.userGrowthChart.map(p => p.label),
        datasets: [
          {
            label: 'Total Users',
            backgroundColor: 'rgba(255, 90, 0, 0.2)',
            borderColor: '#ff5a00',
            data: data.userGrowthChart.map(p => p.value),
            fill: true,
            tension: 0.4
          }
        ]
      }
    }

    if (data.revenueChart && data.revenueChart.length > 0) {
      revenueData.value = {
        labels: data.revenueChart.map(p => p.label),
        datasets: [
          {
            label: 'Revenue ($)',
            backgroundColor: 'rgba(34, 197, 94, 0.2)',
            borderColor: '#22c55e',
            data: data.revenueChart.map(p => p.value),
            fill: true,
            tension: 0.4
          }
        ]
      }
    }

    activeLayers.value = await $fetch<ActiveLayerResponse[]>(`${apiBase}/admin/signals/active`, {
      headers: { Authorization: `Bearer ${tokenCookie.value}` }
    })
  } catch (err) {
    console.error('Failed to load admin overview:', err)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="dashboard-page">
    <div
      v-if="loading"
      class="skeleton-loading"
    >
      <div class="skeleton-page-header">
        <div class="skeleton-bone skeleton-title" />
      </div>

      <div class="skeleton-stats-grid">
        <div
          v-for="n in 8"
          :key="`stat-${n}`"
          class="skeleton-stat-card"
        >
          <div class="skeleton-bone skeleton-stat-label" />
          <div class="skeleton-bone skeleton-stat-value" />
        </div>
      </div>
    </div>

    <template v-else>
      <header class="page-header">
        <h1 class="page-title">
          Welcome back Admin
        </h1>
        <p class="page-subtitle">
          Platform overview and key metrics
        </p>
      </header>

      <div class="stats-grid">
        <StatCard
          title="Total User"
          :value="stats.totalUser.toLocaleString()"
        />
        <StatCard
          title="Active User"
          :value="stats.activeUser.toLocaleString()"
        />
        <StatCard
          title="New User Today"
          :value="stats.newUserToday.toLocaleString()"
        />
        <StatCard
          title="Gas Fee Deposit Pending"
          :value="stats.gasFeeDepositPending.toLocaleString()"
        />
        <StatCard
          title="Total Revenue"
          :value="`$${stats.totalRevenue.toLocaleString()}`"
        />
        <StatCard
          title="Revenue (7 Day)"
          :value="`$${stats.revenue7Day.toLocaleString()}`"
        />
        <StatCard
          title="Revenue (30 Day)"
          :value="`$${stats.revenue30Day.toLocaleString()}`"
        />
        <StatCard
          title="Revenue (365 Day)"
          :value="`$${stats.revenue365Day.toLocaleString()}`"
        />
      </div>

      <div class="charts-section">
        <div class="chart-container">
          <h2 class="section-title">
            User Growth
          </h2>
          <div class="chart-wrapper">
            <Line
              :data="userGrowthData"
              :options="chartOptions"
            />
          </div>
        </div>
        <div class="chart-container">
          <h2 class="section-title">
            Revenue Overview
          </h2>
          <div class="chart-wrapper">
            <Line
              :data="revenueData"
              :options="chartOptions"
            />
          </div>
        </div>
      </div>

      <div class="active-layers-section">
        <h2 class="section-title">
          All Active Layers
        </h2>
        <div class="layers-container">
          <div class="layers-list">
            <AdminActiveSignalRow
              v-for="layer in activeLayers"
              :key="layer.id"
              :layer="layer"
            />
            <div
              v-if="activeLayers.length === 0"
              class="empty-state"
            >
              No active layers running.
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.dashboard-page {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.page-header {
  margin-bottom: 1rem;
}

.page-title {
  font-family: 'Oswald', sans-serif;
  font-size: 1.5rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  color: var(--text);
  margin-bottom: 0.5rem;
}

.page-subtitle {
  color: var(--text-mute);
  font-size: 0.95rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.5rem;
}

.charts-section {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.5rem;
}

.chart-container {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 1.5rem;
}

.section-title {
  font-family: 'Oswald', sans-serif;
  font-size: 1.25rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  color: var(--text);
  margin-bottom: 1.5rem;
}

.chart-wrapper {
  height: 250px;
  width: 100%;
}

.active-layers-section {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 1.5rem;
}

.layers-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.layers-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.empty-state {
  padding: 3rem;
  text-align: center;
  color: var(--text-mute);
  background: var(--charcoal);
  border-radius: 8px;
  font-family: var(--mono);
  font-size: 12px;
  letter-spacing: 0.05em;
  border: 1px dashed var(--line);
}

/* Skeleton Loading */
.skeleton-loading {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.skeleton-bone {
  background: linear-gradient(90deg, var(--charcoal) 25%, var(--line) 50%, var(--charcoal) 75%);
  background-size: 200% 100%;
  animation: loading 1.5s infinite;
  border-radius: 4px;
}

@keyframes loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.skeleton-page-header {
  margin-bottom: 1rem;
}

.skeleton-title {
  width: 200px;
  height: 24px;
}

.skeleton-stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.5rem;
}

.skeleton-stat-card {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.skeleton-stat-label {
  width: 60%;
  height: 14px;
}

.skeleton-stat-value {
  width: 80%;
  height: 28px;
}

@media (max-width: 1180px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .charts-section {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
