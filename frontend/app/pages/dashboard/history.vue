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
    <div
      v-if="loading"
      class="skeleton-loading"
    >
      <div class="skeleton-page-header">
        <div class="skeleton-bone skeleton-title" />
      </div>
      <div class="history-container skeleton-history">
        <div
          v-for="n in 6"
          :key="`hist-${n}`"
          class="skeleton-history-row"
        >
          <div class="skeleton-history-info">
            <div class="skeleton-bone skeleton-history-pair" />
            <div class="skeleton-bone skeleton-history-meta" />
          </div>
          <div class="skeleton-history-stats">
            <div
              v-for="s in 2"
              :key="`hs-${s}`"
              class="skeleton-history-stat"
            >
              <div class="skeleton-bone skeleton-history-stat-label" />
              <div class="skeleton-bone skeleton-history-stat-val" />
            </div>
          </div>
          <div class="skeleton-history-pnl">
            <div class="skeleton-bone skeleton-history-pnl-amount" />
            <div class="skeleton-bone skeleton-history-pnl-label" />
          </div>
        </div>
      </div>
    </div>

    <template v-else>
      <div class="page-header">
        <h2 class="page-title">
          Trading History
        </h2>
      </div>

      <div class="history-container">
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

/* ─── Skeleton Loading ─── */
@keyframes shimmer {
  0% { background-position: -400px 0; }
  100% { background-position: 400px 0; }
}

.skeleton-history {
  animation: skeletonFadeIn 0.4s ease-out;
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
  margin-bottom: 2rem;
}

.skeleton-title {
  width: 180px;
  height: 28px;
}

.skeleton-history-row {
  display: grid;
  grid-template-columns: 2fr 3fr 1.5fr;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--line);
}
.skeleton-history-row:last-child { border-bottom: none; }
.skeleton-history-info { display: flex; flex-direction: column; gap: 0.35rem; }
.skeleton-history-pair { width: 90px; height: 18px; }
.skeleton-history-meta { width: 130px; height: 10px; }
.skeleton-history-stats { display: flex; gap: 2.5rem; }
.skeleton-history-stat { display: flex; flex-direction: column; gap: 0.25rem; }
.skeleton-history-stat-label { width: 50px; height: 9px; }
.skeleton-history-stat-val { width: 70px; height: 13px; }
.skeleton-history-pnl { display: flex; flex-direction: column; align-items: flex-end; gap: 0.25rem; }
.skeleton-history-pnl-amount { width: 65px; height: 18px; }
.skeleton-history-pnl-label { width: 55px; height: 11px; }

@media (max-width: 640px) {
  .dashboard-page { gap: 0.75rem; }
  .page-header { margin-bottom: 0.5rem; }
  .page-title { font-size: 1.3rem; }

  .skeleton-page-header { margin-bottom: 0.5rem; }
  .skeleton-title { width: 130px; height: 22px; }

  .skeleton-history-row {
    grid-template-columns: 1fr auto;
    grid-template-rows: auto auto;
    gap: 0.4rem;
    padding: 0.65rem;
  }
  .skeleton-history-info { grid-column: 1; grid-row: 1; }
  .skeleton-history-pnl { grid-column: 2; grid-row: 1; align-items: flex-end; }
  .skeleton-history-stats {
    grid-column: 1 / -1;
    grid-row: 2;
    justify-content: space-between;
    gap: 0;
    margin-top: 0.25rem;
  }
}

@media (max-width: 380px) {
  .dashboard-page { gap: 0.5rem; }
  .page-title { font-size: 1.15rem; }
}
</style>
