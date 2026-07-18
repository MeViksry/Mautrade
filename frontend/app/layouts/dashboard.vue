<script setup lang="ts">
import DashboardSidebar from '~/components/DashboardSidebar.vue'
import DashboardHeader from '~/components/DashboardHeader.vue'
</script>

<template>
  <div class="dashboard-layout">
    <DashboardSidebar />

    <main class="dashboard-main">
      <DashboardHeader />

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
}

.dashboard-content {
  padding: 3rem;
  flex: 1;
}
</style>
