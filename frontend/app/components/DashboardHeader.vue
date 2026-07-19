<script setup lang="ts">
const isSidebarOpen = useState<boolean | null>('sidebar-open', () => null)
const theme = useState<'dark' | 'light'>('dashboard-theme', () => 'dark')
const isLightMode = computed(() => theme.value === 'light')

const toggleSidebar = () => {
  isSidebarOpen.value = !isSidebarOpen.value
}

const toggleTheme = () => {
  theme.value = isLightMode.value ? 'dark' : 'light'
}
</script>

<template>
  <header class="dashboard-header">
    <div class="header-left">
      <button
        aria-label="Toggle Sidebar"
        class="sidebar-toggle"
        @click="toggleSidebar"
      >
        <svg
          class="sidebar-icon-open"
          :class="{ 'force-show': isSidebarOpen === true, 'force-hide': isSidebarOpen === false }"
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
        <svg
          class="sidebar-icon-closed"
          :class="{ 'force-show': isSidebarOpen === false, 'force-hide': isSidebarOpen === true }"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M6.5 17.5L6.5 6.5M7.8 3H16.2C17.8802 3 18.7202 3 19.362 3.32698C19.9265 3.6146 20.3854 4.07354 20.673 4.63803C21 5.27976 21 6.11984 21 7.8V16.2C21 17.8802 21 18.7202 20.673 19.362C20.3854 19.9265 19.9265 20.3854 19.362 20.673C18.7202 21 17.8802 21 16.2 21H7.8C6.11984 21 5.27976 21 4.63803 20.673C4.07354 20.3854 3.6146 19.9265 3.32698 19.362C3 18.7202 3 17.8802 3 16.2V7.8C3 6.11984 3 5.27976 3.32698 4.63803C3.6146 4.07354 4.07354 3.6146 4.63803 3.32698C5.27976 3 6.11984 3 7.8 3Z"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </button>
    </div>

    <div class="header-right">
      <button
        class="theme-toggle"
        type="button"
        :aria-label="isLightMode ? 'Switch to dark mode' : 'Switch to light mode'"
        :aria-pressed="isLightMode"
        @click="toggleTheme"
      >
        <span class="theme-toggle__track">
          <span
            class="theme-toggle__thumb"
            :class="{ 'theme-toggle__thumb--light': isLightMode }"
          >
            <svg
              v-if="isLightMode"
              class="theme-toggle__icon"
              width="100%"
              height="100%"
              viewBox="0 0 24 24"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                d="M12 2V4M12 20V22M4 12H2M6.31412 6.31412L4.8999 4.8999M17.6859 6.31412L19.1001 4.8999M6.31412 17.69L4.8999 19.1042M17.6859 17.69L19.1001 19.1042M22 12H20M17 12C17 14.7614 14.7614 17 12 17C9.23858 17 7 14.7614 7 12C7 9.23858 9.23858 7 12 7C14.7614 7 17 9.23858 17 12Z"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
            <svg
              v-else
              class="theme-toggle__icon"
              width="100%"
              height="100%"
              viewBox="0 0 24 24"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                d="M18 2L18.6178 3.23558C18.8833 3.76656 19.016 4.03205 19.1934 4.26211C19.3507 4.46626 19.5337 4.64927 19.7379 4.80664C19.9679 4.98397 20.2334 5.11672 20.7644 5.38221L22 6L20.7644 6.61779C20.2334 6.88328 19.9679 7.01603 19.7379 7.19336C19.5337 7.35073 19.3507 7.53374 19.1934 7.73789C19.016 7.96795 18.8833 8.23344 18.6178 8.76442L18 10L17.3822 8.76442C17.1167 8.23344 16.984 7.96795 16.8066 7.73789C16.6493 7.53374 16.4663 7.35073 16.2621 7.19336C16.0321 7.01603 15.7666 6.88328 15.2356 6.61779L14 6L15.2356 5.38221C15.7666 5.11672 16.0321 4.98397 16.2621 4.80664C16.4663 4.64927 16.6493 4.46626 16.8066 4.26211C16.984 4.03205 17.1167 3.76656 17.3822 3.23558L18 2Z"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                d="M21 13.3893C19.689 15.689 17.2145 17.2395 14.3779 17.2395C10.1711 17.2395 6.76075 13.8292 6.76075 9.62233C6.76075 6.78554 8.31149 4.31094 10.6115 3C5.77979 3.45812 2 7.52692 2 12.4785C2 17.7371 6.26292 22 11.5215 22C16.4729 22 20.5415 18.2206 21 13.3893Z"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </span>
        </span>
      </button>

      <div class="user-profile">
        <div class="user-avatar" />
        <div class="user-info">
          <span class="user-name">User Account</span>
          <span class="user-id">UID: 88472910</span>
        </div>
      </div>
    </div>
  </header>
