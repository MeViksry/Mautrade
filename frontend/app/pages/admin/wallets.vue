<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import StatCard from '~/components/StatCard.vue'

definePageMeta({
  layout: 'admin'
})

const seoTitle = 'Wallets Management | Admin Mautrade'
const seoDescription = 'Manage company wallets and view balances.'
useSeoMeta({
  title: seoTitle,
  description: seoDescription
})
const loading = ref(true)
const config = useRuntimeConfig()
const apiBase = config.public.apiBase
const { tokenCookie } = useAdminAuth()

type AdminPersonalWallet = {
  code: string
  displayName: string
  walletAddress: string
  shareRate?: string
  balance?: string | number
  availableBalance?: string
  commissionBalance?: string
  pendingWithdrawalBalance?: string
  withdrawnBalance?: string
  updatedBy?: string
  createdAt?: string
  updatedAt?: string
}

type AdminPersonalWalletWithdrawal = {
  id: string
  walletCode: string
  destinationAddress: string
  amount: string
  asset: string
  status: string
  txId?: string
  failureReason?: string
  requestedAt: string
  updatedAt: string
}

type PersonalWalletCard = AdminPersonalWallet & {
  balance: number
  balanceText: string
  shareRate: string
}

const defaultPersonalWallets: PersonalWalletCard[] = [
  {
    code: 'viksry',
    displayName: 'WALLET VIKSRY',
    walletAddress: '',
    shareRate: '0.10',
    balance: 0,
    balanceText: '0'
  },
  {
    code: 'aryanto_hong',
    displayName: 'WALLET ARYANTO HONG',
    walletAddress: '',
    shareRate: '0.90',
    balance: 0,
    balanceText: '0'
  }
]

const walletStats = ref({
  totalBalance: 0,
  dailyInflow: 0,
  dailyOutflow: 0,
  activeWallets: 0
})

const personalWallets = ref<PersonalWalletCard[]>(defaultPersonalWallets.map(wallet => ({ ...wallet })))
const gasFeeWalletActive = ref(false)
const walletAddressModalOpen = ref(false)
const selectedWallet = ref<PersonalWalletCard | null>(null)
const walletAddressInput = ref('')
const walletAddressError = ref('')
const walletAddressSaving = ref(false)
const withdrawModalOpen = ref(false)
const withdrawWallet = ref<PersonalWalletCard | null>(null)
const withdrawAmountInput = ref('')
const withdrawAddressInput = ref('')
const withdrawError = ref('')
const withdrawSuccess = ref('')
const withdrawSubmitting = ref(false)

const walletAddressPattern = /^0x[a-fA-F0-9]{40}$/

const walletAddressInvalid = computed(() => {
  const value = walletAddressInput.value.trim()
  return value.length > 0 && !walletAddressPattern.test(value)
})

const canSaveWalletAddress = computed(() => {
  return selectedWallet.value !== null
    && walletAddressInput.value.trim().length > 0
    && !walletAddressInvalid.value
    && !walletAddressSaving.value
})

const withdrawAddressInvalid = computed(() => {
  const value = withdrawAddressInput.value.trim()
  return value.length > 0 && !walletAddressPattern.test(value)
})

const decimalToAtomicUnits = (value: string) => {
  let normalized = value.trim()
  if (normalized.startsWith('+')) normalized = normalized.slice(1)
  if (!/^\d+(\.\d{1,18})?$/.test(normalized)) return null

  const [integerPart, fractionPart = ''] = normalized.split('.')
  try {
    return BigInt(integerPart || '0') * 10n ** 18n + BigInt(fractionPart.padEnd(18, '0') || '0')
  } catch {
    return null
  }
}

const withdrawAmountUnits = computed(() => decimalToAtomicUnits(withdrawAmountInput.value))

const withdrawBalanceUnits = computed(() => {
  if (!withdrawWallet.value) return 0n
  return decimalToAtomicUnits(withdrawWallet.value.balanceText) || 0n
})

