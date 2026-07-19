<script setup lang="ts">
import { ref } from 'vue'

definePageMeta({
  layout: 'default'
})

const seoTitle = 'Admin Sign In | Mautrade'
const seoDescription = 'Sign in to Mautrade Admin panel.'

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

const submitLogin = async () => {
  await navigateTo('/admin')
}
</script>

<template>
  <AuthPageShell
    active-tab="login"
    title="Mautrade Admin Access"
    subtitle="Sign in to access the Mautrade administrator dashboard and management tools."
  >
    <form
      class="auth-form"
      autocomplete="off"
      @submit.prevent="submitLogin"
    >
      <div class="auth-field">
        <label>Admin Email</label>
        <input
          v-model="email"
          class="auth-input"
          type="text"
          inputmode="email"
          placeholder="Enter admin email"
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
        <label>Admin Password</label>
        <div class="password-wrap">
          <input
            v-model="password"
            class="auth-input"
            :type="passwordVisible ? 'text' : 'password'"
            placeholder="Enter admin password"
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
        <label class="auth-check">
          <input
            v-model="rememberMe"
            type="checkbox"
          >
          Remember me
        </label>
      </div>

      <button
        class="auth-submit"
        type="submit"
      >
        Open Admin Dashboard
        <UIcon name="lucide:arrow-right" />
      </button>
    </form>
  </AuthPageShell>
</template>

<style scoped>
.auth-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  margin-top: 2.5rem;
}

.auth-field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.auth-field label {
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-mute);
  padding-left: 2px;
}

.password-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.password-wrap button {
  position: absolute;
  right: 1rem;
  background: transparent;
  border: none;
  color: var(--silver);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s;
}

.password-wrap button:hover {
  color: var(--text);
  background: var(--charcoal);
}

.auth-input {
  width: 100%;
  height: 48px;
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 0 1rem;
  color: var(--text);
  font-size: 14px;
  transition: all 300ms var(--ease-quiet);
}

.auth-input:focus {
  outline: none;
  border-color: var(--accent);
  background: var(--charcoal);
  box-shadow: 0 0 0 3px rgba(255, 90, 0, 0.15);
}

.auth-input::placeholder {
  color: var(--silver);
}

.auth-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 0.5rem;
}

.auth-check {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 13px;
  color: var(--text-mute);
  cursor: pointer;
  user-select: none;
}

.auth-check input {
  width: 16px;
  height: 16px;
  accent-color: var(--accent);
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 4px;
}

.auth-submit {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  width: 100%;
  height: 48px;
  margin-top: 1rem;
  background: linear-gradient(135deg, var(--accent) 0%, #ff8a4c 100%);
  color: #fff;
  border: none;
  border-radius: 8px;
  font-family: 'Oswald', sans-serif;
  font-size: 15px;
  font-weight: 500;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  cursor: pointer;
  transition: all 300ms var(--ease-quiet);
  position: relative;
  overflow: hidden;
}

.auth-submit::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #ff8a4c 0%, var(--accent) 100%);
  opacity: 0;
  transition: opacity 300ms var(--ease-quiet);
}

.auth-submit:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(255, 90, 0, 0.3);
}

.auth-submit:hover::before {
  opacity: 1;
}

.auth-submit:active {
  transform: translateY(1px);
}

.auth-submit > * {
  position: relative;
  z-index: 1;
}
</style>
