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

const defaultSelectedCoin = 'BTC/USDT'
const selectedCoinCookie = useCookie<string>('mautrade_admin_execution_selected_coin', {
  default: () => defaultSelectedCoin,
  maxAge: 60 * 60 * 24 * 365,
  path: '/',
  sameSite: 'lax'
})
const route = useRoute()
const activeChartTab = ref('Chart')

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

const normalizeCoinSymbol = (value: unknown) => {
  const rawValue = Array.isArray(value) ? value[0] : value
  if (typeof rawValue !== 'string') return ''

  const compactSymbol = rawValue.trim().toUpperCase()
  if (!compactSymbol) return ''

  const symbol = compactSymbol.includes('/')
    ? compactSymbol
    : compactSymbol.endsWith('USDT')
      ? `${compactSymbol.slice(0, -4)}/USDT`
      : `${compactSymbol}/USDT`

  return coinOptions.value.some(coin => coin.symbol === symbol) ? symbol : ''
}

const initialSelectedCoin = normalizeCoinSymbol(route.query.symbol)
  || normalizeCoinSymbol(route.query.pair)
  || normalizeCoinSymbol(selectedCoinCookie.value)
  || defaultSelectedCoin
const selectedCoin = ref(initialSelectedCoin)
selectedCoinCookie.value = initialSelectedCoin

const selectedExchange = ref('')

interface ExtendedCoinDetail {
  rank: number | string
  circulatingSupply: number
  maxSupply: number | null
  issueDate: string
  ath: number
  atl: number
  dominance: number
  network: string
  description: string
}

