<script setup lang="ts">
defineProps<{
  activeTab: 'login' | 'register'
  title: string
  subtitle: string
}>()
</script>

<template>
  <main class="auth-page">
    <section class="auth-page__form">
      <div class="auth-brand">
        <NuxtLink to="/login">
          MAUTRADE<span class="auth-brand__dot" />
        </NuxtLink>
      </div>

      <div class="auth-tabs">
        <NuxtLink
          class="auth-tab"
          :class="{ active: activeTab === 'login' }"
          to="/login"
        >
          Sign In
        </NuxtLink>
        <NuxtLink
          class="auth-tab"
          :class="{ active: activeTab === 'register' }"
          to="/register"
        >
          Sign Up
        </NuxtLink>
      </div>

      <div class="auth-copy">
        <h1>{{ title }}</h1>
        <p>{{ subtitle }}</p>
      </div>

      <slot />
    </section>

    <AuthMarketPanel />
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
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  padding: 3rem 5rem;
  background:
    linear-gradient(180deg, rgba(255, 90, 0, 0.04), transparent 22%),
    var(--bg);
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
}

.auth-brand__dot {
  width: 6px;
  height: 6px;
  background: var(--accent);
  display: inline-block;
}

.auth-tabs {
  align-self: flex-end;
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  width: 190px;
  margin-top: -2.2rem;
  padding: 0.25rem;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: var(--bg-elevated);
}

.auth-tab {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 38px;
  border-radius: 999px;
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 12px;
  font-weight: 700;
}

.auth-tab.active {
  background: var(--accent);
  color: #070707;
}

.auth-copy {
  margin-top: 5rem;
}

.auth-copy h1 {
  margin: 0;
  color: var(--text);
  font-family: var(--sans);
  font-size: clamp(2rem, 3.4vw, 3rem);
  font-weight: 300;
  line-height: 1;
}

.auth-copy p {
  max-width: 560px;
  margin-top: 0.9rem;
  color: var(--text-mute);
  font-size: 1.05rem;
}

:deep(.auth-form) {
  display: flex;
  flex-direction: column;
  gap: 1.35rem;
  width: 100%;
  max-width: 760px;
  margin-top: 2.4rem;
}

:deep(.auth-field) {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

:deep(.auth-field label),
:deep(.auth-check label) {
  font-family: var(--mono);
  font-size: 12px;
  color: var(--silver);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

:deep(.auth-input) {
  width: 100%;
  height: 52px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--bg-elevated);
  color: var(--text);
  outline: none;
  padding: 0 1rem;
  transition: border-color 220ms var(--ease-quiet), background 220ms var(--ease-quiet);
}

:deep(.auth-input:focus) {
  border-color: var(--accent);
  background: var(--charcoal);
}

:deep(.auth-row) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

:deep(.auth-check) {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  color: var(--text-mute);
}

:deep(.auth-check input) {
  width: 20px;
  height: 20px;
  accent-color: var(--accent);
}

:deep(.auth-link) {
  color: var(--accent);
  font-family: var(--mono);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

:deep(.auth-submit) {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  width: 100%;
  min-height: 56px;
  margin-top: 0.75rem;
  border: none;
  border-radius: 4px;
  background: var(--accent);
  color: #050505;
  font-family: var(--mono);
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  cursor: pointer;
  box-shadow: 0 18px 30px rgba(255, 90, 0, 0.18);
}

:deep(.auth-submit:hover) {
  background: #ff7a1a;
}

:deep(.auth-bottom) {
  margin-top: 0.9rem;
  text-align: center;
  color: #7d858f;
}

@media (max-width: 980px) {
  .auth-page {
    grid-template-columns: 1fr;
  }

  .auth-page__form {
    min-height: auto;
    padding: 2rem;
  }

  .auth-tabs {
    align-self: flex-start;
    margin-top: 2rem;
  }
}

@media (max-width: 560px) {
  .auth-page__form {
    padding: 1.25rem;
  }

  :deep(.auth-row) {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
