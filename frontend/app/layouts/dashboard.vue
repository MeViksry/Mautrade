<script setup lang="ts">
import DashboardSidebar from '~/components/DashboardSidebar.vue'
</script>

<template>
  <div class="dashboard-layout">
    <DashboardSidebar />

    <main class="dashboard-main">
      <header class="dashboard-header">
        <div class="dashboard-header__title">
          <div class="eyebrow-mini">
            Follower Dashboard
          </div>
          <h2>Overview</h2>
        </div>

        <div class="dashboard-header__actions">
          <div class="user-profile">
            <div class="user-avatar" />
            <div class="user-info">
              <span class="user-name">User Account</span>
              <span class="user-id">UID: 88472910</span>
            </div>
          </div>
        </div>
      </header>

      <div class="dashboard-content">
        <slot />
      </div>
    </main>
  </div>
</template>

<style scoped>
.dashboard-layout {
  display: flex;
  min-height: 100vh;
  background: var(--bg);
}

.dashboard-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow-x: hidden;
}

.dashboard-header {
  padding: 2rem 3rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--line);
  background: rgba(10, 10, 10, 0.8);
  backdrop-filter: blur(10px);
  position: sticky;
  top: 0;
  z-index: 10;
}

.eyebrow-mini {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--accent);
  letter-spacing: 0.25em;
  text-transform: uppercase;
  margin-bottom: 0.3rem;
}

.dashboard-header h2 {
  font-size: 2rem;
  margin: 0;
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.5rem 1rem;
  background: var(--charcoal);
  border: 1px solid var(--line);
  border-radius: 40px;
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--accent) 0%, var(--accent-dim) 100%);
}

.user-info {
  display: flex;
  flex-direction: column;
}

.user-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text);
}

.user-id {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--text-mute);
}

.dashboard-content {
  padding: 3rem;
  flex: 1;
}
</style>
