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
  height: 100vh;
  width: 100vw;
  background: var(--bg);
  overflow: hidden;
}

.dashboard-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  height: 100vh;
}

.dashboard-header-fixed {
  flex-shrink: 0;
}

.dashboard-content {
  padding: 3rem;
  flex: 1 0 auto;
}

@media (max-width: 768px) {
  .dashboard-content {
    padding: 1.5rem;
  }
}
</style>
