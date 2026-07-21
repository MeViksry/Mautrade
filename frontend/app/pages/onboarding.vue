<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useCountries } from '~/composables/useCountries'

definePageMeta({
  layout: 'default'
})

const seoTitle = 'Onboarding | Mautrade'
const seoDescription = 'Complete Mautrade onboarding with country profile, age, exchange preference, and initial gas fee deposit.'

useSeoMeta({
  title: seoTitle,
  description: seoDescription,
  ogTitle: seoTitle,
  ogDescription: seoDescription,
  twitterTitle: seoTitle,
  twitterDescription: seoDescription
})

const { countries } = useCountries()

const countrySearch = ref('')
const countryDropdownOpen = ref(false)
const countrySelectRef = ref<HTMLElement | null>(null)
const selectedCountry = ref('ID')
const age = ref<number | null>(null)
const depositAmount = ref(500)
const depositCoinDropdownOpen = ref(false)
const depositCoinSelectRef = ref<HTMLElement | null>(null)
const selectedDepositCoin = ref('USDT')
const selectedExchanges = ref<string[]>([])
const submitAttempted = ref(false)
const depositShake = ref(false)
const ageShake = ref(false)
const countryShake = ref(false)

const exchangeOptions = [
  { id: 'Binance', logo: '/UserDashboard/Binance_logo.svg' },
  { id: 'OKX', logo: '/UserDashboard/OKX_logo_dark.svg' },
  { id: 'Bybit', logo: '/UserDashboard/Bybit_logo_dark.svg' },
  { id: 'Tokocrypto', logo: '/UserDashboard/Tokocrypto_logo.svg' }
]

const depositCoinOptions = [
  { code: 'USDT', name: 'Tether USD', network: 'TRC20 / ERC20 / BEP20', min: 500, icon: '/UserDashboard/USDT_logo.svg' }
]

const selectedCountryData = computed(() => {
  return countries.find(country => country.code === selectedCountry.value)
})

const selectedDepositCoinData = computed(() => {
  return depositCoinOptions.find(coin => coin.code === selectedDepositCoin.value) ?? depositCoinOptions[0]
})

const countrySearchTerm = computed(() => countrySearch.value.trim().toLowerCase())

const filteredCountries = computed(() => {
  if (!countrySearchTerm.value) return countries

  return countries.filter((country) => {
    return `${country.name} ${country.code}`.toLowerCase().includes(countrySearchTerm.value)
  })
})

const countryInvalid = computed(() => !selectedCountry.value)
const ageInvalid = computed(() => !age.value || age.value < 18)
const depositInvalid = computed(() => Number(depositAmount.value) < 500)
const onboardingBlocked = computed(() => countryInvalid.value || ageInvalid.value || depositInvalid.value || selectedExchanges.value.length === 0)

const toggleExchange = (exchange: string) => {
  if (selectedExchanges.value.includes(exchange)) {
    selectedExchanges.value = selectedExchanges.value.filter(item => item !== exchange)
    return
  }

  selectedExchanges.value = [...selectedExchanges.value, exchange]
}

const triggerShake = (target: 'country' | 'age' | 'deposit') => {
  if (target === 'country') {
    countryShake.value = false
    window.requestAnimationFrame(() => {
      countryShake.value = true
    })
  }

  if (target === 'age') {
    ageShake.value = false
    window.requestAnimationFrame(() => {
      ageShake.value = true
    })
  }

  if (target === 'deposit') {
    depositShake.value = false
    window.requestAnimationFrame(() => {
      depositShake.value = true
    })
  }
}

const selectCountry = (countryCode: string) => {
  selectedCountry.value = countryCode
  countrySearch.value = ''
  countryDropdownOpen.value = false
}

const selectDepositCoin = (coinCode: string) => {
  selectedDepositCoin.value = coinCode
  depositCoinDropdownOpen.value = false
}

const handleClickOutside = (event: MouseEvent) => {
  if (!countrySelectRef.value?.contains(event.target as Node)) {
    countryDropdownOpen.value = false
  }

  if (!depositCoinSelectRef.value?.contains(event.target as Node)) {
    depositCoinDropdownOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})

const { completeOnboarding } = useAuth()

const submitOnboarding = async () => {
  submitAttempted.value = true

  if (countryInvalid.value) triggerShake('country')
  if (ageInvalid.value) triggerShake('age')
  if (depositInvalid.value) triggerShake('deposit')

  if (onboardingBlocked.value) return

  try {
    await completeOnboarding({
      age: Number(age.value),
      countryCode: selectedCountry.value,
      exchanges: selectedExchanges.value,
      amount: String(depositAmount.value),
      gasFeeAsset: selectedDepositCoin.value
    })
    await navigateTo('/dashboard')
  } catch (error) {
    console.error('Failed to complete onboarding:', error)
    // could display error toast here
  }
}
</script>

