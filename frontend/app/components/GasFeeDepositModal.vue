<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'submitted': [payload: { amount: number, coin: string, txId: string }]
}>()

const depositStep = ref<'methods' | 'deposit'>('methods')
const depositAmount = ref(500)
const depositCoinDropdownOpen = ref(false)
const depositCoinSelectRef = ref<HTMLElement | null>(null)
const selectedDepositCoin = ref('USDT')
const depositTxId = ref('')
const walletCopied = ref(false)
const depositAmountShake = ref(false)
const depositTxIdShake = ref(false)
const depositSubmitAttempted = ref(false)
const depositSubmitted = ref(false)
const depositWalletAddress = '0xA6574277ABF624DDfd442c5D35B3d7c342416989'
const depositCoinOptions = [
  { code: 'USDT', name: 'Tether USD', network: 'BEP-20', min: 500, icon: '/UserDashboard/USDT_logo.svg' },
  { code: 'USDC', name: 'USD Coin', network: 'ERC20 / Base', min: 500 },
  { code: 'FDUSD', name: 'First Digital USD', network: 'BNB Smart Chain', min: 500 }
]

const selectedDepositCoinData = computed(() => {
  return depositCoinOptions.find(coin => coin.code === selectedDepositCoin.value) ?? depositCoinOptions[0]
})

const resetDepositState = () => {
  depositStep.value = 'methods'
  depositAmount.value = 500
  selectedDepositCoin.value = 'USDT'
  depositCoinDropdownOpen.value = false
  depositTxId.value = ''
  walletCopied.value = false
  depositAmountShake.value = false
  depositTxIdShake.value = false
  depositSubmitAttempted.value = false
  depositSubmitted.value = false
}

watch(() => props.modelValue, (isOpen) => {
  if (isOpen) {
    resetDepositState()
  }
})

const closeDepositModal = () => {
  emit('update:modelValue', false)
}

const selectDepositMethod = () => {
  depositStep.value = 'deposit'
  depositSubmitAttempted.value = false
  depositSubmitted.value = false
}

const selectDepositCoin = (coinCode: string) => {
  selectedDepositCoin.value = coinCode
  depositCoinDropdownOpen.value = false
}

