<script setup lang="ts">
type ActiveLayer = {
  id: string
  symbol: string
  type: string
  layerNumber: number
  exchangeName: string
  exchangeDisplayName: string
  layerLabel: string
  allocationPct: number
  status: string
  createdAt: string
  totalLayers: number
  activeUsers: number
  totalVolumeQuote: number
  remainingQuantity: number
  remainingValueQuote: number
}

defineProps<{
  layer: ActiveLayer
  selling?: boolean
}>()

defineEmits<{
  'sell-layer': [layer: ActiveLayer]
}>()

const formatCurrency = (value: number) => {
  return value.toLocaleString(undefined, { maximumFractionDigits: 2 })
}

const formatQuantity = (value: number) => {
  if (value >= 1) return value.toLocaleString(undefined, { maximumFractionDigits: 6 })
  return value.toLocaleString(undefined, { maximumFractionDigits: 10 })
}

const formatPercent = (value: number) => {
  return value.toLocaleString(undefined, { maximumFractionDigits: 4 })
}

const formatDate = (dateString: string) => {
  const d = new Date(dateString)
  return d.toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

const layerIdentity = (layer: ActiveLayer) => {
  if (Number.isInteger(layer.layerNumber) && layer.layerNumber > 0) return `L${layer.layerNumber}`

  const label = layer.layerLabel?.trim()
  const layerMatch = label?.match(/^L\d+/i)
  if (layerMatch) return layerMatch[0].toUpperCase()

  return 'Layer'
}
</script>

<template>
  <div class="layer-row">
    <div class="layer-row__info">
      <div class="layer-row__pair">
        {{ layer.symbol }}
      </div>
      <div class="layer-row__meta">
        <span>{{ layerIdentity(layer) }}</span>
        <span class="layer-row__dot" />
        <span>{{ layer.activeUsers }} users</span>
        <span class="layer-row__dot" />
        <ClientOnly>
          <span>{{ formatDate(layer.createdAt) }}</span>
        </ClientOnly>
      </div>
    </div>

    <div class="layer-row__stats">
      <div class="layer-row__stat-group">
        <div class="layer-row__label">
          User Layers
        </div>
        <div class="layer-row__val">
          {{ layer.totalLayers }}
        </div>
      </div>
      <div class="layer-row__stat-group">
        <div class="layer-row__label">
          Remaining
        </div>
        <div class="layer-row__val">
          {{ formatQuantity(layer.remainingQuantity) }}
        </div>
      </div>
      <div class="layer-row__stat-group">
        <div class="layer-row__label">
          Allocation
        </div>
        <div class="layer-row__val">
          {{ formatPercent(layer.allocationPct) }}%
        </div>
      </div>
      <div class="layer-row__stat-group">
        <div class="layer-row__label">
          Value
        </div>
        <div class="layer-row__val">
          ${{ formatCurrency(layer.remainingValueQuote) }}
        </div>
      </div>
    </div>

    <div class="layer-row__actions">
      <div class="layer-row__status">
        {{ layer.status.toUpperCase() }}
      </div>
      <button
        type="button"
        class="layer-row__sell"
        :disabled="selling"
        @click="$emit('sell-layer', layer)"
      >
        {{ selling ? 'Selling...' : 'Sell Layer' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.layer-row {
  display: grid;
  grid-template-columns: 1.55fr 3fr 1.1fr;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--line);
  background: var(--bg-elevated);
  transition: background 300ms var(--ease-quiet);
}
.layer-row:hover {
  background: var(--charcoal);
}

.layer-row__info {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}
.layer-row__pair {
  font-family: 'Oswald', sans-serif;
  font-size: 1.2rem;
  color: var(--text);
  letter-spacing: 0.05em;
}
.layer-row__meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  text-transform: uppercase;
}
.layer-row__dot {
  width: 3px; height: 3px;
  background: var(--line-strong);
  border-radius: 50%;
}

.layer-row__stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 1rem;
  min-width: 0;
}
.layer-row__stat-group {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}
.layer-row__label {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}
.layer-row__val {
  min-width: 0;
  overflow: hidden;
  font-family: var(--mono);
  font-size: 13px;
  color: var(--text);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.layer-row__actions {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.5rem;
}
.layer-row__status {
  color: #10b981;
  font-family: var(--mono);
  font-size: 11px;
  font-weight: 800;
}
.layer-row__sell {
  min-width: 108px;
  min-height: 36px;
  border: 1px solid rgba(246, 70, 93, 0.46);
  background: rgba(246, 70, 93, 0.1);
  color: #f6465d;
  border-radius: 4px;
  padding: 0 0.85rem;
  font-family: var(--mono);
  font-size: 0.68rem;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
.layer-row__sell:hover:not(:disabled) {
  background: #f6465d;
  color: #fff;
}
.layer-row__sell:disabled {
  cursor: not-allowed;
  opacity: 0.58;
}

.text-green { color: #10b981; }
.text-red { color: #ef4444; }

@media (max-width: 640px) {
  .layer-row {
    grid-template-columns: 1fr auto;
    grid-template-rows: auto auto;
    gap: 0.4rem;
    padding: 0.65rem;
  }

  .layer-row__info {
    grid-column: 1;
    grid-row: 1;
    gap: 0.1rem;
  }

  .layer-row__pair {
    font-size: 0.9rem;
  }

  .layer-row__meta {
    font-size: 7.5px;
    gap: 0.3rem;
  }

  .layer-row__actions {
    grid-column: 2;
    grid-row: 1;
  }

  .layer-row__status {
    font-size: 8px;
  }

  .layer-row__sell {
    min-width: 88px;
    min-height: 30px;
    padding: 0 0.55rem;
    font-size: 0.56rem;
  }

  .layer-row__stats {
    grid-column: 1 / -1;
    grid-row: 2;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.55rem;
  }

  .layer-row__stat-group {
    gap: 0.1rem;
  }

  .layer-row__label {
    font-size: 7px;
    letter-spacing: 0.04em;
  }

  .layer-row__val {
    font-size: 10px;
  }
}

@media (max-width: 380px) {
  .layer-row {
    gap: 0.3rem;
    padding: 0.5rem;
  }

  .layer-row__pair {
    font-size: 0.85rem;
  }

  .layer-row__meta {
    font-size: 7px;
  }

  .layer-row__status {
    font-size: 7.5px;
  }

  .layer-row__label {
    font-size: 6.5px;
  }

  .layer-row__val {
    font-size: 9px;
  }
}
</style>
