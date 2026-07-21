<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
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
import CoinPairDropdown from '~/components/CoinPairDropdown.vue'
import AdminActiveSignalRow from '~/components/AdminActiveSignalRow.vue'

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

type CoinOption = {
  symbol: string
  name: string
  price: string
  change: string
  volume?: string
}

const coinOptions: CoinOption[] = [
  { symbol: 'BTC/USDT', name: 'Bitcoin', price: '64,718.00', change: '+1.09%', volume: '8.35K BTC' },
  { symbol: 'ETH/USDT', name: 'Ethereum', price: '3,420.12', change: '+1.30%', volume: '158.61K ETH' },
  { symbol: 'SOL/USDT', name: 'Solana', price: '182.33', change: '+2.57%', volume: '2.81M SOL' },
  { symbol: 'BNB/USDT', name: 'BNB', price: '569.31', change: '+0.30%', volume: '421.18K BNB' },
  { symbol: 'PEPE/USDT', name: 'Pepe', price: '0.00002028', change: '+5.69%', volume: '18.4T PEPE' },
  { symbol: 'XRP/USDT', name: 'XRP', price: '1.0967', change: '+0.65%', volume: '620.71M XRP' },
  { symbol: 'DOGE/USDT', name: 'Dogecoin', price: '0.07253', change: '+0.23%', volume: '1.92B DOGE' },
  { symbol: 'ADA/USDT', name: 'Cardano', price: '0.42', change: '+0.92%', volume: '331.44M ADA' },
  { symbol: 'AVAX/USDT', name: 'Avalanche', price: '28.18', change: '-0.36%', volume: '12.08M AVAX' },
  { symbol: 'LINK/USDT', name: 'Chainlink', price: '14.52', change: '+1.82%', volume: '24.66M LINK' },
  { symbol: 'DOT/USDT', name: 'Polkadot', price: '6.18', change: '+0.54%', volume: '41.82M DOT' },
  { symbol: 'LTC/USDT', name: 'Litecoin', price: '83.40', change: '-0.12%', volume: '5.97M LTC' },
  { symbol: 'SHIB/USDT', name: 'Shiba Inu', price: '0.00001891', change: '+4.10%', volume: '9.48T SHIB' },
  { symbol: 'TRX/USDT', name: 'TRON', price: '0.31', change: '+0.48%', volume: '988.12M TRX' },
  { symbol: 'ARB/USDT', name: 'Arbitrum', price: '0.87', change: '-1.14%', volume: '88.24M ARB' },
  { symbol: 'OP/USDT', name: 'Optimism', price: '1.72', change: '+1.76%', volume: '39.11M OP' },
  { symbol: 'NEAR/USDT', name: 'NEAR Protocol', price: '5.41', change: '+2.03%', volume: '52.08M NEAR' },
  { symbol: 'SUI/USDT', name: 'Sui', price: '3.84', change: '-0.82%', volume: '76.55M SUI' }
]

const orderType = ref<'limit' | 'market'>('limit')
const orderSide = ref<'buy' | 'sell'>('buy')
const orderPrice = ref('')
const orderAmount = ref('')

const chartLabels = ref(Array.from({ length: 50 }, (_, i) => `T-${50 - i}`))
const chartDataValues = ref(Array.from({ length: 50 }, () => 65000 + (Math.random() - 0.5) * 100))

const currentPrice = computed(() => chartDataValues.value[chartDataValues.value.length - 1] ?? 65000)
const baseAsset = computed(() => selectedCoin.value.split('/')[0] ?? 'BTC')
const quoteAsset = computed(() => selectedCoin.value.split('/')[1] ?? 'USDT')
const selectedCoinMeta = computed<CoinOption>(() => {
  return coinOptions.find(coin => coin.symbol === selectedCoin.value) ?? coinOptions[0]!
})

const selectedCoinTrend = computed(() => {
  return selectedCoinMeta.value.change.startsWith('-') ? 'down' : 'up'
})

const chartData = computed(() => ({
  labels: chartLabels.value,
  datasets: [
    {
      label: selectedCoin.value,
      backgroundColor: 'rgba(255, 90, 0, 0.1)',
      borderColor: '#ff5a00',
      data: chartDataValues.value,
      fill: false,
      tension: 0.16,
      pointRadius: 0
    }
  ]
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  animation: { duration: 0 },
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: '#121212',
      borderColor: '#ff5a00',
      borderWidth: 1,
      titleColor: '#ebebeb',
      bodyColor: '#c8c8c8'
    }
  },
  scales: {
    y: {
      position: 'right',
      grid: { color: 'rgba(255,255,255,0.05)' },
      ticks: { color: '#888' }
    },
    x: {
      grid: { color: 'rgba(255,255,255,0.025)' },
      ticks: { color: '#666', maxTicksLimit: 8 }
    }
  }
}

