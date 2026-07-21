<script setup lang="ts">
import { ref, onMounted } from 'vue'
import StatCard from '~/components/StatCard.vue'

definePageMeta({
  layout: 'admin'
})

const seoTitle = 'Deposit Management | Admin Mautrade'
const seoDescription = 'Verify and manage user deposits.'
useSeoMeta({
  title: seoTitle,
  description: seoDescription
})

const loading = ref(true)

interface DepositResponse {
  id: string
  userId: string
  amount: string
  asset: string
  txId: string
  status: string
  createdAt: string
  confirmedAt?: string
}

interface FormattedDeposit {
  id: string
  userId: string
  amount: number
  txId: string
  date: string
  status: string
  fullId: string
}

const { tokenCookie } = useAdminAuth()

const depositsStats = ref({
  totalDeposits: 0,
  pending: 0,
  verified: 0,
  rejected: 0
})

const deposits = ref<FormattedDeposit[]>([])

const loadDeposits = async () => {
  try {
    const config = useRuntimeConfig()
    const apiBase = config.public.apiBase
    const data = await $fetch<DepositResponse[]>(`${apiBase}/admin/gas-fee/deposits`, {
      headers: { Authorization: `Bearer ${tokenCookie.value}` }
    })

    deposits.value = data.map(d => ({
      fullId: d.id,
      id: d.id.split('-')[0] ?? '',
      userId: d.userId.split('-')[0] ?? '',
      amount: Number(d.amount),
      txId: d.txId || 'N/A',
      date: new Date(d.createdAt).toISOString().split('T')[0] ?? '',
      status: d.status.charAt(0).toUpperCase() + d.status.slice(1)
    }))

    depositsStats.value = {
      totalDeposits: deposits.value.reduce((sum, d) => sum + d.amount, 0),
      pending: deposits.value.filter(d => d.status === 'Pending').length,
      verified: deposits.value.filter(d => d.status === 'Confirmed').length,
      rejected: deposits.value.filter(d => d.status === 'Rejected').length
    }
  } catch (err) {
    console.error('Failed to load deposits', err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadDeposits()
})

const handleVerify = async (id: string) => {
  const deposit = deposits.value.find(d => d.id === id)
  if (!deposit) return

  try {
    const config = useRuntimeConfig()
    const apiBase = config.public.apiBase
    await $fetch(`${apiBase}/admin/gas-fee/deposits/${deposit.fullId}/status`, {
      method: 'PATCH',
      headers: { Authorization: `Bearer ${tokenCookie.value}` },
      body: { status: 'confirmed' }
    })
    deposit.status = 'Confirmed'
    depositsStats.value.pending--
    depositsStats.value.verified++
  } catch (err) {
    console.error('Failed to verify deposit', err)
    alert('Failed to verify deposit')
  }
}

const handleReject = async (id: string) => {
  const deposit = deposits.value.find(d => d.id === id)
  if (!deposit) return

  try {
    const config = useRuntimeConfig()
    const apiBase = config.public.apiBase
    await $fetch(`${apiBase}/admin/gas-fee/deposits/${deposit.fullId}/status`, {
      method: 'PATCH',
      headers: { Authorization: `Bearer ${tokenCookie.value}` },
      body: { status: 'rejected' }
    })
    deposit.status = 'Rejected'
    depositsStats.value.pending--
    depositsStats.value.rejected++
  } catch (err) {
    console.error('Failed to reject deposit', err)
    alert('Failed to reject deposit')
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
          Deposit Management
        </h1>
        <p class="page-subtitle">
          Verify and manage user gas fee deposits
        </p>
      </header>

      <div class="stats-grid">
        <StatCard
          title="Total Deposits"
          :value="`$${depositsStats.totalDeposits.toLocaleString()}`"
        />
        <StatCard
          title="Pending Verification"
          :value="`$${depositsStats.pending.toLocaleString()}`"
        />
        <StatCard
          title="Verified Deposits"
          :value="`$${depositsStats.verified.toLocaleString()}`"
        />
        <StatCard
          title="Rejected Deposits"
          :value="`$${depositsStats.rejected.toLocaleString()}`"
        />
      </div>

      <div class="table-section">
        <div class="section-header-controls">
          <h2 class="section-title">
            Deposit Requests
          </h2>
          <div class="filters">
            <select class="filter-select">
              <option value="all">
                All Status
              </option>
              <option value="pending">
                Pending
              </option>
              <option value="verified">
                Verified
              </option>
              <option value="rejected">
                Rejected
              </option>
            </select>
          </div>
        </div>

        <div class="table-container">
          <table class="data-table">
            <thead>
              <tr>
                <th>Req ID</th>
                <th>User ID</th>
                <th>Amount (USDT)</th>
                <th>TxID (Hash)</th>
                <th>Date</th>
                <th>Status</th>
                <th class="actions-col">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="deposit in deposits"
                :key="deposit.id"
              >
                <td class="col-id">
                  {{ deposit.id }}
                </td>
                <td class="col-user">
                  #{{ deposit.userId }}
                </td>
                <td class="col-amount">
                  ${{ deposit.amount }}
                </td>
                <td class="col-txid">
                  <span class="tx-hash">{{ deposit.txId }}</span>
                </td>
                <td class="col-date">
                  {{ deposit.date }}
                </td>
                <td class="col-status">
                  <span :class="['status-badge', deposit.status.toLowerCase()]">
                    {{ deposit.status }}
                  </span>
                </td>
                <td class="col-actions">
                  <div
                    v-if="deposit.status === 'Pending'"
                    class="action-buttons"
                  >
                    <button
                      class="action-btn verify-btn"
                      @click="handleVerify(deposit.id)"
                    >
                      Verify
                    </button>
                    <button
                      class="action-btn reject-btn"
                      @click="handleReject(deposit.id)"
                    >
                      Reject
                    </button>
                  </div>
                  <div
                    v-else
                    class="action-buttons"
                  >
                    <span class="action-done">—</span>
                  </div>
                </td>
              </tr>
              <tr v-if="deposits.length === 0">
                <td
                  colspan="7"
                  class="empty-state"
                >
                  No deposit requests found.
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

.table-section {
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

.filter-select {
  background: var(--charcoal);
  border: 1px solid var(--line);
  color: var(--text);
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-size: 0.9rem;
  cursor: pointer;
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

.col-id, .col-user, .col-date {
  font-family: var(--mono);
  font-size: 0.9rem;
  color: var(--silver);
}

.col-amount {
  font-weight: 500;
  color: var(--text);
}

.tx-hash {
  font-family: var(--mono);
  font-size: 0.85rem;
  color: var(--accent);
  background: rgba(255, 90, 0, 0.1);
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
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

.status-badge.verified {
  background: rgba(34, 197, 94, 0.1);
  color: #4ade80;
}

.status-badge.pending {
  background: rgba(245, 158, 11, 0.1);
  color: #fbbf24;
}

.status-badge.rejected {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.actions-col, .col-actions {
  text-align: right;
}

.action-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.action-btn {
  background: transparent;
  border: 1px solid var(--line);
  padding: 0.4rem 0.8rem;
  border-radius: 4px;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 500;
}

.verify-btn {
  color: #4ade80;
  border-color: rgba(74, 222, 128, 0.3);
}

.verify-btn:hover {
  background: rgba(74, 222, 128, 0.1);
  border-color: #4ade80;
}

.reject-btn {
  color: #ef4444;
  border-color: rgba(239, 68, 68, 0.3);
}

.reject-btn:hover {
  background: rgba(239, 68, 68, 0.1);
  border-color: #ef4444;
}

.action-done {
  color: var(--text-mute);
  font-family: var(--mono);
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
}
</style>
