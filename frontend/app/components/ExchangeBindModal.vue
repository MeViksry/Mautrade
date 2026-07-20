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

interface ExtraFieldConfig {
  key: string
  label: string
  type: 'text' | 'password'
  required: boolean
  placeholder: string
}

interface ExchangeConfig {
  extraFields: ExtraFieldConfig[]
}

const props = defineProps<{
  modelValue: boolean
  exchanges: ExchangeBinding[]
  theme: 'dark' | 'light'
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'submitted': [payload: { exchange: string, apiKey: string, apiSecret: string, extras: Record<string, string> }]
}>()

const exchangeConfigs: Record<string, ExchangeConfig> = {
  Binance: {
    extraFields: []
  },
  OKX: {
    extraFields: [
      {
        key: 'passphrase',
        label: 'Passphrase',
        type: 'password',
        required: true,
        placeholder: 'Enter API passphrase'
      }
    ]
  },
  Bybit: {
    extraFields: []
  },
  Tokocrypto: {
    extraFields: []
  }
}

const bindStep = ref<'exchanges' | 'credentials'>('exchanges')
const selectedExchange = ref<ExchangeBinding | null>(null)
const apiKey = ref('')
const apiSecret = ref('')
const extras = reactive<Record<string, string>>({})
const visibleFields = reactive<Record<string, boolean>>({})
const autofillLocked = reactive<Record<string, boolean>>({
  apiKey: true,
  apiSecret: true
})
const submitAttempted = ref(false)
const submitted = ref(false)
const fieldShake = reactive<Record<string, boolean>>({})

const resetBindState = () => {
  bindStep.value = 'exchanges'
  selectedExchange.value = null
  apiKey.value = ''
  apiSecret.value = ''
  submitAttempted.value = false
  submitted.value = false

  Object.keys(extras).forEach((key) => {
    // eslint-disable-next-line @typescript-eslint/no-dynamic-delete
    delete extras[key]
  })
  Object.keys(visibleFields).forEach((key) => {
    // eslint-disable-next-line @typescript-eslint/no-dynamic-delete
    delete visibleFields[key]
  })
  Object.keys(autofillLocked).forEach((key) => {
    // eslint-disable-next-line @typescript-eslint/no-dynamic-delete
    delete autofillLocked[key]
  })
  autofillLocked.apiKey = true
  autofillLocked.apiSecret = true
  Object.keys(fieldShake).forEach((key) => {
    // eslint-disable-next-line @typescript-eslint/no-dynamic-delete
    delete fieldShake[key]
  })
}

watch(() => props.modelValue, (isOpen) => {
  if (isOpen) {
    resetBindState()
  }
})

const closeBindModal = () => {
  emit('update:modelValue', false)
}

const getExchangeLogo = (exchange: ExchangeBinding) => {
  return props.theme === 'dark' && exchange.logoDark ? exchange.logoDark : exchange.logo
}

const isExchangeBound = (exchange: ExchangeBinding) => {
  return exchange.hasApi ?? exchange.status === 'connected'
}

const selectedConfig = computed(() => {
  if (!selectedExchange.value) return { extraFields: [] }

  return exchangeConfigs[selectedExchange.value.name] ?? { extraFields: [] }
})

const selectExchange = (exchange: ExchangeBinding) => {
  if (isExchangeBound(exchange)) return

  selectedExchange.value = exchange
  bindStep.value = 'credentials'
  submitAttempted.value = false
  submitted.value = false

  selectedConfig.value.extraFields.forEach((field) => {
    extras[field.key] = ''
    visibleFields[field.key] = false
    autofillLocked[field.key] = true
  })
}

const fieldErrors = computed(() => {
  const errors: Record<string, boolean> = {
    apiKey: apiKey.value.trim().length === 0,
    apiSecret: apiSecret.value.trim().length === 0
  }

  selectedConfig.value.extraFields.forEach((field) => {
    if (field.required) {
      errors[field.key] = (extras[field.key] ?? '').trim().length === 0
    }
  })

  return errors
})

