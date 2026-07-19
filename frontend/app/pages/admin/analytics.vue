<script setup lang="ts">
import { ref, onMounted } from 'vue'
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

const loading = ref(true)

const analyticsStats = ref({
  totalRevenue: 1540000,
  totalUsers: 12450,
  activeUsers: 8320,
  transactions: 45291,
  depositGasFeeTracker: 12850,
  recentSignups: 42
})

// Chart data
const trafficData = ref({
  labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
  datasets: [
    {
      label: 'Visitors',
      backgroundColor: 'rgba(56, 189, 248, 0.2)',
      borderColor: '#38bdf8',
      data: [1200, 1900, 1500, 2200, 2800, 3100, 4200],
      fill: true,
      tension: 0.4
    },
    {
      label: 'Signups',
      backgroundColor: 'rgba(255, 90, 0, 0.2)',
      borderColor: '#ff5a00',
      data: [42, 68, 55, 89, 120, 145, 198],
      fill: true,
      tension: 0.4
    }
  ]
})

const demographicData = ref({
  labels: ['USA', 'UK', 'Indonesia', 'India', 'Brazil', 'Germany'],
  datasets: [
    {
      label: 'Users by Country',
      backgroundColor: '#ff5a00',
      borderRadius: 4,
      data: [4500, 2100, 1800, 1200, 950, 600]
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

onMounted(() => {
  setTimeout(() => {
    loading.value = false
  }, 1000)
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
            Visitor and signups over time
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
