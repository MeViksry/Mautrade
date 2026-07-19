<script setup lang="ts">
import { ref, onMounted } from 'vue'
import ExchangeBindModal from '~/components/ExchangeBindModal.vue'
import ExchangeManageKeysModal from '~/components/ExchangeManageKeysModal.vue'
import { useDashboardData } from '~/composables/useDashboardData'

definePageMeta({
  layout: 'dashboard'
})

const seoTitle = 'API Keys | Mautrade Dashboard'
const seoDescription = 'Manage Mautrade exchange API keys, connection status, API deletion, and exchange credential verification.'

useSeoMeta({
  title: seoTitle,
  description: seoDescription,
  ogTitle: seoTitle,
  ogDescription: seoDescription,
  twitterTitle: seoTitle,
  twitterDescription: seoDescription
})

interface ExchangeBinding {
  id: number
  name: string
  logo: string
  logoDark?: string
  status: string
  lastSynced: string | null
  balance: number
  hasApi?: boolean
}

const { getExchangeBindings } = useDashboardData()
const exchanges = ref<ExchangeBinding[]>([])
const loading = ref(true)
const theme = useState<'dark' | 'light'>('dashboard-theme', () => 'dark')
const bindModalOpen = ref(false)
const manageModalOpen = ref(false)
const managedExchange = ref<ExchangeBinding | null>(null)
const googleAuthenticatorEnabled = ref(true)

onMounted(async () => {
  loading.value = true
  try {
    exchanges.value = await getExchangeBindings()
  } catch (error) {
    console.error('Error fetching exchange bindings:', error)
  } finally {
    loading.value = false
  }
})

const formatLastSynced = (lastSynced: string | null) => {
  return lastSynced ? new Date(lastSynced).toLocaleString() : 'Never'
}

const getExchangeLogo = (exchange: ExchangeBinding) => {
  return theme.value === 'dark' && exchange.logoDark ? exchange.logoDark : exchange.logo
}

const handleExchangeBindSubmitted = (payload: { exchange: string }) => {
  exchanges.value = exchanges.value.map((exchange) => {
    if (exchange.name !== payload.exchange) return exchange

    return {
      ...exchange,
      status: 'connected',
      lastSynced: new Date().toISOString(),
      hasApi: true
    }
  })
}

const handleDeleteApi = (exchangeId: number) => {
  exchanges.value = exchanges.value.map((exchange) => {
    if (exchange.id !== exchangeId) return exchange

    return {
      ...exchange,
      status: 'disconnected',
      lastSynced: null,
      balance: 0,
      hasApi: false
    }
  })

  if (managedExchange.value?.id === exchangeId) {
    manageModalOpen.value = false
    managedExchange.value = null
  }
}

const openManageKeys = (exchange: ExchangeBinding) => {
  if (exchange.hasApi === false) return

  managedExchange.value = exchange
  manageModalOpen.value = true
}

const handleExchangeStatusChange = (payload: { exchangeId: number, status: 'connected' | 'disconnected' }) => {
  exchanges.value = exchanges.value.map((exchange) => {
    if (exchange.id !== payload.exchangeId) return exchange

    return {
      ...exchange,
      status: payload.status,
      lastSynced: payload.status === 'connected' ? new Date().toISOString() : exchange.lastSynced,
      hasApi: true
    }
  })

  managedExchange.value = exchanges.value.find(exchange => exchange.id === payload.exchangeId) ?? null
}
</script>

<template>
  <div class="dashboard-page">
    <div
      v-if="loading"
      class="skeleton-loading"
    >
      <div class="skeleton-page-header">
        <div class="skeleton-bone skeleton-title" />
        <div class="skeleton-bone skeleton-header-btn" />
      </div>

      <div class="api-keys-grid">
        <div
          v-for="n in 3"
          :key="`skel-exc-${n}`"
          class="exchange-card skeleton-exchange-card"
        >
          <div class="exchange-card__header">
            <div class="skeleton-bone skeleton-exchange-logo" />
            <div class="skeleton-bone skeleton-exchange-status" />
          </div>
          <div class="exchange-card__body">
            <div class="exchange-stat">
              <div class="skeleton-bone skeleton-stat-label" />
              <div class="skeleton-bone skeleton-stat-val" />
            </div>
            <div class="exchange-stat">
              <div class="skeleton-bone skeleton-stat-label" />
              <div class="skeleton-bone skeleton-stat-val-time" />
            </div>
          </div>
          <div class="exchange-card__footer">
            <div class="skeleton-bone skeleton-btn-secondary" />
            <div class="skeleton-bone skeleton-btn-danger" />
          </div>
        </div>
      </div>
    </div>

    <template v-else>
      <div class="page-header">
        <h2 class="page-title">
          API Keys (Exchanges)
        </h2>
        <button
          class="btn-primary"
          type="button"
          @click="bindModalOpen = true"
        >
          + Bind New Exchange
        </button>
      </div>

      <div class="api-keys-grid">
        <div
          v-for="exchange in exchanges"
          :key="exchange.id"
          class="exchange-card"
        >
          <div class="exchange-card__header">
            <div class="exchange-logo-shell">
              <img
                class="exchange-logo"
                :src="getExchangeLogo(exchange)"
                :alt="`${exchange.name} logo`"
              >
            </div>
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

          <div class="exchange-card__footer">
            <button
              class="btn-secondary"
              type="button"
              :disabled="exchange.hasApi === false"
              @click="openManageKeys(exchange)"
            >
              Manage Keys
            </button>
            <button
              class="btn-danger"
              type="button"
              :disabled="exchange.hasApi === false"
              @click="handleDeleteApi(exchange.id)"
            >
              Delete Api
            </button>
          </div>
        </div>
      </div>
    </template>

    <ExchangeBindModal
      v-model="bindModalOpen"
      :exchanges="exchanges"
      :theme="theme"
      @submitted="handleExchangeBindSubmitted"
    />
    <ExchangeManageKeysModal
      v-model="manageModalOpen"
      :exchange="managedExchange"
      :theme="theme"
      :google-authenticator-enabled="googleAuthenticatorEnabled"
      @status-change="handleExchangeStatusChange"
    />
  </div>
