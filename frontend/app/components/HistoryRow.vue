<script setup lang="ts">
defineProps<{
  history: {
    id: string
    pair: string
    exitPrice: number
    pnl: number
    gasFee: number
    closedAt: string
  }
}>()

const formatDate = (dateString: string) => {
  const d = new Date(dateString)
  return d.toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <div class="history-row">
    <div class="history-row__info">
      <div class="history-row__pair">
        {{ history.pair }}
      </div>
      <div class="history-row__meta">
        <span class="history-row__id">{{ history.id }}</span>
        <span class="history-row__dot" />
        <span class="history-row__time">Closed: {{ formatDate(history.closedAt) }}</span>
      </div>
    </div>

    <div class="history-row__stats">
      <div class="history-row__stat-group">
        <div class="history-row__label">
          Exit Price
        </div>
        <div class="history-row__val">
          ${{ history.exitPrice.toLocaleString() }}
        </div>
      </div>
      <div class="history-row__stat-group">
        <div class="history-row__label">
          Gas Fee
        </div>
        <div class="history-row__val">
          ${{ history.gasFee.toFixed(2) }}
        </div>
      </div>
    </div>

    <div
      class="history-row__pnl"
      :class="history.pnl >= 0 ? 'pnl-positive' : 'pnl-negative'"
    >
      <div class="history-row__pnl-amount">
        {{ history.pnl >= 0 ? '+' : '' }}${{ Math.abs(history.pnl).toFixed(2) }}
      </div>
      <div class="history-row__pnl-label">
        Realized PNL
      </div>
    </div>
  </div>
</template>

<style scoped>
.history-row {
  display: grid;
  grid-template-columns: 2fr 3fr 1.5fr;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--line);
  background: var(--bg-elevated);
  transition: background 300ms var(--ease-quiet);
}
.history-row:hover {
  background: var(--charcoal);
}

.history-row__info {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}
.history-row__pair {
  font-family: 'Oswald', sans-serif;
  font-size: 1.2rem;
  color: var(--text);
  letter-spacing: 0.05em;
}
.history-row__meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  text-transform: uppercase;
}
.history-row__dot {
  width: 3px; height: 3px;
  background: var(--line-strong);
  border-radius: 50%;
}

.history-row__stats {
  display: flex;
  gap: 2.5rem;
}
.history-row__stat-group {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}
.history-row__label {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}
.history-row__val {
  font-family: var(--mono);
  font-size: 13px;
  color: var(--text);
}

.history-row__pnl {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.2rem;
}
.history-row__pnl-amount {
  font-family: 'Oswald', sans-serif;
  font-size: 1.3rem;
  letter-spacing: 0.02em;
}
.history-row__pnl-label {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  text-transform: uppercase;
}

.pnl-positive .history-row__pnl-amount {
  color: #10b981;
}
.pnl-negative .history-row__pnl-amount {
  color: #ef4444;
}
</style>
