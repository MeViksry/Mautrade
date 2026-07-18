<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import GasFeeDepositModal from '~/components/GasFeeDepositModal.vue'
import StatCard from '~/components/StatCard.vue'
import { useDashboardData } from '~/composables/useDashboardData'

definePageMeta({
  layout: 'dashboard'
})

const seoTitle = 'Gas Fee | Mautrade Dashboard'
const seoDescription = 'Track Mautrade gas fee balance, deposits, rebates, trading gas fees, and gas fee metric trends.'

useSeoMeta({
  title: seoTitle,
  description: seoDescription,
  ogTitle: seoTitle,
  ogDescription: seoDescription,
  twitterTitle: seoTitle,
  twitterDescription: seoDescription
})

interface UserStats {
  totalBalance: number
  realizedProfit: number
  totalGasFeePaid: number
  activeLayersCount: number
}

interface TradeHistory {
  id: string
  pair: string
  exitPrice: number
  pnl: number
  gasFee: number
  closedAt: string
}

interface GasFeeHistoryEntry {
  id: string
  type: 'deposit' | 'fee' | 'rebate'
  title: string
  reference: string
  occurredAt: string
  change: number
  balanceAfter: number
}

interface GasFeeMetricPoint {
  x: number
  y: number
  value: number
  label: string
}

const { getUserStats, getHistory } = useDashboardData()

const stats = ref<UserStats | null>(null)
const historyItems = ref<TradeHistory[]>([])
const loading = ref(true)
const depositModalOpen = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const [statsData, historyData] = await Promise.all([
      getUserStats(),
      getHistory()
    ])
    stats.value = statsData
    historyItems.value = historyData
  } catch (error) {
    console.error('Error fetching gas fee data:', error)
  } finally {
    loading.value = false
  }
})

const formatCurrency = (value: number) => {
  return value.toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  })
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const gasFeeHistory = computed<GasFeeHistoryEntry[]>(() => {
  if (!stats.value) return []

  const depositEntries = [
    {
      id: 'GF-DEP-003',
      type: 'deposit' as const,
      title: 'Deposit',
      reference: 'USDT Gas Fee Wallet',
      occurredAt: '2026-07-18T10:30:00Z',
      change: 500
    },
    {
      id: 'GF-DEP-002',
      type: 'deposit' as const,
      title: 'Deposit',
      reference: 'USDT Gas Fee Wallet',
      occurredAt: '2026-07-12T08:15:00Z',
      change: 750
    },
    {
      id: 'GF-DEP-001',
      type: 'deposit' as const,
      title: 'Deposit',
      reference: 'USDT Gas Fee Wallet',
      occurredAt: '2026-07-06T07:00:00Z',
      change: 1000
    }
  ]

  const tradeEntries = historyItems.value.map((item) => {
    const isRebate = item.gasFee < 0

    return {
      id: item.id,
      type: isRebate ? 'rebate' as const : 'fee' as const,
      title: isRebate ? 'Gas Fee Rebate' : 'Trading Gas Fee',
      reference: item.pair,
      occurredAt: item.closedAt,
      change: isRebate ? Math.abs(item.gasFee) : -Math.abs(item.gasFee)
    }
  })

  const entries = [...depositEntries, ...tradeEntries].sort((a, b) => {
    return new Date(b.occurredAt).getTime() - new Date(a.occurredAt).getTime()
  })

  let runningBalance = stats.value.totalGasFeePaid

  return entries.map((entry) => {
    const balanceAfter = runningBalance
    runningBalance -= entry.change

    return {
      ...entry,
      balanceAfter
    }
  })
})

const totalGasFeeUsed = computed(() => {
  return historyItems.value.reduce((total, item) => total + Math.max(item.gasFee, 0), 0)
})

const totalGasFeeRebates = computed(() => {
  return historyItems.value.reduce((total, item) => total + Math.max(-item.gasFee, 0), 0)
})

const gasFeeMetricPoints = computed<GasFeeMetricPoint[]>(() => {
  const entries = [...gasFeeHistory.value].reverse()
  if (entries.length === 0) return []

  const width = 640
  const height = 170
  const paddingX = 24
  const paddingY = 18
  const values = entries.map((entry) => entry.balanceAfter)
  const minValue = Math.min(...values)
  const maxValue = Math.max(...values)
  const range = maxValue - minValue || 1

  return entries.map((entry, index) => {
    const x = entries.length === 1
      ? width / 2
      : paddingX + (index / (entries.length - 1)) * (width - paddingX * 2)
    const y = paddingY + ((maxValue - entry.balanceAfter) / range) * (height - paddingY * 2)

    return {
      x,
      y,
      value: entry.balanceAfter,
      label: entry.id
    }
  })
})

const gasFeeMetricLine = computed(() => {
  return gasFeeMetricPoints.value.map((point) => `${point.x},${point.y}`).join(' ')
})

