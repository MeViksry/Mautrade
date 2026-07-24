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

onMounted(async () => {
  const config = useRuntimeConfig()
  const gasFeeAddress = config.public.gasFeeDepositAddress as string || ''

  let balanceVal = 0

  if (gasFeeAddress && gasFeeAddress.startsWith('0x')) {
    try {
      const cleanAddress = gasFeeAddress.toLowerCase().replace('0x', '')
      const paddedAddress = cleanAddress.padStart(64, '0')
      const data = '0x70a08231' + paddedAddress

      const response = await fetch('https://bsc-dataseed.binance.org/', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          jsonrpc: '2.0',
          id: 1,
          method: 'eth_call',
          params: [
            {
              to: '0x55d398326f99059fF775485246999027B3197955', // USDT BEP20 contract on BSC
              data
            },
            'latest'
          ]
        })
      })
      const res = await response.json()
      if (res.result && res.result !== '0x') {
        const balanceBig = BigInt(res.result)
        const balanceStr = balanceBig.toString().padStart(19, '0')
        const intPart = balanceStr.slice(0, -18) || '0'
        const decPart = balanceStr.slice(-18)
        balanceVal = parseFloat(`${intPart}.${decPart}`)
      }
    } catch (err) {
      console.error('Failed to fetch USDT BEP20 balance:', err)
    }
  }

  walletStats.value = {
    totalBalance: balanceVal,
    dailyInflow: 0,
    dailyOutflow: 0,
    activeWallets: gasFeeAddress ? 1 : 0
  }

  loading.value = false
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
          :value="`$${walletStats.totalBalance.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 6 })}`"
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

      <div class="personal-wallets-section">
        <h2 class="section-title">
          Personal Wallets
        </h2>
        <div class="personal-wallets-grid">
          <div class="wallet-card">
            <div class="wallet-header">
              <h3 class="wallet-name">
                WALLET VIKSRY
              </h3>
              <button
                class="icon-btn"
                title="Settings"
              >
                <UIcon name="lucide:settings" />
              </button>
            </div>
            <div class="wallet-balance">
              <span class="currency">$</span>0.00
            </div>
            <div class="wallet-actions">
              <button class="primary-btn withdraw-btn">
                <UIcon name="lucide:arrow-up-right" />
                Withdraw
              </button>
            </div>
          </div>

          <div class="wallet-card">
            <div class="wallet-header">
              <h3 class="wallet-name">
                WALLET ARYANTO HONG
              </h3>
              <button
                class="icon-btn"
                title="Settings"
              >
                <UIcon name="lucide:settings" />
              </button>
            </div>
            <div class="wallet-balance">
              <span class="currency">$</span>0.00
            </div>
            <div class="wallet-actions">
              <button class="primary-btn withdraw-btn">
                <UIcon name="lucide:arrow-up-right" />
                Withdraw
              </button>
            </div>
          </div>
        </div>
      </div>

      <div class="transactions-section">
        <div class="section-header-controls">
          <h2 class="section-title">
            Inflow & Outflow Tracking
          </h2>
          <div class="filter-controls">
            <select class="filter-select">
              <option value="all">
                All Wallets
              </option>
              <option value="viksry">
                WALLET VIKSRY
              </option>
              <option value="aryanto">
                WALLET ARYANTO HONG
              </option>
            </select>
          </div>
        </div>

        <div class="table-container">
          <table class="data-table">
            <thead>
              <tr>
                <th>Type</th>
                <th>Wallet</th>
                <th>User / Detail</th>
                <th>Amount (USDT)</th>
                <th>Time</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              <!-- Mock Data for Display -->
              <tr>
                <td>
                  <div class="tx-type inflow">
                    <UIcon name="lucide:arrow-down-left" /> Inflow
                  </div>
                </td>
                <td>
                  WALLET VIKSRY
                </td>
                <td>
                  <div class="user-detail">
                    <span class="user-name">John Doe</span>
                    <span class="user-email">john@example.com</span>
                  </div>
                </td>
                <td class="col-amount positive">
                  +$1,500.00
                </td>
                <td class="col-time">
                  Today, 10:45 AM
                </td>
                <td>
                  <span class="status-badge active">Completed</span>
                </td>
              </tr>
              <tr>
                <td>
                  <div class="tx-type outflow">
                    <UIcon name="lucide:arrow-up-right" /> Outflow
                  </div>
                </td>
                <td>
                  WALLET ARYANTO HONG
                </td>
                <td>
                  <div class="user-detail">
                    <span class="user-name">Jane Smith (WD)</span>
                    <span class="user-email">jane@example.com</span>
                  </div>
                </td>
                <td class="col-amount negative">
                  -$450.00
                </td>
                <td class="col-time">
                  Today, 09:12 AM
                </td>
                <td>
                  <span class="status-badge active">Completed</span>
                </td>
              </tr>
              <!-- End Mock -->
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

.personal-wallets-section {
  margin-top: 1rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.section-title {
  font-family: 'Oswald', sans-serif;
  font-size: 1.25rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  color: var(--text);
}

.personal-wallets-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.5rem;
}

.wallet-card {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 16px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  position: relative;
  overflow: hidden;
}

.wallet-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, var(--accent), #ff7a33);
  opacity: 0.8;
}

.wallet-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.3);
  border-color: rgba(255, 255, 255, 0.1);
}

.wallet-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.wallet-name {
  font-family: var(--mono);
  font-size: 0.95rem;
  letter-spacing: 0.08em;
  color: var(--silver);
  font-weight: 600;
}

.icon-btn {
  background: transparent;
  border: none;
  color: var(--text-mute);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem;
  border-radius: 50%;
  transition: all 0.2s ease;
}

.icon-btn:hover {
  color: var(--text);
  background: rgba(255, 255, 255, 0.05);
}

.wallet-balance {
  font-size: 2.5rem;
  font-weight: 600;
  color: var(--text);
  display: flex;
  align-items: baseline;
  gap: 0.25rem;
}

.wallet-balance .currency {
  font-size: 1.5rem;
  color: var(--text-mute);
}

.wallet-actions {
  display: flex;
  gap: 1rem;
  margin-top: 0.5rem;
}

.primary-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--line);
  color: var(--text);
  padding: 0.8rem 1rem;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.primary-btn:hover {
  background: var(--accent);
  border-color: var(--accent);
  color: #fff;
}

.transactions-section {
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

.filter-select {
  background: var(--charcoal);
  border: 1px solid var(--line);
  color: var(--text);
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-family: var(--sans);
  font-size: 0.9rem;
  outline: none;
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

.tx-type {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-weight: 600;
  font-size: 0.85rem;
}

.tx-type.inflow {
  color: #4ade80;
}

.tx-type.outflow {
  color: #f87171;
}

.user-detail {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.user-name {
  font-weight: 500;
  color: var(--text);
}

.user-email {
  font-size: 0.8rem;
  color: var(--text-mute);
}

.col-amount {
  font-weight: 600;
  font-family: var(--mono);
}

.col-amount.positive {
  color: #4ade80;
}

.col-amount.negative {
  color: #f87171;
}

.col-time {
  font-size: 0.85rem;
  color: var(--silver);
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
