<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRuntimeConfig } from '#app'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line, Bar } from 'vue-chartjs'
import StatCard from '~/components/StatCard.vue'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

definePageMeta({
  layout: 'admin'
})

const seoTitle = 'Analytics | Admin Mautrade'
const seoDescription = 'Detailed insights and reports for Mautrade administration.'
useSeoMeta({
  title: seoTitle,
  description: seoDescription,
  ogTitle: seoTitle,
  ogDescription: seoDescription,
  twitterTitle: seoTitle,
  twitterDescription: seoDescription
})

interface SignupChartData {
  date: string
  count: number
}

interface CountryDemographicData {
  countryCode: string
  count: number
}

interface AdminAnalyticsResponse {
  totalRevenue: string | number
  totalUsers: number
  activeUsers: number
  transactions: number
  depositGasFeeTracker: string | number
  recentSignups: number
  signupsChartData: SignupChartData[]
  countryDemographics: CountryDemographicData[]
}

const loading = ref(true)

const { tokenCookie } = useAdminAuth()

const analyticsStats = ref({
  totalRevenue: 0,
  totalUsers: 0,
  activeUsers: 0,
  transactions: 0,
  depositGasFeeTracker: 0,
  recentSignups: 0
})

// Chart data
const trafficData = ref({
  labels: [] as string[],
  datasets: [
    {
      label: 'Signups',
      backgroundColor: 'rgba(255, 90, 0, 0.2)',
      borderColor: '#ff5a00',
      data: [] as number[],
      fill: true,
      tension: 0.4
    }
  ]
})

const demographicData = ref({
  labels: [] as string[],
  datasets: [
    {
      label: 'Users by Country',
      backgroundColor: '#ff5a00',
      borderRadius: 4,
      data: [] as number[]
    }
  ]
})

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      labels: { color: '#888' }
    }
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

onMounted(async () => {
  try {
    const config = useRuntimeConfig()
    const apiBase = config.public.apiBase
    const data = await $fetch<AdminAnalyticsResponse>(`${apiBase}/admin/analytics`, {
      headers: { Authorization: `Bearer ${tokenCookie.value}` }
    })

    analyticsStats.value.totalRevenue = typeof data.totalRevenue === 'string' ? parseFloat(data.totalRevenue) : (data.totalRevenue || 0)
    analyticsStats.value.totalUsers = data.totalUsers || 0
    analyticsStats.value.activeUsers = data.activeUsers || 0
    analyticsStats.value.transactions = data.transactions || 0
    analyticsStats.value.depositGasFeeTracker = typeof data.depositGasFeeTracker === 'string' ? parseFloat(data.depositGasFeeTracker) : (data.depositGasFeeTracker || 0)
    analyticsStats.value.recentSignups = data.recentSignups || 0

    if (data.signupsChartData && data.signupsChartData.length > 0) {
      trafficData.value.labels = data.signupsChartData.map((d: SignupChartData) => d.date)
      if (trafficData.value.datasets[0]) {
        trafficData.value.datasets[0].data = data.signupsChartData.map((d: SignupChartData) => d.count)
      }
    }

    if (data.countryDemographics && data.countryDemographics.length > 0) {
      demographicData.value.labels = data.countryDemographics.map((d: CountryDemographicData) => d.countryCode)
      if (demographicData.value.datasets[0]) {
        demographicData.value.datasets[0].data = data.countryDemographics.map((d: CountryDemographicData) => d.count)
      }
    }
  } catch (err) {
    console.error('Failed to load admin analytics:', err)
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
          v-for="n in 6"
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
          Analytics
        </h1>
        <p class="page-subtitle">
          Detailed insights and reports
        </p>
      </header>

      <div class="stats-grid">
        <StatCard
          title="Total Revenue"
          :value="`$${analyticsStats.totalRevenue.toLocaleString()}`"
        />
        <StatCard
          title="Total Users"
          :value="analyticsStats.totalUsers.toLocaleString()"
        />
        <StatCard
          title="Active Users"
          :value="analyticsStats.activeUsers.toLocaleString()"
        />
        <StatCard
          title="Transactions"
          :value="analyticsStats.transactions.toLocaleString()"
        />
        <StatCard
          title="Deposit Gas Fee Tracker"
          :value="`$${analyticsStats.depositGasFeeTracker.toLocaleString()}`"
        />
        <StatCard
          title="Recent Signups"
          :value="analyticsStats.recentSignups.toLocaleString()"
        />
      </div>

      <div class="charts-section">
        <div class="chart-container traffic-chart">
          <h2 class="section-title">
            Traffic Analytics
          </h2>
          <p class="section-desc">
            Signups over time
          </p>
          <div class="chart-wrapper">
            <Line
              :data="trafficData"
              :options="chartOptions"
            />
          </div>
        </div>

        <div class="chart-container map-chart">
          <h2 class="section-title">
            Global Countries Reach
          </h2>
          <p class="section-desc">
            Demographics & heatmaps
          </p>
          <div class="chart-wrapper">
            <Bar
              :data="demographicData"
              :options="chartOptions"
            />
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
  grid-template-columns: repeat(3, 1fr);
  gap: 1.5rem;
}

.charts-section {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.5rem;
}

.chart-container {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
}

.section-title {
  font-family: 'Oswald', sans-serif;
  font-size: 1.25rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  color: var(--text);
  margin-bottom: 0.25rem;
}

.section-desc {
  color: var(--text-mute);
  font-size: 0.9rem;
  margin-bottom: 1.5rem;
}

.chart-wrapper {
  height: 350px;
  width: 100%;
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
  grid-template-columns: repeat(3, 1fr);
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
}

@media (max-width: 640px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
