<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

const compactSidebarQuery = '(max-width: 1180px), (pointer: coarse) and (max-width: 1366px)'
const isSidebarOpen = useState('sidebar-open', () => false)
const isCompactSidebar = ref(false)
const isMounted = ref(false)

let mediaQuery: MediaQueryList | null = null

const syncSidebarMode = (matches: boolean) => {
  isCompactSidebar.value = matches
  isSidebarOpen.value = !matches
}

const handleSidebarModeChange = (event: MediaQueryListEvent) => {
  syncSidebarMode(event.matches)
}

onMounted(() => {
  mediaQuery = window.matchMedia(compactSidebarQuery)
  syncSidebarMode(mediaQuery.matches)
  mediaQuery.addEventListener('change', handleSidebarModeChange)

  // Enable transitions after a short delay to prevent initial load animation
  setTimeout(() => {
    isMounted.value = true
  }, 50)
})

onBeforeUnmount(() => {
  mediaQuery?.removeEventListener('change', handleSidebarModeChange)
})

const navItems = [
  { label: 'Overview', to: '/dashboard', icon: 'lucide:layout-dashboard' },
  { label: 'Active Layers', to: '/dashboard/layers', icon: 'lucide:layers' },
  { label: 'History', to: '/dashboard/history', icon: 'lucide:history' },
  { label: 'Gas Fee', to: '/dashboard/gas-fee', icon: 'lucide:fuel' },
  { label: 'API Keys', to: '/dashboard/api-keys', icon: 'lucide:key' },
  { label: 'Settings', to: '/dashboard/settings', icon: 'lucide:settings' }
]

const closeCompactSidebar = () => {
  if (isCompactSidebar.value) {
    isSidebarOpen.value = false
  }
}
</script>

<template>
  <div
    v-if="isCompactSidebar && isSidebarOpen"
    class="sidebar-overlay"
    @click="isSidebarOpen = false"
  />

  <aside
    class="sidebar"
    :class="{
      'sidebar--closed': !isSidebarOpen,
      'sidebar--compact': isCompactSidebar,
      'sidebar--mounted': isMounted
    }"
  >
    <div class="sidebar__logo">
      <NuxtLink to="/dashboard">
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
        @click="closeCompactSidebar"
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
  flex-shrink: 0;
  margin-left: 0;
  overflow: hidden;
}

.sidebar--mounted {
  transition: margin-left 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.sidebar--closed {
  margin-left: -260px;
}

.sidebar-overlay {
  display: block;
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  z-index: 30;
}

.sidebar--compact {
  position: fixed;
  left: 0;
  top: 0;
  z-index: 40;
  height: 100dvh; /* For mobile browsers */
  box-shadow: 4px 0 24px rgba(0, 0, 0, 0.5);
}

@media (max-width: 420px) {
  .sidebar {
    width: min(260px, 86vw);
  }

  .sidebar--closed {
    margin-left: min(-260px, -86vw);
  }
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