const formBlocked = computed(() => {
  return submitted.value || Object.values(fieldErrors.value).some(Boolean)
})

const triggerFieldShake = (key: string) => {
  fieldShake[key] = false
  window.requestAnimationFrame(() => {
    fieldShake[key] = true
  })
}

const toggleFieldVisibility = (key: string) => {
  visibleFields[key] = !visibleFields[key]
}

const unlockAutofillField = (key: string) => {
  autofillLocked[key] = false
}

watch([apiKey, apiSecret], () => {
  submitted.value = false
})

watch(extras, () => {
  submitted.value = false
})

const submitBindExchange = () => {
  submitAttempted.value = true

  Object.entries(fieldErrors.value).forEach(([key, hasError]) => {
    if (hasError) {
      triggerFieldShake(key)
    }
  })

  if (formBlocked.value || !selectedExchange.value) return

  submitted.value = true
  emit('submitted', {
    exchange: selectedExchange.value.name,
    apiKey: apiKey.value.trim(),
    apiSecret: apiSecret.value.trim(),
    extras: Object.fromEntries(
      selectedConfig.value.extraFields.map(field => [field.key, (extras[field.key] ?? '').trim()])
    )
  })
}
</script>

<template>
  <div
    v-if="modelValue"
    class="bind-modal"
    role="dialog"
    aria-modal="true"
    :aria-label="bindStep === 'exchanges' ? 'Bind New Exchange' : 'Exchange API Credentials'"
    @click.self="closeBindModal"
  >
    <div class="bind-modal__box">
      <div class="bind-modal__header">
        <button
          v-if="bindStep === 'credentials'"
          class="bind-modal__icon-btn"
          type="button"
          aria-label="Back to exchange list"
          @click="bindStep = 'exchanges'"
        >
          <UIcon name="lucide:arrow-left" />
        </button>
        <span
          v-else
          class="bind-modal__spacer"
        />
        <h3>{{ bindStep === 'exchanges' ? 'Bind New Exchange' : 'API Credentials' }}</h3>
        <button
          class="bind-modal__icon-btn"
          type="button"
          aria-label="Close bind modal"
          @click="closeBindModal"
        >
          <UIcon name="lucide:x" />
        </button>
      </div>

      <div
        v-if="bindStep === 'exchanges'"
        class="exchange-picker"
      >
        <button
          v-for="exchange in exchanges"
          :key="exchange.id"
          class="exchange-option"
          :class="{ 'is-bound': isExchangeBound(exchange) }"
          type="button"
          :disabled="isExchangeBound(exchange)"
          @click="selectExchange(exchange)"
        >
          <span class="exchange-option__logo-shell">
            <img
              class="exchange-option__logo"
              :src="getExchangeLogo(exchange)"
              :alt="`${exchange.name} logo`"
            >
          </span>
          <span
            class="exchange-option__status"
            :class="exchange.status === 'connected' ? 'status-active' : 'status-inactive'"
          >
            {{ exchange.status }}
          </span>
          <UIcon
            :name="isExchangeBound(exchange) ? 'lucide:lock' : 'lucide:chevron-right'"
            class="exchange-option__arrow"
          />
        </button>
      </div>

      <form
        v-else
        class="bind-form"
        autocomplete="off"
        data-form-type="other"
        @submit.prevent="submitBindExchange"
      >
        <div
          v-if="selectedExchange"
          class="bind-form__exchange"
        >
          <img
            class="bind-form__logo"
            :src="getExchangeLogo(selectedExchange)"
            :alt="`${selectedExchange.name} logo`"
          >
          <span>{{ selectedExchange.name }}</span>
        </div>

        <label class="bind-field">
          <span>API Key</span>
          <input
            v-model="apiKey"
            :class="{ 'is-invalid': submitAttempted && fieldErrors.apiKey, 'is-shaking': fieldShake.apiKey }"
            :readonly="autofillLocked.apiKey"
            type="text"
            name="mautrade-exchange-api-key"
            placeholder="Enter API key"
            autocomplete="new-password"
            autocapitalize="off"
            autocorrect="off"
            inputmode="text"
            :spellcheck="false"
            data-1p-ignore="true"
            data-lpignore="true"
            data-form-type="other"
            @focus="unlockAutofillField('apiKey')"
            @animationend="fieldShake.apiKey = false"
          >
        </label>

        <label class="bind-field">
          <span>API Secret</span>
          <div
            class="bind-secret"
            :class="{ 'is-invalid': submitAttempted && fieldErrors.apiSecret, 'is-shaking': fieldShake.apiSecret }"
            @animationend="fieldShake.apiSecret = false"
          >
            <input
              v-model="apiSecret"
              :type="visibleFields.apiSecret ? 'text' : 'password'"
              :readonly="autofillLocked.apiSecret"
              name="mautrade-exchange-api-secret"
              placeholder="Enter API secret"
              autocomplete="new-password"
              autocapitalize="off"
              autocorrect="off"
              inputmode="text"
              :spellcheck="false"
              data-1p-ignore="true"
              data-lpignore="true"
              data-form-type="other"
              @focus="unlockAutofillField('apiSecret')"
            >
            <button
              type="button"
              :aria-label="visibleFields.apiSecret ? 'Hide API secret' : 'Show API secret'"
              @click="toggleFieldVisibility('apiSecret')"
            >
              <UIcon :name="visibleFields.apiSecret ? 'lucide:eye-off' : 'lucide:eye'" />
            </button>
          </div>
        </label>

        <label
          v-for="field in selectedConfig.extraFields"
          :key="field.key"
          class="bind-field"
        >
          <span>{{ field.label }}</span>
          <div
            v-if="field.type === 'password'"
            class="bind-secret"
            :class="{ 'is-invalid': submitAttempted && fieldErrors[field.key], 'is-shaking': fieldShake[field.key] }"
            @animationend="fieldShake[field.key] = false"
          >
            <input
              v-model="extras[field.key]"
              :type="visibleFields[field.key] ? 'text' : 'password'"
              :readonly="autofillLocked[field.key]"
              :name="`mautrade-exchange-${field.key}`"
              :placeholder="field.placeholder"
              autocomplete="new-password"
              autocapitalize="off"
              autocorrect="off"
              inputmode="text"
              :spellcheck="false"
              data-1p-ignore="true"
              data-lpignore="true"
              data-form-type="other"
              @focus="unlockAutofillField(field.key)"
            >
            <button
              type="button"
              :aria-label="visibleFields[field.key] ? `Hide ${field.label}` : `Show ${field.label}`"
              @click="toggleFieldVisibility(field.key)"
            >
              <UIcon :name="visibleFields[field.key] ? 'lucide:eye-off' : 'lucide:eye'" />
            </button>
          </div>
          <input
            v-else
            v-model="extras[field.key]"
            :class="{ 'is-invalid': submitAttempted && fieldErrors[field.key], 'is-shaking': fieldShake[field.key] }"
            :readonly="autofillLocked[field.key]"
            :type="field.type"
            :name="`mautrade-exchange-${field.key}`"
            :placeholder="field.placeholder"
            autocomplete="new-password"
            autocapitalize="off"
            autocorrect="off"
            inputmode="text"
            :spellcheck="false"
            data-1p-ignore="true"
            data-lpignore="true"
            data-form-type="other"
            @focus="unlockAutofillField(field.key)"
            @animationend="fieldShake[field.key] = false"
          >
        </label>

        <button
          class="bind-submit"
          :class="{ 'is-blocked': formBlocked }"
          type="submit"
          :aria-disabled="formBlocked"
        >
          <UIcon name="lucide:link" />
          <span>Bind Exchange</span>
        </button>

        <p
          v-if="submitAttempted && formBlocked"
          class="bind-error"
        >
          Complete required credentials
        </p>

        <p
          v-if="submitted"
          class="bind-success"
        >
          Exchange binding submitted
        </p>
      </form>
    </div>
  </div>
