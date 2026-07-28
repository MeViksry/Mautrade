<script setup lang="ts">
import { onMounted, ref } from 'vue'
import ExchangeBindModal from '~/components/ExchangeBindModal.vue'
import ExchangeManageKeysModal from '~/components/ExchangeManageKeysModal.vue'
import { useDashboardData } from '~/composables/useDashboardData'
import type { ExchangeBinding, ExchangeCredentialSummary } from '~/composables/useDashboardData'

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

const {
  getExchangeBindings,
  bindExchange,
  getExchangeBindingCredentials,
  updateExchangeBindingStatus,
  updateExchangeBindingAccountMode,
  deleteExchangeBinding
} = useDashboardData()
const exchanges = ref<ExchangeBinding[]>([])
const exchangeOptions = ref<ExchangeBinding[]>([])
const loading = ref(true)
const theme = useState<'dark' | 'light'>('dashboard-theme', () => 'dark')
const bindModalOpen = ref(false)
const manageModalOpen = ref(false)
const managedExchange = ref<ExchangeBinding | null>(null)
const managedCredentials = ref<ExchangeCredentialSummary | null>(null)
const pageError = ref('')
const bindSubmitting = ref(false)
const bindError = ref('')
const manageSubmitting = ref<string | null>(null)
const manageCredentialsLoading = ref(false)
const manageError = ref('')
const googleAuthenticatorEnabled = ref(true)

const getErrorMessage = (error: unknown, fallback: string) => {
  if (error instanceof Error) {
    return error.message
  }
  return fallback
}

const refreshExchangeBindings = async () => {
  loading.value = true
  pageError.value = ''
  try {
    const allExchanges = await getExchangeBindings({ includeUnbound: true })
    exchangeOptions.value = allExchanges
    exchanges.value = allExchanges.filter(exchange => exchange.hasApi)
  } catch (error) {
    console.error('Error fetching exchange bindings:', error)
    pageError.value = getErrorMessage(error, 'Failed to load exchange bindings')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await refreshExchangeBindings()
})

const formatLastSynced = (lastSynced: string | null) => {
  return lastSynced ? new Date(lastSynced).toLocaleString() : 'Never'
}

const getExchangeLogo = (exchange: ExchangeBinding) => {
  return theme.value === 'dark' && exchange.logoDark ? exchange.logoDark : exchange.logo
}

const openBindModal = () => {
  bindError.value = ''
  bindModalOpen.value = true
}

const handleExchangeBindSubmitted = async (payload: { exchange: string, accountMode: 'real' | 'demo', apiKey: string, apiSecret: string, extras: Record<string, string> }) => {
  bindSubmitting.value = true
  bindError.value = ''
  try {
    await bindExchange({
      exchange: payload.exchange,
      accountMode: payload.accountMode,
      apiKey: payload.apiKey,
      apiSecret: payload.apiSecret,
      passphrase: payload.extras.passphrase
    })
    await refreshExchangeBindings()
    bindModalOpen.value = false
  } catch (error) {
    bindError.value = getErrorMessage(error, 'Failed to bind exchange')
  } finally {
    bindSubmitting.value = false
  }
}

const handleDeleteApi = async (exchange: ExchangeBinding) => {
  if (exchange.hasApi === false || manageSubmitting.value) return

  manageSubmitting.value = exchange.exchange
  pageError.value = ''
  manageError.value = ''
  try {
    await deleteExchangeBinding(exchange.exchange)
    await refreshExchangeBindings()

    if (managedExchange.value?.exchange === exchange.exchange) {
      manageModalOpen.value = false
      managedExchange.value = null
      managedCredentials.value = null
    }
  } catch (error) {
    const message = getErrorMessage(error, 'Failed to delete exchange API')
    pageError.value = message
    manageError.value = message
  } finally {
    manageSubmitting.value = null
  }
}

const openManageKeys = async (exchange: ExchangeBinding) => {
  if (exchange.hasApi === false) return

  managedExchange.value = exchange
  managedCredentials.value = null
  manageError.value = ''
  manageModalOpen.value = true
  manageCredentialsLoading.value = true
  try {
    managedCredentials.value = await getExchangeBindingCredentials(exchange.exchange)
  } catch (error) {
    manageError.value = getErrorMessage(error, 'Failed to load exchange credential')
  } finally {
    manageCredentialsLoading.value = false
  }
}

