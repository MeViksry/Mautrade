<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
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
const depositModalOpen = ref(false)
const depositStep = ref<'methods' | 'deposit'>('methods')
const depositAmount = ref(500)
const depositTxId = ref('')
const walletCopied = ref(false)
const depositAmountShake = ref(false)
const depositTxIdShake = ref(false)
const depositSubmitAttempted = ref(false)
const depositSubmitted = ref(false)
const depositWalletAddress = '0x8F34B7C59A5D4E21F6C789DAB0132E45C67F9012'
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

const openDepositModal = () => {
  depositStep.value = 'methods'
  depositModalOpen.value = true
}

const closeDepositModal = () => {
  depositModalOpen.value = false
}

const selectDepositMethod = () => {
  depositStep.value = 'deposit'
  depositSubmitAttempted.value = false
  depositSubmitted.value = false
}

const copyDepositAddress = async () => {
  await navigator.clipboard.writeText(depositWalletAddress)
  walletCopied.value = true
  window.setTimeout(() => {
    walletCopied.value = false
  }, 1400)
}

const depositQrSvg = computed(() => {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="240" height="240" viewBox="0 0 240 240"><rect width="240" height="240" fill="#f8fafc"/><rect x="20" y="20" width="54" height="54" fill="#111"/><rect x="32" y="32" width="30" height="30" fill="#f8fafc"/><rect x="42" y="42" width="10" height="10" fill="#111"/><rect x="166" y="20" width="54" height="54" fill="#111"/><rect x="178" y="32" width="30" height="30" fill="#f8fafc"/><rect x="188" y="42" width="10" height="10" fill="#111"/><rect x="20" y="166" width="54" height="54" fill="#111"/><rect x="32" y="178" width="30" height="30" fill="#f8fafc"/><rect x="42" y="188" width="10" height="10" fill="#111"/><rect x="92" y="24" width="12" height="12" fill="#111"/><rect x="116" y="24" width="24" height="12" fill="#111"/><rect x="92" y="48" width="36" height="12" fill="#111"/><rect x="140" y="48" width="12" height="12" fill="#111"/><rect x="92" y="84" width="12" height="24" fill="#111"/><rect x="116" y="84" width="12" height="12" fill="#111"/><rect x="152" y="84" width="36" height="12" fill="#111"/><rect x="200" y="96" width="12" height="24" fill="#111"/><rect x="84" y="120" width="24" height="12" fill="#111"/><rect x="120" y="120" width="12" height="36" fill="#111"/><rect x="144" y="120" width="24" height="12" fill="#111"/><rect x="180" y="132" width="36" height="12" fill="#111"/><rect x="88" y="164" width="12" height="12" fill="#111"/><rect x="112" y="164" width="48" height="12" fill="#111"/><rect x="184" y="164" width="12" height="12" fill="#111"/><rect x="92" y="188" width="24" height="12" fill="#111"/><rect x="140" y="188" width="12" height="24" fill="#111"/><rect x="164" y="188" width="48" height="12" fill="#111"/><rect x="104" y="212" width="12" height="12" fill="#111"/><rect x="128" y="212" width="36" height="12" fill="#111"/><rect x="188" y="212" width="12" height="12" fill="#111"/></svg>`
  return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`
})

const depositAmountInvalid = computed(() => Number(depositAmount.value) < 500)
const depositTxIdInvalid = computed(() => depositTxId.value.trim().length === 0)
const depositTxIdErrorVisible = computed(() => depositSubmitAttempted.value && depositTxIdInvalid.value)
const depositFormBlocked = computed(() => depositAmountInvalid.value || depositTxIdInvalid.value)

const triggerDepositAmountShake = () => {
  depositAmountShake.value = false
  window.requestAnimationFrame(() => {
    depositAmountShake.value = true
  })
}

const triggerDepositTxIdShake = () => {
  depositTxIdShake.value = false
  window.requestAnimationFrame(() => {
    depositTxIdShake.value = true
  })
}

watch(depositAmount, () => {
  depositSubmitted.value = false
  if (depositAmountInvalid.value) {
    triggerDepositAmountShake()
  }
})

watch(depositTxId, () => {
  depositSubmitted.value = false
})