</template>

<style scoped>
.dashboard-page {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
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

/* ─── Skeleton Loading ─── */
@keyframes shimmer {
  0% { background-position: -400px 0; }
  100% { background-position: 400px 0; }
}

.skeleton-loading {
  animation: skeletonFadeIn 0.4s ease-out;
  width: 100%;
  max-width: 100%;
  overflow: hidden;
  min-width: 0;
}

@keyframes skeletonFadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

.skeleton-bone {
  background: linear-gradient(
    90deg,
    rgba(255, 255, 255, 0.04) 0%,
    rgba(255, 255, 255, 0.08) 20%,
    rgba(255, 138, 76, 0.12) 40%,
    rgba(255, 138, 76, 0.18) 50%,
    rgba(255, 138, 76, 0.12) 60%,
    rgba(255, 255, 255, 0.08) 80%,
    rgba(255, 255, 255, 0.04) 100%
  );
  background-size: 800px 100%;
  animation: shimmer 1.8s ease-in-out infinite;
  border-radius: 4px;
}

.skeleton-page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.skeleton-title { width: 280px; height: 28px; }
.skeleton-header-btn { width: 160px; height: 38px; border-radius: 0; }
.skeleton-exchange-logo { width: 120px; height: 40px; }
.skeleton-exchange-status { width: 65px; height: 20px; border-radius: 20px; }
.skeleton-stat-label { width: 60px; height: 11px; }
.skeleton-stat-val { width: 50px; height: 14px; }
.skeleton-stat-val-time { width: 120px; height: 11px; }
.skeleton-btn-secondary { width: 85px; height: 26px; border-radius: 0; }
.skeleton-btn-danger { width: 75px; height: 26px; border-radius: 0; }

.api-keys-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 1.5rem;
}

.exchange-card {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  padding: 1.5rem;
  transition: border-color 300ms var(--ease-quiet);
  display: flex;
  flex-direction: column;
}
.exchange-card:hover:not(:has(.exchange-card__footer button:hover)) {
  border-color: var(--accent);
}

.exchange-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.exchange-logo-shell {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 150px;
  height: 52px;
}

.exchange-logo {
  display: block;
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.exchange-status {
  font-family: var(--mono);
  font-size: 10px;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  padding: 0.4rem 0.8rem;
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
  gap: 1.5rem;
  margin-bottom: 2rem;
  flex: 1;
}

.exchange-stat {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 0.5rem;
  border-bottom: 1px dashed var(--line);
}

.exchange-stat__label {
  font-family: var(--mono);
  font-size: 11px;
  color: var(--text-mute);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.exchange-stat__val {
  font-family: var(--mono);
  font-size: 14px;
  color: var(--text);
  font-weight: 500;
}

.exchange-stat__val-time {
  font-family: var(--mono);
  font-size: 11px;
  color: var(--text);
}

.exchange-card__footer {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.btn-primary {
  background: var(--accent);
  color: #000;
  border: none;
  padding: 0.75rem 1.5rem;
  font-family: 'Oswald', sans-serif;
  font-size: 14px;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  cursor: pointer;
  transition: all 0.3s ease;
}
.btn-primary:hover {
  background: #ff7324;
}

.btn-secondary,
.btn-danger {
  background: transparent;
  color: var(--text);
  border: 1px solid var(--line);
  padding: 0.5rem 1rem;
  font-family: var(--mono);
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  cursor: pointer;
  transition: all 0.3s ease;
}
.btn-secondary:hover:not(:disabled) {
  border-color: var(--accent);
  color: var(--accent);
}

.btn-danger {
  color: #ef4444;
  border-color: rgba(239, 68, 68, 0.35);
}

.btn-danger:hover:not(:disabled) {
  border-color: #ef4444;
  background: rgba(239, 68, 68, 0.08);
  color: #ff6b6b;
}

.btn-secondary:disabled,
.btn-danger:disabled {
  cursor: not-allowed;
  opacity: 0.35;
}

@media (max-width: 640px) {
  .dashboard-page { gap: 0.75rem; }
  .page-header { margin-bottom: 0.5rem; flex-direction: column; align-items: flex-start; gap: 0.8rem; }
  .page-title { font-size: 1.3rem; }

  .skeleton-page-header { margin-bottom: 0.5rem; flex-direction: column; align-items: flex-start; gap: 0.8rem; }
  .skeleton-title { width: 180px; height: 22px; }

  .api-keys-grid {
    grid-template-columns: 1fr;
    gap: 0.75rem;
  }
  .exchange-card {
    padding: 1.25rem;
  }
  .exchange-card__header {
    margin-bottom: 1.25rem;
  }
  .exchange-card__body {
    margin-bottom: 1.5rem;
    gap: 1rem;
  }
}

@media (max-width: 380px) {
  .dashboard-page { gap: 0.5rem; }
  .page-title { font-size: 1.15rem; }
  .btn-primary { padding: 0.5rem 1rem; font-size: 12px; }
  .exchange-card {
    padding: 1rem;
  }
  .exchange-card__footer {
    gap: 0.5rem;
  }
}
</style>
