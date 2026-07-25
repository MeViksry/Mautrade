<script setup lang="ts">
import { ref, onMounted } from 'vue'
import StatCard from '~/components/StatCard.vue'

definePageMeta({
  layout: 'admin'
})

const seoTitle = 'Users Management | Admin Mautrade'
const seoDescription = 'Manage all registered users on Mautrade platform.'
useSeoMeta({
  title: seoTitle,
  description: seoDescription
})

const loading = ref(true)

const { tokenCookie } = useAdminAuth()

const usersStats = ref({
  totalUsers: 0,
  verified: 0,
  gasfeeDepleted: 0,
  pendingGasFee: 0
})

interface AdminUserResponse {
  id: string
  displayName?: string
  email: string
  createdAt: string
  status: string
  onboardingCompleted: boolean
  countryCode?: string
  age?: number
  emailVerified: boolean
  gasFeeBalance: string | number
}

interface FormattedUser {
  id: string
  name?: string
  email: string
  registered: string
  status: string
  gasFee: string
  bot: string
  countryCode: string
  age: string | number
  emailVerified: boolean
  gasFeeBalance: string
}

const users = ref<FormattedUser[]>([])
const selectedUser = ref<FormattedUser | null>(null)
const isModalOpen = ref(false)

const openModal = (user: FormattedUser) => {
  selectedUser.value = user
  isModalOpen.value = true
}

const closeModal = () => {
  isModalOpen.value = false
  selectedUser.value = null
}

const searchQuery = ref('')
let searchTimeout: ReturnType<typeof setTimeout>

const fetchUsers = async () => {
  loading.value = true
  try {
    const config = useRuntimeConfig()
    const apiBase = config.public.apiBase
    const params = new URLSearchParams()
    if (searchQuery.value) {
      params.append('search', searchQuery.value)
    }
    const data = await $fetch<AdminUserResponse[]>(`${apiBase}/admin/users?${params.toString()}`, {
      headers: { Authorization: `Bearer ${tokenCookie.value}` }
    })

    users.value = data.map(u => ({
      id: u.id,
      name: u.displayName,
      email: u.email,
      registered: new Date(u.createdAt).toISOString().split('T')[0] ?? '',
      status: u.status === 'active' ? 'Verified' : 'Pending',
      gasFee: 'Paid', // TODO: backend should return this
      bot: u.onboardingCompleted ? 'Active' : 'Inactive',
      countryCode: u.countryCode || 'N/A',
      age: u.age || 'N/A',
      emailVerified: u.emailVerified,
      gasFeeBalance: typeof u.gasFeeBalance === 'string' ? u.gasFeeBalance : (u.gasFeeBalance?.toString() || '0')
    }))

    // Only update stats on initial load (when search is empty) to avoid stats changing based on search
    if (!searchQuery.value) {
      usersStats.value.totalUsers = users.value.length
      usersStats.value.verified = users.value.filter(u => u.status === 'Verified').length
      usersStats.value.gasfeeDepleted = users.value.filter(u => u.gasFee === 'Depleted').length
    }
  } catch (error) {
    console.error('Failed to load users:', error)
  } finally {
    loading.value = false
  }
}

watch(searchQuery, () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    fetchUsers()
  }, 300)
})

const deleteUser = async (id: string) => {
  if (!confirm('Are you sure you want to delete this user? This action cannot be undone.')) return

  try {
    const config = useRuntimeConfig()
    const apiBase = config.public.apiBase
    await $fetch(`${apiBase}/admin/users/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${tokenCookie.value}` }
    })

    closeModal()
    fetchUsers()
  } catch (error) {
    console.error('Failed to delete user:', error)
    alert('Failed to delete user. Check console for details.')
  }
}

