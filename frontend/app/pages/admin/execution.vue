<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { createChart, type IChartApi, type ISeriesApi, type CandlestickData, type Time, CrosshairMode, ColorType, CandlestickSeries } from 'lightweight-charts'
import CoinPairDropdown from '~/components/CoinPairDropdown.vue'
import AdminActiveSignalRow from '~/components/AdminActiveSignalRow.vue'

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

const coinOptions = ref<CoinOption[]>([
  { symbol: 'BTC/USDT', name: 'Bitcoin', price: '...', change: '...', volume: '...' },
  { symbol: 'ETH/USDT', name: 'Ethereum', price: '...', change: '...', volume: '...' },
  { symbol: 'SOL/USDT', name: 'Solana', price: '...', change: '...', volume: '...' },
  { symbol: 'BNB/USDT', name: 'BNB', price: '...', change: '...', volume: '...' },
  { symbol: 'PEPE/USDT', name: 'Pepe', price: '...', change: '...', volume: '...' },
  { symbol: 'XRP/USDT', name: 'XRP', price: '...', change: '...', volume: '...' },
  { symbol: 'DOGE/USDT', name: 'Dogecoin', price: '...', change: '...', volume: '...' },
  { symbol: 'ADA/USDT', name: 'Cardano', price: '...', change: '...', volume: '...' },
  { symbol: 'AVAX/USDT', name: 'Avalanche', price: '...', change: '...', volume: '...' },
  { symbol: 'LINK/USDT', name: 'Chainlink', price: '...', change: '...', volume: '...' },
  { symbol: 'DOT/USDT', name: 'Polkadot', price: '...', change: '...', volume: '...' },
  { symbol: 'LTC/USDT', name: 'Litecoin', price: '...', change: '...', volume: '...' },
  { symbol: 'SHIB/USDT', name: 'Shiba Inu', price: '...', change: '...', volume: '...' },
  { symbol: 'TRX/USDT', name: 'TRON', price: '...', change: '...', volume: '...' },
  { symbol: 'ARB/USDT', name: 'Arbitrum', price: '...', change: '...', volume: '...' },
  { symbol: 'OP/USDT', name: 'Optimism', price: '...', change: '...', volume: '...' },
  { symbol: 'NEAR/USDT', name: 'NEAR Protocol', price: '...', change: '...', volume: '...' },
  { symbol: 'SUI/USDT', name: 'Sui', price: '...', change: '...', volume: '...' }
])

const orderType = ref<'limit' | 'market'>('limit')
const orderSide = ref<'buy' | 'sell'>('buy')
const orderPrice = ref('')
const orderAmount = ref('')

const currentPrice = ref(65000)
const currentCandle = ref<CandlestickData | null>(null)
const binanceSymbol = computed(() => selectedCoin.value.replace('/', '').toLowerCase())
const chartContainer = ref<HTMLElement | null>(null)
let chart: IChartApi | null = null
let candlestickSeries: ISeriesApi<'Candlestick'> | null = null
let ws: WebSocket | null = null
let tickersWs: WebSocket | null = null
const baseAsset = computed(() => selectedCoin.value.split('/')[0] ?? 'BTC')
const quoteAsset = computed(() => selectedCoin.value.split('/')[1] ?? 'USDT')
const selectedCoinMeta = computed<CoinOption>(() => {
  return coinOptions.value.find(coin => coin.symbol === selectedCoin.value) ?? coinOptions.value[0]!
})

const selectedCoinTrend = computed(() => {
  const coin = coinOptions.value.find(c => c.symbol === selectedCoin.value) ?? coinOptions.value[0]!
  return coin.change.startsWith('-') ? 'down' : 'up'
})

const initChart = () => {
  if (!chartContainer.value) return
  chart = createChart(chartContainer.value, {
    layout: {
      background: { type: ColorType.Solid, color: 'transparent' },
      textColor: '#888'
    },
    grid: {
      vertLines: { color: 'rgba(255,255,255,0.05)' },
      horzLines: { color: 'rgba(255,255,255,0.05)' }
    },
    crosshair: {
      mode: CrosshairMode.Normal
    },
    rightPriceScale: {
      borderColor: 'rgba(255,255,255,0.1)'
    },
    timeScale: {
      borderColor: 'rgba(255,255,255,0.1)',
      timeVisible: true,
      secondsVisible: false
    }
  })

  candlestickSeries = chart.addSeries(CandlestickSeries, {
    upColor: '#2ebd85',
    downColor: '#e0294a',
    borderVisible: false,
    wickUpColor: '#2ebd85',
    wickDownColor: '#e0294a'
  })
}

const selectedTimeframe = ref('1d')
const availableTimeframes = [
  { label: '15m', value: '15m' },
  { label: '1h', value: '1h' },
  { label: '1D', value: '1d' },
  { label: '1W', value: '1w' }
]

