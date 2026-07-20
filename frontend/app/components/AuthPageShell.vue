<script setup lang="ts">
defineProps<{
  activeTab: 'signin' | 'signup'
  title: string
  subtitle: string
}>()
</script>

<template>
  <main class="auth-page">
    <section class="auth-page__form">
      <div class="auth-brand">
        <NuxtLink to="/signin">
          MAUTRADE<span class="auth-brand__dot" />
        </NuxtLink>
      </div>

      <div class="auth-tabs">
        <NuxtLink
          class="auth-tab"
          :class="{ active: activeTab === 'signin' }"
          to="/signin"
        >
          Sign In
        </NuxtLink>
        <NuxtLink
          class="auth-tab"
          :class="{ active: activeTab === 'signup' }"
          to="/signup"
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

    <AuthMarketPanel :active-tab="activeTab" />
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
  border-radius: 4px;
  background: var(--bg-elevated);
}

.auth-tab {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 38px;
  border-radius: 3px;
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 12px;
  font-weight: 700;
  transition: background 220ms var(--ease-quiet), color 220ms var(--ease-quiet);
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
  -webkit-text-fill-color: var(--text);
  appearance: none;
  outline: none;
  padding: 0 1rem;
  transition: border-color 220ms var(--ease-quiet), background 220ms var(--ease-quiet);
}

:deep(.auth-input:focus) {
  border-color: var(--accent);
  background: var(--charcoal);
  color: var(--text);
  -webkit-text-fill-color: var(--text);
}

:deep(.auth-input:-webkit-autofill),
:deep(.auth-input:-webkit-autofill:hover),
:deep(.auth-input:-webkit-autofill:focus),
:deep(.auth-input:-webkit-autofill:active) {
  -webkit-box-shadow: 0 0 0 1000px var(--bg-elevated) inset !important;
  -webkit-text-fill-color: var(--text) !important;
  border-color: var(--line) !important;
  caret-color: var(--text);
}

:deep(.auth-input:-webkit-autofill:focus) {
  -webkit-box-shadow: 0 0 0 1000px var(--charcoal) inset !important;
  border-color: var(--accent) !important;
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
  gap: 0.65rem;
  color: var(--text-mute);
  line-height: 1;
  cursor: pointer;
}

:deep(.auth-check input) {
  position: relative;
  flex: 0 0 auto;
  width: 16px;
  height: 16px;
  appearance: none;
  border: 1px solid var(--line-strong);
  border-radius: 3px;
  background: var(--bg-elevated);
  cursor: pointer;
  transition: background 180ms var(--ease-quiet), border-color 180ms var(--ease-quiet), box-shadow 180ms var(--ease-quiet);
}

:deep(.auth-check input::after) {
  content: "";
  position: absolute;
  top: 2px;
  left: 5px;
  width: 4px;
  height: 8px;
  border: solid #050505;
  border-width: 0 2px 2px 0;
  opacity: 0;
  transform: rotate(45deg) scale(0.75);
  transition: opacity 160ms var(--ease-quiet), transform 160ms var(--ease-quiet);
}

:deep(.auth-check input:checked) {
  border-color: var(--accent);
  background: var(--accent);
}

:deep(.auth-check input:checked::after) {
  opacity: 1;
  transform: rotate(45deg) scale(1);
}

:deep(.auth-check input:focus-visible) {
  outline: none;
  box-shadow: 0 0 0 3px rgba(255, 90, 0, 0.18);
}

:deep(.auth-check:hover input) {
  border-color: var(--accent);
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