const withdrawAmountValidationMessage = computed(() => {
  const value = withdrawAmountInput.value.trim()
  if (!value) return ''
  const amount = withdrawAmountUnits.value
  if (amount === null || amount <= 0n) return 'Enter a valid USDT amount.'
  if (amount > withdrawBalanceUnits.value) return 'Amount exceeds wallet balance.'
  return ''
})

const canSubmitWithdrawal = computed(() => {
  const amount = withdrawAmountUnits.value
  return withdrawWallet.value !== null
    && withdrawAddressInput.value.trim().length > 0
    && !withdrawAddressInvalid.value
    && amount !== null
    && amount > 0n
    && amount <= withdrawBalanceUnits.value
    && !withdrawSubmitting.value
})

const parseWalletBalance = (value: string | number | undefined) => {
  if (typeof value === 'number') return Number.isFinite(value) ? value : 0
  if (!value) return 0

  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

const walletBalanceText = (value: string | number | undefined) => {
  if (typeof value === 'number') return Number.isFinite(value) ? String(value) : '0'
  const trimmed = value?.trim()
  return trimmed || '0'
}

const formatDecimalText = (value: string) => {
  const normalized = value.trim()
  if (!/^-?\d+(\.\d+)?$/.test(normalized)) return '0.00'

  const sign = normalized.startsWith('-') ? '-' : ''
  const unsigned = sign ? normalized.slice(1) : normalized
  const parts = unsigned.split('.')
  const rawInteger = parts[0] || '0'
  const rawFraction = parts[1] || ''
  const integer = (rawInteger.replace(/^0+(?=\d)/, '') || '0').replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  const fraction = rawFraction.replace(/0+$/, '')

  return `${sign}${integer}.${(fraction || '').padEnd(2, '0')}`
}

const recomputeActiveWallets = () => {
  walletStats.value.activeWallets = (gasFeeWalletActive.value ? 1 : 0) + personalWallets.value.filter(wallet => wallet.walletAddress.trim() !== '').length
}

const fetchGasFeeWalletBalance = async (gasFeeAddress: string) => {
  let balanceVal = 0

  if (gasFeeAddress && gasFeeAddress.startsWith('0x')) {
    try {
      const cleanAddress = gasFeeAddress.toLowerCase().replace('0x', '')
      const paddedAddress = cleanAddress.padStart(64, '0')
      const data = '0x70a08231' + paddedAddress

      const response = await fetch('https://bsc-dataseed.binance.org/', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          jsonrpc: '2.0',
          id: 1,
          method: 'eth_call',
          params: [
            {
              to: '0x55d398326f99059fF775485246999027B3197955', // USDT BEP20 contract on BSC
              data
            },
            'latest'
          ]
        })
      })
      const res = await response.json()
      if (res.result && res.result !== '0x') {
        const balanceBig = BigInt(res.result)
        const balanceStr = balanceBig.toString().padStart(19, '0')
        const intPart = balanceStr.slice(0, -18) || '0'
        const decPart = balanceStr.slice(-18)
        balanceVal = parseFloat(`${intPart}.${decPart}`)
      }
    } catch (err) {
      console.error('Failed to fetch USDT BEP20 balance:', err)
    }
  }

  return balanceVal
}

const fetchPersonalWallets = async () => {
  if (!tokenCookie.value) return

  try {
    const wallets = await $fetch<AdminPersonalWallet[]>(`${apiBase}/admin/personal-wallets`, {
      headers: {
        Authorization: `Bearer ${tokenCookie.value}`
      }
    })

    personalWallets.value = defaultPersonalWallets.map((wallet) => {
      const remoteWallet = wallets.find(item => item.code === wallet.code)
      const balanceText = walletBalanceText(remoteWallet?.availableBalance ?? remoteWallet?.balance ?? remoteWallet?.commissionBalance ?? wallet.balanceText)
      const balance = parseWalletBalance(balanceText)
      return {
        ...wallet,
        ...remoteWallet,
        walletAddress: remoteWallet?.walletAddress || '',
        shareRate: remoteWallet?.shareRate || wallet.shareRate,
        balance,
        balanceText
      }
    })
  } catch (err) {
    console.error('Failed to fetch personal wallets:', err)
  } finally {
    recomputeActiveWallets()
  }
}

