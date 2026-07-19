<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

type CoinPairOption = {
  symbol: string
  name: string
  price: string
  change: string
}

const props = withDefaults(defineProps<{
  modelValue: string
  options: CoinPairOption[]
  label?: string
  compact?: boolean
}>(), {
  label: 'Select Coin',
  compact: false
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const isOpen = ref(false)
const search = ref('')
const selectRef = ref<HTMLElement | null>(null)

const getBaseAsset = (symbol = '') => symbol.split('/')[0] ?? ''
const getQuoteAsset = (symbol = '') => symbol.split('/')[1] ?? ''

const selectedCoin = computed(() => {
  return props.options.find(coin => coin.symbol === props.modelValue) ?? props.options[0]
})

const selectedBaseAsset = computed(() => getBaseAsset(selectedCoin.value?.symbol))
const selectedQuoteAsset = computed(() => getQuoteAsset(selectedCoin.value?.symbol))
const searchTerm = computed(() => search.value.trim().toLowerCase())

const filteredOptions = computed(() => {
  if (!searchTerm.value) return props.options

  return props.options.filter((coin) => {
    return `${coin.symbol} ${coin.name}`.toLowerCase().includes(searchTerm.value)
  })
})

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
    :class="{ 'coin-pair-select--compact': compact }"
  >
    <button
      class="coin-pair-select__trigger"
      type="button"
      :aria-expanded="isOpen"
      aria-haspopup="listbox"
      @click="isOpen = !isOpen"
      @keydown.escape="isOpen = false"
    >
      <span class="coin-pair-select__mark">
        {{ selectedBaseAsset.slice(0, 3) }}
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
        <span>Coin Market</span>
        <strong>{{ selectedQuoteAsset || 'USDT' }} Pair</strong>
      </div>

      <label class="coin-pair-select__search">
        <UIcon name="lucide:search" />
        <input
          v-model="search"
          type="text"
          placeholder="Search coin or pair"
          autocomplete="off"
          spellcheck="false"
        >
      </label>

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
          <span class="coin-pair-select__option-mark">
            {{ getBaseAsset(coin.symbol).slice(0, 3) }}
          </span>
          <span class="coin-pair-select__option-main">
            <strong>{{ coin.symbol }}</strong>
            <small>{{ coin.name }}</small>
          </span>
          <span class="coin-pair-select__option-market">
            <strong>{{ coin.price }}</strong>
            <small :class="{ 'is-negative': coin.change.startsWith('-') }">
              {{ coin.change }}
            </small>
          </span>
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
  z-index: 14;
  width: min(100%, 360px);
  min-width: 260px;
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
  border: 1px solid rgba(255, 90, 0, 0.34);
  background: rgba(255, 90, 0, 0.12);
  color: var(--accent);
  font-family: var(--mono);
  font-size: 0.68rem;
  font-weight: 900;
  letter-spacing: 0.04em;
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
.coin-pair-select__option-market small {
  color: #00c087;
  font-family: var(--mono);
  font-size: 0.65rem;
  font-weight: 800;
}

.coin-pair-select__market small.is-negative,
.coin-pair-select__option-market small.is-negative {
  color: #f6465d;
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
  top: calc(100% + 0.4rem);
  left: 0;
  width: min(420px, calc(100vw - 2rem));
  overflow: hidden;
  border: 1px solid rgba(255, 90, 0, 0.34);
  background: var(--bg-elevated);
  border-radius: 4px;
  box-shadow: 0 22px 56px rgba(0, 0, 0, 0.42);
  padding: 0.45rem;
}

.coin-pair-select__menu-head {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.25rem 0.35rem 0.55rem;
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 0.62rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.coin-pair-select__menu-head strong {
  color: var(--accent);
  font-weight: 900;
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

.coin-pair-select__list {
  max-height: 322px;
  overflow-y: auto;
  padding-right: 0.15rem;
  scrollbar-color: var(--accent) var(--charcoal);
  scrollbar-width: thin;
}

.coin-pair-select__option {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto 16px;
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
}
</style>
