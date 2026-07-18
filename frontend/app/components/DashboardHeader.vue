<script setup lang="ts">
// Dashboard top header component
</script>

<template>
  <header class="dashboard-header">
    <div class="header-right">
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
  justify-content: flex-end;
  align-items: center;
  border-bottom: 1px solid var(--line);
  background: rgba(10, 10, 10, 0.8);
  backdrop-filter: blur(10px);
  position: sticky;
  top: 0;
  z-index: 10;
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
  background: linear-gradient(135deg, var(--accent) 0%, #ff8a4c 100%);
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
</style>