const formatWalletAddress = (address: string) => {
  const value = address.trim()
  if (!value) return 'No address linked'
  if (value.length <= 14) return value
  return `${value.slice(0, 8)}...${value.slice(-6)}`
}

const openWalletAddressModal = (wallet: PersonalWalletCard) => {
  selectedWallet.value = wallet
  walletAddressInput.value = wallet.walletAddress
  walletAddressError.value = ''
  walletAddressModalOpen.value = true
}

const closeWalletAddressModal = () => {
  if (walletAddressSaving.value) return

  walletAddressModalOpen.value = false
  selectedWallet.value = null
  walletAddressInput.value = ''
  walletAddressError.value = ''
}

const applyUpdatedPersonalWallet = (wallet: AdminPersonalWallet) => {
  personalWallets.value = personalWallets.value.map((item) => {
    if (item.code !== wallet.code) return item
    const balanceText = walletBalanceText(wallet.availableBalance ?? wallet.balance ?? wallet.commissionBalance ?? item.balanceText)
    return {
      ...item,
      ...wallet,
      walletAddress: wallet.walletAddress || '',
      shareRate: wallet.shareRate || item.shareRate,
      balance: parseWalletBalance(balanceText),
      balanceText
    }
  })
  recomputeActiveWallets()
}

const getRequestErrorMessage = (error: unknown, fallback: string) => {
  const apiError = error as {
    data?: { error?: string }
    response?: { _data?: { error?: string } }
    message?: string
  }
  return apiError.data?.error || apiError.response?._data?.error || apiError.message || fallback
}

const submitWalletAddress = async (clearAddress = false) => {
  if (!selectedWallet.value || walletAddressSaving.value) return

  const nextAddress = clearAddress ? '' : walletAddressInput.value.trim().toLowerCase()
  if (!clearAddress && !walletAddressPattern.test(nextAddress)) {
    walletAddressError.value = 'Enter a valid 0x EVM/BEP-20 wallet address.'
    return
  }

  walletAddressSaving.value = true
  walletAddressError.value = ''

  try {
    const wallet = await $fetch<AdminPersonalWallet>(`${apiBase}/admin/personal-wallets/${selectedWallet.value.code}`, {
      method: 'PATCH',
      headers: {
        Authorization: `Bearer ${tokenCookie.value}`
      },
      body: {
        walletAddress: nextAddress
      }
    })
    applyUpdatedPersonalWallet(wallet)
    closeWalletAddressModal()
  } catch (err) {
    console.error('Failed to update personal wallet address:', err)
    walletAddressError.value = 'Failed to save wallet address.'
  } finally {
    walletAddressSaving.value = false
  }
}

const openWithdrawModal = (wallet: PersonalWalletCard) => {
  withdrawWallet.value = wallet
  withdrawAmountInput.value = ''
  withdrawAddressInput.value = wallet.walletAddress
  withdrawError.value = ''
  withdrawSuccess.value = ''
  withdrawModalOpen.value = true
}

const closeWithdrawModal = () => {
  if (withdrawSubmitting.value) return

  withdrawModalOpen.value = false
  withdrawWallet.value = null
  withdrawAmountInput.value = ''
  withdrawAddressInput.value = ''
  withdrawError.value = ''
  withdrawSuccess.value = ''
}