onMounted(() => {
  fetchUsers()
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
          v-for="n in 4"
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
          Users Management
        </h1>
        <p class="page-subtitle">
          Manage all registered users
        </p>
      </header>

      <div class="stats-grid">
        <StatCard
          title="Total Users"
          :value="usersStats.totalUsers.toLocaleString()"
        />
        <StatCard
          title="Verified Users"
          :value="usersStats.verified.toLocaleString()"
        />
        <StatCard
          title="Gasfee Depleted"
          :value="usersStats.gasfeeDepleted.toLocaleString()"
        />
        <StatCard
          title="Pending Gas Fee"
          :value="usersStats.pendingGasFee.toLocaleString()"
        />
      </div>

      <div class="users-section">
        <div class="users-header-controls">
          <h2 class="section-title">
            All Users
          </h2>
          <div class="search-bar">
            <input
              id="user-search"
              v-model="searchQuery"
              name="user-search"
              type="text"
              placeholder="Search by name, email, or ID..."
              class="search-input"
              autocomplete="off"
            >
          </div>
        </div>

        <div class="table-container">
          <table class="data-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>User Details</th>
                <th>Registered</th>
                <th>Verification</th>
                <th>Gas Fee Status</th>
                <th>Bot Status</th>
                <th class="actions-col">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="user in users"
                :key="user.id"
              >
                <td class="col-id">
                  #{{ user.id.split('-')[0] }}
                </td>
                <td class="col-user">
                  <div class="user-info">
                    <span class="u-name">{{ user.name }}</span>
                    <span class="u-email">{{ user.email }}</span>
                  </div>
                </td>
                <td class="col-date">
                  {{ user.registered }}
                </td>
                <td class="col-status">
                  <span :class="['status-badge', user.status.toLowerCase()]">
                    {{ user.status }}
                  </span>
                </td>
                <td class="col-status">
                  <span :class="['status-badge', user.gasFee.toLowerCase() === 'paid' ? 'success' : 'warning']">
                    {{ user.gasFee }}
                  </span>
                </td>
                <td class="col-status">
                  <span :class="['status-badge', user.bot.toLowerCase()]">
                    {{ user.bot }}
                  </span>
                </td>
                <td class="col-actions">
                  <button
                    class="action-btn view-btn"
                    @click="openModal(user)"
                  >
                    View Details
                  </button>
                </td>
              </tr>
              <tr v-if="users.length === 0">
                <td
                  colspan="7"
                  class="empty-state"
                >
                  No users found.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <!-- Details Modal -->
    <div
      v-if="isModalOpen"
      class="modal-overlay"
      @click.self="closeModal"
    >
      <div class="modal-content">
        <div class="modal-header">
          <h3 class="modal-title">
            User Details
          </h3>
          <button
            class="close-btn"
            @click="closeModal"
          >
            &times;
          </button>
        </div>
        <div
          v-if="selectedUser"
          class="modal-body"
        >
          <div class="detail-group">
            <span class="detail-label">ID</span>
            <span class="detail-value">#{{ selectedUser.id.split('-')[0] }}</span>
          </div>
          <div class="detail-group">
            <span class="detail-label">Name</span>
            <span class="detail-value">{{ selectedUser.name || 'N/A' }}</span>
          </div>
          <div class="detail-group">
            <span class="detail-label">Email</span>
            <span class="detail-value">
              {{ selectedUser.email }}
            </span>
          </div>
          <div class="detail-group">
            <span class="detail-label">Country</span>
            <span class="detail-value">{{ selectedUser.countryCode }}</span>
          </div>
          <div class="detail-group">
            <span class="detail-label">Age</span>
            <span class="detail-value">{{ selectedUser.age }}</span>
          </div>
          <div class="detail-group">
            <span class="detail-label">Registered Date</span>
            <span class="detail-value">{{ selectedUser.registered }}</span>
          </div>
          <div class="detail-group">
            <span class="detail-label">Account Status</span>
            <span class="detail-value">{{ selectedUser.status }}</span>
          </div>
          <div class="detail-group">
            <span class="detail-label">Gas Fee Balance</span>
            <span class="detail-value text-accent">${{ selectedUser.gasFeeBalance }}</span>
          </div>

          <div class="modal-actions">
            <button
              class="delete-btn"
              @click="deleteUser(selectedUser.id)"
            >
              <UIcon name="lucide:trash-2" />
              Delete User
            </button>
          </div>
        </div>
      </div>
    </div>
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

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.5rem;
}

.users-section {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.users-header-controls {
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

.search-input {
  background: var(--charcoal);
  border: 1px solid var(--line);
  color: var(--text);
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-size: 0.9rem;
  width: 250px;
  transition: border-color 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: var(--accent);
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

.col-date {
  font-family: var(--mono);
  font-size: 0.9rem;
  color: var(--silver);
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

.status-badge.verified,
.status-badge.success,
.status-badge.active {
  background: rgba(34, 197, 94, 0.1);
  color: #4ade80;
}

.status-badge.pending,
.status-badge.warning {
  background: rgba(245, 158, 11, 0.1);
  color: #fbbf24;
}

.status-badge.inactive {
  background: rgba(156, 163, 175, 0.1);
  color: #9ca3af;
}

.actions-col, .col-actions {
  text-align: right;
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
}

.action-btn:hover {
  border-color: var(--accent);
  color: var(--accent);
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
  grid-template-columns: repeat(4, 1fr);
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
}

.skeleton-stat-label {
  width: 60%;
  height: 14px;
}

.skeleton-stat-value {
  width: 80%;
  height: 28px;
}

@media (max-width: 1180px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  .users-header-controls {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  .search-input {
    width: 100%;
  }
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  box-shadow: 0 10px 25px rgba(0,0,0,0.5);
  animation: modalIn 0.2s ease-out;
}

@keyframes modalIn {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.modal-header {
  padding: 1.5rem;
  border-bottom: 1px solid var(--line);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-title {
  font-family: 'Oswald', sans-serif;
  font-size: 1.25rem;
  color: var(--text);
  margin: 0;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--silver);
  font-size: 1.5rem;
  cursor: pointer;
  line-height: 1;
  padding: 0;
}

.close-btn:hover {
  color: var(--accent);
}

.modal-body {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.detail-group {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}

.detail-group:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.detail-label {
  color: var(--text-mute);
  font-size: 0.9rem;
}

.detail-value {
  color: var(--text);
  font-weight: 500;
  font-size: 0.95rem;
}

.text-success { color: #4ade80; }
.text-warning { color: #fbbf24; }
.text-accent { color: var(--accent); font-weight: 600; font-size: 1.1rem; }

.modal-actions {
  margin-top: 1rem;
  display: flex;
  justify-content: flex-end;
}

.delete-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.2);
  padding: 0.6rem 1rem;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.delete-btn:hover {
  background: #ef4444;
  color: white;
}
</style>