</template>

<style scoped>
.bind-modal {
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

.bind-modal__box {
  width: min(560px, 100%);
  max-height: min(760px, calc(100vh - 4rem));
  overflow-y: auto;
  scrollbar-width: none;
  -ms-overflow-style: none;
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  box-shadow: 0 28px 70px rgba(0, 0, 0, 0.45);
}

.bind-modal__box::-webkit-scrollbar {
  display: none;
}

.bind-modal__header {
  display: grid;
  grid-template-columns: 36px 1fr 36px;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--line);
}

.bind-modal__header h3 {
  margin: 0;
  font-family: 'Oswald', sans-serif;
  font-size: 1.45rem;
  font-weight: 400;
  color: var(--text);
  letter-spacing: 0.04em;
  text-align: center;
  text-transform: uppercase;
}

.bind-modal__icon-btn {
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

.bind-modal__spacer {
  width: 36px;
  height: 36px;
}

.bind-modal__icon-btn:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.exchange-picker,
.bind-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.5rem;
}

.exchange-option {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto 24px;
  align-items: center;
  gap: 1rem;
  width: 100%;
  min-height: 76px;
  padding: 1rem;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text);
  cursor: pointer;
  text-align: left;
  transition: border-color 220ms var(--ease-quiet), background 220ms var(--ease-quiet);
}

