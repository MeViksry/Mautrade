<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAuth } from '~/composables/useAuth'

definePageMeta({
  layout: 'default'
})

const seoTitle = 'Sign In | Mautrade'
const seoDescription = 'Sign in to Mautrade to access Active Layers, Exchange Bindings, Gas Fee, and Trading History.'

useSeoMeta({
  title: seoTitle,
  description: seoDescription,
  ogTitle: seoTitle,
  ogDescription: seoDescription,
  twitterTitle: seoTitle,
  twitterDescription: seoDescription
})

const email = ref('')
const password = ref('')
const rememberMe = ref(false)
const passwordVisible = ref(false)

const unlockReadonlyInput = (event: Event) => {
  const input = event.currentTarget as HTMLInputElement
  input.readOnly = false
}

const loginStep = ref<'credentials' | 'otp'>('credentials')
const otp = ref('')
const otpShake = ref(false)
const submitAttempted = ref(false)

const otpInvalid = computed(() => otp.value.trim().length !== 6)

const { login, verifyOtp: verifyAuthOtp } = useAuth()
const errorMsg = ref('')
const isLoading = ref(false)
const resendLoading = ref(false)
const resendMsg = ref('')

const triggerOtpShake = () => {
  otpShake.value = false
  window.requestAnimationFrame(() => {
    otpShake.value = true
  })
}

const submitLogin = async () => {
  submitAttempted.value = true
  errorMsg.value = ''
  isLoading.value = true
  try {
    const res = await login({ email: email.value, password: password.value })
    if (res.otpRequired) {
      // Need OTP, transition to OTP step
      loginStep.value = 'otp'
      submitAttempted.value = false
      otp.value = ''
    } else {
      await navigateTo('/dashboard')
    }
  } catch (err: unknown) {
    errorMsg.value = (err as Error).message || 'Login failed. Please try again.'
  } finally {
    isLoading.value = false
  }
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
      purpose: 'login_verify'
    })
    await navigateTo('/dashboard')
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
    active-tab="signin"
    :title="loginStep === 'credentials' ? 'Mautrade dashboard access' : 'Verify Mautrade email'"
    :subtitle="loginStep === 'credentials' ? 'Sign in to access Active Layers, Exchange Bindings, Gas Fee, and Trading History for your Mautrade account.' : 'Enter the email OTP to securely log into your Mautrade account.'"
  >
    <form
      v-if="loginStep === 'credentials'"
      class="auth-form"
      autocomplete="off"
      @submit.prevent="submitLogin"
    >
      <div class="auth-field">
        <label for="email">Email Account</label>
        <input
          id="email"
          v-model="email"
          name="email"
          class="auth-input"
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
      </div>

      <div class="auth-field">
        <label for="password">Password Account</label>
        <div class="password-wrap">
          <input
            id="password"
            v-model="password"
            name="password"
            class="auth-input"
            :type="passwordVisible ? 'text' : 'password'"
            placeholder="Enter Mautrade password"
            autocomplete="current-password"
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
      </div>
      <div class="auth-field">
        <label for="email">Email Account</label>
        <input
          id="email"
          v-model="email"
          name="email"
          class="auth-input"
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
      </div>

      <div class="auth-field">
        <label for="password">Password Account</label>
        <div class="password-wrap">
          <input
            id="password"
            v-model="password"
            name="password"
            class="auth-input"
            :type="passwordVisible ? 'text' : 'password'"
            placeholder="Enter Mautrade password"
            autocomplete="current-password"
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
      </div>

      <div class="auth-row">
        <label
          for="rememberMe"
          class="auth-check"
        >
          <input
            id="rememberMe"
            v-model="rememberMe"
            name="rememberMe"
            type="checkbox"
          >
          Remember me
        </label>
        <NuxtLink
          class="auth-link"
          to="/signin"
        >
          Reset account password
        </NuxtLink>
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
        {{ isLoading ? 'Signing In...' : 'Sign In' }}
        <UIcon name="lucide:arrow-right" />
      </button>

      <p class="auth-bottom">
        New to Mautrade?
        <NuxtLink
          class="auth-link"
          to="/signup"
        >
          Sign Up
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
        {{ isLoading ? 'Verifying...' : 'Sign In to Dashboard' }}
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
        @click="loginStep = 'credentials'"
      >
        <UIcon name="lucide:arrow-left" />
        Back to login
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

@keyframes otp-shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-4px); }
  75% { transform: translateX(4px); }
}
</style>