const submitDeposit = () => {
  depositSubmitAttempted.value = true

  if (depositAmountInvalid.value) {
    triggerDepositAmountShake()
  }

  if (depositTxIdInvalid.value) {
    triggerDepositTxIdShake()
  }

  if (depositFormBlocked.value) return

  depositSubmitted.value = true
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
          @action="openDepositModal"
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

    <div
      v-if="depositModalOpen"
      class="deposit-modal"
      role="dialog"
      aria-modal="true"
      :aria-label="depositStep === 'methods' ? 'Deposit Methods' : 'Deposit'"
      @click.self="closeDepositModal"
    >
      <div class="deposit-modal__box">
        <div class="deposit-modal__header">
          <button
            v-if="depositStep === 'deposit'"
            class="deposit-modal__icon-btn"
            type="button"
            aria-label="Back to deposit methods"
            @click="depositStep = 'methods'"
          >
            <UIcon name="lucide:arrow-left" />
          </button>
          <span
            v-else
            class="deposit-modal__spacer"
          />
          <h3>{{ depositStep === 'methods' ? 'Deposit Methods' : 'Deposit' }}</h3>
          <button
            class="deposit-modal__icon-btn"
            type="button"
            aria-label="Close deposit modal"
            @click="closeDepositModal"
          >
            <UIcon name="lucide:x" />
          </button>
        </div>

        <div
          v-if="depositStep === 'methods'"
          class="deposit-methods"
        >
          <button
            class="deposit-method"
            type="button"
            @click="selectDepositMethod"
          >
            <span class="deposit-method__icon">
              <UIcon name="lucide:wallet" />
            </span>
            <span class="deposit-method__content">
              <span class="deposit-method__title">USDT Gas Fee Wallet</span>
              <span class="deposit-method__meta">Minimum deposit 500 USDT</span>
            </span>
            <UIcon
              name="lucide:chevron-right"
              class="deposit-method__arrow"
            />
          </button>
        </div>

        <div
          v-else
          class="deposit-form"
        >
          <div class="deposit-qr">
            <img
              :src="depositQrSvg"
              alt="USDT gas fee deposit QR code"
            >
          </div>

          <a
            class="deposit-download"
            :href="depositQrSvg"
            download="mautrade-gas-fee-deposit-qr.svg"
          >
            <UIcon name="lucide:download" />
            <span>Download QR Code</span>
          </a>

          <label class="deposit-field">
            <span>Wallet Address</span>
            <div class="deposit-copy">
              <input
                :value="depositWalletAddress"
                readonly
              >
              <button
                type="button"
                @click="copyDepositAddress"
              >
                <UIcon :name="walletCopied ? 'lucide:check' : 'lucide:copy'" />
                <span>{{ walletCopied ? 'Copied' : 'Copy' }}</span>
              </button>
            </div>
          </label>

          <label class="deposit-field">
            <span>Amount</span>
            <div
              class="deposit-amount"
              :class="{ 'is-invalid': depositAmountInvalid, 'is-shaking': depositAmountShake }"
              @animationend="depositAmountShake = false"
            >
              <input
                v-model.number="depositAmount"
                type="number"
                min="500"
                step="1"
                aria-describedby="deposit-amount-error"
              >
              <span>USDT</span>
            </div>
            <p
              v-if="depositAmountInvalid"
              id="deposit-amount-error"
              class="deposit-error"
            >
              Minimum deposit is 500 USDT
            </p>
          </label>

          <label class="deposit-field">
            <span>TX ID</span>
            <input
              v-model="depositTxId"
              class="deposit-tx-input"
              :class="{ 'is-invalid': depositTxIdErrorVisible, 'is-shaking': depositTxIdShake }"
              type="text"
              placeholder="Paste transaction ID"
              aria-describedby="deposit-tx-error"
              @animationend="depositTxIdShake = false"
            >
            <p
              v-if="depositTxIdErrorVisible"
              id="deposit-tx-error"
              class="deposit-error"
            >
              TX ID is required
            </p>
          </label>

          <button
            class="deposit-submit"
            :class="{ 'is-blocked': depositFormBlocked }"
            type="button"
            :aria-disabled="depositFormBlocked"
            @click="submitDeposit"
          >
            <UIcon name="lucide:send" />
            <span>Submit Deposit</span>
          </button>

          <p
            v-if="depositSubmitted"
            class="deposit-success"
          >
            Deposit submitted
          </p>
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

.deposit-modal {
  position: fixed;
  inset: 0;
  z-index: 80;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  background: rgba(0, 0, 0, 0.72);
  backdrop-filter: blur(10px);
}

.deposit-modal__box {
  width: min(520px, 100%);
  max-height: min(760px, calc(100vh - 4rem));
  overflow-y: auto;
  scrollbar-width: none;
  -ms-overflow-style: none;
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  box-shadow: 0 28px 70px rgba(0, 0, 0, 0.45);
}

.deposit-modal__box::-webkit-scrollbar {
  display: none;
}

.deposit-modal__header {
  display: grid;
  grid-template-columns: 36px 1fr 36px;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--line);
}