const coinDetails: Record<string, ExtendedCoinDetail> = {
  BTC: {
    rank: 1,
    circulatingSupply: 20060000,
    maxSupply: 21000000,
    issueDate: '2009-01-03',
    ath: 73750.00,
    atl: 0.048,
    dominance: 59.1,
    network: 'Bitcoin',
    description: 'Bitcoin is a decentralized cryptocurrency originally described in a 2008 whitepaper by a person, or group of people, using the alias Satoshi Nakamoto. It was launched soon after, in January 2009.Bitcoin is a peer-to-peer online currency, meaning that all transactions happen directly between equal, independent network participants, without the need for any intermediary to permit or facilitate them. Bitcoin was created, according to Nakamoto’s own words, to allow “online payments to be sent directly from one party to another without going through a financial institution.Some concepts for a similar type of a decentralized electronic currency precede BTC, but Bitcoin holds the distinction of being the first-ever cryptocurrency to come into actual use.'
  },
  ETH: {
    rank: 2,
    circulatingSupply: 120000000,
    maxSupply: null,
    issueDate: '2015-07-30',
    ath: 4891.70,
    atl: 0.420,
    dominance: 17.5,
    network: 'Ethereum (ERC20)',
    description: 'Ethereum is a decentralized open-source blockchain system that features its own cryptocurrency, Ether. ETH works as a platform for numerous other cryptocurrencies, as well as for the execution of decentralized smart contracts.'
  },
  SOL: {
    rank: 5,
    circulatingSupply: 462000000,
    maxSupply: null,
    issueDate: '2020-03-16',
    ath: 260.06,
    atl: 0.5052,
    dominance: 3.2,
    network: 'Solana',
    description: 'Solana is a highly functional open source project that banks on blockchain technology\'s permissionless nature to provide decentralized finance (DeFi) solutions.'
  },
  BNB: {
    rank: 4,
    circulatingSupply: 147500000,
    maxSupply: 200000000,
    issueDate: '2017-07-25',
    ath: 720.67,
    atl: 0.096,
    dominance: 3.8,
    network: 'BNB Smart Chain (BEP20)',
    description: 'BNB was launched through an Initial Coin Offering in 2017. It is the primary coin that powers the Binance ecosystem, allowing users to pay trading fees with a discount.'
  },
  PEPE: {
    rank: 23,
    circulatingSupply: 420690000000000,
    maxSupply: 420690000000000,
    issueDate: '2023-04-14',
    ath: 0.00001718,
    atl: 0.0000000276,
    dominance: 0.12,
    network: 'Ethereum (ERC20)',
    description: 'Pepe is a deflationary memecoin launched on Ethereum. The cryptocurrency was created as a tribute to the Pepe the Frog internet meme created by Matt Furie.'
  },
  XRP: {
    rank: 7,
    circulatingSupply: 55600000000,
    maxSupply: 100000000000,
    issueDate: '2012-01-01',
    ath: 3.84,
    atl: 0.0028,
    dominance: 1.3,
    network: 'XRP Ledger',
    description: 'XRP is a digital asset built for payments. It is the native digital asset on the XRP Ledger—an open-source, permissionless and decentralized blockchain technology.'
  },
  DOGE: {
    rank: 8,
    circulatingSupply: 144000000000,
    maxSupply: null,
    issueDate: '2013-12-06',
    ath: 0.7376,
    atl: 0.00008547,
    dominance: 1.1,
    network: 'Dogecoin',
    description: 'Dogecoin (DOGE) is based on the popular "doge" Internet meme and features a Shiba Inu on its logo. The open-source digital currency was created by Billy Markus and Jackson Palmer.'
  },
  ADA: {
    rank: 10,
    circulatingSupply: 35700000000,
    maxSupply: 45000000000,
    issueDate: '2017-09-01',
    ath: 3.10,
    atl: 0.0173,
    dominance: 0.7,
    network: 'Cardano',
    description: 'Cardano is a proof-of-stake blockchain platform: the first to be founded on peer-reviewed research and developed through evidence-based methods.'
  },
  AVAX: {
    rank: 11,
    circulatingSupply: 393000000,
    maxSupply: 720000000,
    issueDate: '2020-09-21',
    ath: 146.22,
    atl: 2.78,
    dominance: 0.5,
    network: 'Avalanche (C-Chain)',
    description: 'Avalanche is a lightning-fast, low-cost, and eco-friendly open-source smart contract platform built for the scale of global finance and decentralized applications.'
  },
  LINK: {
    rank: 15,
    circulatingSupply: 608000000,
    maxSupply: 1000000000,
    issueDate: '2017-09-20',
    ath: 52.88,
    atl: 0.126,
    dominance: 0.4,
    network: 'Ethereum (ERC20)',
    description: 'Chainlink is a blockchain abstraction layer that enables universally connected smart contracts through a decentralized oracle network.'
  },
  DOT: {
    rank: 16,
    circulatingSupply: 1430000000,
    maxSupply: null,
    issueDate: '2020-08-18',
    ath: 55.00,
    atl: 2.69,
    dominance: 0.38,
    network: 'Polkadot',
    description: 'Polkadot is an open-source sharded multichain protocol that connects and secures a network of specialized blockchains, facilitating cross-chain transfer of any data or asset types.'
  },
  LTC: {
    rank: 21,
    circulatingSupply: 74500000,
    maxSupply: 84000000,
    issueDate: '2011-10-13',
    ath: 412.96,
    atl: 1.11,
    dominance: 0.28,
    network: 'Litecoin',
    description: 'Litecoin (LTC) is a cryptocurrency that was designed to provide fast, secure and low-cost payments by leveraging the unique properties of blockchain technology.'
  },
  SHIB: {
    rank: 12,
    circulatingSupply: 589000000000000,
    maxSupply: null,
    issueDate: '2020-08-01',
    ath: 0.00008845,
    atl: 0.000000000081,
    dominance: 0.45,
    network: 'Ethereum (ERC20)',
    description: 'Shiba Inu coin was created anonymously in August 2020 under the pseudonym "Ryoshi." The meme coin quickly gained speed and value as a community of investors was drawn in by the cute charm of the coin.'
  },
  TRX: {
    rank: 13,
    circulatingSupply: 87000000000,
    maxSupply: null,
    issueDate: '2017-09-13',
    ath: 0.30,
    atl: 0.0010,
    dominance: 0.48,
    network: 'TRON (TRC20)',
    description: 'TRON is a decentralized blockchain-based operating system that aims to build a free, global digital content entertainment ecosystem with distributed storage technology.'
  },
  ARB: {
    rank: 38,
    circulatingSupply: 2650000000,
    maxSupply: 10000000000,
    issueDate: '2023-03-23',
    ath: 2.40,
    atl: 0.74,
    dominance: 0.1,
    network: 'Arbitrum One',
    description: 'Arbitrum is a Layer 2 (L2) scaling solution for Ethereum. It uses optimistic rollups to improve speed, scalability, and cost-efficiency on Ethereum.'
  },
  OP: {
    rank: 41,
    circulatingSupply: 1040000000,
    maxSupply: 4294967296,
    issueDate: '2022-05-31',
    ath: 4.85,
    atl: 0.40,
    dominance: 0.09,
    network: 'Optimism',
    description: 'Optimism is a low-cost and lightning-fast Ethereum L2 blockchain, designed to provide users with near-instant transactions while maintaining L1 security.'
  },
  NEAR: {
    rank: 17,
    circulatingSupply: 1060000000,
    maxSupply: null,
    issueDate: '2020-10-14',
    ath: 20.42,
    atl: 0.52,
    dominance: 0.35,
    network: 'NEAR Protocol',
    description: 'NEAR Protocol is a layer-one blockchain that was designed as a community-run cloud compute platform to eliminate some of the limitations that have been bogging down competing blockchains.'
  },
  SUI: {
    rank: 29,
    circulatingSupply: 2400000000,
    maxSupply: 10000000000,
    issueDate: '2023-05-03',
    ath: 2.18,
    atl: 0.36,
    dominance: 0.2,
    network: 'Sui Network',
    description: 'Sui is a permissionless Layer-1 blockchain designed from the ground up to enable creators and developers to build experiences that cater to the next billion users in Web3.'
  }
}

