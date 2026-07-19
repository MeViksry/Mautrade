<script setup lang="ts">
import { ref, onMounted } from 'vue'

definePageMeta({
  layout: 'admin'
})

const seoTitle = 'Admins Management | Admin Mautrade'
const seoDescription = 'Manage administrators and roles.'
useSeoMeta({
  title: seoTitle,
  description: seoDescription
})

const loading = ref(true)

const admins = ref([
  { id: 'ADM-01', name: 'Super Admin', email: 'super@mautrade.com', role: 'Superadmin', status: 'Active' },
  { id: 'ADM-02', name: 'Finance Manager', email: 'finance@mautrade.com', role: 'Finance', status: 'Active' },
  { id: 'ADM-03', name: 'Support Lead', email: 'support@mautrade.com', role: 'Support', status: 'Inactive' }
])

onMounted(() => {
  setTimeout(() => {
    loading.value = false
  }, 1000)
})
</script>

<template>
  <div class="dashboard-page">
    <div
      v-if="loading"
      class="skeleton-loading"
    >
      <div class="skeleton-page-header">
        <div class="skeleton-bone skeleton-title" />
      </div>

      <div class="skeleton-stats-grid">
        <div
          v-for="n in 1"
          :key="`stat-${n}`"
          class="skeleton-stat-card"
        >
          <div class="skeleton-bone skeleton-stat-label" />
          <div class="skeleton-bone skeleton-stat-value" />
        </div>
      </div>
    </div>

    <template v-else>
      <header class="page-header">
        <h1 class="page-title">
          Admins Management
        </h1>
        <p class="page-subtitle">
          Manage system administrators and their roles
        </p>
      </header>

      <div class="admins-section">
        <div class="section-header-controls">
          <h2 class="section-title">
            Administrators List
          </h2>
          <button class="action-btn add-btn">
            <UIcon name="lucide:user-plus" />
            Add New Admin
          </button>
        </div>

        <div class="table-container">
          <table class="data-table">
            <thead>
              <tr>
                <th>Admin ID</th>
                <th>Name / Email</th>
                <th>Role</th>
                <th>Status</th>
                <th class="actions-col">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="admin in admins"
                :key="admin.id"
              >
                <td class="col-id">
                  {{ admin.id }}
                </td>
                <td class="col-user">
                  <div class="user-info">
                    <span class="u-name">{{ admin.name }}</span>
                    <span class="u-email">{{ admin.email }}</span>
                  </div>
                </td>
                <td class="col-role">
                  <span class="role-badge">{{ admin.role }}</span>
                </td>
                <td class="col-status">
                  <span :class="['status-badge', admin.status.toLowerCase()]">
                    {{ admin.status }}
                  </span>
                </td>
                <td class="col-actions">
                  <div class="action-buttons">
                    <button class="action-btn edit-btn">
                      Edit Role
                    </button>
                    <button class="action-btn delete-btn">
                      Revoke
                    </button>
                  </div>
                </td>
              </tr>
              <tr v-if="admins.length === 0">
                <td
                  colspan="5"
                  class="empty-state"
                >
                  No administrators found.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.dashboard-page {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.page-header {
  margin-bottom: 1rem;
}

.page-title {
  font-family: 'Oswald', sans-serif;
  font-size: 1.5rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  color: var(--text);
  margin-bottom: 0.5rem;
}

.page-subtitle {
  color: var(--text-mute);
  font-size: 0.95rem;
}

.admins-section {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.section-header-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.section-title {
  font-family: 'Oswald', sans-serif;
  font-size: 1.25rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  color: var(--text);
}

.add-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--accent) !important;
  color: #fff !important;
  border: none !important;
}

.add-btn:hover {
  background: #ff7a33 !important;
}

.table-container {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.data-table th,
.data-table td {
  padding: 1rem;
  border-bottom: 1px solid var(--line);
}

.data-table th {
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--text-mute);
  font-weight: normal;
}

.data-table tbody tr:hover {
  background: rgba(255, 255, 255, 0.02);
}

.col-id {
  font-family: var(--mono);
  font-size: 0.9rem;
  color: var(--silver);
}

.user-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.u-name {
  font-weight: 500;
  color: var(--text);
}

.u-email {
  font-size: 0.85rem;
  color: var(--text-mute);
}

.role-badge {
  background: var(--charcoal);
  padding: 0.3rem 0.6rem;
  border-radius: 4px;
  font-family: var(--mono);
  font-size: 0.8rem;
  color: var(--accent);
  border: 1px solid rgba(255, 90, 0, 0.2);
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.6rem;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.status-badge.active {
  background: rgba(34, 197, 94, 0.1);
  color: #4ade80;
}

.status-badge.inactive {
  background: rgba(156, 163, 175, 0.1);
  color: #9ca3af;
}

.actions-col, .col-actions {
  text-align: right;
}

.action-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.action-btn {
  background: transparent;
  border: 1px solid var(--line);
  color: var(--silver);
  padding: 0.4rem 0.8rem;
  border-radius: 4px;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 500;
}

.edit-btn:hover {
  border-color: #38bdf8;
  color: #38bdf8;
}

.delete-btn:hover {
  border-color: #ef4444;
  color: #ef4444;
}

.empty-state {
  text-align: center;
  color: var(--text-mute);
  padding: 3rem !important;
  font-style: italic;
}

/* Skeleton Loading */
.skeleton-loading {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.skeleton-bone {
  background: linear-gradient(90deg, var(--charcoal) 25%, var(--line) 50%, var(--charcoal) 75%);
  background-size: 200% 100%;
  animation: loading 1.5s infinite;
  border-radius: 4px;
}

@keyframes loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.skeleton-page-header {
  margin-bottom: 1rem;
}

.skeleton-title {
  width: 200px;
  height: 24px;
}

.skeleton-stats-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.5rem;
}

.skeleton-stat-card {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  height: 300px;
}

@media (max-width: 640px) {
  .section-header-controls {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
}
</style>