type MarketStat = {
  label: string
  value: string
  tone?: 'up' | 'down'
}

const marketStats = computed<MarketStat[]>(() => [
  { label: '24H Change', value: selectedCoinMeta.value.change, tone: selectedCoinTrend.value },
  { label: 'Last Price', value: selectedCoinMeta.value.price },
  { label: '24H Volume', value: selectedCoinMeta.value.volume ?? '-' },
  { label: 'Quote', value: quoteAsset.value },
  { label: 'Network', value: `${baseAsset.value} (5)` }
])

const marketRows = computed(() => {
  const selected = coinOptions.find(coin => coin.symbol === selectedCoin.value)
  const rest = coinOptions.filter(coin => coin.symbol !== selectedCoin.value)

  return selected ? [selected, ...rest].slice(0, 8) : coinOptions.slice(0, 8)
})

const quickMarketOptions = computed(() => {
  const preferredSymbols = ['BTC/USDT', 'ETH/USDT', 'SOL/USDT', 'BNB/USDT', 'PEPE/USDT', 'XRP/USDT']

  return preferredSymbols
    .map(symbol => coinOptions.find(coin => coin.symbol === symbol))
    .filter((coin): coin is CoinOption => Boolean(coin))
})

const topMovers = [
  { symbol: 'BANK/USDT', change: '+56.89%' },
  { symbol: 'TLM/USDT', change: '+55.88%' },
  { symbol: 'HOME/USDT', change: '+9.47%' }
]

const orderbookAsks = ref(
  Array.from({ length: 13 }, (_, i) => ({
    price: 65100 + i * 8,
    amount: (Math.random() * 2).toFixed(4),
    total: (Math.random() * 900).toFixed(2)
  })).reverse()
)

const orderbookBids = ref(
  Array.from({ length: 13 }, (_, i) => ({
    price: 65090 - i * 8,
    amount: (Math.random() * 2).toFixed(4),
    total: (Math.random() * 900).toFixed(2)
  }))
)

const recentTrades = ref(Array.from({ length: 16 }, (_, i) => ({
  time: new Date(Date.now() - i * 5000).toLocaleTimeString(),
  price: 65090 + (Math.random() - 0.5) * 20,
  amount: (Math.random() * 1.5).toFixed(4),
  type: Math.random() > 0.5 ? 'buy' : 'sell'
})))

interface ActiveLayerResponse {
  id: string
  symbol: string
  type: string
  allocationPct: number
  status: string
  createdAt: string
  totalLayers: number
  totalVolumeQuote: number
}

interface OpenOrderResponse {
  id: string
  symbol: string
  action: string
  quantity: number
  price: number
  status: string
  exchange: string
  createdAt: string
}

const { tokenCookie } = useAdminAuth()

interface CompletedLayer {
  id: string
  pair: string
  entryPrice: number
  closePrice: number
  pnl: number
  date: string
}

const activeLayers = ref<ActiveLayerResponse[]>([])
const openOrders = ref<OpenOrderResponse[]>([])
const completedLayers = ref<CompletedLayer[]>([
  { id: 'layer-eth-c', pair: 'ETH/USDT', entryPrice: 3400, closePrice: 3550, pnl: 4.4, date: '2026-07-18' }
])
const loading = ref(true)

const loadExecutionData = async () => {
  try {
    const config = useRuntimeConfig()
    const apiBase = config.public.apiBase
    const [signalsRes, ordersRes] = await Promise.all([
      $fetch<ActiveLayerResponse[]>(`${apiBase}/admin/signals/active`, {
        headers: { Authorization: `Bearer ${tokenCookie.value}` }
      }),
      $fetch<OpenOrderResponse[]>(`${apiBase}/admin/signals/orders`, {
        headers: { Authorization: `Bearer ${tokenCookie.value}` }
      })
    ])

    activeLayers.value = signalsRes
    openOrders.value = ordersRes
  } catch (err) {
    console.error('Failed to load execution data', err)
  }
}

onMounted(() => {
  loadExecutionData()
  setTimeout(() => {
    loading.value = false
  }, 1000)
})

let intervalId: ReturnType<typeof setInterval>

