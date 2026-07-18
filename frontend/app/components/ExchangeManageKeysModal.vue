<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'

interface ExchangeBinding {
  id: number
  name: string
  logo: string
  logoDark?: string
  status: string
  lastSynced: string | null
  balance: number
  hasApi?: boolean
}

interface CredentialConfig {
  key: string
  label: string
}

interface CredentialPreview {
  key: string
  label: string
  value: string
  maskedValue: string
}

const props = defineProps<{
  modelValue: boolean
  exchange: ExchangeBinding | null
  theme: 'dark' | 'light'
  googleAuthenticatorEnabled: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'status-change': [payload: { exchangeId: number, status: 'connected' | 'disconnected' }]
}>()

const manageStep = ref<'keys' | 'verify'>('keys')
const pendingAction = ref<'connect' | 'disconnect' | null>(null)
const visibleCredentials = reactive<Record<string, boolean>>({})
const emailOtp = ref('')
const googleOtp = ref('')
const submitAttempted = ref(false)
const submitted = ref(false)
const fieldShake = reactive<Record<string, boolean>>({})

const resetManageState = () => {
  manageStep.value = 'keys'
  pendingAction.value = null
  emailOtp.value = ''
  googleOtp.value = ''
  submitAttempted.value = false
  submitted.value = false

  Object.keys(fieldShake).forEach((key) => {
    delete fieldShake[key]
  })
  Object.keys(visibleCredentials).forEach((key) => {
    delete visibleCredentials[key]
  })
}

watch(() => props.modelValue, (isOpen) => {
  if (isOpen) {
    resetManageState()
  }
})

const closeManageModal = () => {
  emit('update:modelValue', false)
}

const getExchangeLogo = (exchange: ExchangeBinding) => {
  return props.theme === 'dark' && exchange.logoDark ? exchange.logoDark : exchange.logo
}

const credentialConfigs: Record<string, CredentialConfig[]> = {
  Binance: [
    { key: 'apiKey', label: 'API Key' },
    { key: 'apiSecret', label: 'API Secret' }
  ],
  OKX: [
    { key: 'apiKey', label: 'API Key' },
    { key: 'apiSecret', label: 'API Secret' },
    { key: 'passphrase', label: 'Passphrase' }
  ],
  Bybit: [
    { key: 'apiKey', label: 'API Key' },
    { key: 'apiSecret', label: 'API Secret' }
  ],
  Tokocrypto: [
    { key: 'apiKey', label: 'API Key' },
    { key: 'apiSecret', label: 'API Secret' }
  ]
}

const getCredentialValue = (exchange: ExchangeBinding, credential: CredentialConfig) => {
  const prefix = exchange.name.toUpperCase().replace(/\s+/g, '')
  const suffix = String(exchange.id).padStart(2, '0')
  const keyLabel = credential.key.replace(/([A-Z])/g, '-$1').toUpperCase()

  return `MAU-${prefix}-${keyLabel}-${suffix}A9F`
}

const maskCredential = (value: string) => {
  return `${value.slice(0, 12)}••••••••${value.slice(-4)}`
}

const credentialPreviews = computed<CredentialPreview[]>(() => {
  if (!props.exchange) return []

  const configs = credentialConfigs[props.exchange.name] ?? credentialConfigs.Binance!

  return configs.map((credential) => {
    const value = getCredentialValue(props.exchange!, credential)

    return {
      ...credential,
      value,
      maskedValue: maskCredential(value)
    }
  })
})

const getDisplayedCredential = (credential: CredentialPreview) => {
  return visibleCredentials[credential.key] ? credential.value : credential.maskedValue
}

const toggleCredentialVisibility = (key: string) => {
  visibleCredentials[key] = !visibleCredentials[key]
}

const actionLabel = computed(() => {
  return pendingAction.value === 'connect' ? 'Connect' : 'Disconnect'
})

const emailOtpInvalid = computed(() => emailOtp.value.trim().length !== 6)
const googleOtpInvalid = computed(() => props.googleAuthenticatorEnabled && googleOtp.value.trim().length !== 6)
const verificationBlocked = computed(() => submitted.value || emailOtpInvalid.value || googleOtpInvalid.value)

const triggerFieldShake = (key: string) => {
  fieldShake[key] = false
  window.requestAnimationFrame(() => {
    fieldShake[key] = true
  })
}

