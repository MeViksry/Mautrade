<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

const compactSidebarQuery = '(max-width: 1180px), (pointer: coarse) and (max-width: 1366px)'
const isSidebarOpen = useState<boolean | null>('sidebar-open', () => null)
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
      'sidebar--closed': isSidebarOpen === false,
      'sidebar--open': isSidebarOpen === true,
      'sidebar--compact': isCompactSidebar,
      'sidebar--mounted': isMounted
    }"
  >
    <div class="sidebar__logo">
      <NuxtLink to="/dashboard">
        MAUTRADE<span class="dot" />
      </NuxtLink>

      <button
        v-if="isCompactSidebar"
        class="sidebar__close-btn"
        aria-label="Close sidebar"
        @click="isSidebarOpen = false"
      >
        <svg
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M17.5 17.5L17.5 6.5M7.8 3H16.2C17.8802 3 18.7202 3 19.362 3.32698C19.9265 3.6146 20.3854 4.07354 20.673 4.63803C21 5.27976 21 6.11984 21 7.8V16.2C21 17.8802 21 18.7202 20.673 19.362C20.3854 19.9265 19.9265 20.3854 19.362 20.673C18.7202 21 17.8802 21 16.2 21H7.8C6.11984 21 5.27976 21 4.63803 20.673C4.07354 20.3854 3.6146 19.9265 3.32698 19.362C3 18.7202 3 17.8802 3 16.2V7.8C3 6.11984 3 5.27976 3.32698 4.63803C3.6146 4.07354 4.07354 3.6146 4.63803 3.32698C5.27976 3 6.11984 3 7.8 3Z"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </button>
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
  margin-left: -260px !important;
}

.sidebar--open {
  margin-left: 0 !important;
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

@media (max-width: 1180px), (pointer: coarse) and (max-width: 1366px) {
  .sidebar {
    margin-left: -260px; /* Mobile default closed */
  }
}

@media (max-width: 420px) {
  .sidebar {
    width: min(260px, 86vw);
  }

  .sidebar--closed {
    margin-left: min(-260px, -86vw) !important;
  }
}

.sidebar__logo {
  height: 72px;
  padding: 0 1rem 0 2rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
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

.sidebar__close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: transparent;
  border: none;
  color: var(--text-mute);
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.2s;
}

.sidebar__close-btn:hover {
  background: var(--charcoal);
  color: var(--text);
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