<template>
  <main class="onboarding-page">
    <section class="onboarding-panel">
      <div class="onboarding-brand">
        MAUTRADE<span />
      </div>

      <div class="onboarding-heading">
        <p>Welcome to Mautrade</p>
        <h1>Configure Your Trading Environment</h1>
      </div>

      <form
        class="onboarding-form"
        @submit.prevent="submitOnboarding"
      >
        <div class="onboarding-field">
          <label>Region & Compliance</label>
          <div
            ref="countrySelectRef"
            class="country-select"
            :class="{ 'is-invalid': submitAttempted && countryInvalid, 'is-shaking': countryShake }"
            @animationend="countryShake = false"
          >
            <button
              class="country-select__trigger"
              type="button"
              @click="countryDropdownOpen = !countryDropdownOpen"
            >
              <span>
                <em>{{ selectedCountryData?.flag }}</em>
                {{ selectedCountryData?.name }}
              </span>
              <UIcon name="lucide:chevrons-up-down" />
            </button>

            <div
              v-if="countryDropdownOpen"
              class="country-select__dropdown"
            >
              <div class="country-search">
                <UIcon name="lucide:search" />
                <input
                  v-model="countrySearch"
                  type="text"
                  placeholder="Search country for compliance sync"
                  autocomplete="off"
                >
              </div>

              <div class="country-list">
                <button
                  v-for="country in filteredCountries"
                  :key="country.code"
                  class="country-option"
                  :class="{ 'is-selected': country.code === selectedCountry }"
                  type="button"
                  @click="selectCountry(country.code)"
                >
                  <span>
                    <em>{{ country.flag }}</em>
                    {{ country.name }}
                  </span>
                  <strong>{{ country.code }}</strong>
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="onboarding-field">
          <label>Age Verification</label>
          <input
            v-model.number="age"
            class="onboarding-input"
            :class="{ 'is-invalid': submitAttempted && ageInvalid, 'is-shaking': ageShake }"
            type="number"
            min="18"
            placeholder="e.g., 25"
            @animationend="ageShake = false"
          >
        </div>

        <div class="onboarding-field">
          <label>Frequently Used Exchanges</label>
          <div class="exchange-choice-grid">
            <button
              v-for="exchange in exchangeOptions"
              :key="exchange.id"
              class="exchange-choice"
              :class="{ 'is-selected': selectedExchanges.includes(exchange.id) }"
              type="button"
              @click="toggleExchange(exchange.id)"
            >
              <img
                v-if="exchange.logo"
                :src="exchange.logo"
                :alt="`${exchange.id} logo`"
              >
              <span v-else>{{ exchange.label }}</span>
            </button>
          </div>
        </div>

        <div class="onboarding-field">
          <label>Required Gas Fee Allocation</label>
          <div class="deposit-compose">
            <div
              class="deposit-input"
              :class="{ 'is-invalid': submitAttempted && depositInvalid, 'is-shaking': depositShake }"
              @animationend="depositShake = false"
            >
              <input
                v-model.number="depositAmount"
                type="number"
                min="500"
                step="1"
              >
              <span>{{ selectedDepositCoin }}</span>
            </div>

            <div
              ref="depositCoinSelectRef"
              class="coin-select"
            >
              <button
                class="coin-select__trigger"
                type="button"
                :aria-expanded="depositCoinDropdownOpen"
                aria-haspopup="listbox"
                @click="depositCoinDropdownOpen = !depositCoinDropdownOpen"
              >
                <span class="coin-select__asset">
                  <img
                    v-if="selectedDepositCoinData?.icon"
                    :src="selectedDepositCoinData.icon"
                    :alt="`${selectedDepositCoinData.code} logo`"
                  >
                  <strong>{{ selectedDepositCoinData?.code }}</strong>
                  <small>{{ selectedDepositCoinData?.network }}</small>
                </span>
                <UIcon name="lucide:chevrons-up-down" />
              </button>

              <div
                v-if="depositCoinDropdownOpen"
                class="coin-select__dropdown"
                role="listbox"
              >
                <button
                  v-for="coin in depositCoinOptions"
                  :key="coin.code"
                  class="coin-option"
                  :class="{ 'is-selected': coin.code === selectedDepositCoin }"
                  type="button"
                  role="option"
                  :aria-selected="coin.code === selectedDepositCoin"
                  @click="selectDepositCoin(coin.code)"
                >
                  <span class="coin-option__identity">
                    <img
                      v-if="coin.icon"
                      :src="coin.icon"
                      :alt="`${coin.code} logo`"
                    >
                    <span>
                      <strong>{{ coin.code }}</strong>
                      <small>{{ coin.name }}</small>
                    </span>
                  </span>
                  <em>Min {{ coin.min }}</em>
                </button>
              </div>
            </div>
          </div>
        </div>

        <button
          class="onboarding-submit"
          :class="{ 'is-blocked': onboardingBlocked }"
          type="submit"
        >
          Initialize Trading Dashboard
          <UIcon name="lucide:arrow-right" />
        </button>
      </form>
    </section>

    <section class="onboarding-side">
      <div class="onboarding-card">
        <UIcon name="lucide:shield-check" />
        <h2>Seamless Trading Synchronization</h2>
        <p>We collect these essential details to securely synchronize your exchange APIs, calibrate algorithmic trading layers, and establish your dedicated gas fee pool for automated trade execution.</p>
      </div>
    </section>
  </main>
