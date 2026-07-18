<script setup lang="ts">
import { ref, onMounted } from 'vue'
import StatCard from '~/components/StatCard.vue'
import LayerRow from '~/components/LayerRow.vue'
import { useDashboardData } from '~/composables/useDashboardData'

definePageMeta({
  layout: 'dashboard'
})

interface UserStats {
  totalBalance: number
  realizedProfit: number
  totalGasFeePaid: number
  activeLayersCount: number
}

interface ExchangeBinding {
  id: number
  name: string
  status: string
  lastSynced: string
  balance: number
}

interface Layer {
  id: string
  pair: string
  entryPrice: number
  currentPrice: number
  allocationPct: number
  allocatedUsdt: number
  unrealizedPnl: number
  unrealizedPnlPct: number
  openedAt: string
  status: string
}

const { getUserStats, getExchangeBindings, getActiveLayers } = useDashboardData()

const stats = ref<UserStats | null>(null)
const exchanges = ref<ExchangeBinding[]>([])
const layers = ref<Layer[]>([])
const loading = ref(true)

onMounted(async () => {
  loading.value = true
  try {
    const [statsData, exchangesData, layersData] = await Promise.all([
      getUserStats(),
      getExchangeBindings(),
      getActiveLayers()
    ])
    stats.value = statsData
    exchanges.value = exchangesData
    layers.value = layersData
  } catch (error) {
    console.error('Error fetching dashboard data:', error)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="dashboard-page">
    <div
      v-if="loading"
      class="loading-state"
    >
      Loading dashboard...
    </div>

    <div v-else-if="stats">
      <div class="page-header">
        <h2 class="page-title">
          Overview
        </h2>
      </div>

      <!-- Stats Row -->
      <div class="stats-grid">
        <StatCard
          title="Total Balance"
          :value="stats.totalBalance.toLocaleString()"
          unit="USDT"
          :trend="5.2"
        />
        <StatCard
          title="Active Layers"
          :value="stats.activeLayersCount"
        />
        <StatCard
          title="Realized Profit"
          :value="stats.realizedProfit.toLocaleString()"
          unit="USDT"
          :trend="12.4"
        />
        <StatCard
          title="Total Gas Fee Paid"
          :value="stats.totalGasFeePaid.toLocaleString()"
          unit="USDT"
        />
      </div>

      <div class="dashboard-grid">
        <!-- Main Column: Active Layers -->
        <div class="main-column">
          <div class="section-header">
            <h3>Active Layers</h3>
            <NuxtLink
              to="/dashboard/layers"
              class="view-all"
            >View All →</NuxtLink>
          </div>

          <div class="layers-container">
            <LayerRow
              v-for="layer in layers"
              :key="layer.id"
              :layer="layer"
            />
            <div
              v-if="layers.length === 0"
              class="empty-state"
            >
              No active layers. Waiting for Master Signal.
            </div>
          </div>
        </div>

        <!-- Sidebar Column: Exchange Status -->
        <div class="side-column">
          <div class="section-header">
            <h3>Exchange Bindings</h3>
            <NuxtLink
              to="/dashboard/api-keys"
              class="view-all"
            >Manage →</NuxtLink>
          </div>

          <div class="exchange-list">
            <div
              v-for="exchange in exchanges"
              :key="exchange.id"
              class="exchange-card"
            >
              <div class="exchange-card__header">
                <span class="exchange-name">{{ exchange.name }}</span>
                <span
                  class="exchange-status"
                  :class="exchange.status === 'connected' ? 'status-active' : 'status-inactive'"
                >
                  {{ exchange.status }}
                </span>
              </div>
              <div class="exchange-card__body">
                <div class="exchange-stat">
                  <span class="exchange-stat__label">Balance</span>
                  <span class="exchange-stat__val">${{ exchange.balance.toLocaleString() }}</span>
                </div>
                <div class="exchange-stat">
                  <span class="exchange-stat__label">Last Synced</span>
                  <span class="exchange-stat__val-time">{{ new Date(exchange.lastSynced).toLocaleString() }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard-page {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.page-header {
  margin-bottom: 2rem;
}

.page-title {
  font-family: 'Oswald', sans-serif;
  font-size: 2.5rem;
  font-weight: 300;
  text-transform: uppercase;
  color: var(--text);
  margin: 0;
  letter-spacing: 0.05em;
}

.loading-state {
  font-family: var(--mono);
  color: var(--text-mute);
  padding: 4rem;
  text-align: center;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.5rem;
}

.dashboard-grid {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 2rem;
  margin-top: 2.5rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.section-header h3 {
  font-family: 'Oswald', sans-serif;
  font-size: 1.5rem;
  font-weight: 400;
  text-transform: uppercase;
  color: var(--text);
  margin: 0;
  letter-spacing: 0.02em;
}

.view-all {
  font-family: var(--mono);
  font-size: 11px;
  color: var(--accent);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}
.view-all:hover {
  text-decoration: underline;
}

.layers-container {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--line);
}

.empty-state {
  padding: 3rem;
  text-align: center;
  font-family: var(--mono);
  font-size: 12px;
  color: var(--text-mute);
  background: var(--bg-elevated);
}

.exchange-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.exchange-card {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  padding: 1.5rem;
  transition: border-color 300ms var(--ease-quiet);
}
.exchange-card:hover {
  border-color: var(--accent);
}

.exchange-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.exchange-name {
  font-family: 'Oswald', sans-serif;
  font-size: 1.2rem;
  color: var(--text);
  letter-spacing: 0.05em;
}

.exchange-status {
  font-family: var(--mono);
  font-size: 9px;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  padding: 0.3rem 0.6rem;
  border-radius: 20px;
}
.status-active {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.2);
}
.status-inactive {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.2);
}

.exchange-card__body {
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
}

.exchange-stat {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.exchange-stat__label {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  text-transform: uppercase;
}

.exchange-stat__val {
  font-family: var(--mono);
  font-size: 12px;
  color: var(--text);
}

.exchange-stat__val-time {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--silver);
}
</style>
