<script setup lang="ts">
import DashboardSidebar from '~/components/DashboardSidebar.vue'
import DashboardHeader from '~/components/DashboardHeader.vue'
</script>

<template>
  <div class="dashboard-layout">
    <DashboardSidebar />

    <main class="dashboard-main">
      <DashboardHeader class="dashboard-header-fixed" />

      <div class="dashboard-content">
        <slot />
      </div>
    </main>
  </div>
</template>

<style scoped>
.dashboard-layout {
  display: flex;
  height: 100dvh;
  width: 100%;
  background: var(--bg);
  overflow: hidden;
}

.dashboard-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  overflow-x: hidden;
  height: 100dvh;
}

.dashboard-header-fixed {
  flex-shrink: 0;
}

.dashboard-content {
  padding: 3rem;
  flex: 1 0 auto;
}

@media (max-width: 1180px) {
  .dashboard-main {
    width: 100%;
  }

  .dashboard-content {
    padding: 2rem;
  }
}

@media (pointer: coarse) and (max-width: 1366px) {
  .dashboard-main {
    width: 100%;
  }

  .dashboard-content {
    padding: 2rem;
  }
}

@media (max-width: 640px) {
  .dashboard-content {
    padding: 0.65rem;
  }
}

@media (max-width: 380px) {
  .dashboard-content {
    padding: 0.5rem;
  }
}
</style>
