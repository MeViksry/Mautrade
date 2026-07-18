<script setup lang="ts">
import { ref } from 'vue'

definePageMeta({
  layout: 'default'
})

const email = ref('')
const password = ref('')
const rememberMe = ref(false)
const passwordVisible = ref(false)

const submitLogin = async () => {
  await navigateTo('/dashboard')
}
</script>

<template>
  <AuthPageShell
    active-tab="login"
    title="Mautrade dashboard access"
    subtitle="Sign in to access Active Layers, Exchange Bindings, Gas Fee, and Trading History for your Mautrade account."
  >
    <form
      class="auth-form"
      autocomplete="off"
      @submit.prevent="submitLogin"
    >
      <div class="auth-field">
        <label>Email Account</label>
        <input
          v-model="email"
          class="auth-input"
          type="email"
          placeholder="Enter Mautrade email"
          autocomplete="email"
        >
      </div>

      <div class="auth-field">
        <label>Password Account</label>
        <div class="password-wrap">
          <input
            v-model="password"
            class="auth-input"
            :type="passwordVisible ? 'text' : 'password'"
            placeholder="Enter Mautrade password"
            autocomplete="current-password"
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
        <label class="auth-check">
          <input
            v-model="rememberMe"
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

      <button
        class="auth-submit"
        type="submit"
      >
        Open Dashboard
        <UIcon name="lucide:arrow-right" />
      </button>

      <p class="auth-bottom">
        New to Mautrade?
        <NuxtLink
          class="auth-link"
          to="/register"
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
