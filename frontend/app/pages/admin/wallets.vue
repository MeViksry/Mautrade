<script setup lang="ts">
import { ref, onMounted } from 'vue'
import StatCard from '~/components/StatCard.vue'

definePageMeta({
  layout: 'admin'
})

const seoTitle = 'Wallets Management | Admin Mautrade'
const seoDescription = 'Manage company wallets and view balances.'
useSeoMeta({
  title: seoTitle,
  description: seoDescription
})

const loading = ref(true)

const walletStats = ref({
  totalBalance: 0,
  dailyInflow: 0,
  dailyOutflow: 0,
  activeWallets: 0
})

interface CompanyWallet {
  id: string
  network: string
  address: string
  balance: number
  status: string
}

const companyWallets = ref<CompanyWallet[]>([])

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
          v-for="n in 4"
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
          Company Wallets
        </h1>
        <p class="page-subtitle">
          Manage internal wallets and balances
        </p>
      </header>

      <div class="stats-grid">
        <StatCard
          title="Total Balance"
          :value="`$${walletStats.totalBalance.toLocaleString()}`"
        />
        <StatCard
          title="Daily Inflow"
          :value="`+$${walletStats.dailyInflow.toLocaleString()}`"
        />
        <StatCard
          title="Daily Outflow"
          :value="`-$${walletStats.dailyOutflow.toLocaleString()}`"
        />
        <StatCard
          title="Active Wallets"
          :value="walletStats.activeWallets.toString()"
        />
      </div>

      <div class="wallets-section">
        <div class="section-header-controls">
          <h2 class="section-title">
            Wallet Addresses
          </h2>
          <button class="action-btn add-btn">
            <UIcon name="lucide:plus" />
            Add New Wallet
          </button>
        </div>

        <div class="table-container">
          <table class="data-table">
            <thead>
              <tr>
                <th>Wallet ID</th>
                <th>Network</th>
                <th>Address</th>
                <th>Balance (USDT)</th>
                <th>Status</th>
                <th class="actions-col">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="wallet in companyWallets"
                :key="wallet.id"
              >
                <td class="col-id">
                  {{ wallet.id }}
                </td>
                <td class="col-network">
                  <span class="network-badge">{{ wallet.network }}</span>
                </td>
                <td class="col-address">
                  <div class="address-wrap">
                    <span class="addr-text">{{ wallet.address }}</span>
                    <button
                      class="copy-btn"
                      title="Copy Address"
                    >
                      <UIcon name="lucide:copy" />
                    </button>
                  </div>
                </td>
                <td class="col-amount">
                  ${{ wallet.balance.toLocaleString() }}
                </td>
                <td class="col-status">
                  <span :class="['status-badge', wallet.status.toLowerCase()]">
                    {{ wallet.status }}
                  </span>
                </td>
                <td class="col-actions">
                  <button class="action-btn view-btn">
                    Manage
                  </button>
                </td>
              </tr>
              <tr v-if="companyWallets.length === 0">
                <td
                  colspan="6"
                  class="empty-state"
                >
                  No company wallets found.
                </td>
              </tr>
            </tbody>
          </table>
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

.wallets-section {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.section-header-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.section-title {
  font-family: 'Oswald', sans-serif;
  font-size: 1.25rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  color: var(--text);
}

.add-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--accent) !important;
  color: #fff !important;
  border: none !important;
}

.add-btn:hover {
  background: #ff7a33 !important;
}

.table-container {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.data-table th,
.data-table td {
  padding: 1rem;
  border-bottom: 1px solid var(--line);
}

.data-table th {
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--text-mute);
  font-weight: normal;
}

.data-table tbody tr:hover {
  background: rgba(255, 255, 255, 0.02);
}

.col-id {
  font-family: var(--mono);
  font-size: 0.9rem;
  color: var(--silver);
}

.network-badge {
  background: var(--charcoal);
  padding: 0.3rem 0.6rem;
  border-radius: 4px;
  font-family: var(--mono);
  font-size: 0.8rem;
  color: var(--text);
  border: 1px solid var(--line);
}

.address-wrap {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.addr-text {
  font-family: var(--mono);
  font-size: 0.9rem;
  color: var(--silver);
}

.copy-btn {
  background: transparent;
  border: none;
  color: var(--text-mute);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.2s;
}

.copy-btn:hover {
  color: var(--accent);
}

.col-amount {
  font-weight: 500;
  color: var(--text);
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.6rem;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.status-badge.active {
  background: rgba(34, 197, 94, 0.1);
  color: #4ade80;
}

.actions-col, .col-actions {
  text-align: right;
}

.action-btn {
  background: transparent;
  border: 1px solid var(--line);
  color: var(--silver);
  padding: 0.4rem 0.8rem;
  border-radius: 4px;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 500;
}

.action-btn:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.empty-state {
  text-align: center;
  color: var(--text-mute);
  padding: 3rem !important;
  font-style: italic;
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
}

@media (max-width: 640px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  .section-header-controls {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
}
</style>
