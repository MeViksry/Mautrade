<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

type CoinPairOption = {
  symbol: string
  name: string
  price: string
  change: string
  volume?: string
}

type CoinPairCategory = 'all' | 'major' | 'meme' | 'alt'

const props = withDefaults(defineProps<{
  modelValue: string
  options: CoinPairOption[]
  label?: string
  compact?: boolean
  fullWidth?: boolean
}>(), {
  label: 'Select Coin',
  compact: false,
  fullWidth: false
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const isOpen = ref(false)
const search = ref('')
const activeCategory = ref<CoinPairCategory>('all')
const activeQuote = ref('all')
const selectRef = ref<HTMLElement | null>(null)
const searchInputRef = ref<HTMLInputElement | null>(null)

const getBaseAsset = (symbol = '') => symbol.split('/')[0] ?? ''
const getQuoteAsset = (symbol = '') => symbol.split('/')[1] ?? ''
const majorAssets = ['BTC', 'ETH', 'BNB', 'SOL', 'XRP']
const memeAssets = ['DOGE', 'PEPE', 'SHIB']
const coinColors: Record<string, string> = {
  BTC: '#f7931a',
  ETH: '#627eea',
  SOL: '#14f195',
  BNB: '#f3ba2f',
  PEPE: '#6fd15a',
  XRP: '#9aa4b2',
  DOGE: '#c2a633',
  ADA: '#2a71d0',
  AVAX: '#e84142',
  LINK: '#2a5ada',
  DOT: '#e6007a',
  LTC: '#345d9d',
  SHIB: '#f05d23',
  TRX: '#ef0027',
  ARB: '#28a0f0',
  OP: '#ff0420',
  NEAR: '#00c08b',
  SUI: '#4da2ff'
}
const coinIcons: Record<string, string> = {
  BTC: 'simple-icons:bitcoin',
  ETH: 'simple-icons:ethereum',
  SOL: 'simple-icons:solana',
  BNB: 'simple-icons:binance',
  XRP: 'simple-icons:ripple',
  DOGE: 'simple-icons:dogecoin',
  ADA: 'simple-icons:cardano',
  LINK: 'simple-icons:chainlink',
  DOT: 'simple-icons:polkadot',
  LTC: 'simple-icons:litecoin',
  OP: 'simple-icons:optimism',
  NEAR: 'simple-icons:near',
  SUI: 'simple-icons:sui'
}
const customCoinLogoSvgs: Record<string, string> = {
  PEPE: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><circle cx="32" cy="32" r="30" fill="#69c95a"/><path fill="#b8ef83" d="M14 31c0-12 8-21 20-21 11 0 19 8 19 20 0 14-10 23-22 23-10 0-17-8-17-22Z"/><circle cx="25" cy="27" r="8" fill="#fff"/><circle cx="41" cy="27" r="8" fill="#fff"/><circle cx="27" cy="29" r="3.5" fill="#101820"/><circle cx="43" cy="29" r="3.5" fill="#101820"/><path fill="none" stroke="#101820" stroke-linecap="round" stroke-width="4" d="M22 42c6 5 15 5 22 0"/><path fill="#2b7c38" opacity=".72" d="M15 24c6-11 15-15 28-10-14 1-24 5-28 10Z"/></svg>`,
  AVAX: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><circle cx="32" cy="32" r="30" fill="#e84142"/><path fill="#fff" d="M30 17c1-2 4-2 5 0l17 30c1 2 0 4-3 4H15c-3 0-4-2-3-4l18-30Zm2 10L22 44h20L32 27Z"/><path fill="#e84142" d="M32 27 22 44h20L32 27Z"/><path fill="#fff" d="M44 35c1-2 3-2 4 0l7 12c1 2 0 4-2 4h-13c-2 0-3-2-2-4l6-12Z"/></svg>`,
  SHIB: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><circle cx="32" cy="32" r="30" fill="#f05d23"/><path fill="#ffe6b7" d="M16 26 13 13l13 6c4-2 8-2 12 0l13-6-3 13c4 6 4 15-1 21-7 8-23 8-30 0-5-6-5-15-1-21Z"/><path fill="#8b2d15" d="m15 15 12 11-14 2 2-13Zm34 0-12 11 14 2-2-13Z"/><circle cx="25" cy="32" r="3.5" fill="#101820"/><circle cx="39" cy="32" r="3.5" fill="#101820"/><path fill="#101820" d="M32 37 27 34h10l-5 3Z"/><path fill="none" stroke="#101820" stroke-linecap="round" stroke-width="3" d="M24 42c5 5 11 5 16 0"/></svg>`,
  TRX: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><circle cx="32" cy="32" r="30" fill="#ef0027"/><path fill="none" stroke="#fff" stroke-linejoin="round" stroke-width="4" d="M14 13 51 21 29 52 14 13Z"/><path fill="none" stroke="#fff" stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M14 13 30 29m21-8-21 8m-1 23 1-23"/></svg>`,
  ARB: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><path fill="#213147" d="M32 4 56 18v28L32 60 8 46V18L32 4Z"/><path fill="#28a0f0" d="M32 9 51 20v23L32 54 13 43V20L32 9Z"/><path fill="#fff" d="m29 46 5-1 11-25-5-3-11 29Zm-10-9 4 3 10-23-4-2-10 22Z"/><path fill="#213147" d="m38 47 4-2-12-26-4 2 12 26Z"/></svg>`
}
const customCoinLogoSources = Object.fromEntries(
  Object.entries(customCoinLogoSvgs).map(([asset, svg]) => [asset, `data:image/svg+xml,${encodeURIComponent(svg)}`])
) as Record<string, string>

const categoryOptions: Array<{ id: CoinPairCategory, label: string }> = [
  { id: 'all', label: 'All' },
  { id: 'major', label: 'Major' },
  { id: 'meme', label: 'Meme' },
  { id: 'alt', label: 'Alt' }
]

const selectedCoin = computed(() => {
  return props.options.find(coin => coin.symbol === props.modelValue) ?? props.options[0]
})

const selectedBaseAsset = computed(() => getBaseAsset(selectedCoin.value?.symbol))
const selectedQuoteAsset = computed(() => getQuoteAsset(selectedCoin.value?.symbol))
const searchTerm = computed(() => search.value.trim().toLowerCase())
const quoteOptions = computed(() => {
  const quotes = new Set(props.options.map(coin => getQuoteAsset(coin.symbol)).filter(Boolean))

  return ['all', ...Array.from(quotes)]
})

const quickOptions = computed(() => {
  const preferredSymbols = ['BTC/USDT', 'ETH/USDT', 'SOL/USDT', 'BNB/USDT', 'XRP/USDT', 'DOGE/USDT']
  const preferredOptions = preferredSymbols
    .map(symbol => props.options.find(coin => coin.symbol === symbol))
    .filter((coin): coin is CoinPairOption => Boolean(coin))
  const fallbackOptions = props.options.filter(coin => !preferredSymbols.includes(coin.symbol))

  return [...preferredOptions, ...fallbackOptions].slice(0, 6)
})

const filteredOptions = computed(() => {
  const quoteFilteredOptions = props.options.filter((coin) => {
    if (activeQuote.value === 'all') return true

    return getQuoteAsset(coin.symbol) === activeQuote.value
  })

  const categoryFilteredOptions = quoteFilteredOptions.filter((coin) => {
    const baseAsset = getBaseAsset(coin.symbol)

    if (activeCategory.value === 'major') return majorAssets.includes(baseAsset)
    if (activeCategory.value === 'meme') return memeAssets.includes(baseAsset)
    if (activeCategory.value === 'alt') return !majorAssets.includes(baseAsset) && !memeAssets.includes(baseAsset)

    return true
  })

  if (!searchTerm.value) return categoryFilteredOptions

  return categoryFilteredOptions.filter((coin) => {
    return `${coin.symbol} ${coin.name}`.toLowerCase().includes(searchTerm.value)
  })
})

const selectedTokenColor = computed(() => getCoinColor(selectedBaseAsset.value))

const getCoinColor = (asset = '') => {
  return coinColors[asset] ?? '#ff5a00'
}

const getCoinIconName = (asset = '') => {
  return coinIcons[asset] ?? ''
}

const getCustomCoinLogoSrc = (asset = '') => {
  return customCoinLogoSources[asset] ?? ''
}

const chooseCoin = (symbol: string) => {
  emit('update:modelValue', symbol)
  search.value = ''
  isOpen.value = false
}

const closeOnOutsideClick = (event: MouseEvent) => {
  if (!selectRef.value?.contains(event.target as Node)) {
    isOpen.value = false
  }
}

watch(isOpen, async (open) => {
  if (!open) return

  await nextTick()
  searchInputRef.value?.focus()
})

onMounted(() => {
  document.addEventListener('click', closeOnOutsideClick)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', closeOnOutsideClick)
})
</script>

<template>
  <div
    ref="selectRef"
    class="coin-pair-select"
    :class="{ 'coin-pair-select--compact': compact, 'coin-pair-select--full': fullWidth, 'coin-pair-select--open': isOpen }"
  >
    <button
      class="coin-pair-select__trigger"
      type="button"
      :aria-expanded="isOpen"
      aria-haspopup="listbox"
      @click="isOpen = !isOpen"
      @keydown.enter.prevent="isOpen = !isOpen"
      @keydown.escape="isOpen = false"
    >
      <span
        class="coin-pair-select__mark"
        :style="{ '--coin-color': selectedTokenColor }"
      >
        <UIcon
          v-if="getCoinIconName(selectedBaseAsset)"
          :name="getCoinIconName(selectedBaseAsset)"
          class="coin-pair-select__icon"
        />
        <img
          v-else-if="getCustomCoinLogoSrc(selectedBaseAsset)"
          :src="getCustomCoinLogoSrc(selectedBaseAsset)"
          :alt="`${selectedBaseAsset} logo`"
          class="coin-pair-select__logo-image"
        >
        <strong v-else>{{ selectedBaseAsset.slice(0, 1) }}</strong>
      </span>
      <span class="coin-pair-select__main">
        <span>{{ label }}</span>
        <strong>{{ selectedCoin?.symbol }}</strong>
        <small>{{ selectedCoin?.name }}</small>
      </span>
      <span class="coin-pair-select__market">
        <strong>{{ selectedCoin?.price }}</strong>
        <small :class="{ 'is-negative': (selectedCoin?.change ?? '').startsWith('-') }">
          {{ selectedCoin?.change }}
        </small>
      </span>
      <UIcon
        name="lucide:chevron-down"
        class="coin-pair-select__chevron"
        :class="{ 'coin-pair-select__chevron--open': isOpen }"
      />
    </button>

    <div
      v-if="isOpen"
      class="coin-pair-select__menu"
      role="listbox"
    >
      <div class="coin-pair-select__menu-head">
        <div>
          <span>Coin Market</span>
          <strong>{{ filteredOptions.length }} Pairs</strong>
        </div>
        <em>Spot {{ selectedQuoteAsset || 'USDT' }}</em>
      </div>

      <label class="coin-pair-select__search">
        <UIcon name="lucide:search" />
        <input
          ref="searchInputRef"
          v-model="search"
          type="text"
          placeholder="Search coin or pair"
          autocomplete="off"
          spellcheck="false"
        >
      </label>

      <div
        class="coin-pair-select__quick"
        aria-label="Quick coin selection"
      >
        <button
          v-for="coin in quickOptions"
          :key="`quick-${coin.symbol}`"
          type="button"
          :class="{ 'is-active': coin.symbol === modelValue }"
          @click="chooseCoin(coin.symbol)"
        >
          <span
            class="coin-pair-select__quick-mark"
            :style="{ '--coin-color': getCoinColor(getBaseAsset(coin.symbol)) }"
          >
            <UIcon
              v-if="getCoinIconName(getBaseAsset(coin.symbol))"
              :name="getCoinIconName(getBaseAsset(coin.symbol))"
            />
            <img
              v-else-if="getCustomCoinLogoSrc(getBaseAsset(coin.symbol))"
              :src="getCustomCoinLogoSrc(getBaseAsset(coin.symbol))"
              :alt="`${getBaseAsset(coin.symbol)} logo`"
              class="coin-pair-select__logo-image"
            >
            <strong v-else>{{ getBaseAsset(coin.symbol).slice(0, 1) }}</strong>
          </span>
          <strong>{{ getBaseAsset(coin.symbol) }}</strong>
          <small :class="{ 'is-negative': coin.change.startsWith('-') }">{{ coin.change }}</small>
        </button>
      </div>

      <div class="coin-pair-select__filters">
        <div
          class="coin-pair-select__quote-tabs"
          aria-label="Quote market"
        >
          <button
            v-for="quote in quoteOptions"
            :key="quote"
            type="button"
            :class="{ 'is-active': activeQuote === quote }"
            @click="activeQuote = quote"
          >
            {{ quote === 'all' ? 'All' : quote }}
          </button>
        </div>

        <div
          class="coin-pair-select__tabs"
          aria-label="Coin categories"
        >
          <button
            v-for="category in categoryOptions"
            :key="category.id"
            type="button"
            :class="{ 'is-active': activeCategory === category.id }"
            @click="activeCategory = category.id"
          >
            {{ category.label }}
          </button>
        </div>
      </div>

      <div class="coin-pair-select__table-head">
        <span>Pair</span>
        <span>Last Price</span>
        <span>24H</span>
      </div>

      <div class="coin-pair-select__list">
        <button
          v-for="coin in filteredOptions"
          :key="coin.symbol"
          class="coin-pair-select__option"
          :class="{ 'is-selected': coin.symbol === modelValue }"
          type="button"
          role="option"
          :aria-selected="coin.symbol === modelValue"
          @click="chooseCoin(coin.symbol)"
        >
          <span
            class="coin-pair-select__option-mark"
            :style="{ '--coin-color': getCoinColor(getBaseAsset(coin.symbol)) }"
          >
            <UIcon
              v-if="getCoinIconName(getBaseAsset(coin.symbol))"
              :name="getCoinIconName(getBaseAsset(coin.symbol))"
              class="coin-pair-select__icon"
            />
            <img
              v-else-if="getCustomCoinLogoSrc(getBaseAsset(coin.symbol))"
              :src="getCustomCoinLogoSrc(getBaseAsset(coin.symbol))"
              :alt="`${getBaseAsset(coin.symbol)} logo`"
              class="coin-pair-select__logo-image"
            >
            <strong v-else>{{ getBaseAsset(coin.symbol).slice(0, 1) }}</strong>
          </span>
          <span class="coin-pair-select__option-main">
            <strong>{{ coin.symbol }}</strong>
            <small>{{ coin.name }}{{ coin.volume ? ` / Vol ${coin.volume}` : '' }}</small>
          </span>
          <span class="coin-pair-select__option-market">
            <strong>{{ coin.price }}</strong>
          </span>
          <small
            class="coin-pair-select__change"
            :class="{ 'is-negative': coin.change.startsWith('-') }"
          >
            {{ coin.change }}
          </small>
          <UIcon
            v-if="coin.symbol === modelValue"
            name="lucide:check"
            class="coin-pair-select__check"
          />
        </button>

        <div
          v-if="filteredOptions.length === 0"
          class="coin-pair-select__empty"
        >
          No coin found
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.coin-pair-select {
  position: relative;
  z-index: 40;
  width: min(100%, 360px);
  min-width: 260px;
}

.coin-pair-select--open {
  z-index: 260;
}

.coin-pair-select__trigger {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto auto;
  align-items: center;
  gap: 0.7rem;
  width: 100%;
  min-height: 56px;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text);
  border-radius: 4px;
  padding: 0 0.8rem;
  text-align: left;
  transition: border-color 180ms var(--ease-quiet), background 180ms var(--ease-quiet);
}

.coin-pair-select__trigger:hover,
.coin-pair-select__trigger[aria-expanded='true'] {
  border-color: rgba(255, 90, 0, 0.58);
  background: rgba(255, 90, 0, 0.08);
}

.coin-pair-select__mark,
.coin-pair-select__option-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border: 1px solid color-mix(in srgb, var(--coin-color, var(--accent)) 58%, var(--line));
  background: color-mix(in srgb, var(--coin-color, var(--accent)) 18%, var(--bg-elevated));
  color: color-mix(in srgb, var(--coin-color, var(--accent)) 78%, var(--text));
  font-family: var(--mono);
  font-size: 0.72rem;
  font-weight: 900;
}

.coin-pair-select__icon {
  width: 20px;
  height: 20px;
  color: currentColor;
}

.coin-pair-select__logo-image {
  display: block;
  width: 24px;
  height: 24px;
  object-fit: contain;
}

.coin-pair-select__mark strong,
.coin-pair-select__option-mark strong {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 999px;
  background: currentColor;
  color: var(--bg-elevated);
  font-size: 0.66rem;
  line-height: 1;
}

.coin-pair-select__main,
.coin-pair-select__market,
.coin-pair-select__option-main,
.coin-pair-select__option-market {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.coin-pair-select__main > span {
  color: var(--accent);
  font-family: var(--mono);
  font-size: 0.58rem;
  font-weight: 900;
  letter-spacing: 0.12em;
  line-height: 1;
  text-transform: uppercase;
}

.coin-pair-select__main strong {
  color: var(--text);
  font-family: 'Oswald', sans-serif;
  font-size: 1.08rem;
  font-weight: 500;
  line-height: 1.05;
  white-space: nowrap;
}

.coin-pair-select__main small,
.coin-pair-select__option-main small {
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.66rem;
  text-transform: uppercase;
}

.coin-pair-select__market,
.coin-pair-select__option-market {
  align-items: flex-end;
}

.coin-pair-select__market strong,
.coin-pair-select__option-market strong {
  color: var(--text);
  font-family: var(--mono);
  font-size: 0.75rem;
  font-weight: 700;
  white-space: nowrap;
}

.coin-pair-select__market small,
.coin-pair-select__change {
  color: #00c087;
  font-family: var(--mono);
  font-size: 0.65rem;
  font-weight: 800;
}

.coin-pair-select__market small.is-negative,
.coin-pair-select__change.is-negative {
  color: #f6465d;
}

.coin-pair-select__change {
  justify-self: end;
  white-space: nowrap;
}

.coin-pair-select__chevron {
  width: 16px;
  height: 16px;
  color: var(--text-mute);
  transition: transform 180ms var(--ease-quiet), color 180ms var(--ease-quiet);
}

.coin-pair-select__chevron--open {
  color: var(--accent);
  transform: rotate(180deg);
}

.coin-pair-select__menu {
  position: absolute;
  z-index: 260;
  top: calc(100% + 0.4rem);
  left: 0;
  width: min(520px, calc(100vw - 2rem));
  overflow: hidden;
  border: 1px solid rgba(255, 90, 0, 0.34);
  background: var(--bg-elevated);
  border-radius: 4px;
  box-shadow: 0 22px 56px rgba(0, 0, 0, 0.42);
  padding: 0.45rem;
}

.coin-pair-select__menu-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.25rem 0.35rem 0.55rem;
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.62rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.coin-pair-select__menu-head div {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.coin-pair-select__menu-head strong,
.coin-pair-select__menu-head em {
  color: var(--accent);
  font-weight: 900;
  font-style: normal;
}

.coin-pair-select__search {
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr);
  align-items: center;
  gap: 0.55rem;
  margin-bottom: 0.4rem;
  min-height: 40px;
  padding: 0 0.65rem;
  border: 1px solid var(--line);
  background: var(--charcoal);
}

.coin-pair-select__search svg {
  color: var(--accent);
}

.coin-pair-select__search input {
  width: 100%;
  min-width: 0;
  border: 0;
  background: transparent;
  color: var(--text);
  font-family: var(--mono);
  font-size: 0.72rem;
  outline: none;
}

.coin-pair-select__search input::placeholder {
  color: var(--text-mute);
}

.coin-pair-select__quick {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.35rem;
  margin-bottom: 0.45rem;
}

.coin-pair-select__quick button {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.4rem;
  min-height: 34px;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text);
  border-radius: 4px;
  padding: 0 0.45rem;
  text-align: left;
  transition: border-color 160ms var(--ease-quiet), background 160ms var(--ease-quiet), color 160ms var(--ease-quiet);
}

.coin-pair-select__quick button:hover,
.coin-pair-select__quick button.is-active {
  border-color: rgba(255, 90, 0, 0.46);
  background: rgba(255, 90, 0, 0.13);
  color: var(--accent);
}

.coin-pair-select__quick button > span {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border: 1px solid color-mix(in srgb, var(--coin-color, var(--accent)) 48%, var(--line));
  background: color-mix(in srgb, var(--coin-color, var(--accent)) 14%, var(--bg-elevated));
  color: color-mix(in srgb, var(--coin-color, var(--accent)) 82%, var(--text));
}

.coin-pair-select__quick button > span svg {
  width: 12px;
  height: 12px;
}

.coin-pair-select__quick .coin-pair-select__logo-image {
  width: 14px;
  height: 14px;
}

.coin-pair-select__quick button > span strong {
  font-size: 0.52rem;
  line-height: 1;
}

.coin-pair-select__quick strong,
.coin-pair-select__quick small {
  min-width: 0;
  overflow: hidden;
  font-family: var(--mono);
  font-size: 0.66rem;
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.coin-pair-select__quick small {
  color: #00c087;
  justify-self: end;
}

.coin-pair-select__quick small.is-negative {
  color: #f6465d;
}

.coin-pair-select__filters {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.45rem;
  margin-bottom: 0.45rem;
}

.coin-pair-select__quote-tabs,
.coin-pair-select__tabs {
  display: grid;
  gap: 0.3rem;
}

.coin-pair-select__quote-tabs {
  grid-auto-flow: column;
  grid-auto-columns: minmax(52px, max-content);
  overflow-x: auto;
  scrollbar-width: none;
}

.coin-pair-select__quote-tabs::-webkit-scrollbar {
  display: none;
}

.coin-pair-select__tabs {
  grid-template-columns: repeat(4, minmax(50px, 1fr));
}

.coin-pair-select__quote-tabs button,
.coin-pair-select__tabs button {
  height: 30px;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text-mute);
  border-radius: 4px;
  font-family: var(--mono);
  font-size: 0.64rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  transition: border-color 160ms var(--ease-quiet), background 160ms var(--ease-quiet), color 160ms var(--ease-quiet);
}

.coin-pair-select__quote-tabs button:hover,
.coin-pair-select__quote-tabs button.is-active,
.coin-pair-select__tabs button:hover,
.coin-pair-select__tabs button.is-active {
  border-color: rgba(255, 90, 0, 0.42);
  background: rgba(255, 90, 0, 0.14);
  color: var(--accent);
}

.coin-pair-select__table-head {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 96px 62px;
  gap: 0.6rem;
  padding: 0.35rem 0.55rem;
  border-top: 1px solid var(--line);
  border-bottom: 1px solid var(--line);
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.58rem;
  text-transform: uppercase;
}

.coin-pair-select__table-head span:nth-child(2),
.coin-pair-select__table-head span:nth-child(3) {
  text-align: right;
}

.coin-pair-select__list {
  max-height: 322px;
  overflow-y: auto;
  padding-right: 0.15rem;
  scrollbar-color: var(--accent) var(--charcoal);
  scrollbar-width: thin;
}

.coin-pair-select__option {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) 96px 62px 16px;
  align-items: center;
  gap: 0.7rem;
  width: 100%;
  min-height: 56px;
  border: 1px solid transparent;
  background: transparent;
  color: var(--text);
  border-radius: 4px;
  padding: 0 0.65rem;
  text-align: left;
  transition: border-color 160ms var(--ease-quiet), background 160ms var(--ease-quiet);
}

.coin-pair-select__option:hover,
.coin-pair-select__option.is-selected {
  border-color: rgba(255, 90, 0, 0.34);
  background: rgba(255, 90, 0, 0.08);
}

.coin-pair-select__option-main strong {
  color: var(--text);
  font-family: var(--mono);
  font-size: 0.78rem;
  font-weight: 800;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.coin-pair-select__option-main small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.coin-pair-select__check {
  width: 16px;
  height: 16px;
  color: var(--accent);
}

.coin-pair-select__empty {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 64px;
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.72rem;
}

.coin-pair-select--compact {
  width: min(100%, 300px);
  min-width: 230px;
}

.coin-pair-select--full {
  width: 100%;
  min-width: 0;
}

.coin-pair-select--compact .coin-pair-select__trigger {
  min-height: 42px;
}

.coin-pair-select--compact .coin-pair-select__mark {
  width: 28px;
  height: 28px;
  font-size: 0.58rem;
}

.coin-pair-select--compact .coin-pair-select__main small,
.coin-pair-select--compact .coin-pair-select__market {
  display: none;
}

.coin-pair-select--compact .coin-pair-select__menu {
  width: min(460px, calc(100vw - 2rem));
}

@media (max-width: 640px) {
  .coin-pair-select {
    width: 100%;
    min-width: 0;
  }

  .coin-pair-select__trigger {
    grid-template-columns: auto minmax(0, 1fr) auto;
  }

  .coin-pair-select__market {
    display: none;
  }

  .coin-pair-select__menu {
    width: min(100vw - 2rem, 420px);
  }

  .coin-pair-select__filters {
    grid-template-columns: 1fr;
  }

  .coin-pair-select__quick {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .coin-pair-select__tabs {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .coin-pair-select__table-head {
    grid-template-columns: minmax(0, 1fr) 80px 58px;
  }

  .coin-pair-select__option {
    grid-template-columns: auto minmax(0, 1fr) 80px 58px;
  }

  .coin-pair-select__check {
    display: none;
  }
}
</style>
