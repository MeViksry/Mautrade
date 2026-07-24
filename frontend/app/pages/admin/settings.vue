<script setup lang="ts">
import { ref, onMounted } from 'vue'

definePageMeta({
  layout: 'admin'
})

const seoTitle = 'Global Settings | Admin Mautrade'
const seoDescription = 'Configure global platform settings.'
useSeoMeta({
  title: seoTitle,
  description: seoDescription
})

const loading = ref(true)
const config = useRuntimeConfig()
const apiBase = config.public.apiBase

const settings = ref({
  maintenanceMode: false,
  allowRegistrations: true,
  gasFeePercentage: 20,
  minDepositUsdt: 500,
  maxActiveLayersPerUser: 10,
  supportEmail: 'support@mautrade.com'
})

const fetchSettings = async () => {
  try {
    const data = await $fetch<any>(`${apiBase}/settings`)
    if (data) {
      settings.value = {
        maintenanceMode: data.maintenanceMode,
        allowRegistrations: data.allowRegistrations,
        gasFeePercentage: parseFloat(data.gasFeePercentage || 20),
        minDepositUsdt: parseFloat(data.minDepositUsdt || 500),
        maxActiveLayersPerUser: data.maxActiveLayersPerUser,
        supportEmail: data.supportEmail
      }
    }
  } catch (err) {
    console.error('Failed to fetch settings', err)
  } finally {
    loading.value = false
  }
}