const loadHistoricalData = async (symbol: string) => {
  try {
    const res = await fetch(`https://api.binance.com/api/v3/klines?symbol=${symbol.toUpperCase()}&interval=${selectedTimeframe.value}&limit=500`)
    const data = await res.json()
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const formattedData = data.map((d: any) => ({
      time: (d[0] / 1000) as Time,
      open: parseFloat(d[1]),
      high: parseFloat(d[2]),
      low: parseFloat(d[3]),
      close: parseFloat(d[4])
    }))
    if (candlestickSeries) {
      candlestickSeries.setData(formattedData)
    }
    if (formattedData.length > 0) {
      const latest = formattedData[formattedData.length - 1]
      currentPrice.value = latest.close
      currentCandle.value = latest
    }
  } catch (err) {
    console.error('Failed to load historical data', err)
  }
}

const connectWebSocket = (symbol: string) => {
  if (ws) {
    ws.onclose = null
    ws.onerror = null
    ws.onmessage = null
    ws.close()
  }
  ws = new WebSocket(`wss://stream.binance.com:9443/ws/${symbol}@kline_${selectedTimeframe.value}`)
  ws.onmessage = (event) => {
    const data = JSON.parse(event.data)
    if (data.e === 'kline') {
      const kline = data.k
      const candle: CandlestickData = {
        time: (kline.t / 1000) as Time,
        open: parseFloat(kline.o),
        high: parseFloat(kline.h),
        low: parseFloat(kline.l),
        close: parseFloat(kline.c)
      }
      if (candlestickSeries) {
        candlestickSeries.update(candle)
      }
      currentPrice.value = candle.close
      currentCandle.value = candle
    }
  }
}

watch([selectedCoin, selectedTimeframe], async ([newCoin]) => {
  const sym = newCoin.replace('/', '').toLowerCase()
  if (candlestickSeries) candlestickSeries.setData([])
  await loadHistoricalData(sym)
  connectWebSocket(sym)
})

type MarketStat = {
  label: string
  value: string
  tone?: 'up' | 'down'
}

const formattedLivePrice = computed(() => {
  return currentPrice.value < 1
    ? currentPrice.value.toLocaleString(undefined, { maximumFractionDigits: 8 })
    : currentPrice.value.toLocaleString(undefined, { maximumFractionDigits: 2, minimumFractionDigits: 2 })
})

const marketStats = computed<MarketStat[]>(() => {
  const coin = coinOptions.value.find(c => c.symbol === selectedCoin.value) ?? coinOptions.value[0]!
  return [
    { label: '24H Change', value: coin.change, tone: selectedCoinTrend.value },
    { label: 'Last Price', value: formattedLivePrice.value },
    { label: '24H Volume', value: coin.volume ?? '-' },
    { label: 'Quote', value: quoteAsset.value }
  ]
})

const marketRows = computed(() => {
  const selected = coinOptions.value.find(coin => coin.symbol === selectedCoin.value)
  const rest = coinOptions.value.filter(coin => coin.symbol !== selectedCoin.value)

  return selected ? [selected, ...rest].slice(0, 8) : coinOptions.value.slice(0, 8)
})

const orderbookAsks = computed(() => {
  return Array.from({ length: 13 }, (_, i) => ({
    price: currentPrice.value + (i + 1) * 2,
    amount: (Math.random() * 2).toFixed(4),
    total: (Math.random() * 900).toFixed(2)
  })).reverse()
})

const orderbookBids = computed(() => {
  return Array.from({ length: 13 }, (_, i) => ({
    price: currentPrice.value - (i + 1) * 2,
    amount: (Math.random() * 2).toFixed(4),
    total: (Math.random() * 900).toFixed(2)
  }))
})

const recentTrades = ref(Array.from({ length: 16 }, (_, i) => ({
  time: new Date(Date.now() - i * 5000).toLocaleTimeString(),
  price: 65090 + (Math.random() - 0.5) * 20,
  amount: (Math.random() * 1.5).toFixed(4),
  type: Math.random() > 0.5 ? 'buy' : 'sell'
})))

watch(currentPrice, (newVal, oldVal) => {
  if (newVal !== oldVal && oldVal !== 0) {
    recentTrades.value.unshift({
      time: new Date().toLocaleTimeString(),
      price: newVal,
      amount: (Math.random() * 1.5).toFixed(4),
      type: newVal >= oldVal ? 'buy' : 'sell'
    })
    if (recentTrades.value.length > 16) {
      recentTrades.value.pop()
    }
  }
})

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

