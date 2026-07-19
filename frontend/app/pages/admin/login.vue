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
  <main class="auth-page">
    <section class="auth-page__form">
      <div class="auth-brand">
        <NuxtLink to="/admin">
          MAUTRADE<span class="auth-brand__dot" />
        </NuxtLink>
        <span class="auth-brand__subtitle">ADMIN PORTAL</span>
      </div>

      <div class="auth-copy">
        <h1>Administrator Access</h1>
        <p>Sign in to the secure portal to manage users, track metrics, and oversee the Mautrade platform infrastructure.</p>
      </div>

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
          Secure Sign In
          <UIcon name="lucide:arrow-right" />
        </button>
      </form>
    </section>

    <!-- Admin Specific Panel -->
    <aside class="auth-market admin-panel">
      <div class="auth-market__status">
        <span />
        Mautrade Admin Network
      </div>

      <div class="auth-market__metrics">
        <div class="metric-box">
          <strong>12.4K</strong>
          <span>Total Users</span>
          <em>Global registered accounts</em>
        </div>
        <div class="metric-box">
          <strong>$1.5M</strong>
          <span>Total Revenue</span>
          <em>Annual gross volume</em>
        </div>
        <div class="metric-box metric-box--glow">
          <strong>99.9%</strong>
          <span>System Uptime</span>
          <em>All exchange endpoints healthy</em>
        </div>
        <div class="metric-box">
          <strong>450</strong>
          <span>Pending Gas Fees</span>
          <em>Awaiting admin verification</em>
        </div>
      </div>

      <div class="auth-market__footer">
        <h3>Master Control Panel.</h3>
        <h3 class="accent">
          Oversee every layer.
        </h3>
        <p>Mautrade requires rigorous authentication for administrative access. Every action is logged and monitored for compliance.</p>
      </div>
    </aside>
  </main>
</template>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: grid;
  grid-template-columns: minmax(500px, 1fr) minmax(520px, 1fr);
  background: var(--bg);
}

.auth-page__form {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-height: 100vh;
  padding: 3rem 5rem;
  background:
    linear-gradient(180deg, rgba(255, 90, 0, 0.04), transparent 22%),
    var(--bg);
}

.auth-brand {
  position: absolute;
  top: 3rem;
  left: 5rem;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.auth-brand a {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--text);
  font-family: 'Oswald', sans-serif;
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 0.15em;
  text-decoration: none;
}

.auth-brand__dot {
  width: 6px;
  height: 6px;
  background: var(--accent);
  display: inline-block;
}

.auth-brand__subtitle {
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.2em;
  color: var(--accent);
  border-left: 1px solid var(--line);
  padding-left: 1rem;
}

.auth-copy {
  margin-top: 2.5rem;
}

.auth-copy h1 {
  margin: 0;
  color: var(--text);
  font-family: var(--sans);
  font-size: clamp(2rem, 3.4vw, 3rem);
  font-weight: 300;
  line-height: 1.1;
}

.auth-copy p {
  max-width: 560px;
  margin-top: 0.9rem;
  color: var(--text-mute);
  font-size: 1.05rem;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 1.35rem;
  width: 100%;
  max-width: 760px;
  margin-top: 2.4rem;
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

/* Admin Market Panel */
.admin-panel {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 3rem 5rem;
  border-left: 1px solid var(--line);
  background:
    linear-gradient(180deg, rgba(255, 90, 0, 0.1), transparent 38%),
    var(--bg-elevated);
  position: relative;
  overflow: hidden;
}

.admin-panel::before {
  content: "";
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(var(--grid-line) 1px, transparent 1px),
    linear-gradient(90deg, var(--grid-line) 1px, transparent 1px);
  background-size: 52px 52px;
  mask-image: linear-gradient(to bottom, #000, transparent 88%);
  pointer-events: none;
}

.auth-market__status {
  position: absolute;
  top: 3rem;
  left: 5rem;
  display: inline-flex;
  align-items: center;
  gap: 0.75rem;
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-mute);
}

.auth-market__status span {
  width: 8px;
  height: 8px;
  background: var(--accent);
  border-radius: 50%;
  box-shadow: 0 0 10px rgba(255, 90, 0, 0.5);
  animation: pulse-admin 2s infinite ease-in-out;
}

@keyframes pulse-admin {
  0% { transform: scale(0.95); opacity: 0.8; }
  50% { transform: scale(1.1); opacity: 1; }
  100% { transform: scale(0.95); opacity: 0.8; }
}

.auth-market__metrics {
  position: relative;
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.5rem;
  margin-bottom: 4rem;
}

.metric-box {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 1.5rem;
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 8px;
}

.metric-box--glow {
  border-color: rgba(255, 90, 0, 0.3);
  background: linear-gradient(145deg, var(--bg-elevated), rgba(255, 90, 0, 0.03));
}

.metric-box strong {
  font-family: 'Oswald', sans-serif;
  font-size: 2rem;
  font-weight: 400;
  color: var(--text);
  line-height: 1.1;
}

.metric-box span {
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--silver);
}

.metric-box em {
  font-style: normal;
  font-size: 12px;
  color: var(--text-mute);
  margin-top: 0.25rem;
}

.auth-market__footer {
  position: relative;
}

.auth-market__footer h3 {
  font-family: var(--sans);
  font-size: 2.5rem;
  font-weight: 300;
  color: var(--text);
  line-height: 1.1;
  margin: 0;
}

.auth-market__footer h3.accent {
  color: var(--accent);
}

.auth-market__footer p {
  margin-top: 1rem;
  color: var(--text-mute);
  font-size: 1rem;
  max-width: 480px;
  line-height: 1.5;
}

@media (max-width: 1024px) {
  .auth-page {
    grid-template-columns: 1fr;
  }
  .admin-panel {
    display: none;
  }
}

@media (max-width: 640px) {
  .auth-page__form {
    padding: 2rem;
  }

  .auth-brand {
    position: relative;
    top: auto;
    left: auto;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }

  .auth-brand__subtitle {
    border-left: none;
    padding-left: 0;
  }

  .auth-copy h1 {
    font-size: 1.8rem;
  }
}
</style>