const gasFeeMetricArea = computed(() => {
  const points = gasFeeMetricPoints.value
  if (points.length === 0) return ''

  const bottom = 152
  const first = points[0]!
  const last = points[points.length - 1]!
  const line = points.map((point) => `L ${point.x} ${point.y}`).join(' ')

  return `M ${first.x} ${bottom} L ${first.x} ${first.y} ${line} L ${last.x} ${bottom} Z`
})

const gasFeeMetricTrend = computed(() => {
  const points = gasFeeMetricPoints.value
  if (points.length < 2) return 0

  return points[points.length - 1]!.value - points[0]!.value
})
</script>

<template>
  <div class="dashboard-page">
    <div class="page-header">
      <h2 class="page-title">
        Gas Fee
      </h2>
    </div>

    <div
      v-if="loading"
      class="loading-state"
    >
      Loading gas fee...
    </div>

    <div
      v-else-if="stats"
      class="gas-fee-content"
    >
      <div class="gas-fee-summary">
        <StatCard
          class="gas-fee-balance-card"
          title="Gas Fee Balance"
          :value="stats.totalGasFeePaid.toLocaleString()"
          unit="USDT"
          action-label="Deposit"
          action-icon="lucide:plus"
          @action="depositModalOpen = true"
        />

        <div class="gas-fee-metric">
          <div class="gas-fee-metric__header">
            <div>
              <div class="gas-fee-metric__label">
                Gas Fee Metric
              </div>
              <div class="gas-fee-metric__value">
                ${{ formatCurrency(stats.totalGasFeePaid) }}<span>USDT</span>
              </div>
            </div>
            <div
              class="gas-fee-metric__trend"
              :class="gasFeeMetricTrend >= 0 ? 'change-positive' : 'change-negative'"
            >
              {{ gasFeeMetricTrend >= 0 ? '+' : '-' }}${{ formatCurrency(Math.abs(gasFeeMetricTrend)) }}
            </div>
          </div>

          <div class="gas-fee-chart">
            <svg
              viewBox="0 0 640 170"
              preserveAspectRatio="none"
              aria-hidden="true"
            >
              <defs>
                <linearGradient
                  id="gasFeeArea"
                  x1="0"
                  y1="0"
                  x2="0"
                  y2="1"
                >
                  <stop
                    offset="0%"
                    stop-color="var(--accent)"
                    stop-opacity="0.34"
                  />
                  <stop
                    offset="100%"
                    stop-color="var(--accent)"
                    stop-opacity="0"
                  />
                </linearGradient>
              </defs>
              <line
                x1="24"
                y1="38"
                x2="616"
                y2="38"
                class="gas-fee-chart__grid"
              />
              <line
                x1="24"
                y1="84"
                x2="616"
                y2="84"
                class="gas-fee-chart__grid"
              />
              <line
                x1="24"
                y1="130"
                x2="616"
                y2="130"
                class="gas-fee-chart__grid"
              />
              <path
                :d="gasFeeMetricArea"
                class="gas-fee-chart__area"
              />
              <polyline
                :points="gasFeeMetricLine"
                class="gas-fee-chart__line"
              />
              <circle
                v-for="point in gasFeeMetricPoints"
                :key="point.label"
                :cx="point.x"
                :cy="point.y"
                r="3.5"
                class="gas-fee-chart__dot"
              />
            </svg>
          </div>

          <div class="gas-fee-metric__footer">
            <div>
              <span>Fees Used</span>
              <strong>${{ formatCurrency(totalGasFeeUsed) }}</strong>
            </div>
            <div>
              <span>Rebates</span>
              <strong>${{ formatCurrency(totalGasFeeRebates) }}</strong>
            </div>
          </div>
        </div>
      </div>

      <section class="gas-fee-history">
        <div class="section-header">
          <h3>Gas Fee History</h3>
        </div>

        <div class="gas-fee-list">
          <div
            v-for="entry in gasFeeHistory"
            :key="entry.id"
            class="gas-fee-row"
          >
            <div class="gas-fee-row__info">
              <div class="gas-fee-row__title">
                {{ entry.title }}
              </div>
              <div class="gas-fee-row__meta">
                <span>{{ entry.id }}</span>
                <span class="gas-fee-row__dot" />
                <span>{{ formatDate(entry.occurredAt) }}</span>
              </div>
            </div>

            <div class="gas-fee-row__reference">
              <div class="gas-fee-row__label">
                Reference
              </div>
              <div class="gas-fee-row__value">
                {{ entry.reference }}
              </div>
            </div>

            <div class="gas-fee-row__balance">
              <div class="gas-fee-row__label">
                Balance After
              </div>
              <div class="gas-fee-row__value">
                ${{ formatCurrency(entry.balanceAfter) }}
              </div>
            </div>

            <div
              class="gas-fee-row__change"
              :class="entry.change >= 0 ? 'change-positive' : 'change-negative'"
            >
              {{ entry.change >= 0 ? '+' : '-' }}${{ formatCurrency(Math.abs(entry.change)) }}
            </div>
          </div>
        </div>
      </section>
    </div>

    <GasFeeDepositModal v-model="depositModalOpen" />
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

