<script setup lang="ts">
const navItems = [
  { label: 'Overview', to: '/dashboard', icon: 'lucide:layout-dashboard' },
  { label: 'Active Layers', to: '/dashboard/layers', icon: 'lucide:layers' },
  { label: 'History', to: '/dashboard/history', icon: 'lucide:history' },
  { label: 'API Keys', to: '/dashboard/api-keys', icon: 'lucide:key' },
  { label: 'Settings', to: '/dashboard/settings', icon: 'lucide:settings' }
]
</script>

<template>
  <aside class="sidebar">
    <div class="sidebar__logo">
      <NuxtLink to="/">
        MAUTRADE<span class="dot" />
      </NuxtLink>
    </div>

    <nav class="sidebar__nav">
      <NuxtLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="sidebar__link"
        active-class="sidebar__link--active"
      >
        <UIcon
          :name="item.icon"
          class="sidebar__icon"
        />
        {{ item.label }}
      </NuxtLink>
    </nav>

    <div class="sidebar__bottom">
      <button class="sidebar__logout">
        <UIcon
          name="lucide:log-out"
          class="sidebar__icon"
        />
        Sign Out
      </button>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 260px;
  background: var(--bg-elevated);
  border-right: 1px solid var(--line);
  display: flex;
  flex-direction: column;
  height: 100vh;
  position: sticky;
  top: 0;
}

.sidebar__logo {
  height: 72px;
  padding: 0 2rem;
  display: flex;
  align-items: center;
  font-family: 'Oswald', sans-serif;
  font-weight: 700;
  font-size: 20px;
  letter-spacing: 0.15em;
  border-bottom: 1px solid var(--line);
}
.sidebar__logo a {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.sidebar__logo .dot {
  width: 6px; height: 6px;
  background: var(--accent);
  display: inline-block;
}

.sidebar__nav {
  padding: 2rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  flex: 1;
}

.sidebar__link {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.85rem 1rem;
  border-radius: 4px;
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--silver);
  transition: all 300ms var(--ease-quiet);
}

.sidebar__link:hover {
  background: var(--charcoal);
  color: var(--text);
}

.sidebar__link--active {
  background: rgba(255, 90, 0, 0.1);
  color: var(--accent);
  font-weight: 500;
}

.sidebar__icon {
  width: 1.2rem;
  height: 1.2rem;
}

.sidebar__bottom {
  padding: 2rem;
  border-top: 1px solid var(--line);
}

.sidebar__logout {
  display: flex;
  align-items: center;
  gap: 1rem;
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--text-dim);
  background: none;
  border: none;
  transition: color 300ms var(--ease-quiet);
}

.sidebar__logout:hover {
  color: #ff4d4d;
}
</style>