onMounted(() => {
  intervalId = setInterval(() => {
    chartLabels.value.push('Now')
    chartLabels.value.shift()

    const lastVal = chartDataValues.value[chartDataValues.value.length - 1] ?? 65000
    chartDataValues.value.push(lastVal + (Math.random() - 0.5) * 20)
    chartDataValues.value.shift()

    if (orderbookAsks.value[12]) {
      orderbookAsks.value[12].amount = (Math.random() * 2).toFixed(4)
    }
    if (orderbookBids.value[0]) {
      orderbookBids.value[0].amount = (Math.random() * 2).toFixed(4)
    }

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

const handleExecuteOrder = (side = orderSide.value) => {
  orderSide.value = side
  console.log(`Executing ${orderSide.value} ${orderType.value} for ${selectedCoin.value}`)
}

const cancelAllLayers = () => {
  if (confirm('Cancel all active master layers?')) {
    activeLayers.value = []
  }
}
</script>

<template>
  <div class="execution-page">
    <section class="market-strip">
      <div class="market-identity">
        <div class="market-picker-shell">
          <CoinPairDropdown
            v-model="selectedCoin"
            :options="coinOptions"
            label="Market Pair"
            full-width
          />

          <div class="market-picker-shell__quick">
            <button
              v-for="coin in quickMarketOptions"
              :key="`strip-${coin.symbol}`"
              type="button"
              :class="{ 'is-active': coin.symbol === selectedCoin }"
              @click="selectedCoin = coin.symbol"
            >
              {{ coin.symbol }}
            </button>
          </div>
        </div>

        <div class="market-price">
          <strong>{{ selectedCoinMeta.price }}</strong>
          <span>{{ selectedCoinMeta.name }} spot market</span>
        </div>
      </div>

      <div class="market-stats">
        <div
          v-for="stat in marketStats"
          :key="stat.label"
          class="market-stat"
        >
          <span>{{ stat.label }}</span>
          <strong :class="{ 'text-success': stat.tone === 'up', 'text-danger': stat.tone === 'down' }">{{ stat.value }}</strong>
        </div>
      </div>

      <div class="market-actions">
        <button type="button">
          <UIcon name="lucide:activity" />
          Live Feed
        </button>
        <button type="button">
          <UIcon name="lucide:settings-2" />
          Controls
        </button>
      </div>
    </section>

    <section class="terminal-grid">
      <aside class="orderbook-panel terminal-panel">
        <div class="terminal-panel__header">
          <h2>Order Book</h2>
          <span>0.01</span>
        </div>

        <div class="book-table">
          <div class="book-head">
            <span>Price</span>
            <span>Amount</span>
            <span>Total</span>
          </div>

          <div class="book-side book-side--asks">
            <div
              v-for="(ask, index) in orderbookAsks"
              :key="`ask-${index}`"
              class="book-row"
            >
              <span class="price-sell">{{ ask.price.toFixed(2) }}</span>
              <span>{{ ask.amount }}</span>
              <span>{{ ask.total }}</span>
            </div>
          </div>

          <div class="book-spread">
            <strong>{{ currentPrice.toLocaleString(undefined, { maximumFractionDigits: 2 }) }}</strong>
            <span>Spread 0.09%</span>
          </div>

          <div class="book-side book-side--bids">
            <div
              v-for="(bid, index) in orderbookBids"
              :key="`bid-${index}`"
              class="book-row"
            >
              <span class="price-buy">{{ bid.price.toFixed(2) }}</span>
              <span>{{ bid.amount }}</span>
              <span>{{ bid.total }}</span>
            </div>
          </div>
        </div>
      </aside>

      <main class="trade-zone">
        <section class="chart-panel terminal-panel">
          <div class="terminal-panel__header chart-header">
            <div class="chart-tabs">
              <button class="active">
                Chart
              </button>
              <button>
                Info
              </button>
              <button>
                Data
              </button>
              <button>
                Analysis
              </button>
            </div>
            <div class="timeframe-tabs">
              <button>
                15m
              </button>
              <button>
                1h
              </button>
              <button class="active">
                1D
              </button>
              <button>
                1W
              </button>
            </div>
          </div>

          <div class="chart-meta">
            <span>Open <strong>64,834.21</strong></span>
            <span>High <strong>64,967.25</strong></span>
            <span>Low <strong>63,887.73</strong></span>
            <span>MA(7) <strong>64,206.98</strong></span>
          </div>

          <div class="chart-wrapper">
            <Line
              :data="chartData"
              :options="chartOptions as any"
            />
          </div>
        </section>

        <section class="order-entry terminal-panel">
          <div class="order-entry__bar">
            <div class="order-entry__tabs">
              <button
                type="button"
                :class="{ active: orderType === 'limit' }"
                @click="orderType = 'limit'"
              >
                Limit
              </button>
              <button
                type="button"
                :class="{ active: orderType === 'market' }"
                @click="orderType = 'market'"
              >
                Market
              </button>
              <button type="button">
                Stop Limit
              </button>
            </div>

            <CoinPairDropdown
              v-model="selectedCoin"
              :options="coinOptions"
              label="Order Coin"
              compact
              class="order-entry__coin-select"
            />
          </div>

          <div class="order-ticket-grid">
            <div class="order-ticket order-ticket--buy">
              <label>Price</label>
              <div class="ticket-input">
                <input
                  v-model="orderPrice"
                  type="number"
                  placeholder="Market price"
                >
                <span>{{ quoteAsset }}</span>
              </div>

              <label>Amount</label>
              <div class="ticket-input">
                <input
                  v-model="orderAmount"
                  type="number"
                  placeholder="0.00"
                >
                <span>{{ baseAsset }}</span>
              </div>

              <div class="ticket-summary">
                <span>Available</span>
                <strong>8,177.18 USDT</strong>
              </div>

              <button
                class="submit-order submit-order--buy"
                type="button"
                @click="handleExecuteOrder('buy')"
              >
                Buy {{ baseAsset }}
              </button>
            </div>

            <div class="order-ticket order-ticket--sell">
              <label>Price</label>
              <div class="ticket-input">
                <input
                  v-model="orderPrice"
                  type="number"
                  placeholder="Market price"
                >
                <span>{{ quoteAsset }}</span>
              </div>

              <label>Amount</label>
              <div class="ticket-input">
                <input
                  v-model="orderAmount"
                  type="number"
                  placeholder="0.00"
                >
                <span>{{ baseAsset }}</span>
              </div>

              <div class="ticket-summary">
                <span>Locked</span>
                <strong>0.00000000 {{ baseAsset }}</strong>
              </div>

              <button
                class="submit-order submit-order--sell"
                type="button"
                @click="handleExecuteOrder('sell')"
              >
                Sell {{ baseAsset }}
              </button>
            </div>
          </div>
        </section>
      </main>

      <aside class="market-rail">
        <section class="watchlist-panel terminal-panel">
          <div class="terminal-panel__header">
            <h2>Markets</h2>
            <span>USDT</span>
          </div>

          <div class="market-selector">
            <CoinPairDropdown
              v-model="selectedCoin"
              :options="coinOptions"
              label="Choose Coin"
              compact
              full-width
              class="market-selector__dropdown"
            />

            <div class="selected-market-card">
              <span>Selected Pair</span>
              <strong>{{ selectedCoin }}</strong>
              <em :class="{ 'is-negative': selectedCoinTrend === 'down' }">
                {{ selectedCoinMeta.price }} / {{ selectedCoinMeta.change }}
              </em>
            </div>
          </div>

          <div class="watchlist">
            <button
              v-for="item in marketRows"
              :key="item.symbol"
              type="button"
              class="watch-row"
              :class="{ 'is-active': item.symbol === selectedCoin }"
              @click="selectedCoin = item.symbol"
            >
              <span>{{ item.symbol }}</span>
              <strong>{{ item.price }}</strong>
              <em :class="{ 'is-negative': item.change.startsWith('-') }">{{ item.change }}</em>
            </button>
          </div>
        </section>

        <section class="recent-trades-panel terminal-panel">
          <div class="terminal-panel__header">
            <h2>Market Trades</h2>
            <span>Live</span>
          </div>

          <div class="trade-table">
            <div class="trade-head">
              <span>Price</span>
              <span>Amount</span>
              <span>Time</span>
            </div>
            <div
              v-for="(trade, index) in recentTrades"
              :key="`trade-${index}`"
              class="trade-row"
            >
              <span :class="trade.type === 'buy' ? 'price-buy' : 'price-sell'">{{ trade.price.toFixed(2) }}</span>
              <span>{{ trade.amount }}</span>
              <span>{{ trade.time }}</span>
            </div>
          </div>
        </section>

        <section class="top-movers-panel terminal-panel">
          <div class="terminal-panel__header">
            <h2>Top Movers</h2>
            <span>24H</span>
          </div>

          <div class="mover-list">
            <div
              v-for="item in topMovers"
              :key="item.symbol"
              class="mover-row"
            >
              <span>{{ item.symbol }}</span>
              <strong>{{ item.change }}</strong>
            </div>
          </div>
        </section>
      </aside>
    </section>

    <section class="bottom-desk terminal-panel">
      <div class="bottom-tabs">
        <button class="active">
          Open Orders({{ openOrders.length }})
        </button>
        <button>
          Active Layers({{ activeLayers.length }})
        </button>
        <button>
          Completed History
        </button>
        <button>
          Risk Queue
        </button>
        <div class="bottom-actions">
          <button
            type="button"
            class="cancel-all"
            @click="cancelAllLayers"
          >
            Cancel All
          </button>
        </div>
      </div>

      <div class="orders-table-wrap">
        <table class="orders-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Symbol</th>
              <th>Action</th>
              <th>Quantity</th>
              <th>Price</th>
              <th>Exchange</th>
              <th>Status</th>
              <th>Created</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="order in openOrders"
              :key="order.id"
            >
              <td>{{ order.id.split('-')[0] }}</td>
              <td>{{ order.symbol }}</td>
              <td :class="order.action === 'buy' ? 'price-buy' : 'price-sell'">
                {{ order.action.toUpperCase() }}
              </td>
              <td>{{ order.quantity }}</td>
              <td>{{ order.price }}</td>
              <td>{{ order.exchange }}</td>
              <td>{{ order.status }}</td>
              <td>{{ new Date(order.createdAt).toLocaleDateString() }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="layers-list">
        <AdminActiveSignalRow
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

      <div class="completed-strip">
        <span
          v-for="item in completedLayers"
          :key="item.id"
        >
          {{ item.pair }} closed at {{ item.closePrice }} <strong class="text-success">+{{ item.pnl }}%</strong>
        </span>
      </div>
    </section>
  </div>
</template>

<style scoped>
.execution-page {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  min-width: 0;
}

.market-strip,
.terminal-panel {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 4px;
}

.market-strip {
  position: relative;
  z-index: 30;
  display: grid;
  grid-template-columns: minmax(260px, 1.1fr) minmax(420px, 2fr) auto;
  align-items: center;
  gap: 1rem;
  overflow: visible;
  padding: 0.85rem 1rem;
}

.market-identity {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  min-width: 0;
}

.market-picker-shell {
  display: grid;
  flex: 0 1 390px;
  gap: 0.45rem;
  min-width: min(100%, 300px);
}

.market-picker-shell :deep(.coin-pair-select__menu) {
  width: min(560px, calc(100vw - 2rem));
}

.market-picker-shell__quick {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.35rem;
}

.market-picker-shell__quick button {
  min-width: 0;
  height: 28px;
  overflow: hidden;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text-mute);
  border-radius: 4px;
  padding: 0 0.45rem;
  font-family: var(--mono);
  font-size: 0.62rem;
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: border-color 160ms var(--ease-quiet), background 160ms var(--ease-quiet), color 160ms var(--ease-quiet);
}

.market-picker-shell__quick button:hover,
.market-picker-shell__quick button.is-active {
  border-color: rgba(255, 90, 0, 0.48);
  background: rgba(255, 90, 0, 0.14);
  color: var(--accent);
}

.pair-dropdown {
  position: relative;
  z-index: 8;
  min-width: 220px;
}

.pair-trigger {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.65rem;
  width: 100%;
  min-height: 48px;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text);
  border-radius: 4px;
  padding: 0 0.8rem;
  text-align: left;
  transition: border-color 180ms var(--ease-quiet), background 180ms var(--ease-quiet);
}

.pair-trigger:hover,
.pair-trigger[aria-expanded='true'] {
  border-color: rgba(255, 90, 0, 0.55);
  background: rgba(255, 90, 0, 0.08);
}

.pair-trigger__main {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.pair-trigger__main > span {
  color: var(--accent);
  font-family: var(--mono);
  font-size: 0.58rem;
  font-weight: 900;
  letter-spacing: 0.12em;
  line-height: 1;
  text-transform: uppercase;
}

.pair-trigger__main strong {
  color: var(--text);
  font-family: 'Oswald', sans-serif;
  font-size: 1.06rem;
  font-weight: 500;
  line-height: 1.05;
  white-space: nowrap;
}

.pair-trigger__main small {
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.65rem;
  text-transform: uppercase;
}

.pair-trigger__icon {
  width: 16px;
  height: 16px;
  color: var(--text-mute);
  transition: transform 180ms var(--ease-quiet), color 180ms var(--ease-quiet);
}

.pair-trigger__icon--open {
  color: var(--accent);
  transform: rotate(180deg);
}

.pair-menu {
  position: absolute;
  top: calc(100% + 0.35rem);
  left: 0;
  width: min(390px, 88vw);
  overflow: hidden;
  border: 1px solid rgba(255, 90, 0, 0.32);
  background: var(--bg-elevated);
  border-radius: 4px;
  box-shadow: 0 18px 44px rgba(0, 0, 0, 0.38);
  padding: 0.4rem;
}

.pair-menu__search {
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr);
  align-items: center;
  gap: 0.55rem;
  margin-bottom: 0.35rem;
  min-height: 40px;
  padding: 0 0.65rem;
  border: 1px solid var(--line);
  background: var(--charcoal);
}

.pair-menu__search svg {
  color: var(--accent);
}

.pair-menu__search input {
  width: 100%;
  min-width: 0;
  border: 0;
  background: transparent;
  color: var(--text);
  font-family: var(--mono);
  font-size: 0.72rem;
  outline: none;
}

.pair-menu__list {
  max-height: 322px;
  overflow-y: auto;
  padding-right: 0.15rem;
  scrollbar-color: var(--accent) var(--charcoal);
  scrollbar-width: thin;
}

.pair-option {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 1rem;
  width: 100%;
  min-height: 52px;
  border: 1px solid transparent;
  background: transparent;
  color: var(--text);
  border-radius: 4px;
  padding: 0 0.65rem;
  text-align: left;
  transition: border-color 160ms var(--ease-quiet), background 160ms var(--ease-quiet);
}

.pair-option:hover,
.pair-option--active {
  border-color: rgba(255, 90, 0, 0.32);
  background: rgba(255, 90, 0, 0.08);
}

.pair-option__asset,
.pair-option__market {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.pair-option__asset strong,
.pair-option__market strong {
  color: var(--text);
  font-family: var(--mono);
  font-size: 0.78rem;
  font-weight: 700;
  white-space: nowrap;
}

.pair-option__asset small,
.pair-option__market small {
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.66rem;
}

.pair-option__market {
  align-items: flex-end;
}

.pair-option__market small {
  color: #00c087;
  font-weight: 700;
}

.pair-option__market small.is-negative {
  color: #f6465d;
}

.pair-menu__empty {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 64px;
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.72rem;
}

.pair-dot {
  width: 9px;
  height: 9px;
  background: var(--accent);
  display: inline-block;
  min-width: 0;
}

.market-price {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.market-price strong {
  color: #00c087;
  font-family: 'Oswald', sans-serif;
  font-size: 1.45rem;
  font-weight: 500;
  line-height: 1;
}

.market-price span {
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.7rem;
}

.market-stats {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 0.8rem;
  min-width: 0;
}

.market-stat {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  min-width: 0;
}

.market-stat span,
.terminal-panel__header span,
.chart-meta,
.ticket-summary,
.completed-strip {
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.68rem;
}

.market-stat strong {
  color: var(--text);
  font-family: var(--mono);
  font-size: 0.78rem;
  white-space: nowrap;
}

.market-actions {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.market-actions button {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  height: 34px;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text);
  border-radius: 4px;
  padding: 0 0.75rem;
  font-family: var(--mono);
  font-size: 0.72rem;
}

.terminal-grid {
  display: grid;
  grid-template-columns: minmax(240px, 300px) minmax(0, 1fr) minmax(270px, 330px);
  gap: 0.35rem;
  align-items: stretch;
}

.trade-zone,
.market-rail {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  min-width: 0;
}

.trade-zone {
  overflow: visible;
}

.terminal-panel {
  min-width: 0;
  overflow: hidden;
}

.watchlist-panel {
  position: relative;
  z-index: 36;
  overflow: visible;
}

.terminal-panel__header {
  min-height: 42px;
  padding: 0 0.85rem;
  border-bottom: 1px solid var(--line);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.terminal-panel__header h2,
.terminal-panel__header h3 {
  margin: 0;
  color: var(--text);
  font-family: 'Oswald', sans-serif;
  font-size: 0.95rem;
  font-weight: 400;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.book-table,
.trade-table {
  padding: 0.65rem 0.85rem;
  font-family: var(--mono);
  font-size: 0.72rem;
}

.book-head,
.book-row,
.trade-head,
.trade-row,
.watch-row,
.mover-row {
  display: grid;
  align-items: center;
  gap: 0.6rem;
}

.book-head,
.book-row {
  grid-template-columns: 1fr 0.8fr 0.8fr;
}

.book-head,
.trade-head {
  color: var(--text-mute);
  padding-bottom: 0.35rem;
}

.book-row,
.trade-row {
  position: relative;
  min-height: 22px;
  color: var(--silver);
}

.book-row span:nth-child(2),
.book-row span:nth-child(3),
.trade-row span:nth-child(2),
.trade-row span:nth-child(3) {
  text-align: right;
}

.book-spread {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 0.45rem -0.85rem;
  padding: 0.65rem 0.85rem;
  border-top: 1px solid var(--line);
  border-bottom: 1px solid var(--line);
  background: rgba(255, 90, 0, 0.06);
}

.book-spread strong {
  color: #00c087;
  font-family: 'Oswald', sans-serif;
  font-size: 1.25rem;
}

.book-spread span {
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.68rem;
}

.chart-panel {
  position: relative;
  z-index: 1;
  min-height: 480px;
}

.chart-header {
  align-items: stretch;
  padding: 0 0.75rem;
}

.chart-tabs,
.timeframe-tabs,
.order-entry__tabs,
.bottom-tabs {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  min-width: 0;
}

.chart-tabs button,
.timeframe-tabs button,
.order-entry__tabs button,
.bottom-tabs button {
  height: 42px;
  border: 0;
  border-bottom: 2px solid transparent;
  background: transparent;
  color: var(--text-mute);
  padding: 0 0.65rem;
  font-family: var(--mono);
  font-size: 0.72rem;
}

.chart-tabs button.active,
.timeframe-tabs button.active,
.order-entry__tabs button.active,
.bottom-tabs button.active {
  color: var(--accent);
  border-bottom-color: var(--accent);
}

.chart-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  padding: 0.7rem 0.85rem 0;
}

.chart-meta strong {
  color: var(--accent);
  font-weight: 500;
}

.chart-wrapper {
  height: 390px;
  padding: 0.5rem 0.75rem 0.85rem;
}

.order-entry {
  position: relative;
  z-index: 24;
  min-height: 214px;
  overflow: visible;
}

.order-entry__bar {
  position: relative;
  z-index: 30;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  min-height: 58px;
  border-bottom: 1px solid var(--line);
  padding: 0 0.85rem;
}

.order-entry__bar .order-entry__tabs {
  min-height: 58px;
}

.order-entry__coin-select {
  flex: 0 1 300px;
  z-index: 16;
}

.order-entry__coin-select :deep(.coin-pair-select__menu) {
  right: 0;
  left: auto;
}

.order-ticket-grid {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
  padding: 0.9rem;
}

.order-ticket {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.55rem;
  min-width: 0;
}

.order-ticket label {
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.68rem;
  text-transform: uppercase;
}

.ticket-input {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  border: 1px solid var(--line);
  background: var(--charcoal);
  border-radius: 4px;
  overflow: hidden;
}

.ticket-input input {
  width: 100%;
  height: 38px;
  border: 0;
  background: transparent;
  color: var(--text);
  padding: 0 0.7rem;
  outline: none;
  font-family: var(--mono);
}

.ticket-input span {
  padding-right: 0.7rem;
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.72rem;
}

.ticket-summary {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
}

.ticket-summary strong {
  color: var(--silver);
  font-weight: 500;
}

.submit-order {
  height: 40px;
  border: 0;
  border-radius: 4px;
  color: #030303;
  font-family: var(--mono);
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.submit-order--buy {
  background: #00c087;
}

.submit-order--sell {
  background: #f6465d;
  color: #fff;
}

.watchlist,
.mover-list {
  padding: 0.55rem;
}

.market-selector {
  display: grid;
  gap: 0.55rem;
  padding: 0.65rem;
  border-bottom: 1px solid var(--line);
}

.market-selector__dropdown {
  width: 100%;
  min-width: 0;
}

.market-selector__dropdown :deep(.coin-pair-select__trigger) {
  min-height: 46px;
  background: color-mix(in srgb, var(--charcoal) 88%, var(--accent) 12%);
}

.market-selector__dropdown :deep(.coin-pair-select__menu) {
  left: auto;
  right: 0;
  width: min(520px, calc(100vw - 2rem));
}

.selected-market-card {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: end;
  gap: 0.25rem 0.75rem;
  min-height: 62px;
  border: 1px solid rgba(255, 90, 0, 0.28);
  background:
    linear-gradient(135deg, rgba(255, 90, 0, 0.12), transparent 62%),
    var(--charcoal);
  border-radius: 4px;
  padding: 0.65rem 0.75rem;
}

.selected-market-card span {
  grid-column: 1 / -1;
  color: var(--accent);
  font-family: var(--mono);
  font-size: 0.58rem;
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.selected-market-card strong {
  min-width: 0;
  color: var(--text);
  font-family: 'Oswald', sans-serif;
  font-size: 1.35rem;
  font-weight: 500;
  line-height: 1;
}

.selected-market-card em {
  color: #00c087;
  font-family: var(--mono);
  font-size: 0.68rem;
  font-style: normal;
  font-weight: 800;
  text-align: right;
  white-space: nowrap;
}

.selected-market-card em.is-negative {
  color: #f6465d;
}

.watch-row {
  width: 100%;
  grid-template-columns: minmax(0, 1fr) auto auto;
  min-height: 32px;
  border: 0;
  background: transparent;
  color: var(--text);
  border-radius: 4px;
  padding: 0 0.45rem;
  font-family: var(--mono);
  font-size: 0.72rem;
  text-align: left;
}

.watch-row:hover {
  background: var(--charcoal);
}

.watch-row.is-active {
  border: 1px solid rgba(255, 90, 0, 0.42);
  background: rgba(255, 90, 0, 0.1);
  color: var(--accent);
}

.watch-row strong,
.watch-row em {
  font-weight: 500;
  font-style: normal;
  white-space: nowrap;
}

.watch-row em,
.mover-row strong,
.text-success,
.price-buy {
  color: #00c087;
}

.watch-row em.is-negative {
  color: #f6465d;
}

.recent-trades-panel {
  flex: 1;
}

.trade-head,
.trade-row {
  grid-template-columns: 1fr 0.8fr 0.8fr;
}

.top-movers-panel {
  min-height: 138px;
}

.mover-row {
  grid-template-columns: minmax(0, 1fr) auto;
  min-height: 30px;
  padding: 0 0.45rem;
  color: var(--silver);
  font-family: var(--mono);
  font-size: 0.72rem;
}

.bottom-desk {
  min-height: 240px;
}

.bottom-tabs {
  border-bottom: 1px solid var(--line);
  padding: 0 0.85rem;
}

.bottom-actions {
  margin-left: auto;
}

.cancel-all {
  color: #f6465d !important;
}

.orders-table-wrap {
  overflow-x: auto;
}

.orders-table {
  width: 100%;
  min-width: 760px;
  border-collapse: collapse;
}

.orders-table th,
.orders-table td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--line);
  color: var(--silver);
  font-family: var(--mono);
  font-size: 0.72rem;
  text-align: left;
}

.orders-table th {
  color: var(--text-mute);
  font-size: 0.65rem;
  font-weight: 500;
  text-transform: uppercase;
}

.layers-list {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  padding: 0.8rem;
  border-top: 1px solid var(--line);
}

.completed-strip {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  padding: 0.8rem 1rem;
  border-top: 1px solid var(--line);
}

.price-sell,
.text-danger {
  color: #f6465d;
}

.empty-state {
  text-align: center;
  padding: 2rem;
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.8rem;
}

@media (max-width: 1380px) {
  .market-strip {
    grid-template-columns: 1fr;
  }

  .market-stats {
    grid-template-columns: repeat(5, minmax(110px, 1fr));
    overflow-x: auto;
  }

  .terminal-grid {
    grid-template-columns: minmax(220px, 280px) minmax(0, 1fr);
  }

  .market-rail {
    grid-column: 1 / -1;
    display: grid;
    grid-template-columns: 1.1fr 1.2fr 0.8fr;
  }
}

@media (max-width: 980px) {
  .terminal-grid,
  .market-rail,
  .order-ticket-grid {
    grid-template-columns: 1fr;
  }

  .order-entry__bar {
    align-items: stretch;
    flex-direction: column;
    gap: 0.65rem;
    padding: 0.75rem;
  }

  .order-entry__bar .order-entry__tabs {
    min-height: 42px;
  }

  .order-entry__coin-select {
    width: 100%;
    flex-basis: auto;
  }

  .order-entry__coin-select :deep(.coin-pair-select__menu) {
    right: auto;
    left: 0;
  }

  .orderbook-panel {
    order: 2;
  }

  .trade-zone {
    order: 1;
  }

  .market-rail {
    order: 3;
    display: flex;
  }

  .chart-panel {
    min-height: 420px;
  }

  .chart-wrapper {
    height: 320px;
  }
}

@media (max-width: 640px) {
  .market-strip {
    padding: 0.75rem;
  }

  .market-actions,
  .chart-header,
  .bottom-tabs {
    flex-wrap: wrap;
  }

  .market-identity {
    align-items: stretch;
    flex-direction: column;
  }

  .market-picker-shell {
    flex-basis: auto;
    width: 100%;
  }

  .market-picker-shell__quick {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .market-stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    overflow: visible;
  }

  .chart-tabs,
  .timeframe-tabs,
  .order-entry__tabs,
  .bottom-tabs {
    overflow-x: auto;
  }

  .chart-wrapper {
    height: 260px;
  }
}
</style>