</template>

<style scoped>
.onboarding-page {
  min-height: 100vh;
  display: grid;
  grid-template-columns: minmax(560px, 0.95fr) minmax(420px, 1fr);
  background:
    linear-gradient(180deg, rgba(255, 90, 0, 0.05), transparent 26%),
    var(--bg);
}

.onboarding-panel {
  display: flex;
  flex-direction: column;
  padding: 3rem 5rem;
}

.onboarding-brand {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--text);
  font-family: 'Oswald', sans-serif;
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 0.15em;
}

.onboarding-brand span {
  width: 6px;
  height: 6px;
  background: var(--accent);
  display: inline-block;
}

.onboarding-heading {
  margin-top: 5rem;
}

.onboarding-heading p {
  margin: 0 0 0.75rem;
  color: var(--accent);
  font-family: var(--mono);
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.onboarding-heading h1 {
  max-width: 640px;
  margin: 0;
  color: var(--text);
  font-family: var(--sans);
  font-size: clamp(2.4rem, 5vw, 4.4rem);
  font-weight: 300;
  line-height: 1;
}

.onboarding-form {
  display: flex;
  flex-direction: column;
  gap: 1.4rem;
  max-width: 760px;
  margin-top: 2.6rem;
}

.onboarding-field {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.onboarding-field label {
  font-family: var(--mono);
  font-size: 12px;
  color: var(--silver);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.onboarding-input,
.country-select__trigger,
.deposit-input {
  width: 100%;
  min-height: 52px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--bg-elevated);
  color: var(--text);
}

.onboarding-input {
  outline: none;
  padding: 0 1rem;
}

.country-select {
  position: relative;
}

.country-select__trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0 1rem;
  text-align: left;
}

.country-select__trigger span,
.country-option span {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 0;
}

.country-select__trigger em,
.country-option em {
  flex: 0 0 auto;
  font-size: 1.35rem;
  font-style: normal;
}

.country-select__dropdown {
  position: absolute;
  z-index: 30;
  top: calc(100% + 0.55rem);
  left: 0;
  right: 0;
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--bg-elevated);
  box-shadow: 0 24px 60px rgba(0, 0, 0, 0.52);
}

.country-search {
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr);
  align-items: center;
  gap: 0.65rem;
  padding: 0.85rem 1rem;
  border-bottom: 1px solid var(--line);
  background: var(--charcoal);
}

.country-search input {
  min-width: 0;
  border: none;
  outline: none;
  background: transparent;
  color: var(--text);
  font-family: var(--mono);
}

.country-list {
  max-height: 292px;
  overflow-y: auto;
  scrollbar-color: var(--accent) var(--charcoal);
  scrollbar-width: thin;
}

.country-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  width: 100%;
  min-height: 44px;
  padding: 0 1rem;
  border: none;
  border-bottom: 1px solid var(--line);
  background: var(--bg-elevated);
  color: var(--text);
  text-align: left;
}

.country-option:hover,
.country-option.is-selected {
  background: rgba(255, 90, 0, 0.12);
  color: var(--accent);
}

.country-option strong {
  font-family: var(--mono);
  font-size: 11px;
}

.exchange-choice-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.8rem;
}

.exchange-choice {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 58px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--bg-elevated);
  color: var(--text);
  font-family: var(--mono);
  font-size: 12px;
  font-weight: 800;
}

.exchange-choice img {
  display: block;
  max-width: 92px;
  max-height: 26px;
  object-fit: contain;
}

.exchange-choice svg {
  position: absolute;
  top: -6px;
  right: 8px;
  width: 18px;
  height: 18px;
  padding: 3px;
  border-radius: 50%;
  background: var(--accent);
  color: #050505;
  opacity: 0;
}

.exchange-choice.is-selected {
  border-color: var(--accent);
  background: rgba(255, 90, 0, 0.12);
}

.exchange-choice.is-selected svg {
  opacity: 1;
}

.deposit-input {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  overflow: hidden;
}

.deposit-compose {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(220px, 0.34fr);
  gap: 0.75rem;
}

