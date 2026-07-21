<script setup lang="ts">
import { computed, ref } from 'vue'
import { useAuth } from '~/composables/useAuth'

definePageMeta({
  layout: 'default'
})

const seoTitle = 'Create Account | Mautrade'
const seoDescription = 'Create your Mautrade account with name, email, password confirmation, and email OTP verification.'

useSeoMeta({
  title: seoTitle,
  description: seoDescription,
  ogTitle: seoTitle,
  ogDescription: seoDescription,
  twitterTitle: seoTitle,
  twitterDescription: seoDescription
})

const registerStep = ref<'credentials' | 'otp'>('credentials')
const fullName = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const passwordVisible = ref(false)
const confirmPasswordVisible = ref(false)
const otp = ref('')
const registerShake = ref(false)
const otpShake = ref(false)
const submitAttempted = ref(false)

const fullNameInvalid = computed(() => fullName.value.trim().length < 2)
const emailInvalid = computed(() => !email.value.trim())
const passwordInvalid = computed(() => password.value.length < 8)
const confirmPasswordInvalid = computed(() => confirmPassword.value !== password.value || !confirmPassword.value)
const registerInvalid = computed(() => fullNameInvalid.value || emailInvalid.value || passwordInvalid.value || confirmPasswordInvalid.value)
const otpInvalid = computed(() => otp.value.trim().length !== 6)

const unlockReadonlyInput = (event: Event) => {
  const input = event.currentTarget as HTMLInputElement
  input.readOnly = false
}

const { register, login, verifyOtp: verifyAuthOtp } = useAuth()
const errorMsg = ref('')
const isLoading = ref(false)
const resendLoading = ref(false)
const resendMsg = ref('')

const submitRegister = async () => {
  submitAttempted.value = true

  if (registerInvalid.value) {
    registerShake.value = false
    window.requestAnimationFrame(() => {
      registerShake.value = true
    })
    return
  }

  errorMsg.value = ''
  isLoading.value = true
  try {
    await register({
      email: email.value,
      name: fullName.value,
      password: password.value,
      confirm_password: confirmPassword.value
    })
    registerStep.value = 'otp'
    submitAttempted.value = false
    otp.value = ''
  } catch (err: unknown) {
    if ((err as Error).message === 'Account already exists') {
      try {
        const res = await login({ email: email.value, password: password.value })
        if (res?.otpRequired) {
          registerStep.value = 'otp'
          submitAttempted.value = false
          otp.value = ''
          return
        }
      } catch {
        // If login fails (e.g. wrong password), just show the original error
      }
    }
    errorMsg.value = (err as Error).message
  } finally {
    isLoading.value = false
  }
}

const triggerOtpShake = () => {
  otpShake.value = false
  window.requestAnimationFrame(() => {
    otpShake.value = true
  })
}

const handleResendOtp = async () => {
  resendMsg.value = ''
  errorMsg.value = ''
  resendLoading.value = true
  try {
    // Calling login will generate and send a new OTP if email is unverified
    await login({
      email: email.value,
      password: password.value
    })
    resendMsg.value = 'A new OTP has been sent to your email.'
  } catch (err: unknown) {
    errorMsg.value = (err as Error).message
  } finally {
    resendLoading.value = false
  }
}