const submitWithdrawal = async () => {
  if (!withdrawWallet.value || !canSubmitWithdrawal.value) return

  withdrawSubmitting.value = true
  withdrawError.value = ''
  withdrawSuccess.value = ''

  try {
    const withdrawal = await $fetch<AdminPersonalWalletWithdrawal>(`${apiBase}/admin/personal-wallets/${withdrawWallet.value.code}/withdrawals`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${tokenCookie.value}`
      },
      body: {
        amount: withdrawAmountInput.value.trim(),
        walletAddress: withdrawAddressInput.value.trim().toLowerCase()
      }
    })
    withdrawSuccess.value = `Withdraw queued: ${formatDecimalText(withdrawal.amount)} ${withdrawal.asset}`
    withdrawAmountInput.value = ''
    await fetchPersonalWallets()
    const refreshedWallet = personalWallets.value.find(wallet => wallet.code === withdrawal.walletCode)
    if (refreshedWallet) {
      withdrawWallet.value = refreshedWallet
    }
  } catch (err) {
    console.error('Failed to create personal wallet withdrawal:', err)
    withdrawError.value = getRequestErrorMessage(err, 'Failed to create withdrawal.')
  } finally {
    withdrawSubmitting.value = false
  }
}

onMounted(async () => {
  const gasFeeAddress = config.public.gasFeeDepositAddress as string || ''
  const balanceVal = await fetchGasFeeWalletBalance(gasFeeAddress)

  await fetchPersonalWallets()
  gasFeeWalletActive.value = Boolean(gasFeeAddress)
  walletStats.value = {
    totalBalance: balanceVal,
    dailyInflow: 0,
    dailyOutflow: 0,
    activeWallets: walletStats.value.activeWallets
  }
  recomputeActiveWallets()

  loading.value = false
})
</script>

<template>
  <div class="dashboard-page">
    <div
      v-if="loading"
      class="skeleton-loading"
    >
      <div class="skeleton-page-header">
        <div class="skeleton-bone skeleton-title" />
      </div>

      <div class="skeleton-stats-grid">
        <div
          v-for="n in 4"
          :key="`stat-${n}`"
          class="skeleton-stat-card"
        >
          <div class="skeleton-bone skeleton-stat-label" />
          <div class="skeleton-bone skeleton-stat-value" />
        </div>
      </div>
    </div>

    <template v-else>
      <header class="page-header">
        <h1 class="page-title">
          Company Wallets
        </h1>
        <p class="page-subtitle">
          Manage internal wallets and balances
        </p>
      </header>

      <div class="stats-grid">
        <StatCard
          title="Total Balance"
          :value="`$${walletStats.totalBalance.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 6 })}`"
        />
        <StatCard
          title="Daily Inflow"
          :value="`+$${walletStats.dailyInflow.toLocaleString()}`"
        />
        <StatCard
          title="Daily Outflow"
          :value="`-$${walletStats.dailyOutflow.toLocaleString()}`"
        />
        <StatCard
          title="Active Wallets"
          :value="walletStats.activeWallets.toString()"
        />
      </div>

      <div class="personal-wallets-section">
        <h2 class="section-title">
          Personal Wallets
        </h2>
        <div class="personal-wallets-grid">
          <div
            v-for="wallet in personalWallets"
            :key="wallet.code"
            class="wallet-card"
          >
            <div class="wallet-header">
              <h3 class="wallet-name">
                {{ wallet.displayName }}
              </h3>
              <button
                class="icon-btn"
                type="button"
                title="Link wallet address"
                :aria-label="`Link address for ${wallet.displayName}`"
                @click="openWalletAddressModal(wallet)"
              >
                <UIcon name="lucide:settings" />
              </button>
            </div>
            <div class="wallet-balance">
              <span class="currency">$</span>{{ formatDecimalText(wallet.balanceText) }}
            </div>
            <div
              class="wallet-address-status"
              :class="{ 'wallet-address-status--linked': wallet.walletAddress }"
            >
              <UIcon :name="wallet.walletAddress ? 'lucide:link' : 'lucide:unlink'" />
              <span>{{ formatWalletAddress(wallet.walletAddress) }}</span>
            </div>
            <div class="wallet-actions">
              <button
                class="primary-btn withdraw-btn"
                type="button"
                :disabled="wallet.balance <= 0"
                :title="wallet.balance <= 0 ? 'No balance available' : 'Withdraw'"
                @click="openWithdrawModal(wallet)"
              >
                <UIcon name="lucide:arrow-up-right" />
                Withdraw
              </button>
            </div>
          </div>
        </div>
      </div>
    </template>

    <Teleport to="body">
      <div
        v-if="walletAddressModalOpen && selectedWallet"
        class="wallet-modal"
        role="dialog"
        aria-modal="true"
        aria-label="Link personal wallet address"
        @click.self="closeWalletAddressModal"
      >
        <div class="wallet-modal__box">
          <div class="wallet-modal__header">
            <span class="wallet-modal__spacer" />
            <h3>Wallet Address</h3>
            <button
              class="wallet-modal__icon-btn"
              type="button"
              aria-label="Close wallet address modal"
              @click="closeWalletAddressModal"
            >
              <UIcon name="lucide:x" />
            </button>
          </div>

          <form
            class="wallet-modal__body"
            autocomplete="off"
            @submit.prevent="submitWalletAddress(false)"
          >
            <div class="wallet-modal__target">
              <span>{{ selectedWallet.displayName }}</span>
              <strong>{{ selectedWallet.walletAddress ? 'Linked' : 'Not Linked' }}</strong>
            </div>

            <label class="wallet-field">
              <span>USDT BEP-20 / EVM Address</span>
              <input
                v-model="walletAddressInput"
                type="text"
                inputmode="text"
                autocomplete="off"
                autocapitalize="off"
                spellcheck="false"
                placeholder="0x0000000000000000000000000000000000000000"
                :class="{ 'is-invalid': walletAddressInvalid || walletAddressError }"
                :disabled="walletAddressSaving"
              >
            </label>

            <p
              v-if="walletAddressError"
              class="wallet-modal__error"
            >
              {{ walletAddressError }}
            </p>
            <p
              v-else-if="walletAddressInvalid"
              class="wallet-modal__error"
            >
              Enter a valid 0x EVM/BEP-20 wallet address.
            </p>

            <div class="wallet-modal__actions">
              <button
                v-if="selectedWallet.walletAddress"
                class="secondary-btn"
                type="button"
                :disabled="walletAddressSaving"
                @click="submitWalletAddress(true)"
              >
                <UIcon name="lucide:unlink" />
                Clear Address
              </button>
              <button
                class="save-wallet-btn"
                type="submit"
                :disabled="!canSaveWalletAddress"
              >
                <UIcon name="lucide:link" />
                {{ walletAddressSaving ? 'Saving...' : 'Save Address' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="withdrawModalOpen && withdrawWallet"
        class="wallet-modal"
        role="dialog"
        aria-modal="true"
        aria-label="Withdraw personal wallet balance"
        @click.self="closeWithdrawModal"
      >
        <div class="wallet-modal__box">
          <div class="wallet-modal__header">
            <span class="wallet-modal__spacer" />
            <h3>Withdraw</h3>
            <button
              class="wallet-modal__icon-btn"
              type="button"
              aria-label="Close withdraw modal"
              @click="closeWithdrawModal"
            >
              <UIcon name="lucide:x" />
            </button>
          </div>

          <form
            class="wallet-modal__body"
            autocomplete="off"
            @submit.prevent="submitWithdrawal"
          >
            <div class="wallet-modal__target">
              <span>{{ withdrawWallet.displayName }}</span>
              <strong>{{ formatDecimalText(withdrawWallet.balanceText) }} USDT</strong>
            </div>

            <label class="wallet-field">
              <span>Wallet Address</span>
              <input
                v-model="withdrawAddressInput"
                type="text"
                inputmode="text"
                autocomplete="off"
                autocapitalize="off"
                spellcheck="false"
                placeholder="0x0000000000000000000000000000000000000000"
                :class="{ 'is-invalid': withdrawAddressInvalid || withdrawError }"
                :disabled="withdrawSubmitting"
              >
            </label>

            <label class="wallet-field">
              <span>Amount (USDT)</span>
              <input
                v-model="withdrawAmountInput"
                type="text"
                inputmode="decimal"
                autocomplete="off"
                placeholder="0.00"
                :class="{ 'is-invalid': Boolean(withdrawAmountValidationMessage) || withdrawError }"
                :disabled="withdrawSubmitting"
              >
            </label>

            <p
              v-if="withdrawError"
              class="wallet-modal__error"
            >
              {{ withdrawError }}
            </p>
            <p
              v-else-if="withdrawAddressInvalid"
              class="wallet-modal__error"
            >
              Enter a valid 0x EVM/BEP-20 wallet address.
            </p>
            <p
              v-else-if="withdrawAmountValidationMessage"
              class="wallet-modal__error"
            >
              {{ withdrawAmountValidationMessage }}
            </p>
            <p
              v-else-if="withdrawSuccess"
              class="wallet-modal__success"
            >
              {{ withdrawSuccess }}
            </p>

            <div class="wallet-modal__actions">
              <button
                class="secondary-neutral-btn"
                type="button"
                :disabled="withdrawSubmitting"
                @click="closeWithdrawModal"
              >
                Cancel
              </button>
              <button
                class="save-wallet-btn"
                type="submit"
                :disabled="!canSubmitWithdrawal"
              >
                <UIcon name="lucide:arrow-up-right" />
                {{ withdrawSubmitting ? 'Withdrawing...' : 'Withdraw' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.dashboard-page {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.page-header {
  margin-bottom: 1rem;
}

.page-title {
  font-family: 'Oswald', sans-serif;
  font-size: 1.5rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  color: var(--text);
  margin-bottom: 0.5rem;
}

.page-subtitle {
  color: var(--text-mute);
  font-size: 0.95rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.5rem;
}

/* Skeleton Loading */
.skeleton-loading {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.skeleton-bone {
  background: linear-gradient(90deg, var(--charcoal) 25%, var(--line) 50%, var(--charcoal) 75%);
  background-size: 200% 100%;
  animation: loading 1.5s infinite;
  border-radius: 4px;
}

@keyframes loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.skeleton-page-header {
  margin-bottom: 1rem;
}

.skeleton-title {
  width: 200px;
  height: 24px;
}

.skeleton-stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.5rem;
}

.skeleton-stat-card {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.skeleton-stat-label {
  width: 60%;
  height: 14px;
}

.skeleton-stat-value {
  width: 80%;
  height: 28px;
}

.personal-wallets-section {
  margin-top: 1rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.section-title {
  font-family: 'Oswald', sans-serif;
  font-size: 1.25rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  color: var(--text);
}

.personal-wallets-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.5rem;
}

.wallet-card {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 16px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  position: relative;
  overflow: hidden;
  transition: none; /* No hover effect */
  cursor: default; /* No pointer */
}

.wallet-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, var(--accent), #ff7a33);
  opacity: 0.8;
}

.wallet-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.wallet-name {
  font-family: var(--mono);
  font-size: 0.95rem;
  letter-spacing: 0.08em;
  color: var(--silver);
  font-weight: 600;
}

.icon-btn {
  background: transparent;
  border: none;
  color: var(--text-mute);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem;
  border-radius: 50%;
  transition: all 0.2s ease;
}

.icon-btn:hover {
  color: var(--text);
  background: rgba(255, 255, 255, 0.05);
}

.wallet-balance {
  font-size: 2.5rem;
  font-weight: 600;
  color: var(--text);
  display: flex;
  align-items: baseline;
  gap: 0.25rem;
}

.wallet-balance .currency {
  font-size: 1.5rem;
  color: var(--text-mute);
}

.wallet-address-status {
  min-height: 40px;
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.75rem 0.85rem;
  border: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.025);
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  overflow-wrap: anywhere;
}

.wallet-address-status--linked {
  border-color: rgba(16, 185, 129, 0.24);
  background: rgba(16, 185, 129, 0.08);
  color: #10b981;
}

.wallet-actions {
  display: flex;
  gap: 1rem;
  margin-top: 0.5rem;
}

.primary-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--line);
  color: var(--text);
  padding: 0.8rem 1rem;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.primary-btn:hover {
  background: var(--accent);
  border-color: var(--accent);
  color: #fff;
}

.primary-btn:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.primary-btn:disabled:hover {
  background: rgba(255, 255, 255, 0.03);
  border-color: var(--line);
  color: var(--text);
}

.wallet-modal {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  background: rgba(0, 0, 0, 0.88);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
}

.wallet-modal__box {
  width: min(540px, 100%);
  max-height: min(640px, calc(100vh - 4rem));
  overflow-y: auto;
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  box-shadow: 0 28px 70px rgba(0, 0, 0, 0.45);
}

.wallet-modal__header {
  display: grid;
  grid-template-columns: 36px minmax(0, 1fr) 36px;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--line);
}

.wallet-modal__header h3 {
  margin: 0;
  font-family: 'Oswald', sans-serif;
  font-size: 1.45rem;
  font-weight: 400;
  color: var(--text);
  letter-spacing: 0.04em;
  text-align: center;
  text-transform: uppercase;
}

.wallet-modal__spacer {
  width: 36px;
  height: 36px;
}

.wallet-modal__icon-btn {
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

.wallet-modal__icon-btn:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.wallet-modal__body {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.5rem;
}

.wallet-modal__target {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  min-height: 56px;
  padding: 0.9rem 1rem;
  border: 1px solid var(--line);
  background: var(--charcoal);
}

.wallet-modal__target span,
.wallet-modal__target strong {
  font-family: var(--mono);
  font-size: 10px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.wallet-modal__target span {
  color: var(--text);
}

.wallet-modal__target strong {
  color: var(--accent);
}

.wallet-field {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.wallet-field span {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.wallet-field input {
  width: 100%;
  min-width: 0;
  height: 44px;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text);
  font-family: var(--mono);
  font-size: 12px;
  outline: none;
  padding: 0 0.85rem;
}

.wallet-field input:focus {
  border-color: var(--accent);
}

.wallet-field input.is-invalid {
  border-color: #ef4444;
}

.wallet-modal__error {
  margin: -0.25rem 0 0;
  color: #ef4444;
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.wallet-modal__success {
  margin: -0.25rem 0 0;
  color: #10b981;
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.wallet-modal__actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.save-wallet-btn,
.secondary-neutral-btn,
.secondary-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  min-height: 44px;
  padding: 0 1rem;
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

.secondary-neutral-btn {
  border-color: var(--line);
  background: transparent;
  color: var(--text);
}

.secondary-neutral-btn:hover {
  border-color: var(--accent);
  color: var(--accent);
  transform: translateY(-1px);
}

.save-wallet-btn:hover {
  background: #ff7324;
  border-color: #ff7324;
  transform: translateY(-1px);
}

.secondary-btn {
  border-color: rgba(239, 68, 68, 0.45);
  background: transparent;
  color: #ef4444;
}

.secondary-btn:hover {
  border-color: #ef4444;
  background: rgba(239, 68, 68, 0.08);
  color: #ff6b6b;
  transform: translateY(-1px);
}

.save-wallet-btn:disabled,
.secondary-neutral-btn:disabled,
.secondary-btn:disabled {
  cursor: not-allowed;
  opacity: 0.45;
  transform: none;
}

@media (max-width: 1180px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }

  .wallet-modal {
    padding: 1rem;
  }

  .wallet-modal__actions {
    flex-direction: column-reverse;
  }

  .save-wallet-btn,
  .secondary-neutral-btn,
  .secondary-btn {
    width: 100%;
  }
}
</style>
