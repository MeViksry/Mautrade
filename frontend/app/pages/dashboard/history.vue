<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useDashboardData } from '~/composables/useDashboardData'
import HistoryRow from '~/components/HistoryRow.vue'

definePageMeta({
  layout: 'dashboard'
})

const seoTitle = 'Trading History | Mautrade Dashboard'
const seoDescription = 'Review Mautrade trading history with closed layers, exit prices, gas fees, and realized PNL.'

useSeoMeta({
  title: seoTitle,
  description: seoDescription,
  ogTitle: seoTitle,
  ogDescription: seoDescription,
  twitterTitle: seoTitle,
  twitterDescription: seoDescription
})

interface TradeHistory {
  id: string
  pair: string
  exitPrice: number
  pnl: number
  gasFee: number
  closedAt: string
}

const { getHistory } = useDashboardData()
const historyItems = ref<TradeHistory[]>([])
const loading = ref(true)

onMounted(async () => {
  loading.value = true
  try {
    historyItems.value = await getHistory()
  } catch (error) {
    console.error('Error fetching history:', error)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="dashboard-page">
    <div class="page-header">
      <h2 class="page-title">
        Trading History
      </h2>
    </div>

    <div
      v-if="loading"
      class="loading-state"
    >
      Loading history...
    </div>

    <div
      v-else
      class="history-container"
    >
      <div class="history-list">
        <HistoryRow
          v-for="item in historyItems"
          :key="item.id"
          :history="item"
        />
        <div
          v-if="historyItems.length === 0"
          class="empty-state"
        >
          No trading history available yet.
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

.history-container {
  display: flex;
  flex-direction: column;
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 4px;
  overflow: hidden;
}

.history-list {
  display: flex;
  flex-direction: column;
  max-height: calc(6.45rem * 12);
  overflow-y: auto;
  scrollbar-color: var(--accent) var(--charcoal);
  scrollbar-gutter: stable;
  scrollbar-width: thin;
}

.history-list::-webkit-scrollbar {
  width: 10px;
}

.history-list::-webkit-scrollbar-track {
  background: var(--charcoal);
  border-left: 1px solid var(--line);
}

.history-list::-webkit-scrollbar-thumb {
  background: var(--accent);
  border: 2px solid var(--charcoal);
  border-radius: 999px;
}

.history-list::-webkit-scrollbar-thumb:hover {
  background: #ff7324;
}

.empty-state {
  padding: 3rem;
  text-align: center;
  font-family: var(--mono);
  font-size: 12px;
  color: var(--text-mute);
}
</style>
