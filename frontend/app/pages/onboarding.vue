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

const currentStep = ref(1)
const txid = ref('')
const txidShake = ref(false)
const walletAddress = '0xA6574277ABF624DDfd442c5D35B3d7c342416989'

const selectedExchanges = ref<string[]>([])
const submitAttempted = ref(false)
const depositShake = ref(false)
const ageShake = ref(false)
const countryShake = ref(false)
const exchangeShake = ref(false)

const exchangeOptions = [
  { id: 'Binance', logo: '/UserDashboard/Binance_logo.svg' },
  { id: 'OKX', logo: '/UserDashboard/OKX_logo_dark.svg' },
  { id: 'Bybit', logo: '/UserDashboard/Bybit_logo_dark.svg' },
  { id: 'Tokocrypto', logo: '/UserDashboard/Tokocrypto_logo.svg' }
]

const selectedCountryData = computed(() => {
  return countries.find(country => country.code === selectedCountry.value)
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
const exchangeInvalid = computed(() => selectedExchanges.value.length === 0)
const onboardingBlocked = computed(() => countryInvalid.value || ageInvalid.value || exchangeInvalid.value)
const txidInvalid = computed(() => currentStep.value === 2 && !txid.value.trim())

const toggleExchange = (exchange: string) => {
  if (selectedExchanges.value.includes(exchange)) {
    selectedExchanges.value = selectedExchanges.value.filter(item => item !== exchange)
    return
  }

  selectedExchanges.value = [...selectedExchanges.value, exchange]
}

const triggerShake = (target: 'country' | 'age' | 'deposit' | 'exchange') => {
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

  if (target === 'exchange') {
    exchangeShake.value = false
    window.requestAnimationFrame(() => {
      exchangeShake.value = true
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

const { completeOnboarding } = useAuth()

const handleDepositBlur = () => {
  if (!depositAmount.value || depositAmount.value < 500) {
    depositAmount.value = 500
    triggerShake('deposit')
  }
}

const copyWalletAddress = async () => {
  try {
    await navigator.clipboard.writeText(walletAddress)
  } catch (err) {
    console.error('Failed to copy', err)
  }
}

const triggerTxidShake = () => {
  txidShake.value = false
  window.requestAnimationFrame(() => {
    txidShake.value = true
  })
}

const nextStep = () => {
  submitAttempted.value = true

  if (countryInvalid.value) triggerShake('country')
  if (ageInvalid.value) triggerShake('age')
  if (exchangeInvalid.value) triggerShake('exchange')

  if (onboardingBlocked.value) return

  submitAttempted.value = false
  currentStep.value = 2
}

const submitPayment = async () => {
  submitAttempted.value = true

  if (depositInvalid.value) triggerShake('deposit')
  if (txidInvalid.value) triggerTxidShake()

  if (depositInvalid.value || txidInvalid.value) return

  try {
    await completeOnboarding({
      age: Number(age.value),
      countryCode: selectedCountry.value,
      exchanges: selectedExchanges.value,
      amount: String(depositAmount.value),
      gasFeeAsset: 'USDT',
      txId: txid.value
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
        <p v-if="currentStep === 1">
          Welcome to Mautrade
        </p>
        <p v-else>
          Gas Fee Deposit
        </p>
        <h1 v-if="currentStep === 1">
          Configure Your Trading Environment
        </h1>
        <h1 v-else>
          Deposit Required Gas Fee
        </h1>
      </div>

      <!-- Step 1 -->
      <form
        v-if="currentStep === 1"
        class="onboarding-form"
        @submit.prevent="nextStep"
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
                  id="countrySearch"
                  v-model="countrySearch"
                  name="countrySearch"
                  aria-label="Search country"
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
          <label for="age">Age Verification</label>
          <input
            id="age"
            v-model.number="age"
            name="age"
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
          <div
            class="exchange-choice-grid"
            :class="{ 'is-invalid-grid': submitAttempted && exchangeInvalid, 'is-shaking': exchangeShake }"
            @animationend="exchangeShake = false"
          >
            <button
              v-for="exchange in exchangeOptions"
              :key="exchange.id"
              class="exchange-choice"
              :class="{ 'is-selected': selectedExchanges.includes(exchange.id) }"
              type="button"
              @click="toggleExchange(exchange.id)"
            >
              <img
                :src="exchange.logo"
                :alt="`${exchange.id} logo`"
              >
            </button>
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

      <!-- Step 2 -->
      <form
        v-if="currentStep === 2"
        class="onboarding-form payment-step"
        @submit.prevent="submitPayment"
      >
        <div class="payment-instructions">
          <p>
            To initialize your dashboard, please deposit a minimum of <strong>{{ depositAmount }} USDT</strong> (BEP-20) to the following address:
          </p>
        </div>

        <div class="qr-container">
          <img
            :src="`https://api.qrserver.com/v1/create-qr-code/?size=180x180&data=${walletAddress}&color=FF5A00&bgcolor=000`"
            alt="Wallet QR Code"
            class="qr-image"
          >
        </div>

        <div class="wallet-address-container">
          <label>Wallet Address</label>
          <div class="wallet-input-group">
            <input
              class="onboarding-input wallet-input"
              type="text"
              readonly
              :value="walletAddress"
            >
            <button
              class="copy-button"
              type="button"
              @click="copyWalletAddress"
            >
              <UIcon name="lucide:copy" />
            </button>
          </div>
        </div>

        <div class="onboarding-field">
          <label for="depositAmount">Deposit Amount</label>
          <div
            class="deposit-input"
            :class="{ 'is-invalid': submitAttempted && depositInvalid, 'is-shaking': depositShake }"
            @animationend="depositShake = false"
          >
            <input
              id="depositAmount"
              v-model.number="depositAmount"
              name="depositAmount"
              type="number"
              min="500"
              step="1"
              @blur="handleDepositBlur"
            >
            <span>
              <img
                src="/UserDashboard/USDT_logo.svg"
                alt="USDT"
                style="width: 24px; height: 24px;"
              >
            </span>
          </div>
        </div>

        <div class="onboarding-field">
          <label for="txidInput">Transaction ID (TXID)</label>
          <input
            id="txidInput"
            v-model="txid"
            type="text"
            class="onboarding-input"
            :class="{ 'is-invalid': submitAttempted && txidInvalid, 'is-shaking': txidShake }"
            placeholder="Enter the transaction hash"
            @animationend="txidShake = false"
          >
        </div>

        <div class="payment-actions">
          <button
            type="button"
            class="btn-back"
            @click="currentStep = 1"
          >
            Back
          </button>
          <button
            class="onboarding-submit"
            type="submit"
          >
            Verify Payment
            <UIcon name="lucide:check" />
          </button>
        </div>
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

.is-invalid-grid {
  border-radius: 6px;
  box-shadow: 0 0 0 1px #ef4444;
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

/* Payment Step UI */
.payment-step {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.payment-instructions {
  color: var(--text-mute);
  font-size: 0.95rem;
  line-height: 1.5;
  text-align: center;
}

.payment-instructions strong {
  color: var(--accent);
}

.qr-container {
  display: flex;
  justify-content: center;
  margin: 1rem 0;
}

.qr-image {
  border-radius: 8px;
  border: 1px solid var(--line);
  background: #fff;
  padding: 8px;
  width: 180px;
  height: 180px;
}

.wallet-address-container label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  color: var(--text-mute);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.wallet-input-group {
  display: flex;
  background: var(--charcoal);
  border: 1px solid var(--line);
  border-radius: 4px;
  overflow: hidden;
}

.wallet-input {
  flex: 1;
  border: none;
  background: transparent;
  color: var(--text);
  padding: 0.8rem 1rem;
  font-family: var(--mono);
  font-size: 0.9rem;
  outline: none;
}

.copy-button {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 1rem;
  background: rgba(255, 90, 0, 0.1);
  border: none;
  border-left: 1px solid var(--line);
  color: var(--accent);
  cursor: pointer;
  transition: all 0.2s ease;
}

.copy-button:hover {
  background: var(--accent);
  color: #000;
}

.copy-button svg {
  width: 18px;
  height: 18px;
}

.payment-actions {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 1rem;
  margin-top: 1rem;
}

.btn-back {
  padding: 0 1.5rem;
  background: transparent;
  border: 1px solid var(--line);
  color: var(--text-mute);
  border-radius: 4px;
  font-family: var(--mono);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-back:hover {
  background: var(--charcoal);
  color: var(--text);
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

  .onboarding-panel {
    padding: 2rem;
  }

  .onboarding-side {
    display: none;
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
