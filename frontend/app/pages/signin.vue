<script setup lang="ts">
import { ref } from 'vue'
import { useAuth } from '../composables/useAuth'

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

const { login } = useAuth()
const errorMsg = ref('')
const isLoading = ref(false)

const submitLogin = async () => {
  errorMsg.value = ''
  isLoading.value = true
  try {
    const res = await login({ email: email.value, password: password.value })
    if (res.otpRequired) {
      // Need OTP, navigate to signup page which handles OTP verification
      await navigateTo('/signup?step=otp&email=' + encodeURIComponent(email.value))
    } else {
      await navigateTo('/dashboard')
    }
  } catch (err: unknown) {
    errorMsg.value = (err as Error).message || 'Login failed. Please try again.'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <AuthPageShell
    active-tab="signin"
    title="Mautrade dashboard access"
    subtitle="Sign in to access Active Layers, Exchange Bindings, Gas Fee, and Trading History for your Mautrade account."
  >
    <form
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
          to="/login"
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
</style>
