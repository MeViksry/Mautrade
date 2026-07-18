<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useDashboardData } from '~/composables/useDashboardData'
import LayerRow from '~/components/LayerRow.vue'

definePageMeta({
  layout: 'dashboard'
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
    <div class="page-header">
      <h2 class="page-title">
        Active Layers
      </h2>
    </div>

    <div
      v-if="loading"
      class="loading-state"
    >
      Loading layers...
    </div>

    <div
      v-else
      class="layers-container"
    >
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
}

.layers-list {
  display: flex;
  flex-direction: column;
}

.empty-state {
  padding: 3rem;
  text-align: center;
  font-family: var(--mono);
  font-size: 12px;
  color: var(--text-mute);
}
</style>