.gas-fee-content {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.gas-fee-summary {
  display: grid;
  grid-template-columns: minmax(280px, 420px) minmax(420px, 1fr);
  align-items: stretch;
  gap: 1.5rem;
}

.gas-fee-balance-card :deep(.stat-card__action) {
  margin-top: auto;
}

.gas-fee-metric {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-width: 0;
  min-height: 100%;
  padding: 1.5rem;
  background: var(--charcoal);
  border: 1px solid var(--line);
  transition: border-color 300ms var(--ease-quiet);
}

.gas-fee-metric:hover {
  border-color: var(--accent);
}

.gas-fee-metric__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.gas-fee-metric__label {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  letter-spacing: 0.25em;
  text-transform: uppercase;
}

.gas-fee-metric__value {
  display: flex;
  align-items: baseline;
  margin-top: 0.75rem;
  font-family: 'Oswald', sans-serif;
  font-size: 2.1rem;
  font-weight: 300;
  line-height: 1;
  color: var(--text);
}

.gas-fee-metric__value span {
  margin-left: 0.35rem;
  font-size: 0.8rem;
  color: var(--text-mute);
  font-weight: 300;
}

.gas-fee-metric__trend {
  flex: 0 0 auto;
  font-family: var(--mono);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.gas-fee-chart {
  width: 100%;
  height: 170px;
  margin: 0.75rem 0;
}

.gas-fee-chart svg {
  display: block;
  width: 100%;
  height: 100%;
}

.gas-fee-chart__grid {
  stroke: var(--line);
  stroke-width: 1;
  vector-effect: non-scaling-stroke;
}

.gas-fee-chart__area {
  fill: url("#gasFeeArea");
}

.gas-fee-chart__line {
  fill: none;
  stroke: var(--accent);
  stroke-width: 3;
  stroke-linecap: round;
  stroke-linejoin: round;
  vector-effect: non-scaling-stroke;
}

.gas-fee-chart__dot {
  fill: var(--bg-elevated);
  stroke: var(--accent);
  stroke-width: 2;
  vector-effect: non-scaling-stroke;
}

.gas-fee-metric__footer {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
  border-top: 1px solid var(--line);
  padding-top: 1rem;
}

.gas-fee-metric__footer div {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 0;
}

.gas-fee-metric__footer span {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.gas-fee-metric__footer strong {
  font-family: var(--mono);
  font-size: 13px;
  color: var(--text);
  font-weight: 500;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--line);
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

.gas-fee-history {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.gas-fee-list {
  display: flex;
  flex-direction: column;
  max-height: calc(6.1rem * 12);
  overflow-y: auto;
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  scrollbar-color: var(--accent) var(--charcoal);
  scrollbar-gutter: stable;
  scrollbar-width: thin;
}

.gas-fee-list::-webkit-scrollbar {
  width: 10px;
}

.gas-fee-list::-webkit-scrollbar-track {
  background: var(--charcoal);
  border-left: 1px solid var(--line);
}

.gas-fee-list::-webkit-scrollbar-thumb {
  background: var(--accent);
  border: 2px solid var(--charcoal);
  border-radius: 999px;
}

.gas-fee-list::-webkit-scrollbar-thumb:hover {
  background: #ff7324;
}

.gas-fee-row {
  display: grid;
  grid-template-columns: 2fr 1.4fr 1.4fr 1fr;
  align-items: center;
  gap: 1.5rem;
  min-height: 6.1rem;
  padding: 1.35rem 1.5rem;
  border-bottom: 1px solid var(--line);
  background: var(--bg-elevated);
  transition: background 300ms var(--ease-quiet);
}

.gas-fee-row:hover {
  background: var(--charcoal);
}

.gas-fee-row__info {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  min-width: 0;
}

.gas-fee-row__title {
  font-family: 'Oswald', sans-serif;
  font-size: 1.2rem;
  color: var(--text);
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.gas-fee-row__meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  text-transform: uppercase;
  min-width: 0;
}

.gas-fee-row__dot {
  width: 3px;
  height: 3px;
  background: var(--line-strong);
  border-radius: 50%;
  flex: 0 0 auto;
}

.gas-fee-row__reference,
.gas-fee-row__balance {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 0;
}

.gas-fee-row__label {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.gas-fee-row__value {
  font-family: var(--mono);
  font-size: 13px;
  color: var(--text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.gas-fee-row__change {
  justify-self: end;
  font-family: 'Oswald', sans-serif;
  font-size: 1.35rem;
  letter-spacing: 0.02em;
}

.change-positive {
  color: #10b981;
}

.change-negative {
  color: #ef4444;
}

@media (max-width: 900px) {
  .gas-fee-summary {
    grid-template-columns: 1fr;
  }

  .gas-fee-row {
    grid-template-columns: 1fr;
    align-items: start;
    gap: 0.85rem;
  }

  .gas-fee-row__change {
    justify-self: start;
  }
}
</style>