const currentCoinDetail = computed<ExtendedCoinDetail>(() => {
  const symbol = selectedCoin.value.split('/')[0] || ''
  return coinDetails[symbol] || {
    rank: '-',
    circulatingSupply: 1000000000,
    maxSupply: null,
    issueDate: '-',
    ath: currentPrice.value * 2,
    atl: currentPrice.value * 0.1,
    dominance: 0.01,
    network: symbol,
    description: 'This cryptocurrency is traded live on the Mautrade market. This information is presented as a standard reference.'
  }
})

const buyAllocationPct = ref('10')
const dispatchingSignal = ref(false)
const sellingLayerKey = ref('')
const signalMessage = ref('')
const signalMessageTone = ref<'success' | 'error' | ''>('')

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
const nextLayerLabel = computed(() => `Next ${baseAsset.value} Layer`)
const selectedCoinMeta = computed<CoinOption>(() => {
  return coinOptions.value.find(coin => coin.symbol === selectedCoin.value) ?? coinOptions.value[0]!
})

const selectedCoinTrend = computed(() => {
  const coin = coinOptions.value.find(c => c.symbol === selectedCoin.value) ?? coinOptions.value[0]!
  return coin.change.startsWith('-') ? 'down' : 'up'
})

watch(selectedCoin, (symbol) => {
  const normalizedSymbol = normalizeCoinSymbol(symbol) || defaultSelectedCoin
  if (normalizedSymbol !== symbol) {
    selectedCoin.value = normalizedSymbol
    return
  }

  selectedCoinCookie.value = normalizedSymbol
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

const priceStep = computed(() => {
  const p = currentPrice.value
  if (p > 1000) return 1.0
  if (p > 100) return 0.1
  if (p > 1) return 0.01
  if (p > 0.01) return 0.0001
  return 0.000001
})

const orderbookTickSize = computed(() => {
  return priceStep.value.toString()
})

const formatPrice = (price: number) => {
  return price < 1
    ? price.toLocaleString(undefined, { maximumFractionDigits: 8, minimumFractionDigits: 4 })
    : price.toLocaleString(undefined, { maximumFractionDigits: 2, minimumFractionDigits: 2 })
}

const orderbookAsks = computed(() => {
  return Array.from({ length: 15 }).map((_, i) => ({
    price: currentPrice.value + (i + 1) * priceStep.value,
    amount: (Math.random() * (5000 / currentPrice.value)).toFixed(4),
    total: (Math.random() * 800 + 100).toFixed(2)
  })).reverse()
})

const orderbookBids = computed(() => {
  return Array.from({ length: 15 }).map((_, i) => ({
    price: currentPrice.value - (i + 1) * priceStep.value,
    amount: (Math.random() * (5000 / currentPrice.value)).toFixed(4),
    total: (Math.random() * 800 + 100).toFixed(2)
  }))
})

const generateRecentTrades = (basePrice: number) => {
  return Array.from({ length: 16 }, (_, i) => ({
    time: new Date(Date.now() - i * 5000).toLocaleTimeString(),
    price: basePrice + (Math.random() - 0.5) * (priceStep.value * 10),
    amount: (Math.random() * (5000 / basePrice)).toFixed(4),
    type: Math.random() > 0.5 ? 'buy' : 'sell'
  }))
}

const recentTrades = ref(generateRecentTrades(65000))

watch(selectedCoin, () => {
  setTimeout(() => {
    recentTrades.value = generateRecentTrades(currentPrice.value)
  }, 500)
})

watch(currentPrice, (newVal, oldVal) => {
  if (newVal !== oldVal && oldVal !== 0) {
    recentTrades.value.unshift({
      time: new Date().toLocaleTimeString(),
      price: newVal,
      amount: (Math.random() * (5000 / newVal)).toFixed(4),
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

interface CreateAdminSignalResponse {
  signalId: string
  status: string
  idempotencyKey: string
  jobsCreated: number
  jobsSkipped: number
  jobsPublished: number
  queueState: string
}

const { tokenCookie } = useAdminAuth()
const runtimeConfig = useRuntimeConfig()
const apiBase = runtimeConfig.public.apiBase

interface CompletedLayer {
  id: string
  pair: string
  entryPrice: number
  closePrice: number
  pnl: number
  date: string
}

const activeTab = ref('active')
const activeLayers = ref<ActiveLayerResponse[]>([])
const completedLayers = ref<CompletedLayer[]>([])
const loading = ref(true)

const loadCompletedData = async () => {
  try {
    completedLayers.value = await $fetch<CompletedLayer[]>(`${apiBase}/admin/signals/completed`, {
      headers: { Authorization: `Bearer ${tokenCookie.value}` }
    })
  } catch (err) {
    console.error('Failed to load completed layers', err)
  }
}

const formatDate = (dateString: string) => {
  const d = new Date(dateString)
  return d.toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

const loadExecutionData = async () => {
  try {
    activeLayers.value = await $fetch<ActiveLayerResponse[]>(`${apiBase}/admin/signals/active`, {
      headers: { Authorization: `Bearer ${tokenCookie.value}` }
    })
  } catch (err) {
    console.error('Failed to load execution data', err)
  }
}

onMounted(() => {
  loadExecutionData()
  loadCompletedData()
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

  const streams = coinOptions.value.map(c => c.symbol.replace('/', '').toLowerCase() + '@ticker').join('/')
  tickersWs = new WebSocket(`wss://stream.binance.com:9443/stream?streams=${streams}`)
  tickersWs.onmessage = (event) => {
    try {
      const payload = JSON.parse(event.data)
      if (payload.data && payload.data.s) {
        const ticker = payload.data
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

type APIErrorLike = {
  data?: { error?: string }
  response?: { _data?: { error?: string } }
  message?: string
}

const signalErrorMessage = (error: unknown) => {
  const apiError = error as APIErrorLike
  return apiError.data?.error || apiError.response?._data?.error || apiError.message || 'Failed to dispatch signal'
}

const isPercentInputValid = (value: string) => {
  const normalized = value.trim().replace(',', '.')
  if (normalized === '') return false
  const numberValue = Number(normalized)
  return Number.isFinite(numberValue) && numberValue > 0 && numberValue <= 100
}

const createSignalIdempotencyKey = (side: 'buy' | 'sell', symbol = selectedCoin.value, layerNumber?: number, exchangeName = '') => {
  const randomPart = globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(36).slice(2)}`
  const layerPart = layerNumber ? `:layer-${layerNumber}` : ''
  const exchangePart = exchangeName ? `:${exchangeName}` : ''
  return `admin-signal:${side}:${symbol}${layerPart}${exchangePart}:${randomPart}`
}

const layerSellKey = (layer: ActiveLayerResponse) => {
  return `${layer.symbol}:${layer.layerNumber}:${layer.exchangeName || 'all'}`
}

const dispatchAdminSignal = async (body: Record<string, string | number>, idempotencyKey: string) => {
  return await $fetch<CreateAdminSignalResponse>(`${apiBase}/admin/signals`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${tokenCookie.value}`,
      'Idempotency-Key': idempotencyKey
    },
    body
  })
}

const handleBuySignal = async () => {
  signalMessage.value = ''
  signalMessageTone.value = ''

  if (!tokenCookie.value) {
    signalMessage.value = 'Admin session expired. Please login again.'
    signalMessageTone.value = 'error'
    return
  }

  const idempotencyKey = createSignalIdempotencyKey('buy', selectedCoin.value, undefined, selectedExchange.value)
  const body: Record<string, string | number> = {
    type: 'buy',
    symbol: selectedCoin.value,
    idempotency_key: idempotencyKey
  }
  if (selectedExchange.value) {
    body.exchange_name = selectedExchange.value
  }

  const allocation = buyAllocationPct.value.trim().replace(',', '.')
  if (!isPercentInputValid(allocation)) {
    signalMessage.value = 'Allocation must be greater than 0 and at most 100.'
    signalMessageTone.value = 'error'
    return
  }
  body.allocation_pct = allocation

  try {
    dispatchingSignal.value = true
    const response = await dispatchAdminSignal(body, idempotencyKey)
    const actionLabel = `BUY ${allocation}%`
    signalMessage.value = `${actionLabel} accepted. Published ${response.jobsPublished}/${response.jobsCreated} jobs, skipped ${response.jobsSkipped}. Queue ${response.queueState}.`
    signalMessageTone.value = 'success'
    await loadExecutionData()
  } catch (error) {
    signalMessage.value = signalErrorMessage(error)
    signalMessageTone.value = 'error'
  } finally {
    dispatchingSignal.value = false
  }
}

const handleSellLayer = async (layer: ActiveLayerResponse) => {
  signalMessage.value = ''
  signalMessageTone.value = ''

  if (!tokenCookie.value) {
    signalMessage.value = 'Admin session expired. Please login again.'
    signalMessageTone.value = 'error'
    return
  }

  const layerNumber = Number(layer.layerNumber)
  if (!Number.isInteger(layerNumber) || layerNumber <= 0) {
    signalMessage.value = 'Layer number is missing for this active layer.'
    signalMessageTone.value = 'error'
    return
  }

  const exchangeName = layer.exchangeName?.trim() || ''
  const idempotencyKey = createSignalIdempotencyKey('sell', layer.symbol, layerNumber, exchangeName)
  const body: Record<string, string | number> = {
    type: 'sell',
    symbol: layer.symbol,
    layer_number: layerNumber,
    sell_pct: '100',
    idempotency_key: idempotencyKey
  }
  if (exchangeName) {
    body.exchange_name = exchangeName
  }

  try {
    sellingLayerKey.value = layerSellKey(layer)
    const response = await dispatchAdminSignal(body, idempotencyKey)
    signalMessage.value = `SELL ${layer.symbol} ${layer.layerLabel || `L${layerNumber}`} accepted. Published ${response.jobsPublished}/${response.jobsCreated} jobs, skipped ${response.jobsSkipped}. Queue ${response.queueState}.`
    signalMessageTone.value = 'success'
    await loadExecutionData()
  } catch (error) {
    signalMessage.value = signalErrorMessage(error)
    signalMessageTone.value = 'error'
  } finally {
    sellingLayerKey.value = ''
  }
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
          <span>{{ orderbookTickSize }}</span>
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
                <span class="price-sell">{{ formatPrice(ask.price) }}</span>
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
                <span class="price-buy">{{ formatPrice(bid.price) }}</span>
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
              <button
                :class="{ active: activeChartTab === 'Chart' }"
                @click="activeChartTab = 'Chart'"
              >
                Chart
              </button>
              <button
                :class="{ active: activeChartTab === 'Info' }"
                @click="activeChartTab = 'Info'"
              >
                Info
              </button>
              <button
                :class="{ active: activeChartTab === 'Data' }"
                @click="activeChartTab = 'Data'"
              >
                Data
              </button>
              <button
                :class="{ active: activeChartTab === 'Analysis' }"
                @click="activeChartTab = 'Analysis'"
              >
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

          <div
            v-show="activeChartTab === 'Chart'"
            class="chart-meta"
          >
            <span>Open <strong v-if="currentCandle">{{ currentCandle.open.toLocaleString(undefined, { maximumFractionDigits: 4 }) }}</strong><strong v-else>...</strong></span>
            <span>High <strong v-if="currentCandle">{{ currentCandle.high.toLocaleString(undefined, { maximumFractionDigits: 4 }) }}</strong><strong v-else>...</strong></span>
            <span>Low <strong v-if="currentCandle">{{ currentCandle.low.toLocaleString(undefined, { maximumFractionDigits: 4 }) }}</strong><strong v-else>...</strong></span>
            <span>Close <strong v-if="currentCandle">{{ currentCandle.close.toLocaleString(undefined, { maximumFractionDigits: 4 }) }}</strong><strong v-else>...</strong></span>
          </div>

          <div
            v-show="activeChartTab === 'Chart'"
            class="chart-wrapper"
          >
            <div
              ref="chartContainer"
              style="width: 100%; height: 100%;"
            />
          </div>

          <div
            v-if="activeChartTab === 'Info'"
            class="tab-info-content"
          >
            <div class="info-card">
              <div class="info-header">
                <div class="coin-icon">
                  <img
                    :src="`https://assets.coincap.io/assets/icons/${selectedCoin.split('/')[0]?.toLowerCase() || 'btc'}@2x.png`"
                    :alt="selectedCoin"
                    @error="$event.target.src='https://assets.coincap.io/assets/icons/btc@2x.png'"
                  >
                </div>
                <div>
                  <h3>{{ coinOptions.find(c => c.symbol === selectedCoin)?.name || selectedCoin.split('/')[0] }}</h3>
                  <span>{{ selectedCoin }}</span>
                </div>
              </div>

              <div class="info-grid extended">
                <div class="info-item">
                  <label>Rank</label>
                  <strong>No. {{ currentCoinDetail.rank }}</strong>
                </div>
                <div class="info-item">
                  <label>Market Cap</label>
                  <strong>{{ '$' + (currentCoinDetail.circulatingSupply * currentPrice).toLocaleString(undefined, { maximumFractionDigits: 0 }) }}</strong>
                </div>
                <div class="info-item">
                  <label>Market Dominance</label>
                  <strong>{{ currentCoinDetail.dominance }}%</strong>
                </div>
                <div class="info-item">
                  <label>24H Volume</label>
                  <strong>{{ coinOptions.find(c => c.symbol === selectedCoin)?.volume || '...' }}</strong>
                </div>
                <div class="info-item">
                  <label>Circulating Supply</label>
                  <strong>{{ currentCoinDetail.circulatingSupply.toLocaleString() }} {{ selectedCoin.split('/')[0] }}</strong>
                </div>
                <div class="info-item">
                  <label>Max Supply</label>
                  <strong>{{ currentCoinDetail.maxSupply ? currentCoinDetail.maxSupply.toLocaleString() + ' ' + selectedCoin.split('/')[0] : 'Unlimited' }}</strong>
                </div>
                <div class="info-item">
                  <label>Issue Date</label>
                  <strong>{{ currentCoinDetail.issueDate }}</strong>
                </div>
                <div class="info-item">
                  <label>Network</label>
                  <strong>{{ currentCoinDetail.network }}</strong>
                </div>
                <div class="info-item">
                  <label>All-Time High</label>
                  <strong class="text-green">${{ formatPrice(currentCoinDetail.ath) }}</strong>
                </div>
                <div class="info-item">
                  <label>All-Time Low</label>
                  <strong class="text-red">${{ formatPrice(currentCoinDetail.atl) }}</strong>
                </div>
              </div>

              <div class="info-project">
                <h4>Introduction</h4>
                <p class="info-desc">
                  {{ currentCoinDetail.description }}
                </p>
              </div>
            </div>
          </div>

          <div
            v-if="!['Chart', 'Info'].includes(activeChartTab)"
            class="tab-placeholder"
          >
            <p>{{ activeChartTab }} features are currently under development.</p>
          </div>
        </section>

        <section class="order-entry terminal-panel">
          <div class="order-entry__bar">
            <div class="order-entry__mode">
              Buy Signal
            </div>
            <div class="signal-scope">
              <span>{{ selectedCoin }}</span>
              <strong>All Active Users</strong>
            </div>
          </div>

          <div class="order-ticket-grid order-ticket-grid--signal">
            <div class="order-ticket order-ticket--buy">
              <label for="buy-allocation-pct">Allocation Per User</label>
              <div class="ticket-input">
                <input
                  id="buy-allocation-pct"
                  v-model="buyAllocationPct"
                  name="buy-allocation-pct"
                  type="text"
                  inputmode="decimal"
                  placeholder="10"
                >
                <span>% {{ quoteAsset }} Spot</span>
              </div>
              <label
                for="buy-exchange-select"
                style="margin-top: 1rem;"
              >Exchange</label>
              <div class="ticket-input">
                <select
                  id="buy-exchange-select"
                  v-model="selectedExchange"
                  class="ticket-select"
                >
                  <option value="">
                    All Exchanges
                  </option>
                  <option value="binance">
                    Binance
                  </option>
                  <option value="bybit">
                    Bybit
                  </option>
                  <option value="okx">
                    OKX
                  </option>
                  <option value="tokocrypto">
                    Tokocrypto
                  </option>
                </select>
              </div>

              <div class="ticket-summary">
                <span>Layer</span>
                <strong>{{ nextLayerLabel }}</strong>
              </div>

              <button
                class="submit-order submit-order--buy"
                type="button"
                :disabled="dispatchingSignal"
                @click="handleBuySignal"
              >
                {{ dispatchingSignal ? 'Dispatching...' : `Dispatch Buy ${baseAsset}` }}
              </button>
            </div>
            <p
              v-if="signalMessage"
              class="signal-message"
              :class="{
                'signal-message--success': signalMessageTone === 'success',
                'signal-message--error': signalMessageTone === 'error'
              }"
            >
              {{ signalMessage }}
            </p>
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
                <span :class="trade.type === 'buy' ? 'price-buy' : 'price-sell'">{{ formatPrice(trade.price) }}</span>
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
        <button
          :class="{ active: activeTab === 'active' }"
          @click="activeTab = 'active'"
        >
          Active Layers({{ activeLayers.length }})
        </button>
        <button
          :class="{ active: activeTab === 'completed' }"
          @click="activeTab = 'completed'"
        >
          Completed History
        </button>
        <button
          :class="{ active: activeTab === 'risk' }"
          @click="activeTab = 'risk'"
        >
          Risk Queue
        </button>
        <div
          v-if="activeTab === 'active'"
          class="bottom-actions"
        >
          <button
            type="button"
            class="cancel-all"
            @click="cancelAllLayers"
          >
            Cancel All
          </button>
        </div>
      </div>

      <div
        v-if="activeTab === 'active'"
        class="layers-list"
      >
        <AdminActiveSignalRow
          v-for="layer in activeLayers"
          :key="layer.id"
          :layer="layer"
          :selling="sellingLayerKey === layerSellKey(layer)"
          @sell-layer="handleSellLayer"
        />
        <div
          v-if="activeLayers.length === 0"
          class="empty-state"
        >
          No active layers running.
        </div>
      </div>

      <div
        v-if="activeTab === 'completed'"
        class="layers-list"
      >
        <div
          v-for="item in completedLayers"
          :key="item.id"
          class="layer-row"
        >
          <div class="layer-row__info">
            <div class="layer-row__pair">
              {{ item.pair }}
            </div>
            <div class="layer-row__meta">
              <span>{{ formatDate(item.date) }}</span>
            </div>
          </div>
          <div class="layer-row__stats">
            <div class="layer-row__stat-group">
              <div class="layer-row__label">
                Entry Price
              </div>
              <div class="layer-row__val">
                {{ item.entryPrice }}
              </div>
            </div>
            <div class="layer-row__stat-group">
              <div class="layer-row__label">
                Close Price
              </div>
              <div class="layer-row__val">
                {{ item.closePrice }}
              </div>
            </div>
            <div class="layer-row__stat-group">
              <div class="layer-row__label">
                PnL
              </div>
              <div
                class="layer-row__val"
                :style="{ color: item.pnl >= 0 ? 'var(--text-success)' : 'var(--text-danger)' }"
              >
                {{ item.pnl > 0 ? '+' : '' }}{{ item.pnl.toFixed(2) }}%
              </div>
            </div>
          </div>
        </div>
        <div
          v-if="completedLayers.length === 0"
          class="empty-state"
        >
          No completed history.
        </div>
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

.book-head span:nth-child(2),
.book-head span:nth-child(3),
.book-row span:nth-child(2),
.book-row span:nth-child(3),
.trade-head span:nth-child(2),
.trade-head span:nth-child(3),
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
.bottom-tabs {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  min-width: 0;
}

.chart-tabs button,
.timeframe-tabs button,
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
  min-height: 380px;
  padding: 0.5rem 0.75rem 0.85rem;
}

.tab-info-content {
  flex: 1;
  display: flex;
  min-height: 380px;
  padding: 1.5rem;
}

.info-card {
  width: 100%;
  background: var(--charcoal);
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.info-header {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.info-header h3 {
  margin: 0 0 0.2rem 0;
  color: var(--text);
  font-family: 'Oswald', sans-serif;
  font-size: 1.25rem;
}

.info-header span {
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.85rem;
}

.coin-icon {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.05);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.coin-icon img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.info-desc {
  color: var(--silver);
  font-size: 0.9rem;
  line-height: 1.5;
  margin: 0;
}

.info-project {
  margin-top: 0.5rem;
  padding-top: 1.25rem;
  border-top: 1px solid var(--line);
}

.info-project h4 {
  color: var(--text);
  font-family: 'Oswald', sans-serif;
  margin: 0 0 0.5rem 0;
}

.info-project h5 {
  color: var(--text);
  margin: 0.5rem 0;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.25rem;
  margin-top: 0.5rem;
  padding-top: 1.25rem;
  border-top: 1px solid var(--line);
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.info-item label {
  color: var(--text-mute);
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.info-item strong {
  color: var(--text);
  font-family: var(--mono);
  font-size: 1.1rem;
}

.tab-placeholder {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 380px;
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.85rem;
  background: rgba(0, 0, 0, 0.2);
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

.order-entry__mode {
  display: flex;
  align-items: center;
  min-height: 58px;
  border-bottom: 2px solid var(--accent);
  color: var(--accent);
  font-family: var(--mono);
  font-size: 0.72rem;
  font-weight: 800;
  text-transform: uppercase;
}

.order-entry__coin-select {
  flex: 0 1 300px;
  z-index: 16;
}

.order-entry__coin-select :deep(.coin-pair-select__menu) {
  right: 0;
  left: auto;
}

.signal-scope {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.65rem;
  text-transform: uppercase;
}

.signal-scope span,
.signal-scope strong {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.signal-scope strong {
  color: var(--accent);
  font-weight: 800;
}

.order-ticket-grid {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
  padding: 0.9rem;
}

.order-ticket-grid--signal {
  align-items: stretch;
  grid-template-columns: minmax(0, 1fr);
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

.ticket-input input,
.ticket-input select {
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

.signal-message {
  grid-column: 1 / -1;
  margin: 0;
  border: 1px solid var(--line);
  background: var(--charcoal);
  border-radius: 4px;
  padding: 0.65rem 0.75rem;
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.68rem;
}

.signal-message--success {
  border-color: rgba(0, 192, 135, 0.34);
  color: #00c087;
}

.signal-message--error {
  border-color: rgba(246, 70, 93, 0.38);
  color: #f6465d;
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

.submit-order:disabled {
  cursor: not-allowed;
  opacity: 0.62;
}

.submit-order--buy {
  background: #00c087;
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

  .signal-scope {
    flex-wrap: wrap;
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
  .bottom-tabs {
    overflow-x: auto;
  }

  .chart-wrapper {
    min-height: 260px;
  }
}
</style>