const handleClickOutside = (event: MouseEvent) => {
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

const copyDepositAddress = async () => {
  await navigator.clipboard.writeText(depositWalletAddress)
  walletCopied.value = true
  window.setTimeout(() => {
    walletCopied.value = false
  }, 1400)
}

const depositQrSvg = computed(() => {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="240" height="240" viewBox="0 0 240 240"><rect width="240" height="240" fill="#f8fafc"/><rect x="20" y="20" width="54" height="54" fill="#111"/><rect x="32" y="32" width="30" height="30" fill="#f8fafc"/><rect x="42" y="42" width="10" height="10" fill="#111"/><rect x="166" y="20" width="54" height="54" fill="#111"/><rect x="178" y="32" width="30" height="30" fill="#f8fafc"/><rect x="188" y="42" width="10" height="10" fill="#111"/><rect x="20" y="166" width="54" height="54" fill="#111"/><rect x="32" y="178" width="30" height="30" fill="#f8fafc"/><rect x="42" y="188" width="10" height="10" fill="#111"/><rect x="92" y="24" width="12" height="12" fill="#111"/><rect x="116" y="24" width="24" height="12" fill="#111"/><rect x="92" y="48" width="36" height="12" fill="#111"/><rect x="140" y="48" width="12" height="12" fill="#111"/><rect x="92" y="84" width="12" height="24" fill="#111"/><rect x="116" y="84" width="12" height="12" fill="#111"/><rect x="152" y="84" width="36" height="12" fill="#111"/><rect x="200" y="96" width="12" height="24" fill="#111"/><rect x="84" y="120" width="24" height="12" fill="#111"/><rect x="120" y="120" width="12" height="36" fill="#111"/><rect x="144" y="120" width="24" height="12" fill="#111"/><rect x="180" y="132" width="36" height="12" fill="#111"/><rect x="88" y="164" width="12" height="12" fill="#111"/><rect x="112" y="164" width="48" height="12" fill="#111"/><rect x="184" y="164" width="12" height="12" fill="#111"/><rect x="92" y="188" width="24" height="12" fill="#111"/><rect x="140" y="188" width="12" height="24" fill="#111"/><rect x="164" y="188" width="48" height="12" fill="#111"/><rect x="104" y="212" width="12" height="12" fill="#111"/><rect x="128" y="212" width="36" height="12" fill="#111"/><rect x="188" y="212" width="12" height="12" fill="#111"/></svg>`
  return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`
})

const depositAmountInvalid = computed(() => Number(depositAmount.value) < 500)
const depositTxIdInvalid = computed(() => depositTxId.value.trim().length === 0)
const depositTxIdErrorVisible = computed(() => depositSubmitAttempted.value && depositTxIdInvalid.value)
const depositFormBlocked = computed(() => depositAmountInvalid.value || depositTxIdInvalid.value)

const triggerDepositAmountShake = () => {
  depositAmountShake.value = false
  window.requestAnimationFrame(() => {
    depositAmountShake.value = true
  })
}

const triggerDepositTxIdShake = () => {
  depositTxIdShake.value = false
  window.requestAnimationFrame(() => {
    depositTxIdShake.value = true
  })
}

watch(depositAmount, () => {
  depositSubmitted.value = false
  if (depositAmountInvalid.value) {
    triggerDepositAmountShake()
  }
})

watch(depositTxId, () => {
  depositSubmitted.value = false
})

const submitDeposit = () => {
  depositSubmitAttempted.value = true

  if (depositAmountInvalid.value) {
    triggerDepositAmountShake()
  }

  if (depositTxIdInvalid.value) {
    triggerDepositTxIdShake()
  }

  if (depositFormBlocked.value) return

  depositSubmitted.value = true
  emit('submitted', {
    amount: Number(depositAmount.value),
    coin: selectedDepositCoin.value,
    txId: depositTxId.value.trim()
  })
}
</script>

<template>
  <div
    v-if="modelValue"
    class="deposit-modal"
    role="dialog"
    aria-modal="true"
    :aria-label="depositStep === 'methods' ? 'Deposit Methods' : 'Deposit'"
    @click.self="closeDepositModal"
  >
    <div class="deposit-modal__box">
      <div class="deposit-modal__header">
        <button
          v-if="depositStep === 'deposit'"
          class="deposit-modal__icon-btn"
          type="button"
          aria-label="Back to deposit methods"
          @click="depositStep = 'methods'"
        >
          <UIcon name="lucide:arrow-left" />
        </button>
        <span
          v-else
          class="deposit-modal__spacer"
        />
        <h3>{{ depositStep === 'methods' ? 'Deposit Methods' : 'Deposit' }}</h3>
        <button
          class="deposit-modal__icon-btn"
          type="button"
          aria-label="Close deposit modal"
          @click="closeDepositModal"
        >
          <UIcon name="lucide:x" />
        </button>
      </div>

      <div
        v-if="depositStep === 'methods'"
        class="deposit-methods"
      >
        <button
          class="deposit-method"
          type="button"
          @click="selectDepositMethod"
        >
          <span class="deposit-method__icon">
            <UIcon name="lucide:wallet" />
          </span>
          <span class="deposit-method__content">
            <span class="deposit-method__title">USDT Gas Fee Wallet</span>
            <span class="deposit-method__meta">Minimum deposit 500 USDT</span>
          </span>
          <UIcon
            name="lucide:chevron-right"
            class="deposit-method__arrow"
          />
        </button>
      </div>

      <div
        v-else
        class="deposit-form"
      >
        <div class="deposit-qr">
          <img
            :src="depositQrSvg"
            alt="USDT gas fee deposit QR code"
          >
        </div>

        <a
          class="deposit-download"
          :href="depositQrSvg"
          download="mautrade-gas-fee-deposit-qr.svg"
        >
          <UIcon name="lucide:download" />
          <span>Download QR Code</span>
        </a>

        <label class="deposit-field">
          <span>Deposit Coin</span>
          <div
            ref="depositCoinSelectRef"
            class="deposit-coin-select"
          >
            <button
              class="deposit-coin-select__trigger"
              type="button"
              :aria-expanded="depositCoinDropdownOpen"
              aria-haspopup="listbox"
              @click="depositCoinDropdownOpen = !depositCoinDropdownOpen"
            >
              <span class="deposit-coin-select__identity">
                <img
                  v-if="selectedDepositCoinData?.icon"
                  :src="selectedDepositCoinData.icon"
                  :alt="`${selectedDepositCoinData.code} logo`"
                >
                <span>
                  <strong>{{ selectedDepositCoinData?.code }}</strong>
                  <small>{{ selectedDepositCoinData?.network }}</small>
                </span>
              </span>
              <UIcon name="lucide:chevrons-up-down" />
            </button>

            <div
              v-if="depositCoinDropdownOpen"
              class="deposit-coin-select__dropdown"
              role="listbox"
            >
              <button
                v-for="coin in depositCoinOptions"
                :key="coin.code"
                class="deposit-coin-option"
                :class="{ 'is-selected': coin.code === selectedDepositCoin }"
                type="button"
                role="option"
                :aria-selected="coin.code === selectedDepositCoin"
                @click="selectDepositCoin(coin.code)"
              >
                <span class="deposit-coin-option__identity">
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
        </label>

        <label class="deposit-field">
          <span>Wallet Address</span>
          <div class="deposit-copy">
            <input
              :value="depositWalletAddress"
              readonly
            >
            <button
              type="button"
              @click="copyDepositAddress"
            >
              <UIcon :name="walletCopied ? 'lucide:check' : 'lucide:copy'" />
              <span>{{ walletCopied ? 'Copied' : 'Copy' }}</span>
            </button>
          </div>
        </label>

        <label class="deposit-field">
          <span>Amount</span>
          <div
            class="deposit-amount"
            :class="{ 'is-invalid': depositAmountInvalid, 'is-shaking': depositAmountShake }"
            @animationend="depositAmountShake = false"
          >
            <input
              v-model.number="depositAmount"
              type="number"
              min="500"
              step="1"
              aria-describedby="deposit-amount-error"
            >
            <span>{{ selectedDepositCoin }}</span>
          </div>
          <p
            v-if="depositAmountInvalid"
            id="deposit-amount-error"
            class="deposit-error"
          >
            Minimum deposit is 500 USDT
          </p>
        </label>

        <label class="deposit-field">
          <span>TX ID</span>
          <input
            v-model="depositTxId"
            class="deposit-tx-input"
            :class="{ 'is-invalid': depositTxIdErrorVisible, 'is-shaking': depositTxIdShake }"
            type="text"
            placeholder="Paste transaction ID"
            aria-describedby="deposit-tx-error"
            @animationend="depositTxIdShake = false"
          >
          <p
            v-if="depositTxIdErrorVisible"
            id="deposit-tx-error"
            class="deposit-error"
          >
            TX ID is required
          </p>
        </label>

        <button
          class="deposit-submit"
          :class="{ 'is-blocked': depositFormBlocked }"
          type="button"
          :aria-disabled="depositFormBlocked"
          @click="submitDeposit"
        >
          <UIcon name="lucide:send" />
          <span>Submit Deposit</span>
        </button>

        <p
          v-if="depositSubmitted"
          class="deposit-success"
        >
          Deposit submitted
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.deposit-modal {
  position: fixed;
  inset: 0;
  z-index: 80;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  background: rgba(0, 0, 0, 0.72);
  backdrop-filter: blur(10px);
}

.deposit-modal__box {
  width: min(520px, 100%);
  max-height: min(760px, calc(100vh - 4rem));
  overflow-y: auto;
  scrollbar-width: none;
  -ms-overflow-style: none;
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  box-shadow: 0 28px 70px rgba(0, 0, 0, 0.45);
}

.deposit-modal__box::-webkit-scrollbar {
  display: none;
}

.deposit-modal__header {
  display: grid;
  grid-template-columns: 36px 1fr 36px;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--line);
}

