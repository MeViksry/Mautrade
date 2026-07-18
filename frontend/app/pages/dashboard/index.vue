<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
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
  logo: string
  logoDark?: string
  status: string
  lastSynced: string | null
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
const theme = useState<'dark' | 'light'>('dashboard-theme', () => 'dark')
const layersContainer = ref<HTMLElement | null>(null)
const exchangeListHeight = ref<number | null>(null)
const activeLayerPage = ref(1)
const activeLayersPerPage = 6
let layersResizeObserver: ResizeObserver | null = null

const syncExchangeListHeight = () => {
  if (!layersContainer.value) return
  exchangeListHeight.value = Math.round(layersContainer.value.getBoundingClientRect().height)
}

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
    await nextTick()
    syncExchangeListHeight()

    if (layersContainer.value && 'ResizeObserver' in window) {
      layersResizeObserver = new ResizeObserver(syncExchangeListHeight)
      layersResizeObserver.observe(layersContainer.value)
    }

    window.addEventListener('resize', syncExchangeListHeight)
  }
})

onBeforeUnmount(() => {
  layersResizeObserver?.disconnect()
  window.removeEventListener('resize', syncExchangeListHeight)
})

const formatLastSynced = (lastSynced: string | null) => {
  return lastSynced ? new Date(lastSynced).toLocaleString() : 'Never'
}

const getExchangeLogo = (exchange: ExchangeBinding) => {
  return theme.value === 'dark' && exchange.logoDark ? exchange.logoDark : exchange.logo
}

const totalActiveLayerPages = computed(() => Math.max(1, Math.ceil(layers.value.length / activeLayersPerPage)))

const visibleActiveLayers = computed(() => {
  const start = (activeLayerPage.value - 1) * activeLayersPerPage
  return layers.value.slice(start, start + activeLayersPerPage)
})

const activeLayerPages = computed(() => {
  return Array.from({ length: totalActiveLayerPages.value }, (_, index) => index + 1)
})

const exchangeListStyle = computed(() => {
  return exchangeListHeight.value ? { height: `${exchangeListHeight.value}px` } : undefined
})

const setActiveLayerPage = async (page: number) => {
  activeLayerPage.value = Math.min(Math.max(page, 1), totalActiveLayerPages.value)
  await nextTick()
  syncExchangeListHeight()
}

const goToPreviousActiveLayerPage = () => {
  void setActiveLayerPage(activeLayerPage.value - 1)
}

const goToNextActiveLayerPage = () => {
  void setActiveLayerPage(activeLayerPage.value + 1)
}
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
          title="Gas Fee Balance"
          :value="stats.totalGasFeePaid.toLocaleString()"
          unit="USDT"
          action-label="Deposit"
          action-icon="lucide:plus"
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

          <div
            ref="layersContainer"
            class="layers-container"
          >
            <LayerRow
              v-for="layer in visibleActiveLayers"
              :key="layer.id"
              :layer="layer"
            />
            <div
              v-if="layers.length === 0"
              class="empty-state"
            >
              No active layers. Waiting for Master Signal.
            </div>

            <div
              v-if="layers.length > activeLayersPerPage"
              class="layer-pagination"
              aria-label="Active layers pagination"
            >
              <button
                class="layer-pagination__nav"
                type="button"
                :disabled="activeLayerPage === 1"
                aria-label="Previous active layers page"
                @click="goToPreviousActiveLayerPage"
              >
                <svg
                  width="100%"
                  height="100%"
                  viewBox="0 0 24 24"
                  fill="none"
                  xmlns="http://www.w3.org/2000/svg"
                  aria-hidden="true"
                >
                  <path
                    d="M20.6621 17C18.933 19.989 15.7013 22 11.9999 22C6.47703 22 1.99988 17.5228 1.99988 12C1.99988 6.47715 6.47703 2 11.9999 2C15.7013 2 18.933 4.01099 20.6621 7M11.9999 8L7.99995 12M7.99995 12L11.9999 16M7.99995 12H21.9999"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                </svg>
              </button>

              <button
                v-for="page in activeLayerPages"
                :key="page"
                class="layer-pagination__page"
                :class="{ 'is-active': activeLayerPage === page }"
                type="button"
                :aria-current="activeLayerPage === page ? 'page' : undefined"
                @click="setActiveLayerPage(page)"
              >
                {{ page }}
              </button>

              <button
                class="layer-pagination__nav"
                type="button"
                :disabled="activeLayerPage === totalActiveLayerPages"
                aria-label="Next active layers page"
                @click="goToNextActiveLayerPage"
              >
                <svg
                  width="100%"
                  height="100%"
                  viewBox="0 0 24 24"
                  fill="none"
                  xmlns="http://www.w3.org/2000/svg"
                  aria-hidden="true"
                >
                  <path
                    d="M3.33789 7C5.06694 4.01099 8.29866 2 12.0001 2C17.5229 2 22.0001 6.47715 22.0001 12C22.0001 17.5228 17.5229 22 12.0001 22C8.29866 22 5.06694 19.989 3.33789 17M12 16L16 12M16 12L12 8M16 12H2"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                </svg>
              </button>
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

          <div
            class="exchange-list"
            :style="exchangeListStyle"
          >
            <div
              v-for="exchange in exchanges"
              :key="exchange.id"
              class="exchange-card"
            >
              <div class="exchange-card__header">
                <img
                  class="exchange-logo"
                  :src="getExchangeLogo(exchange)"
                  :alt="`${exchange.name} logo`"
                >
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
                  <span class="exchange-stat__val-time">{{ formatLastSynced(exchange.lastSynced) }}</span>
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
  grid-template-rows: auto 1fr;
  column-gap: 2rem;
  row-gap: 2rem;
  margin-top: 2.5rem;
  align-items: stretch;
}

.main-column,
.side-column {
  display: contents;
}

.main-column > .section-header {
  grid-column: 1;
  grid-row: 1;
}

.side-column > .section-header {
  grid-column: 2;
  grid-row: 1;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  margin-bottom: 0;
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
  grid-column: 1;
  grid-row: 2;
  align-self: stretch;
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

.layer-pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.65rem;
  min-height: 42px;
  border-top: 1px solid var(--line);
  background: var(--bg-elevated);
}

.layer-pagination__nav,
.layer-pagination__page {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--line);
  background: var(--bg-elevated);
  color: var(--text);
  cursor: pointer;
  transition: border-color 220ms var(--ease-quiet), color 220ms var(--ease-quiet), background 220ms var(--ease-quiet);
}

.layer-pagination__nav {
  width: 28px;
  height: 28px;
  padding: 0.35rem;
}

.layer-pagination__nav:hover:not(:disabled) {
  border-color: var(--accent);
  color: var(--accent);
}

.layer-pagination__nav:disabled {
  cursor: not-allowed;
  opacity: 0.35;
}

.layer-pagination__page {
  width: 28px;
  height: 28px;
  padding: 0;
  font-family: var(--mono);
  font-size: 12px;
  font-weight: 700;
}

.layer-pagination__page:hover,
.layer-pagination__page.is-active {
  border-color: var(--accent);
  background: var(--accent);
  color: #000;
}

.exchange-list {
  display: grid;
  grid-column: 2;
  grid-row: 2;
  grid-template-rows: repeat(4, minmax(0, 1fr));
  align-self: stretch;
  gap: 1rem;
}

.exchange-card {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  padding: 1.5rem;
  transition: border-color 300ms var(--ease-quiet);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-height: 0;
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

.exchange-logo {
  display: block;
  width: 118px;
  height: 30px;
  object-fit: contain;
  object-position: left center;
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