const verifyOtp = async () => {
  submitAttempted.value = true

  if (otpInvalid.value) {
    triggerOtpShake()
    return
  }

  errorMsg.value = ''
  isLoading.value = true
  try {
    await verifyAuthOtp({
      email: email.value,
      code: otp.value,
      purpose: 'register_verify'
    })
    await navigateTo('/onboarding')
  } catch (err: unknown) {
    errorMsg.value = (err as Error).message
    triggerOtpShake()
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <AuthPageShell
    active-tab="signup"
    :title="registerStep === 'credentials' ? 'Create Mautrade access' : 'Verify Mautrade email'"
    :subtitle="registerStep === 'credentials' ? 'Create access with name, email, and password, then verify OTP before setting country, exchange, and gas fee.' : 'Enter the email OTP to continue into Mautrade onboarding.'"
  >
    <form
      v-if="registerStep === 'credentials'"
      class="auth-form"
      :class="{ 'is-shaking': registerShake }"
      autocomplete="off"
      @submit.prevent="submitRegister"
      @animationend="registerShake = false"
    >
      <div class="auth-field">
        <label for="fullName">Full Name</label>
        <input
          id="fullName"
          v-model="fullName"
          name="fullName"
          class="auth-input"
          :class="{ 'is-invalid': submitAttempted && fullNameInvalid }"
          type="text"
          placeholder="Enter full name"
          autocomplete="name"
          readonly
          @focus="unlockReadonlyInput"
          @pointerdown="unlockReadonlyInput"
        >
        <small
          v-if="submitAttempted && fullNameInvalid"
          class="field-error"
        >
          Full name is required.
        </small>
      </div>

      <div class="auth-field">
        <label for="email">Email Account</label>
        <input
          id="email"
          v-model="email"
          name="email"
          class="auth-input"
          :class="{ 'is-invalid': submitAttempted && emailInvalid }"
          type="text"
          inputmode="email"
          placeholder="Enter Mautrade email"
          autocomplete="off"
          autocapitalize="none"
          autocorrect="off"
          readonly
          spellcheck="false"
          @focus="unlockReadonlyInput"
          @pointerdown="unlockReadonlyInput"
        >
        <small
          v-if="submitAttempted && emailInvalid"
          class="field-error"
        >
          Email account is required.
        </small>
      </div>

      <div class="auth-field">
        <label for="password">Create Password</label>
        <div class="password-wrap">
          <input
            id="password"
            v-model="password"
            name="password"
            class="auth-input"
            :class="{ 'is-invalid': submitAttempted && passwordInvalid }"
            :type="passwordVisible ? 'text' : 'password'"
            placeholder="Create Mautrade password"
            autocomplete="new-password"
            readonly
            @focus="unlockReadonlyInput"
            @pointerdown="unlockReadonlyInput"
          >
          <button
            type="button"
            :aria-label="passwordVisible ? 'Hide password' : 'Show password'"
            @click="passwordVisible = !passwordVisible"
          >
            <UIcon :name="passwordVisible ? 'lucide:eye-off' : 'lucide:eye'" />
          </button>
        </div>
        <small
          v-if="submitAttempted && passwordInvalid"
          class="field-error"
        >
          Use at least 8 characters.
        </small>
      </div>

      <div class="auth-field">
        <label for="confirmPassword">Confirm Password</label>
        <div class="password-wrap">
          <input
            id="confirmPassword"
            v-model="confirmPassword"
            name="confirmPassword"
            class="auth-input"
            :class="{ 'is-invalid': submitAttempted && confirmPasswordInvalid }"
            :type="confirmPasswordVisible ? 'text' : 'password'"
            placeholder="Confirm Mautrade password"
            autocomplete="new-password"
            readonly
            @focus="unlockReadonlyInput"
            @pointerdown="unlockReadonlyInput"
          >
          <button
            type="button"
            :aria-label="confirmPasswordVisible ? 'Hide password' : 'Show password'"
            @click="confirmPasswordVisible = !confirmPasswordVisible"
          >
            <UIcon :name="confirmPasswordVisible ? 'lucide:eye-off' : 'lucide:eye'" />
          </button>
        </div>
        <small
          v-if="submitAttempted && confirmPasswordInvalid"
          class="field-error"
        >
          Password confirmation must match.
        </small>
      </div>

      <div
        v-if="errorMsg"
        class="auth-error"
        style="color: red; margin-bottom: 1rem; font-size: 0.875rem;"
      >
        {{ errorMsg }}
      </div>

      <button
        class="auth-submit"
        type="submit"
        :disabled="isLoading"
      >
        {{ isLoading ? 'Sending...' : 'Send Email OTP' }}
        <UIcon name="lucide:mail-check" />
      </button>

      <p class="auth-bottom">
        Already have Mautrade access?
        <NuxtLink
          class="auth-link"
          to="/signin"
        >
          Sign In
        </NuxtLink>
      </p>
    </form>

    <form
      v-else
      class="auth-form"
      autocomplete="off"
      @submit.prevent="verifyOtp"
    >
      <div class="otp-card">
        <UIcon name="lucide:shield-check" />
        <span>{{ email || 'user@email.com' }}</span>
      </div>

      <div class="auth-field">
        <label for="otp">Mautrade Email OTP</label>
        <input
          id="otp"
          v-model="otp"
          name="otp"
          class="auth-input otp-input"
          :class="{ 'is-invalid': submitAttempted && otpInvalid, 'is-shaking': otpShake }"
          type="text"
          inputmode="numeric"
          maxlength="6"
          placeholder="000000"
          autocomplete="one-time-code"
          @animationend="otpShake = false"
        >
      </div>

      <div
        v-if="errorMsg"
        class="auth-error"
        style="color: red; margin-bottom: 1rem; font-size: 0.875rem; text-align: center;"
      >
        {{ errorMsg }}
      </div>

      <div
        v-if="resendMsg"
        class="auth-success"
        style="color: #10b981; margin-bottom: 1rem; font-size: 0.875rem; text-align: center;"
      >
        {{ resendMsg }}
      </div>

      <button
        class="auth-submit"
        type="submit"
        :disabled="isLoading || resendLoading"
      >
        {{ isLoading ? 'Verifying...' : 'Continue Mautrade Onboarding' }}
        <UIcon name="lucide:arrow-right" />
      </button>

      <button
        class="auth-submit"
        type="button"
        style="margin-top: 0.5rem; background: transparent; border: 1px solid rgba(255, 255, 255, 0.1); color: #94a3b8;"
        :disabled="isLoading || resendLoading"
        @click="handleResendOtp"
      >
        {{ resendLoading ? 'Sending...' : 'Resend OTP' }}
        <UIcon
          name="lucide:refresh-cw"
          :class="{ 'animate-spin': resendLoading }"
        />
      </button>

      <button
        class="back-link"
        type="button"
        @click="registerStep = 'credentials'"
      >
        <UIcon name="lucide:arrow-left" />
        Back to account setup
      </button>
    </form>
  </AuthPageShell>
</template>

<style scoped>
.password-wrap {
  position: relative;
}

.password-wrap .auth-input {
  padding-right: 3.25rem;
}

.password-wrap button {
  position: absolute;
  top: 50%;
  right: 1rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  color: #8d95a1;
  transform: translateY(-50%);
}

.otp-card {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  min-height: 58px;
  padding: 0 1rem;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--bg-elevated);
  color: var(--silver);
  font-family: var(--mono);
}

.otp-card svg {
  color: var(--accent);
}

.otp-input {
  letter-spacing: 0.35em;
  text-align: center;
}

.otp-input.is-invalid {
  border-color: #ef4444;
}

.auth-input.is-invalid {
  border-color: #ef4444;
}

.otp-input.is-shaking {
  animation: otp-shake 260ms ease-in-out;
}

.auth-form.is-shaking {
  animation: otp-shake 260ms ease-in-out;
}

.field-error {
  color: #ef4444;
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.04em;
}

.back-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border: none;
  background: transparent;
  color: var(--accent);
  font-family: var(--mono);
  font-size: 12px;
  font-weight: 700;
}

@keyframes otp-shake {
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
</style>