.deposit-modal__header h3 {
  margin: 0;
  font-family: 'Oswald', sans-serif;
  font-size: 1.45rem;
  font-weight: 400;
  color: var(--text);
  letter-spacing: 0.04em;
  text-align: center;
  text-transform: uppercase;
}

.deposit-modal__icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text);
  cursor: pointer;
  transition: border-color 220ms var(--ease-quiet), color 220ms var(--ease-quiet);
}

.deposit-modal__spacer {
  width: 36px;
  height: 36px;
}

.deposit-modal__icon-btn:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.deposit-methods,
.deposit-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.5rem;
}

.deposit-method {
  display: grid;
  grid-template-columns: 44px 1fr 24px;
  align-items: center;
  gap: 1rem;
  width: 100%;
  padding: 1rem;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text);
  cursor: pointer;
  text-align: left;
  transition: border-color 220ms var(--ease-quiet), background 220ms var(--ease-quiet);
}

.deposit-method:hover {
  border-color: var(--accent);
  background: rgba(255, 90, 0, 0.08);
}

.deposit-method__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  background: var(--accent);
  color: #000;
}

.deposit-method__content {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.deposit-method__title {
  font-family: 'Oswald', sans-serif;
  font-size: 1.1rem;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.deposit-method__meta {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.deposit-method__arrow {
  color: var(--accent);
}

.deposit-qr {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  border: 1px solid var(--line);
  background: var(--charcoal);
}

.deposit-qr img {
  display: block;
  width: 220px;
  height: 220px;
  object-fit: contain;
}

.deposit-download,
.deposit-copy button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border: 1px solid var(--accent);
  background: var(--accent);
  color: #000;
  font-family: var(--mono);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  cursor: pointer;
  transition: background 220ms var(--ease-quiet), border-color 220ms var(--ease-quiet);
}

.deposit-download {
  align-self: center;
  padding: 0.65rem 1rem;
}

.deposit-download:hover,
.deposit-copy button:hover {
  background: #ff7324;
  border-color: #ff7324;
}

.deposit-field {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.deposit-field > span {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.deposit-field input {
  width: 100%;
  min-width: 0;
  height: 42px;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text);
  font-family: var(--mono);
  font-size: 12px;
  outline: none;
  padding: 0 0.85rem;
}

.deposit-field input:focus {
  border-color: var(--accent);
}

.deposit-coin-select {
  position: relative;
}

.deposit-coin-select__trigger {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 18px;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  min-height: 44px;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text);
  padding: 0 0.85rem;
  text-align: left;
  transition: border-color 220ms var(--ease-quiet), background 220ms var(--ease-quiet);
}

.deposit-coin-select__trigger:hover,
.deposit-coin-select__trigger[aria-expanded='true'] {
  border-color: var(--accent);
  background: rgba(255, 90, 0, 0.08);
}

.deposit-coin-select__trigger > svg {
  color: var(--accent);
}

.deposit-coin-select__identity,
.deposit-coin-option__identity {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  min-width: 0;
}

.deposit-coin-select__identity img,
.deposit-coin-option__identity img {
  width: 24px;
  height: 24px;
  object-fit: contain;
}

.deposit-coin-select__identity strong,
.deposit-coin-option strong {
  display: block;
  color: var(--text);
  font-family: var(--mono);
  font-size: 12px;
  font-weight: 900;
  line-height: 1;
}

.deposit-coin-select__identity small,
.deposit-coin-option small {
  display: block;
  overflow: hidden;
  max-width: 280px;
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 9px;
  line-height: 1.25;
  text-overflow: ellipsis;
  text-transform: uppercase;
  white-space: nowrap;
}

.deposit-coin-select__dropdown {
  position: absolute;
  z-index: 4;
  top: calc(100% + 0.45rem);
  left: 0;
  right: 0;
  overflow: hidden;
  border: 1px solid rgba(255, 90, 0, 0.34);
  background: var(--bg-elevated);
  box-shadow: 0 22px 54px rgba(0, 0, 0, 0.44);
}

.deposit-coin-option {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  min-height: 50px;
  border: none;
  border-bottom: 1px solid var(--line);
  background: var(--bg-elevated);
  color: var(--text);
  padding: 0 0.85rem;
  text-align: left;
}

.deposit-coin-option:last-child {
  border-bottom: none;
}

.deposit-coin-option:hover,
.deposit-coin-option.is-selected {
  background: rgba(255, 90, 0, 0.12);
}

.deposit-coin-option em {
  color: var(--accent);
  font-family: var(--mono);
  font-size: 10px;
  font-style: normal;
  font-weight: 900;
  white-space: nowrap;
}

.deposit-copy,
.deposit-amount {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
}

.deposit-copy button {
  height: 42px;
  padding: 0 0.85rem;
  border-left: none;
}

.deposit-amount span {
  display: inline-flex;
  align-items: center;
  height: 42px;
  padding: 0 0.85rem;
  border: 1px solid var(--line);
  border-left: none;
  background: var(--charcoal);
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 11px;
}

.deposit-amount.is-invalid input,
.deposit-amount.is-invalid span {
  border-color: #ef4444;
}

.deposit-tx-input.is-invalid {
  border-color: #ef4444;
}

.deposit-amount.is-shaking,
.deposit-tx-input.is-shaking {
  animation: deposit-shake 260ms ease-in-out;
}

.deposit-error,
.deposit-success {
  margin: -0.2rem 0 0;
  font-family: var(--mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.deposit-error {
  color: #ef4444;
}

.deposit-success {
  color: #10b981;
  text-align: center;
}

.deposit-submit {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  width: 100%;
  height: 44px;
  border: 1px solid var(--accent);
  background: var(--accent);
  color: #000;
  font-family: var(--mono);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  cursor: pointer;
  transition: background 220ms var(--ease-quiet), border-color 220ms var(--ease-quiet), transform 220ms var(--ease-quiet);
}

.deposit-submit:hover {
  background: #ff7324;
  border-color: #ff7324;
  transform: translateY(-1px);
}

.deposit-submit.is-blocked {
  box-shadow: inset 0 0 0 1px rgba(239, 68, 68, 0.45);
}

@keyframes deposit-shake {
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

@media (max-width: 640px), (max-height: 740px) {
  .deposit-modal {
    padding: 0.5rem;
  }

  .deposit-modal__box {
    max-height: 98vh;
  }

  .deposit-modal__header {
    padding: 0.75rem 1rem;
    gap: 0.25rem;
  }

  .deposit-modal__header h3 {
    font-size: 1rem;
  }

  .deposit-modal__icon-btn,
  .deposit-modal__spacer {
    width: 28px;
    height: 28px;
  }

  .deposit-methods,
  .deposit-form {
    padding: 0.75rem 1rem;
    gap: 0.5rem;
  }

  .deposit-method {
    padding: 0.5rem 0.75rem;
    gap: 0.5rem;
  }

  .deposit-method__icon {
    width: 32px;
    height: 32px;
  }

  .deposit-method__title {
    font-size: 0.9rem;
  }

  .deposit-qr {
    padding: 0.5rem;
  }

  .deposit-qr img {
    width: 120px;
    height: 120px;
  }

  .deposit-download {
    padding: 0.4rem 0.75rem;
    font-size: 9px;
  }

  .deposit-field {
    gap: 0.2rem;
  }

  .deposit-field > span {
    font-size: 9px;
  }

  .deposit-field input {
    height: 34px;
    font-size: 11px;
    padding: 0 0.5rem;
  }

  .deposit-copy button,
  .deposit-amount span {
    height: 34px;
    font-size: 9px;
    padding: 0 0.5rem;
  }

  .deposit-copy {
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 0;
  }

  .deposit-submit {
    height: 36px;
    font-size: 10px;
  }

  .deposit-error,
  .deposit-success {
    font-size: 9px;
  }
}
</style>
