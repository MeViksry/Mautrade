<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
} from 'chart.js'
import { Line } from 'vue-chartjs'
import LayerRow from '~/components/LayerRow.vue'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
)

definePageMeta({
  layout: 'admin'
})

const seoTitle = 'Execution Hub | Admin Mautrade'
const seoDescription = 'Live market execution and layer management.'
useSeoMeta({
  title: seoTitle,
  description: seoDescription
})

const selectedCoin = ref('BTC/USDT')
const coins = ['BTC/USDT', 'ETH/USDT', 'SOL/USDT', 'BNB/USDT']

const orderType = ref<'limit' | 'market'>('limit')
const orderSide = ref<'buy' | 'sell'>('buy')
const orderPrice = ref('')
const orderAmount = ref('')

// Dummy data for live chart
const chartLabels = ref(Array.from({ length: 50 }, (_, i) => `T-${50 - i}`))
const chartDataValues = ref(Array.from({ length: 50 }, () => 65000 + (Math.random() - 0.5) * 100))

const chartData = computed(() => ({
  labels: chartLabels.value,
  datasets: [
    {
      label: selectedCoin.value,
      backgroundColor: 'rgba(255, 90, 0, 0.1)',
      borderColor: '#ff5a00',
      data: chartDataValues.value,
      fill: false,
      tension: 0.1,
      pointRadius: 0
    }
  ]
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  animation: { duration: 0 },
  plugins: { legend: { display: false } },
  scales: {
    y: {
      position: 'right',
      grid: { color: 'rgba(255,255,255,0.05)' },
      ticks: { color: '#888' }
    },
    x: {
      grid: { display: false },
      ticks: { display: false }
    }
  }
}

// Orderbook Mock
const orderbookAsks = ref(Array.from({ length: 15 }, (_, i) => ({ price: 65100 + i * 10, amount: (Math.random() * 2).toFixed(4) })).reverse())
const orderbookBids = ref(Array.from({ length: 15 }, (_, i) => ({ price: 65090 - i * 10, amount: (Math.random() * 2).toFixed(4) })))

// Live Trades Mock
const recentTrades = ref(Array.from({ length: 20 }, (_, i) => ({
  time: new Date(Date.now() - i * 5000).toLocaleTimeString(),
  price: 65090 + (Math.random() - 0.5) * 20,
  amount: (Math.random() * 1.5).toFixed(4),
  type: Math.random() > 0.5 ? 'buy' : 'sell'
})))

const activeLayers = ref([
  {
    id: 'layer-btc-1',
    pair: 'BTC/USDT',
    entryPrice: 65400,
    currentPrice: 66200,
    allocationPct: 15,
    allocatedUsdt: 1500,
    unrealizedPnl: 18.3,
    unrealizedPnlPct: 1.2,
    openedAt: new Date(Date.now() - 3600000 * 2).toISOString(),
    status: 'ACTIVE'
  }
])

const completedLayers = ref([
  { id: 'layer-eth-c', pair: 'ETH/USDT', entryPrice: 3400, closePrice: 3550, pnl: 4.4, date: '2026-07-18' }
])

let intervalId: ReturnType<typeof setInterval>

onMounted(() => {
  // Simulate live data websocket
  intervalId = setInterval(() => {
    // Update chart
    chartLabels.value.push('Now')
    chartLabels.value.shift()
    const lastVal = chartDataValues.value[chartDataValues.value.length - 1] ?? 65000
    chartDataValues.value.push(lastVal + (Math.random() - 0.5) * 20)
    chartDataValues.value.shift()

    // Update orderbook randomly
    if (orderbookAsks.value[14]) {
      orderbookAsks.value[14].amount = (Math.random() * 2).toFixed(4)
    }
    if (orderbookBids.value[0]) {
      orderbookBids.value[0].amount = (Math.random() * 2).toFixed(4)
    }

    // Add recent trade
    recentTrades.value.unshift({
      time: new Date().toLocaleTimeString(),
      price: lastVal + (Math.random() - 0.5) * 10,
      amount: (Math.random() * 1.5).toFixed(4),
      type: Math.random() > 0.5 ? 'buy' : 'sell'
    })
    recentTrades.value.pop()
  }, 1000)
})

onUnmounted(() => {
  clearInterval(intervalId)
})

const handleExecuteOrder = () => {
  console.log(`Executing ${orderSide.value} ${orderType.value} for ${selectedCoin.value}`)
  // Implementation for order execution goes here
}

const cancelAllLayers = () => {
  if (confirm('Are you sure you want to cancel all active layers?')) {
    activeLayers.value = []
  }
}
</script>

<template>
  <div class="execution-page">
    <header class="page-header">
      <h1 class="page-title">
        Execution Hub
      </h1>
      <div class="header-controls">
        <select
          v-model="selectedCoin"
          class="coin-select"
        >
          <option
            v-for="coin in coins"
            :key="coin"
            :value="coin"
          >
            {{ coin }}
          </option>
        </select>
      </div>
    </header>

    <div class="trading-layout">
      <!-- Left Column: Chart and Orders -->
      <div class="main-column">
        <div class="chart-panel panel">
          <div class="panel-header">
            <h3>Live Chart (Custom feed)</h3>
          </div>
          <div class="chart-wrapper">
            <Line
              :data="chartData"
              :options="chartOptions as any"
            />
          </div>
        </div>

        <div class="layers-panel panel">
          <div class="panel-header">
            <h3>Active Master Layers</h3>
            <div class="panel-actions">
              <button
                class="action-btn cancel-all"
                @click="cancelAllLayers"
              >
                Cancel All
              </button>
              <button class="action-btn sell-limit">
                Bulk Sell Limit
              </button>
            </div>
          </div>
          <div class="layers-list">
            <LayerRow
              v-for="layer in activeLayers"
              :key="layer.id"
              :layer="layer"
            />
            <div
              v-if="activeLayers.length === 0"
              class="empty-state"
            >
              No active layers running.
            </div>
          </div>
        </div>

        <div class="history-panel panel">
          <div class="panel-header">
            <h3>Completed Layers (History)</h3>
          </div>
          <table class="history-table">
            <thead>
              <tr>
                <th>Pair</th>
                <th>Entry</th>
                <th>Close</th>
                <th>PnL (%)</th>
                <th>Date</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="h in completedLayers"
                :key="h.id"
              >
                <td>{{ h.pair }}</td>
                <td>{{ h.entryPrice }}</td>
                <td>{{ h.closePrice }}</td>
                <td :class="h.pnl >= 0 ? 'text-success' : 'text-danger'">
                  +{{ h.pnl }}%
                </td>
                <td>{{ h.date }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Right Column: Orderbook and Execution -->
      <div class="side-column">
        <div class="execution-panel panel">
          <div class="order-types">
            <button
              :class="{ active: orderType === 'limit' }"
              @click="orderType = 'limit'"
            >
              Limit
            </button>
            <button
              :class="{ active: orderType === 'market' }"
              @click="orderType = 'market'"
            >
              Market
            </button>
          </div>
          <div class="order-sides">
            <button
              :class="['side-btn buy', { active: orderSide === 'buy' }]"
              @click="orderSide = 'buy'"
            >
              Buy
            </button>
            <button
              :class="['side-btn sell', { active: orderSide === 'sell' }]"
              @click="orderSide = 'sell'"
            >
              Sell
            </button>
          </div>

          <div class="order-form">
            <div
              v-if="orderType === 'limit'"
              class="form-group"
            >
              <label>Price (USDT)</label>
              <input
                v-model="orderPrice"
                type="number"
                placeholder="0.00"
                class="form-input"
              >
            </div>
            <div class="form-group">
              <label>Amount ({{ selectedCoin.split('/')[0] }})</label>
              <input
                v-model="orderAmount"
                type="number"
                placeholder="0.00"
                class="form-input"
              >
            </div>
            <button
              :class="['execute-btn', orderSide]"
              @click="handleExecuteOrder"
            >
              Execute {{ orderSide.toUpperCase() }} Layer
            </button>
          </div>
        </div>

        <div class="orderbook-panel panel">
          <div class="panel-header">
            <h3>Order Book</h3>
          </div>
          <div class="ob-container">
            <div class="ob-header">
              <span>Price(USDT)</span>
              <span>Amount</span>
            </div>
            <div class="ob-asks">
              <div
                v-for="(ask, i) in orderbookAsks"
                :key="'a'+i"
                class="ob-row ask"
              >
                <span class="price">{{ ask.price.toFixed(2) }}</span>
                <span>{{ ask.amount }}</span>
              </div>
            </div>
            <div class="ob-spread">
              <span class="spread-price">{{ chartDataValues[chartDataValues.length - 1]?.toFixed(2) ?? '0.00' }}</span>
            </div>
            <div class="ob-bids">
              <div
                v-for="(bid, i) in orderbookBids"
                :key="'b'+i"
                class="ob-row bid"
              >
                <span class="price">{{ bid.price.toFixed(2) }}</span>
                <span>{{ bid.amount }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="recent-trades-panel panel">
          <div class="panel-header">
            <h3>Recent Trades</h3>
          </div>
          <div class="trades-container">
            <div class="trade-header">
              <span>Price</span>
              <span>Amount</span>
              <span>Time</span>
            </div>
            <div class="trades-list">
              <div
                v-for="(trade, i) in recentTrades"
                :key="'t'+i"
                class="trade-row"
              >
                <span :class="trade.type === 'buy' ? 'text-success' : 'text-danger'">{{ trade.price.toFixed(2) }}</span>
                <span>{{ trade.amount }}</span>
                <span class="text-mute">{{ trade.time }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.execution-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-title {
  font-family: 'Oswald', sans-serif;
  font-size: 1.5rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  color: var(--text);
}

.coin-select {
  background: var(--charcoal);
  border: 1px solid var(--line);
  color: var(--text);
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-family: 'Oswald', sans-serif;
  font-size: 1.1rem;
  cursor: pointer;
  outline: none;
}

.trading-layout {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 1.5rem;
}

.main-column, .side-column {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.panel {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 12px;
  display: flex;
  flex-direction: column;
}

.panel-header {
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--line);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.panel-header h3 {
  font-family: 'Oswald', sans-serif;
  font-size: 1rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  color: var(--text);
  margin: 0;
}

.chart-wrapper {
  height: 400px;
  padding: 1rem;
}

.layers-list {
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.panel-actions {
  display: flex;
  gap: 0.5rem;
}

.action-btn {
  background: transparent;
  border: 1px solid var(--line);
  color: var(--text);
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s;
}

.cancel-all {
  color: #ef4444;
  border-color: rgba(239, 68, 68, 0.3);
}

.cancel-all:hover {
  background: rgba(239, 68, 68, 0.1);
  border-color: #ef4444;
}

.sell-limit {
  color: #fbbf24;
  border-color: rgba(245, 158, 11, 0.3);
}

.sell-limit:hover {
  background: rgba(245, 158, 11, 0.1);
  border-color: #fbbf24;
}

.history-table {
  width: 100%;
  border-collapse: collapse;
}

.history-table th, .history-table td {
  padding: 0.75rem 1.5rem;
  text-align: left;
  border-bottom: 1px solid var(--line);
  font-size: 0.9rem;
}

.history-table th {
  font-family: var(--mono);
  font-size: 10px;
  text-transform: uppercase;
  color: var(--text-mute);
  border-bottom: 2px solid var(--line);
}

.order-types {
  display: flex;
  border-bottom: 1px solid var(--line);
}

.order-types button {
  flex: 1;
  background: transparent;
  border: none;
  padding: 1rem;
  color: var(--text-mute);
  font-size: 0.9rem;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  font-weight: 500;
}

.order-types button.active {
  color: var(--accent);
  border-bottom-color: var(--accent);
}

.order-sides {
  display: flex;
  gap: 0.5rem;
  padding: 1rem;
}

.side-btn {
  flex: 1;
  padding: 0.5rem;
  border-radius: 4px;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text-mute);
  cursor: pointer;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  transition: all 0.2s;
}

.side-btn.buy.active {
  background: rgba(34, 197, 94, 0.1);
  color: #4ade80;
  border-color: #4ade80;
}

.side-btn.sell.active {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border-color: #ef4444;
}

.order-form {
  padding: 0 1rem 1rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-group label {
  display: block;
  font-size: 0.8rem;
  color: var(--text-mute);
  margin-bottom: 0.25rem;
}

.form-input {
  width: 100%;
  background: var(--charcoal);
  border: 1px solid var(--line);
  color: var(--text);
  padding: 0.75rem;
  border-radius: 4px;
  font-family: var(--mono);
}

.execute-btn {
  width: 100%;
  padding: 0.85rem;
  border-radius: 4px;
  border: none;
  color: #fff;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  cursor: pointer;
  transition: opacity 0.2s;
}

.execute-btn.buy {
  background: #22c55e;
}

.execute-btn.sell {
  background: #ef4444;
}

.execute-btn:hover {
  opacity: 0.9;
}

.ob-container, .trades-container {
  padding: 1rem;
  font-family: var(--mono);
  font-size: 0.85rem;
}

.ob-header, .trade-header {
  display: flex;
  justify-content: space-between;
  color: var(--text-mute);
  font-size: 0.75rem;
  margin-bottom: 0.5rem;
}

.ob-row, .trade-row {
  display: flex;
  justify-content: space-between;
  padding: 0.15rem 0;
}

.ob-row.ask .price { color: #ef4444; }
.ob-row.bid .price { color: #4ade80; }

.ob-spread {
  text-align: center;
  padding: 0.5rem 0;
  font-size: 1.1rem;
  font-weight: bold;
  border-top: 1px solid var(--line);
  border-bottom: 1px solid var(--line);
  margin: 0.5rem 0;
}

.text-success { color: #4ade80; }
.text-danger { color: #ef4444; }
.text-mute { color: var(--text-mute); }

.empty-state {
  text-align: center;
  padding: 2rem;
  color: var(--text-mute);
  font-style: italic;
  font-size: 0.9rem;
}

@media (max-width: 1180px) {
  .trading-layout {
    grid-template-columns: 1fr;
  }
}
</style>