.exchange-option:hover {
  border-color: var(--accent);
  background: rgba(255, 90, 0, 0.08);
}

.exchange-option.is-bound {
  cursor: not-allowed;
  opacity: 0.55;
}

.exchange-option.is-bound:hover {
  border-color: var(--line);
  background: var(--charcoal);
}

.exchange-option__logo-shell {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  width: 150px;
  height: 42px;
}

.exchange-option__logo {
  display: block;
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.exchange-option__status {
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

.exchange-option__arrow {
  color: var(--accent);
}

.bind-form__exchange {
  display: flex;
  align-items: center;
  gap: 1rem;
  min-height: 58px;
  padding: 0.85rem 1rem;
  border: 1px solid var(--line);
  background: var(--charcoal);
}

.bind-form__logo {
  display: block;
  width: 136px;
  height: 34px;
  object-fit: contain;
  object-position: left center;
}

.bind-form__exchange span {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.bind-field {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.bind-field > span {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.bind-field input,
.bind-secret input {
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

.bind-field input:focus,
.bind-secret input:focus {
  border-color: var(--accent);
}

.bind-secret {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 42px;
}

.bind-secret input {
  border-right: none;
}

.bind-secret button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 42px;
  border: 1px solid var(--line);
  background: var(--charcoal);
  color: var(--text);
  transition: border-color 220ms var(--ease-quiet), color 220ms var(--ease-quiet);
}

.bind-secret button:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.bind-field input.is-invalid,
.bind-secret.is-invalid input,
.bind-secret.is-invalid button {
  border-color: #ef4444;
}

.bind-field input.is-shaking,
.bind-secret.is-shaking {
  animation: bind-shake 260ms ease-in-out;
}

.bind-submit {
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

.bind-submit:hover {
  background: #ff7324;
  border-color: #ff7324;
  transform: translateY(-1px);
}

.bind-submit.is-blocked {
  box-shadow: inset 0 0 0 1px rgba(239, 68, 68, 0.45);
}

.bind-error,
.bind-success {
  margin: -0.2rem 0 0;
  font-family: var(--mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  text-align: center;
  text-transform: uppercase;
}

.bind-error {
  color: #ef4444;
}

.bind-success {
  color: #10b981;
}

@keyframes bind-shake {
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
  .bind-modal {
    padding: 1rem;
  }

  .exchange-option {
    grid-template-columns: 1fr 24px;
  }

  .exchange-option__status {
    display: none;
  }
}
</style>