const handleSave = async () => {
  try {
    const adminToken = useCookie('admin_token')
    const res = await $fetch<any>(`${apiBase}/admin/settings`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${adminToken.value}`
      },
      body: {
        maintenanceMode: settings.value.maintenanceMode,
        allowRegistrations: settings.value.allowRegistrations,
        gasFeePercentage: settings.value.gasFeePercentage.toString(),
        minDepositUsdt: settings.value.minDepositUsdt.toString(),
        maxActiveLayersPerUser: settings.value.maxActiveLayersPerUser,
        supportEmail: settings.value.supportEmail
      }
    })
    alert('Settings saved successfully!')
  } catch (err) {
    console.error('Failed to save settings', err)
    alert('Failed to save settings. Check console for details.')
  }
}

onMounted(() => {
  fetchSettings()
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

      <div class="skeleton-settings-form">
        <div
          v-for="i in 5"
          :key="i"
          class="skeleton-bone skeleton-row"
        />
      </div>
    </div>

    <template v-else>
      <header class="page-header">
        <h1 class="page-title">
          Global Settings
        </h1>
        <p class="page-subtitle">
          Configure core platform parameters and fees
        </p>
      </header>

      <div class="settings-grid">
        <!-- General System Settings -->
        <div class="settings-panel">
          <h2 class="panel-title">
            System State
          </h2>
          <div class="settings-form">
            <div class="form-group toggle-group">
              <div class="toggle-info">
                <label>Maintenance Mode</label>
                <span class="help-text">Disable all user access except for admins.</span>
              </div>
              <label class="switch">
                <input
                  v-model="settings.maintenanceMode"
                  type="checkbox"
                >
                <span class="slider round" />
              </label>
            </div>

            <div class="form-group toggle-group">
              <div class="toggle-info">
                <label>Allow Registrations</label>
                <span class="help-text">Allow new users to create accounts.</span>
              </div>
              <label class="switch">
                <input
                  v-model="settings.allowRegistrations"
                  type="checkbox"
                >
                <span class="slider round" />
              </label>
            </div>

            <div class="form-group">
              <label>Support Email Address</label>
              <input
                v-model="settings.supportEmail"
                type="email"
                class="form-input"
              >
            </div>
          </div>
        </div>

        <!-- Trading & Financial Parameters -->
        <div class="settings-panel">
          <h2 class="panel-title">
            Trading Parameters
          </h2>
          <div class="settings-form">
            <div class="form-group">
              <label>Profit Share / Gas Fee (%)</label>
              <div class="input-with-addon">
                <input
                  v-model="settings.gasFeePercentage"
                  type="number"
                  class="form-input"
                >
                <span class="addon">%</span>
              </div>
              <span class="help-text">Percentage taken from user profits as gas fee.</span>
            </div>

            <div class="form-group">
              <label>Minimum Gas Fee Deposit (USDT)</label>
              <div class="input-with-addon">
                <span class="addon">$</span>
                <input
                  v-model="settings.minDepositUsdt"
                  type="number"
                  class="form-input"
                >
              </div>
            </div>

            <div class="form-group">
              <label>Max Active Layers (Per User)</label>
              <input
                v-model="settings.maxActiveLayersPerUser"
                type="number"
                class="form-input"
              >
            </div>
          </div>
        </div>
      </div>

      <div class="page-actions">
        <button
          class="action-btn save-btn"
          @click="handleSave"
        >
          Save All Settings
        </button>
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

.settings-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
}

.settings-panel {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.panel-title {
  font-family: 'Oswald', sans-serif;
  font-size: 1.2rem;
  font-weight: 500;
  letter-spacing: 0.05em;
  color: var(--text);
  border-bottom: 1px solid var(--line);
  padding-bottom: 0.75rem;
}

.settings-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.toggle-group {
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
}

.toggle-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.form-group label {
  font-size: 0.9rem;
  color: var(--text);
  font-weight: 500;
}

.help-text {
  font-size: 0.8rem;
  color: var(--text-mute);
}

.form-input {
  width: 100%;
  background: var(--charcoal);
  border: 1px solid var(--line);
  color: var(--text);
  padding: 0.75rem 1rem;
  border-radius: 6px;
  font-size: 0.95rem;
  font-family: var(--mono);
  outline: none;
  transition: border-color 0.2s;
}

.form-input:focus {
  border-color: var(--accent);
}

.input-with-addon {
  display: flex;
  align-items: stretch;
}

.input-with-addon .form-input {
  border-radius: 0;
}

.input-with-addon .form-input:first-child {
  border-top-left-radius: 6px;
  border-bottom-left-radius: 6px;
}

.input-with-addon .form-input:last-child {
  border-top-right-radius: 6px;
  border-bottom-right-radius: 6px;
}

.addon {
  background: var(--line);
  color: var(--text-mute);
  padding: 0 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: var(--mono);
  font-size: 0.9rem;
  border: 1px solid var(--line);
}

.input-with-addon .addon:first-child {
  border-top-left-radius: 6px;
  border-bottom-left-radius: 6px;
  border-right: none;
}

.input-with-addon .addon:last-child {
  border-top-right-radius: 6px;
  border-bottom-right-radius: 6px;
  border-left: none;
}

/* Switch styling */
.switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 24px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--charcoal);
  transition: .4s;
  border: 1px solid var(--line);
}

.slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 3px;
  bottom: 3px;
  background-color: var(--silver);
  transition: .4s;
}

input:checked + .slider {
  background-color: var(--accent);
  border-color: var(--accent);
}

input:focus + .slider {
  box-shadow: 0 0 1px var(--accent);
}

input:checked + .slider:before {
  transform: translateX(26px);
  background-color: #fff;
}

.slider.round {
  border-radius: 24px;
}

.slider.round:before {
  border-radius: 50%;
}

.page-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
}

.save-btn {
  background: var(--accent);
  color: #fff;
  border: none;
  padding: 0.75rem 2rem;
  border-radius: 6px;
  font-family: 'Oswald', sans-serif;
  font-size: 1.1rem;
  letter-spacing: 0.05em;
  cursor: pointer;
  transition: opacity 0.2s;
}

.save-btn:hover {
  opacity: 0.9;
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

.skeleton-settings-form {
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.skeleton-row {
  width: 100%;
  height: 48px;
}

@media (max-width: 900px) {
  .settings-grid {
    grid-template-columns: 1fr;
  }
}
</style>