.deposit-input input {
  width: 100%;
  min-width: 0;
  border: none;
  outline: none;
  background: transparent;
  color: var(--text);
  padding: 0 1rem;
}

.deposit-input span {
  display: inline-flex;
  align-items: center;
  padding: 0 1rem;
  border-left: 1px solid var(--line);
  color: var(--accent);
  font-family: var(--mono);
  font-size: 12px;
  font-weight: 800;
}

.coin-select {
  position: relative;
  min-width: 0;
}

.coin-select__trigger {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 18px;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  min-height: 52px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--bg-elevated);
  color: var(--text);
  padding: 0 0.85rem;
  text-align: left;
}

.coin-select__trigger:hover,
.coin-select__trigger[aria-expanded='true'] {
  border-color: var(--accent);
  background: rgba(255, 90, 0, 0.08);
}

.coin-select__trigger svg {
  color: var(--accent);
}

.coin-select__asset,
.coin-option__identity {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  min-width: 0;
}

.coin-select__asset img,
.coin-option__identity img {
  width: 24px;
  height: 24px;
  object-fit: contain;
}

.coin-select__asset strong,
.coin-option strong {
  display: block;
  color: var(--text);
  font-family: var(--mono);
  font-size: 12px;
  font-weight: 900;
  line-height: 1;
}

.coin-select__asset small,
.coin-option small {
  display: block;
  overflow: hidden;
  max-width: 128px;
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 9px;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.coin-select__dropdown {
  position: absolute;
  z-index: 32;
  top: calc(100% + 0.5rem);
  right: 0;
  width: min(320px, 84vw);
  overflow: hidden;
  border: 1px solid rgba(255, 90, 0, 0.34);
  border-radius: 4px;
  background: var(--bg-elevated);
  box-shadow: 0 24px 60px rgba(0, 0, 0, 0.45);
}

.coin-option {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.8rem;
  width: 100%;
  min-height: 52px;
  padding: 0 0.85rem;
  border: none;
  border-bottom: 1px solid var(--line);
  background: var(--bg-elevated);
  color: var(--text);
  text-align: left;
}

.coin-option:last-child {
  border-bottom: none;
}

.coin-option:hover,
.coin-option.is-selected {
  background: rgba(255, 90, 0, 0.12);
}

.coin-option em {
  color: var(--accent);
  font-family: var(--mono);
  font-size: 10px;
  font-style: normal;
  font-weight: 900;
  white-space: nowrap;
}

.is-invalid {
  border-color: #ef4444;
}

.is-shaking {
  animation: onboarding-shake 260ms ease-in-out;
}

.onboarding-submit {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  width: 100%;
  min-height: 58px;
  margin-top: 0.6rem;
  border: none;
  border-radius: 4px;
  background: var(--accent);
  color: #050505;
  font-family: var(--mono);
  font-weight: 900;
  cursor: pointer;
  box-shadow: 0 18px 30px rgba(255, 90, 0, 0.18);
}

.onboarding-submit:hover {
  background: #ff7a1a;
}

.onboarding-submit.is-blocked {
  box-shadow: inset 0 0 0 1px rgba(239, 68, 68, 0.35);
}

.onboarding-side {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  border-left: 1px solid var(--line);
  background:
    linear-gradient(var(--grid-line) 1px, transparent 1px),
    linear-gradient(90deg, var(--grid-line) 1px, transparent 1px),
    var(--bg-elevated);
  background-size: 52px 52px;
}

.onboarding-card {
  max-width: 520px;
  padding: 2.2rem;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--charcoal);
}

.onboarding-card svg {
  width: 42px;
  height: 42px;
  color: var(--accent);
}

.onboarding-card h2 {
  margin: 1.5rem 0 0;
  color: var(--text);
  font-size: 2.4rem;
}

.onboarding-card p {
  margin-top: 1rem;
  color: var(--text-mute);
  line-height: 1.7;
}

@keyframes onboarding-shake {
  0%, 100% {
    transform: translateX(0);
  }

  20% {
    transform: translateX(-7px);
  }

  40% {
    transform: translateX(7px);
  }

  60% {
    transform: translateX(-5px);
  }

  80% {
    transform: translateX(5px);
  }
}

@media (max-width: 1040px) {
  .onboarding-page {
    grid-template-columns: 1fr;
  }

  .onboarding-panel,
  .onboarding-side {
    padding: 2rem;
  }
}

@media (max-width: 640px) {
  .onboarding-panel,
  .onboarding-side {
    padding: 1.25rem;
  }

  .exchange-choice-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .deposit-compose {
    grid-template-columns: 1fr;
  }

  .coin-select__dropdown {
    left: 0;
    right: 0;
    width: 100%;
  }
}
</style>