const handleExchangeStatusChange = async (payload: { exchange: string, status: 'connected' | 'disconnected' }) => {
  if (manageSubmitting.value) return

  manageSubmitting.value = payload.exchange
  manageError.value = ''
  try {
    await updateExchangeBindingStatus(payload.exchange, payload.status)
    await refreshExchangeBindings()
    managedExchange.value = exchanges.value.find(exchange => exchange.exchange === payload.exchange) ?? null
    if (managedExchange.value?.hasApi) {
      managedCredentials.value = await getExchangeBindingCredentials(payload.exchange)
    }
    manageModalOpen.value = false
  } catch (error) {
    manageError.value = getErrorMessage(error, 'Failed to update exchange status')
  } finally {
    manageSubmitting.value = null
  }
}

const handleExchangeAccountModeChange = async (payload: { exchange: string, accountMode: 'real' | 'demo' }) => {
  if (manageSubmitting.value) return

  manageSubmitting.value = payload.exchange
  manageError.value = ''
  try {
    await updateExchangeBindingAccountMode(payload.exchange, payload.accountMode)
    await refreshExchangeBindings()
    managedExchange.value = exchanges.value.find(exchange => exchange.exchange === payload.exchange) ?? null
    if (managedExchange.value?.hasApi) {
      managedCredentials.value = await getExchangeBindingCredentials(payload.exchange)
    }
  } catch (error) {
    manageError.value = getErrorMessage(error, 'Failed to update account mode')
  } finally {
    manageSubmitting.value = null
  }
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
          @click="openBindModal"
        >
          + Bind New Exchange
        </button>
      </div>

      <p
        v-if="pageError"
        class="page-error"
      >
        {{ pageError }}
      </p>

      <div
        v-if="exchanges.length === 0"
        class="api-empty-state"
      >
        <div class="api-empty-state__icon">
          <UIcon name="lucide:key-round" />
        </div>
        <h3>No Exchange Connected</h3>
        <button
          class="btn-primary"
          type="button"
          @click="openBindModal"
        >
          + Bind New Exchange
        </button>
      </div>

      <div
        v-else
        class="api-keys-grid"
      >
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
              <span class="exchange-stat__label">Account</span>
              <span class="exchange-stat__val-time">{{ exchange.accountMode }}</span>
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
              :disabled="exchange.hasApi === false || manageSubmitting === exchange.exchange"
              @click="handleDeleteApi(exchange)"
            >
              {{ manageSubmitting === exchange.exchange ? 'Deleting...' : 'Delete Api' }}
            </button>
          </div>
        </div>
      </div>
    </template>

    <ExchangeBindModal
      v-model="bindModalOpen"
      :exchanges="exchangeOptions"
      :theme="theme"
      :submitting="bindSubmitting"
      :error-message="bindError"
      @submitted="handleExchangeBindSubmitted"
    />
    <ExchangeManageKeysModal
      v-model="manageModalOpen"
      :exchange="managedExchange"
      :theme="theme"
      :google-authenticator-enabled="googleAuthenticatorEnabled"
      :credentials="managedCredentials"
      :credentials-loading="manageCredentialsLoading"
      :submitting="Boolean(manageSubmitting)"
      :error-message="manageError"
      @status-change="handleExchangeStatusChange"
      @account-mode-change="handleExchangeAccountModeChange"
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

.page-error {
  margin: -0.35rem 0 0.35rem;
  color: #ef4444;
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
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

.api-empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 1rem;
  min-height: 280px;
  border: 1px solid var(--line);
  background: var(--bg-elevated);
}

.api-empty-state__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 46px;
  height: 46px;
  border: 1px solid rgba(255, 90, 0, 0.28);
  color: var(--accent);
}

.api-empty-state h3 {
  margin: 0;
  color: var(--text);
  font-family: var(--mono);
  font-size: 12px;
  letter-spacing: 0.14em;
  text-transform: uppercase;
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
