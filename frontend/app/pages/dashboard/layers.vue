<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useDashboardData } from '~/composables/useDashboardData'
import LayerRow from '~/components/LayerRow.vue'

definePageMeta({
  layout: 'dashboard'
})

const seoTitle = 'Active Layers | Mautrade Dashboard'
const seoDescription = 'View active Mautrade trading layers, entries, allocations, current prices, and unrealized PNL.'

useSeoMeta({
  title: seoTitle,
  description: seoDescription,
  ogTitle: seoTitle,
  ogDescription: seoDescription,
  twitterTitle: seoTitle,
  twitterDescription: seoDescription
})

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

const { getActiveLayers } = useDashboardData()
const layers = ref<Layer[]>([])
const loading = ref(true)

onMounted(async () => {
  loading.value = true
  try {
    layers.value = await getActiveLayers()
  } catch (error) {
    console.error('Error fetching layers:', error)
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
      <div class="layers-container skeleton-layers">
        <div
          v-for="n in 6"
          :key="`layer-${n}`"
          class="skeleton-layer-row"
        >
          <div class="skeleton-layer-info">
            <div class="skeleton-bone skeleton-layer-pair" />
            <div class="skeleton-bone skeleton-layer-meta" />
          </div>
          <div class="skeleton-layer-stats">
            <div
              v-for="s in 3"
              :key="`ls-${s}`"
              class="skeleton-layer-stat"
            >
              <div class="skeleton-bone skeleton-layer-stat-label" />
              <div class="skeleton-bone skeleton-layer-stat-val" />
            </div>
          </div>
          <div class="skeleton-layer-pnl">
            <div class="skeleton-bone skeleton-layer-pnl-amount" />
            <div class="skeleton-bone skeleton-layer-pnl-pct" />
          </div>
        </div>
      </div>
    </div>

    <template v-else>
      <div class="page-header">
        <h2 class="page-title">
          Active Layers
        </h2>
      </div>

      <div class="layers-container">
        <div class="layers-list">
          <LayerRow
            v-for="layer in layers"
            :key="layer.id"
            :layer="layer"
          />
          <div
            v-if="layers.length === 0"
            class="empty-state"
          >
            No active layers. Waiting for Master Signal.
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

.layers-container {
  display: flex;
  flex-direction: column;
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 4px;
  overflow: hidden;
}

.layers-list {
  display: flex;
  flex-direction: column;
  max-height: calc(6.45rem * 12);
  overflow-y: auto;
  scrollbar-color: var(--accent) var(--charcoal);
  scrollbar-gutter: stable;
  scrollbar-width: thin;
}

.layers-list::-webkit-scrollbar {
  width: 10px;
}

.layers-list::-webkit-scrollbar-track {
  background: var(--charcoal);
  border-left: 1px solid var(--line);
}

.layers-list::-webkit-scrollbar-thumb {
  background: var(--accent);
  border: 2px solid var(--charcoal);
  border-radius: 999px;
}

.layers-list::-webkit-scrollbar-thumb:hover {
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

.skeleton-layers {
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
  width: 140px;
  height: 28px;
}

.skeleton-layer-row {
  display: grid;
  grid-template-columns: 2fr 3fr 1.5fr;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--line);
}
.skeleton-layer-row:last-child { border-bottom: none; }
.skeleton-layer-info { display: flex; flex-direction: column; gap: 0.35rem; }
.skeleton-layer-pair { width: 90px; height: 18px; }
.skeleton-layer-meta { width: 130px; height: 10px; }
.skeleton-layer-stats { display: flex; gap: 2.5rem; }
.skeleton-layer-stat { display: flex; flex-direction: column; gap: 0.25rem; }
.skeleton-layer-stat-label { width: 50px; height: 9px; }
.skeleton-layer-stat-val { width: 70px; height: 13px; }
.skeleton-layer-pnl { display: flex; flex-direction: column; align-items: flex-end; gap: 0.25rem; }
.skeleton-layer-pnl-amount { width: 65px; height: 18px; }
.skeleton-layer-pnl-pct { width: 45px; height: 11px; }

@media (max-width: 640px) {
  .dashboard-page { gap: 0.75rem; }
  .page-header { margin-bottom: 0.5rem; }
  .page-title { font-size: 1.3rem; }

  .skeleton-page-header { margin-bottom: 0.5rem; }
  .skeleton-title { width: 100px; height: 22px; }

  .skeleton-layer-row {
    grid-template-columns: 1fr auto;
    grid-template-rows: auto auto;
    gap: 0.4rem;
    padding: 0.65rem;
  }
  .skeleton-layer-info { grid-column: 1; grid-row: 1; }
  .skeleton-layer-pnl { grid-column: 2; grid-row: 1; align-items: flex-end; }
  .skeleton-layer-stats {
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