const startVerification = (action: 'connect' | 'disconnect') => {
  pendingAction.value = action
  manageStep.value = 'verify'
  emailOtp.value = ''
  googleOtp.value = ''
  submitAttempted.value = false
  submitted.value = false
}

watch([emailOtp, googleOtp], () => {
  submitted.value = false
})

const submitVerification = () => {
  submitAttempted.value = true

  if (emailOtpInvalid.value) {
    triggerFieldShake('emailOtp')
  }

  if (googleOtpInvalid.value) {
    triggerFieldShake('googleOtp')
  }

  if (verificationBlocked.value || !props.exchange || !pendingAction.value) return

  submitted.value = true
  emit('status-change', {
    exchangeId: props.exchange.id,
    status: pendingAction.value === 'connect' ? 'connected' : 'disconnected'
  })
}
</script>

<template>
  <div
    v-if="modelValue && exchange"
    class="manage-modal"
    role="dialog"
    aria-modal="true"
    :aria-label="manageStep === 'keys' ? 'Manage API Keys' : `${actionLabel} Verification`"
    @click.self="closeManageModal"
  >
    <div class="manage-modal__box">
      <div class="manage-modal__header">
        <button
          v-if="manageStep === 'verify'"
          class="manage-modal__icon-btn"
          type="button"
          aria-label="Back to API key details"
          @click="manageStep = 'keys'"
        >
          <UIcon name="lucide:arrow-left" />
        </button>
        <span
          v-else
          class="manage-modal__spacer"
        />
        <h3>{{ manageStep === 'keys' ? 'Manage Keys' : `${actionLabel} Verification` }}</h3>
        <button
          class="manage-modal__icon-btn"
          type="button"
          aria-label="Close manage keys modal"
          @click="closeManageModal"
        >
          <UIcon name="lucide:x" />
        </button>
      </div>

      <div
        v-if="manageStep === 'keys'"
        class="key-panel"
      >
        <div class="key-panel__exchange">
          <img
            class="key-panel__logo"
            :src="getExchangeLogo(exchange)"
            :alt="`${exchange.name} logo`"
          >
          <span
            class="key-panel__status"
            :class="exchange.status === 'connected' ? 'status-active' : 'status-inactive'"
          >
            {{ exchange.status }}
          </span>
        </div>

        <label
          v-for="credential in credentialPreviews"
          :key="credential.key"
          class="manage-field"
        >
          <span>{{ credential.label }}</span>
          <div class="api-key-view">
            <input
              :value="getDisplayedCredential(credential)"
              readonly
              type="text"
            >
            <button
              type="button"
              :aria-label="visibleCredentials[credential.key] ? `Hide ${credential.label}` : `Show ${credential.label}`"
              @click="toggleCredentialVisibility(credential.key)"
            >
              <UIcon :name="visibleCredentials[credential.key] ? 'lucide:eye-off' : 'lucide:eye'" />
            </button>
          </div>
        </label>

        <div class="key-actions">
          <button
            v-if="exchange.status === 'connected'"
            class="btn-danger"
            type="button"
            @click="startVerification('disconnect')"
          >
            <UIcon name="lucide:unlink" />
            <span>Disconnect</span>
          </button>
          <button
            v-else
            class="btn-primary"
            type="button"
            @click="startVerification('connect')"
          >
            <UIcon name="lucide:link" />
            <span>Connect</span>
          </button>
        </div>
      </div>

      <form
        v-else
        class="verify-form"
        autocomplete="off"
        data-form-type="other"
        @submit.prevent="submitVerification"
      >
        <div class="verify-form__target">
          <span>{{ exchange.name }}</span>
          <strong>{{ actionLabel }}</strong>
        </div>

        <label class="manage-field">
          <span>Email OTP</span>
          <input
            v-model="emailOtp"
            class="otp-input"
            :class="{ 'is-invalid': submitAttempted && emailOtpInvalid, 'is-shaking': fieldShake.emailOtp }"
            type="text"
            inputmode="numeric"
            autocomplete="one-time-code"
            maxlength="6"
            placeholder="000000"
            @animationend="fieldShake.emailOtp = false"
          >
        </label>

        <label
          v-if="googleAuthenticatorEnabled"
          class="manage-field"
        >
          <span>Google Authenticator</span>
          <input
            v-model="googleOtp"
            class="otp-input"
            :class="{ 'is-invalid': submitAttempted && googleOtpInvalid, 'is-shaking': fieldShake.googleOtp }"
            type="text"
            inputmode="numeric"
            autocomplete="one-time-code"
            maxlength="6"
            placeholder="000000"
            @animationend="fieldShake.googleOtp = false"
          >
        </label>

        <button
          class="verify-submit"
          :class="{ 'is-blocked': verificationBlocked }"
          type="submit"
          :aria-disabled="verificationBlocked"
        >
          <UIcon :name="pendingAction === 'connect' ? 'lucide:link' : 'lucide:unlink'" />
          <span>{{ actionLabel }}</span>
        </button>

        <p
          v-if="submitAttempted && verificationBlocked"
          class="manage-error"
        >
          Complete verification codes
        </p>

        <p
          v-if="submitted"
          class="manage-success"
        >
          Status updated
        </p>
      </form>
    </div>
  </div>