.deposit-modal__header h3 {
  margin: 0;
  font-family: 'Oswald', sans-serif;
  font-size: 1.45rem;
  font-weight: 400;
  color: var(--text);
  letter-spacing: 0.04em;
  text-align: center;
  text-transform: uppercase;
}

.deposit-modal__icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text);
  cursor: pointer;
  transition: border-color 220ms var(--ease-quiet), color 220ms var(--ease-quiet);
}

.deposit-modal__spacer {
  width: 36px;
  height: 36px;
}

.deposit-modal__icon-btn:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.deposit-methods,
.deposit-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.5rem;
}

.deposit-method {
  display: grid;
  grid-template-columns: 44px 1fr 24px;
  align-items: center;
  gap: 1rem;
  width: 100%;
  padding: 1rem;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text);
  cursor: pointer;
  text-align: left;
  transition: border-color 220ms var(--ease-quiet), background 220ms var(--ease-quiet);
}

.deposit-method:hover {
  border-color: var(--accent);
  background: rgba(255, 90, 0, 0.08);
}

.deposit-method__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  background: var(--accent);
  color: #000;
}

.deposit-method__content {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.deposit-method__title {
  font-family: 'Oswald', sans-serif;
  font-size: 1.1rem;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.deposit-method__meta {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.deposit-method__arrow {
  color: var(--accent);
}

.deposit-qr {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  border: 1px solid var(--line);
  background: var(--charcoal);
}

.deposit-qr img {
  display: block;
  width: 220px;
  height: 220px;
  object-fit: contain;
}

.deposit-download,
.deposit-copy button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border: 1px solid var(--accent);
  background: var(--accent);
  color: #000;
  font-family: var(--mono);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  cursor: pointer;
  transition: background 220ms var(--ease-quiet), border-color 220ms var(--ease-quiet);
}

.deposit-download {
  align-self: center;
  padding: 0.65rem 1rem;
}

.deposit-download:hover,
.deposit-copy button:hover {
  background: #ff7324;
  border-color: #ff7324;
}

.deposit-field {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.deposit-field > span {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.deposit-field input {
  width: 100%;
  min-width: 0;
  height: 42px;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text);
  font-family: var(--mono);
  font-size: 12px;
  outline: none;
  padding: 0 0.85rem;
}

.deposit-field input:focus {
  border-color: var(--accent);
}

.deposit-copy,
.deposit-amount {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
}

.deposit-copy button {
  height: 42px;
  padding: 0 0.85rem;
  border-left: none;
}

.deposit-amount span {
  display: inline-flex;
  align-items: center;
  height: 42px;
  padding: 0 0.85rem;
  border: 1px solid var(--line);
  border-left: none;
  background: var(--charcoal);
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 11px;
}

.deposit-amount.is-invalid input,
.deposit-amount.is-invalid span {
  border-color: #ef4444;
}

.deposit-tx-input.is-invalid {
  border-color: #ef4444;
}

.deposit-amount.is-shaking,
.deposit-tx-input.is-shaking {
  animation: deposit-shake 260ms ease-in-out;
}

.deposit-error,
.deposit-success {
  margin: -0.2rem 0 0;
  font-family: var(--mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.deposit-error {
  color: #ef4444;
}

.deposit-success {
  color: #10b981;
  text-align: center;
}

.deposit-submit {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  width: 100%;
  height: 44px;
  border: 1px solid var(--accent);
  background: var(--accent);
  color: #000;
  font-family: var(--mono);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  cursor: pointer;
  transition: background 220ms var(--ease-quiet), border-color 220ms var(--ease-quiet), transform 220ms var(--ease-quiet);
}

.deposit-submit:hover {
  background: #ff7324;
  border-color: #ff7324;
  transform: translateY(-1px);
}

.deposit-submit.is-blocked {
  box-shadow: inset 0 0 0 1px rgba(239, 68, 68, 0.45);
}

@keyframes deposit-shake {
  0%, 100% {
    transform: translateX(0);
  }

  20% {
    transform: translateX(-7px);
  }

  40% {
    transform: translateX(7px);
  }

  60% {
    transform: translateX(-5px);
  }

  80% {
    transform: translateX(5px);
  }
}

@media (max-width: 640px) {
  .deposit-modal {
    padding: 1rem;
  }

  .deposit-copy {
    grid-template-columns: 1fr;
    gap: 0.5rem;
  }

  .deposit-copy button {
    border-left: 1px solid var(--accent);
  }
}
</style>
