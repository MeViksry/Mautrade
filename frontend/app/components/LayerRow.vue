<script setup lang="ts">
defineProps<{
  layer: {
    id: string
    pair: string
    entryPrice: number
    currentPrice: number
    allocationPct: number
    allocatedUsdt: number
    unrealizedPnl: number
    unrealizedPnlPct: number
    openedAt: string
  }
}>()

const formatDate = (dateString: string) => {
  const d = new Date(dateString)
  return d.toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <div class="layer-row">
    <div class="layer-row__info">
      <div class="layer-row__pair">
        {{ layer.pair }}
      </div>
      <div class="layer-row__meta">
        <span class="layer-row__id">{{ layer.id }}</span>
        <span class="layer-row__dot" />
        <span class="layer-row__time">{{ formatDate(layer.openedAt) }}</span>
      </div>
    </div>

    <div class="layer-row__stats">
      <div class="layer-row__stat-group">
        <div class="layer-row__label">
          Entry
        </div>
        <div class="layer-row__val">
          ${{ layer.entryPrice.toLocaleString() }}
        </div>
      </div>
      <div class="layer-row__stat-group">
        <div class="layer-row__label">
          Current
        </div>
        <div class="layer-row__val">
          ${{ layer.currentPrice.toLocaleString() }}
        </div>
      </div>
      <div class="layer-row__stat-group">
        <div class="layer-row__label">
          Allocation ({{ layer.allocationPct }}%)
        </div>
        <div class="layer-row__val">
          ${{ layer.allocatedUsdt.toLocaleString() }}
        </div>
      </div>
    </div>

    <div
      class="layer-row__pnl"
      :class="layer.unrealizedPnl >= 0 ? 'pnl-positive' : 'pnl-negative'"
    >
      <div class="layer-row__pnl-amount">
        {{ layer.unrealizedPnl >= 0 ? '+' : '' }}${{ Math.abs(layer.unrealizedPnl).toFixed(2) }}
      </div>
      <div class="layer-row__pnl-pct">
        {{ layer.unrealizedPnlPct >= 0 ? '+' : '' }}{{ layer.unrealizedPnlPct.toFixed(2) }}%
      </div>
    </div>
  </div>
</template>

<style scoped>
.layer-row {
  display: grid;
  grid-template-columns: 2fr 3fr 1.5fr;
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
  display: flex;
  gap: 2.5rem;
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
  font-family: var(--mono);
  font-size: 13px;
  color: var(--text);
}

.layer-row__pnl {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.2rem;
}
.layer-row__pnl-amount {
  font-family: 'Oswald', sans-serif;
  font-size: 1.3rem;
  letter-spacing: 0.02em;
}
.layer-row__pnl-pct {
  font-family: var(--mono);
  font-size: 11px;
}

.pnl-positive .layer-row__pnl-amount,
.pnl-positive .layer-row__pnl-pct {
  color: #10b981; /* Emerald 500 */
}
.pnl-negative .layer-row__pnl-amount,
.pnl-negative .layer-row__pnl-pct {
  color: #ef4444; /* Red 500 */
}

@media (max-width: 640px) {
  .layer-row {
    grid-template-columns: 1fr auto;
    grid-template-rows: auto auto;
    gap: 0.5rem;
    padding: 0.75rem;
  }

  .layer-row__info {
    grid-column: 1;
    grid-row: 1;
    gap: 0.15rem;
  }

  .layer-row__pair {
    font-size: 0.95rem;
  }

  .layer-row__meta {
    font-size: 8px;
    gap: 0.35rem;
  }

  .layer-row__pnl {
    grid-column: 2;
    grid-row: 1;
  }

  .layer-row__pnl-amount {
    font-size: 1rem;
  }

  .layer-row__pnl-pct {
    font-size: 9px;
  }

  .layer-row__stats {
    grid-column: 1 / -1;
    grid-row: 2;
    gap: 0;
    justify-content: space-between;
  }

  .layer-row__stat-group {
    gap: 0.1rem;
  }

  .layer-row__label {
    font-size: 8px;
    letter-spacing: 0.05em;
  }

  .layer-row__val {
    font-size: 11px;
  }
}
</style>