</template>

<style scoped>
.manage-modal {
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

.manage-modal__box {
  width: min(540px, 100%);
  max-height: min(720px, calc(100vh - 4rem));
  overflow-y: auto;
  scrollbar-width: none;
  -ms-overflow-style: none;
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  box-shadow: 0 28px 70px rgba(0, 0, 0, 0.45);
}

.manage-modal__box::-webkit-scrollbar {
  display: none;
}

.manage-modal__header {
  display: grid;
  grid-template-columns: 36px 1fr 36px;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--line);
}

.manage-modal__header h3 {
  margin: 0;
  font-family: 'Oswald', sans-serif;
  font-size: 1.45rem;
  font-weight: 400;
  color: var(--text);
  letter-spacing: 0.04em;
  text-align: center;
  text-transform: uppercase;
}

.manage-modal__icon-btn {
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

.manage-modal__spacer {
  width: 36px;
  height: 36px;
}

.manage-modal__icon-btn:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.key-panel,
.verify-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.5rem;
}

.key-panel__exchange {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  min-height: 62px;
  padding: 0.9rem 1rem;
  border: 1px solid var(--line);
  background: var(--charcoal);
}

.key-panel__logo {
  display: block;
  width: 150px;
  height: 34px;
  object-fit: contain;
  object-position: left center;
}

.key-panel__status {
  font-family: var(--mono);
  font-size: 9px;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  padding: 0.3rem 0.6rem;
  border-radius: 20px;
}

.status-active {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.2);
}

.status-inactive {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.2);
}

.manage-field {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.manage-field > span {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.manage-field input,
.api-key-view input {
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

.api-key-view {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 42px;
}

.api-key-view input {
  border-right: none;
}

.api-key-view button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 42px;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text);
  transition: border-color 220ms var(--ease-quiet), color 220ms var(--ease-quiet);
}

.api-key-view button:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.key-actions {
  display: flex;
  justify-content: flex-end;
}

.btn-primary,
.btn-danger,
.verify-submit {
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

.btn-primary:hover,
.verify-submit:hover {
  background: #ff7324;
  border-color: #ff7324;
  transform: translateY(-1px);
}

.btn-danger {
  border-color: rgba(239, 68, 68, 0.45);
  background: transparent;
  color: #ef4444;
}

.btn-danger:hover {
  border-color: #ef4444;
  background: rgba(239, 68, 68, 0.08);
  color: #ff6b6b;
  transform: translateY(-1px);
}

.verify-form__target {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.9rem 1rem;
  border: 1px solid var(--line);
  background: var(--charcoal);
}

.verify-form__target span,
.verify-form__target strong {
  font-family: var(--mono);
  font-size: 10px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.verify-form__target span {
  color: var(--text-mute);
}

.verify-form__target strong {
  color: var(--accent);
}

.otp-input {
  letter-spacing: 0.28em;
}

.manage-field input.is-invalid {
  border-color: #ef4444;
}

.manage-field input.is-shaking {
  animation: manage-shake 260ms ease-in-out;
}

.verify-submit {
  width: 100%;
}

.verify-submit.is-blocked {
  box-shadow: inset 0 0 0 1px rgba(239, 68, 68, 0.45);
}

.manage-error,
.manage-success {
  margin: -0.2rem 0 0;
  font-family: var(--mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  text-align: center;
  text-transform: uppercase;
}

.manage-error {
  color: #ef4444;
}

.manage-success {
  color: #10b981;
}

@keyframes manage-shake {
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

@media (max-width: 640px) {
  .manage-modal {
    padding: 1rem;
  }

  .key-panel__exchange {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