</template>

<style scoped>
.dashboard-header {
  height: 72px;
  padding: 0 3rem;
  display: flex;
  gap: 1rem;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--line);
  background: var(--header-bg);
  backdrop-filter: blur(10px);
  position: sticky;
  top: 0;
  z-index: 20;
}

.header-left {
  display: flex;
  align-items: center;
}

.sidebar-toggle {
  width: 40px;
  height: 40px;
  background: var(--bg-elevated);
  border: 1px solid transparent;
  color: var(--text-mute);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem;
  border-radius: 4px;
  transition: all 0.2s;
  margin-left: -1rem; /* Adjust alignment to perfectly align with padding */
}

.sidebar-toggle:hover {
  color: var(--text);
  background: var(--charcoal);
  border-color: var(--line);
}

/* Sidebar Toggle Icons Responsive CSS */
.sidebar-icon-open {
  display: block;
}
.sidebar-icon-closed {
  display: none;
}

@media (max-width: 1180px) {
  .sidebar-icon-open {
    display: none;
  }
  .sidebar-icon-closed {
    display: block;
  }
}

@media (pointer: coarse) and (max-width: 1366px) {
  .sidebar-icon-open {
    display: none;
  }
  .sidebar-icon-closed {
    display: block;
  }
}

.sidebar-icon-open.force-show,
.sidebar-icon-closed.force-show {
  display: block !important;
}

.sidebar-icon-open.force-hide,
.sidebar-icon-closed.force-hide {
  display: none !important;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 1rem;
  min-width: 0;
}

.theme-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 58px;
  height: 34px;
  padding: 0;
  color: var(--text);
  background: transparent;
  border: none;
}

.theme-toggle__track {
  position: relative;
  display: block;
  width: 58px;
  height: 34px;
  background: var(--charcoal);
  border: 1px solid var(--line);
  border-radius: 999px;
  transition: background 220ms var(--ease-quiet), border-color 220ms var(--ease-quiet);
}

.theme-toggle:hover .theme-toggle__track {
  border-color: var(--line-strong);
}

.theme-toggle__thumb {
  position: absolute;
  top: 50%;
  left: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  color: var(--silver);
  background: var(--bg);
  border: 1px solid var(--line);
  border-radius: 50%;
  line-height: 1;
  transform: translateY(-50%);
  transition: transform 220ms var(--ease-quiet), color 220ms var(--ease-quiet), background 220ms var(--ease-quiet);
}

.theme-toggle__thumb--light {
  color: var(--accent);
  transform: translate(24px, -50%);
}

.theme-toggle__icon {
  display: block;
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0 0.75rem 0 2px;
  height: 34px;
  background: var(--charcoal);
  border: 1px solid var(--line);
  border-radius: 999px;
  min-width: 0;
}

.user-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--accent) 0%, #ff8a4c 100%);
}

.user-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
  justify-content: center;
  gap: 2px;
}

.user-name {
  font-size: 12px;
  line-height: 1;
  font-weight: 500;
  color: var(--text);
  white-space: nowrap;
}

.user-id {
  font-family: var(--mono);
  font-size: 9px;
  line-height: 1;
  color: var(--text-mute);
  white-space: nowrap;
}

@media (max-width: 1180px) {
  .dashboard-header {
    height: 64px;
    padding: 0 1.5rem;
  }

  .sidebar-toggle {
    margin-left: -0.5rem;
  }
}

@media (pointer: coarse) and (max-width: 1366px) {
  .dashboard-header {
    height: 64px;
    padding: 0 1.5rem;
  }

  .sidebar-toggle {
    margin-left: -0.5rem;
  }
}

@media (max-width: 640px) {
  .dashboard-header {
    padding: 0 1rem;
  }

  .header-right {
    gap: 0.65rem;
  }

  .theme-toggle,
  .theme-toggle__track {
    width: 52px;
    height: 32px;
  }

  .theme-toggle__thumb {
    width: 22px;
    height: 22px;
  }

  .theme-toggle__thumb--light {
    transform: translate(22px, -50%);
  }

  .theme-toggle__icon {
    width: 14px;
    height: 14px;
  }

  .user-profile {
    gap: 0;
    padding: 0 3px;
    justify-content: center;
  }

  .user-info {
    display: none;
  }
}
</style>