onMounted(async () => {
  initChart()
  const resizeObserver = new ResizeObserver(() => {
    if (chart && chartContainer.value) {
      chart.applyOptions({ width: chartContainer.value.clientWidth, height: chartContainer.value.clientHeight })
    }
  })
  if (chartContainer.value) {
    resizeObserver.observe(chartContainer.value)
  }

  const sym = binanceSymbol.value
  await loadHistoricalData(sym)
  connectWebSocket(sym)

  tickersWs = new WebSocket('wss://stream.binance.com:9443/ws/!ticker@arr')
  tickersWs.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (Array.isArray(data)) {
        data.forEach((ticker: { s: string, c: string, P: string, v: string }) => {
          const coin = coinOptions.value.find(c => c.symbol.replace('/', '') === ticker.s)
          if (coin) {
            const val = parseFloat(ticker.c)
            coin.price = val >= 1000 ? val.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) : val.toString()
            const changeVal = parseFloat(ticker.P)
            coin.change = (changeVal > 0 ? '+' : '') + changeVal.toFixed(2) + '%'
            const vol = parseFloat(ticker.v)
            let formattedVol = vol.toString()
            if (vol >= 1e9) formattedVol = (vol / 1e9).toFixed(2) + 'B'
            else if (vol >= 1e6) formattedVol = (vol / 1e6).toFixed(2) + 'M'
            else if (vol >= 1e3) formattedVol = (vol / 1e3).toFixed(2) + 'K'
            coin.volume = formattedVol + ' ' + coin.symbol.split('/')[0]
          }
        })
      }
    } catch (error) {
      console.error('Ticker parsing error', error)
    }
  }
})

onUnmounted(() => {
  if (ws) {
    ws.onclose = null
    ws.onerror = null
    ws.onmessage = null
    ws.close()
  }
  if (tickersWs) {
    tickersWs.onclose = null
    tickersWs.onerror = null
    tickersWs.onmessage = null
    tickersWs.close()
  }
  if (chart) chart.remove()
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
        <div class="market-price">
          <strong>{{ formattedLivePrice }}</strong>
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
    </section>

    <section class="terminal-grid">
      <aside class="orderbook-panel terminal-panel">
        <div class="terminal-panel__header">
          <h2>Order Book</h2>
          <span>0.01</span>
        </div>

        <ClientOnly>
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
        </ClientOnly>
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
              <button
                v-for="tf in availableTimeframes"
                :key="tf.value"
                type="button"
                :class="{ active: selectedTimeframe === tf.value }"
                @click="selectedTimeframe = tf.value"
              >
                {{ tf.label }}
              </button>
            </div>
          </div>

          <div class="chart-meta">
            <span>Open <strong v-if="currentCandle">{{ currentCandle.open.toLocaleString(undefined, { maximumFractionDigits: 4 }) }}</strong><strong v-else>...</strong></span>
            <span>High <strong v-if="currentCandle">{{ currentCandle.high.toLocaleString(undefined, { maximumFractionDigits: 4 }) }}</strong><strong v-else>...</strong></span>
            <span>Low <strong v-if="currentCandle">{{ currentCandle.low.toLocaleString(undefined, { maximumFractionDigits: 4 }) }}</strong><strong v-else>...</strong></span>
            <span>Close <strong v-if="currentCandle">{{ currentCandle.close.toLocaleString(undefined, { maximumFractionDigits: 4 }) }}</strong><strong v-else>...</strong></span>
          </div>

          <div class="chart-wrapper">
            <div
              ref="chartContainer"
              style="width: 100%; height: 100%;"
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
          </div>

          <div class="order-ticket-grid">
            <div class="order-ticket order-ticket--buy">
              <label for="buy-order-price">Price</label>
              <div class="ticket-input">
                <input
                  id="buy-order-price"
                  v-model="orderPrice"
                  name="buy-order-price"
                  type="number"
                  placeholder="Market price"
                >
                <span>{{ quoteAsset }}</span>
              </div>

              <label for="buy-order-amount">Amount</label>
              <div class="ticket-input">
                <input
                  id="buy-order-amount"
                  v-model="orderAmount"
                  name="buy-order-amount"
                  type="number"
                  placeholder="0.00"
                >
                <span>{{ baseAsset }}</span>
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
              <label for="sell-order-price">Price</label>
              <div class="ticket-input">
                <input
                  id="sell-order-price"
                  v-model="orderPrice"
                  name="sell-order-price"
                  type="number"
                  placeholder="Market price"
                >
                <span>{{ quoteAsset }}</span>
              </div>

              <label for="sell-order-amount">Amount</label>
              <div class="ticket-input">
                <input
                  id="sell-order-amount"
                  v-model="orderAmount"
                  name="sell-order-amount"
                  type="number"
                  placeholder="0.00"
                >
                <span>{{ baseAsset }}</span>
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

          <ClientOnly>
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
          </ClientOnly>
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
  z-index: 10;
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
  z-index: 15;
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

.chart-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.chart-wrapper {
  flex: 1;
  min-height: 390px;
  padding: 0.5rem 0.75rem 0.85rem;
}

.order-entry {
  position: relative;
  z-index: 14;
  min-height: 214px;
  overflow: visible;
}

.order-entry__bar {
  position: relative;
  z-index: 15;
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
    min-height: 320px;
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
    min-height: 260px;
  }
}
</style>
