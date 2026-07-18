<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useDashboardData } from '~/composables/useDashboardData'

definePageMeta({
  layout: 'dashboard'
})

interface ExchangeBinding {
  id: number
  name: string
  logo: string
  status: string
  lastSynced: string | null
  balance: number
}

const { getExchangeBindings } = useDashboardData()
const exchanges = ref<ExchangeBinding[]>([])
const loading = ref(true)

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
</script>

<template>
  <div class="dashboard-page">
    <div class="page-header">
      <h2 class="page-title">
        API Keys (Exchanges)
      </h2>
      <button class="btn-primary">
        + Bind New Exchange
      </button>
    </div>

    <div
      v-if="loading"
      class="loading-state"
    >
      Loading API keys...
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
              :src="exchange.logo"
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
          <button class="btn-secondary">
            Manage Keys
          </button>
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

.loading-state {
  font-family: var(--mono);
  color: var(--text-mute);
  padding: 4rem;
  text-align: center;
}

.api-keys-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
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
.exchange-card:hover {
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

.btn-secondary {
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
.btn-secondary:hover {
  border-color: var(--text);
}
</style>
