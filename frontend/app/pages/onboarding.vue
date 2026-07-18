<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useCountries } from '~/composables/useCountries'

definePageMeta({
  layout: 'default'
})

const { countries } = useCountries()

const countrySearch = ref('')
const countryDropdownOpen = ref(false)
const countrySelectRef = ref<HTMLElement | null>(null)
const selectedCountry = ref('ID')
const age = ref<number | null>(null)
const depositAmount = ref(500)
const selectedExchanges = ref<string[]>(['Binance'])
const submitAttempted = ref(false)
const depositShake = ref(false)
const ageShake = ref(false)
const countryShake = ref(false)

const exchangeOptions = [
  { id: 'Binance', logo: '/UserDashboard/Binance_logo.svg' },
  { id: 'OKX', logo: '/UserDashboard/OKX_logo_dark.svg' },
  { id: 'Bybit', logo: '/UserDashboard/Bybit_logo_dark.svg' },
  { id: 'Tokocrypto', logo: '/UserDashboard/Tokocrypto_logo.svg' },
  { id: 'KuCoin', label: 'KuCoin' },
  { id: 'Coinbase', label: 'Coinbase' },
  { id: 'Kraken', label: 'Kraken' },
  { id: 'Bitget', label: 'Bitget' }
]

const selectedCountryData = computed(() => {
  return countries.find((country) => country.code === selectedCountry.value)
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
    selectedExchanges.value = selectedExchanges.value.filter((item) => item !== exchange)
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

const handleClickOutside = (event: MouseEvent) => {
  if (!countrySelectRef.value?.contains(event.target as Node)) {
    countryDropdownOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})

const submitOnboarding = async () => {
  submitAttempted.value = true

  if (countryInvalid.value) triggerShake('country')
  if (ageInvalid.value) triggerShake('age')
  if (depositInvalid.value) triggerShake('deposit')

  if (onboardingBlocked.value) return

  await navigateTo('/dashboard')
}
</script>

<template>
  <main class="onboarding-page">
    <section class="onboarding-panel">
      <div class="onboarding-brand">
        MAUTRADE<span />
      </div>

      <div class="onboarding-heading">
        <p>Mautrade Onboarding</p>
        <h1>Prepare your dashboard profile</h1>
      </div>

      <form
        class="onboarding-form"
        @submit.prevent="submitOnboarding"
      >
        <div class="onboarding-field">
          <label>Country Profile</label>
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
                  placeholder="Search country for timezone sync"
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
          <label>User Age</label>
          <input
            v-model.number="age"
            class="onboarding-input"
            :class="{ 'is-invalid': submitAttempted && ageInvalid, 'is-shaking': ageShake }"
            type="number"
            min="18"
            placeholder="Enter user age"
            @animationend="ageShake = false"
          >
        </div>

        <div class="onboarding-field">
          <label>Exchange Preference</label>
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
              <UIcon name="lucide:check" />
            </button>
          </div>
        </div>

        <div class="onboarding-field">
          <label>Initial Gas Fee Deposit</label>
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
            <span>USDT</span>
          </div>
        </div>

        <button
          class="onboarding-submit"
          :class="{ 'is-blocked': onboardingBlocked }"
          type="submit"
        >
          Enter Mautrade Dashboard
          <UIcon name="lucide:arrow-right" />
        </button>
      </form>
    </section>

    <section class="onboarding-side">
      <div class="onboarding-card">
        <UIcon name="lucide:clock-3" />
        <h2>Profile data for dashboard sync</h2>
        <p>Country, age, exchange preference, and minimum gas fee are prepared here so Mautrade can align Active Layers, Trading History, and Gas Fee records to the user account.</p>
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
    linear-gradient(rgba(255, 255, 255, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.035) 1px, transparent 1px),
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
}
</style>
