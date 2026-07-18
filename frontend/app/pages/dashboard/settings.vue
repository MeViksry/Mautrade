<script setup lang="ts">
import { ref } from 'vue'

definePageMeta({
  layout: 'dashboard'
})

const activeTab = ref('profile')

const profileForm = ref({
  fullName: 'User Account',
  email: 'user@mautrade.com',
  timezone: 'UTC+07:00 (Jakarta)'
})

const notifSettings = ref({
  emailNewTrade: true,
  emailDailyReport: false,
  telegramAlerts: true
})
</script>

<template>
  <div class="dashboard-page">
    <div class="page-header">
      <h2 class="page-title">
        Settings
      </h2>
    </div>

    <div class="settings-container">
      <div class="settings-sidebar">
        <button
          class="settings-tab"
          :class="{ active: activeTab === 'profile' }"
          @click="activeTab = 'profile'"
        >
          <UIcon
            name="lucide:user"
            class="tab-icon"
          />
          Profile Details
        </button>
        <button
          class="settings-tab"
          :class="{ active: activeTab === 'notifications' }"
          @click="activeTab = 'notifications'"
        >
          <UIcon
            name="lucide:bell"
            class="tab-icon"
          />
          Notifications
        </button>
        <button
          class="settings-tab"
          :class="{ active: activeTab === 'security' }"
          @click="activeTab = 'security'"
        >
          <UIcon
            name="lucide:shield"
            class="tab-icon"
          />
          Security
        </button>
      </div>

      <div class="settings-content">
        <!-- Profile Tab -->
        <div
          v-if="activeTab === 'profile'"
          class="tab-pane"
        >
          <h3 class="pane-title">
            Profile Information
          </h3>
          <p class="pane-subtitle">
            Update your account's profile information and timezone.
          </p>

          <form
            class="settings-form"
            @submit.prevent
          >
            <div class="form-group">
              <label>Full Name</label>
              <input
                v-model="profileForm.fullName"
                type="text"
                class="form-input"
              >
            </div>

            <div class="form-group">
              <label>Email Address</label>
              <input
                v-model="profileForm.email"
                type="email"
                class="form-input"
                disabled
              >
              <span class="input-help">Email cannot be changed directly. Please contact support.</span>
            </div>

            <div class="form-group">
              <label>Timezone</label>
              <select
                v-model="profileForm.timezone"
                class="form-input"
              >
                <option value="UTC+07:00 (Jakarta)">
                  UTC+07:00 (Jakarta)
                </option>
                <option value="UTC+00:00 (London)">
                  UTC+00:00 (London)
                </option>
                <option value="UTC-05:00 (New York)">
                  UTC-05:00 (New York)
                </option>
              </select>
            </div>

            <div class="form-actions">
              <button class="btn-primary">
                Save Changes
              </button>
            </div>
          </form>
        </div>

        <!-- Notifications Tab -->
        <div
          v-if="activeTab === 'notifications'"
          class="tab-pane"
        >
          <h3 class="pane-title">
            Notification Preferences
          </h3>
          <p class="pane-subtitle">
            Choose what updates you want to receive from Mautrade.
          </p>

          <div class="settings-form">
            <div class="toggle-group">
              <div class="toggle-info">
                <h4>New Trade Executed</h4>
                <p>Get notified when a master signal triggers a trade.</p>
              </div>
              <label class="toggle-switch">
                <input
                  v-model="notifSettings.emailNewTrade"
                  type="checkbox"
                >
                <span class="slider" />
              </label>
            </div>

            <div class="toggle-group">
              <div class="toggle-info">
                <h4>Daily PNL Report</h4>
                <p>Receive a daily summary of your realized and unrealized profit.</p>
              </div>
              <label class="toggle-switch">
                <input
                  v-model="notifSettings.emailDailyReport"
                  type="checkbox"
                >
                <span class="slider" />
              </label>
            </div>

            <div class="toggle-group">
              <div class="toggle-info">
                <h4>Telegram Alerts</h4>
                <p>Instant push notifications via our Telegram bot.</p>
              </div>
              <label class="toggle-switch">
                <input
                  v-model="notifSettings.telegramAlerts"
                  type="checkbox"
                >
                <span class="slider" />
              </label>
            </div>
          </div>
        </div>

        <!-- Security Tab -->
        <div
          v-if="activeTab === 'security'"
          class="tab-pane"
        >
          <h3 class="pane-title">
            Security Settings
          </h3>
          <p class="pane-subtitle">
            Manage your password and 2-factor authentication.
          </p>

          <form
            class="settings-form"
            @submit.prevent
          >
            <div class="form-group">
              <label>Current Password</label>
              <input
                type="password"
                class="form-input"
                placeholder="Enter current password"
              >
            </div>

            <div class="form-group">
              <label>New Password</label>
              <input
                type="password"
                class="form-input"
                placeholder="Enter new password"
              >
            </div>

            <div class="form-actions">
              <button class="btn-primary">
                Update Password
              </button>
            </div>

            <hr class="divider">

            <div class="two-fa-section">
              <div class="two-fa-info">
                <h4>Two-Factor Authentication (2FA)</h4>
                <p>Secure your account with an authenticator app.</p>
              </div>
              <button class="btn-secondary">
                Enable 2FA
              </button>
            </div>
          </form>
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
  font-size: 2.5rem;
  font-weight: 300;
  text-transform: uppercase;
  color: var(--text);
  margin: 0;
  letter-spacing: 0.05em;
}

.settings-container {
  display: grid;
  grid-template-columns: 280px 1fr;
  background: var(--bg-elevated);
  border: 1px solid var(--line);
  border-radius: 4px;
  min-height: 600px;
}

.settings-sidebar {
  border-right: 1px solid var(--line);
  padding: 2rem 0;
  display: flex;
  flex-direction: column;
}

.settings-tab {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 2rem;
  background: transparent;
  border: none;
  color: var(--text-mute);
  font-family: var(--mono);
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  cursor: pointer;
  text-align: left;
  border-left: 3px solid transparent;
  transition: all 0.2s ease;
}

.settings-tab:hover {
  background: rgba(255, 255, 255, 0.02);
  color: var(--text);
}

.settings-tab.active {
  background: var(--charcoal);
  color: var(--accent);
  border-left-color: var(--accent);
}

.tab-icon {
  width: 18px;
  height: 18px;
  opacity: 0.8;
}

.settings-content {
  padding: 3rem;
}

.pane-title {
  font-family: 'Oswald', sans-serif;
  font-size: 1.5rem;
  color: var(--text);
  margin-bottom: 0.5rem;
  letter-spacing: 0.05em;
}

.pane-subtitle {
  color: var(--text-mute);
  font-size: 14px;
  margin-bottom: 3rem;
}

.settings-form {
  max-width: 500px;
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  font-family: var(--mono);
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-mute);
}

.form-input {
  background: var(--bg);
  border: 1px solid var(--line);
  padding: 0.8rem 1rem;
  color: var(--text);
  font-family: var(--sans);
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s ease;
}

.form-input:focus {
  border-color: var(--accent);
}

.form-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.input-help {
  font-size: 12px;
  color: var(--text-mute);
}

.form-actions {
  margin-top: 1rem;
}

.btn-primary {
  background: var(--accent);
  color: #000;
  border: none;
  padding: 0.75rem 2rem;
  font-family: 'Oswald', sans-serif;
  font-size: 14px;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-primary:hover {
  background: #ff7324;
}

.btn-secondary {
  background: transparent;
  color: var(--text);
  border: 1px solid var(--line);
  padding: 0.75rem 1.5rem;
  font-family: 'Oswald', sans-serif;
  font-size: 14px;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-secondary:hover {
  border-color: var(--text);
}

/* Toggle Switch Styles */
.toggle-group {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding-bottom: 1.5rem;
  border-bottom: 1px dashed var(--line);
}

.toggle-info h4 {
  font-family: var(--sans);
  font-size: 15px;
  font-weight: 500;
  color: var(--text);
  margin: 0 0 0.25rem 0;
}

.toggle-info p {
  margin: 0;
  font-size: 13px;
  color: var(--text-mute);
}

.toggle-switch {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
}

.toggle-switch input {
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
  background-color: var(--bg);
  border: 1px solid var(--line);
  transition: .3s;
  border-radius: 34px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 3px;
  bottom: 3px;
  background-color: var(--text-mute);
  transition: .3s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: rgba(255, 90, 0, 0.2);
  border-color: var(--accent);
}

input:checked + .slider:before {
  background-color: var(--accent);
  transform: translateX(20px);
}

.divider {
  border: 0;
  border-top: 1px solid var(--line);
  margin: 3rem 0;
}

.two-fa-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.two-fa-info h4 {
  font-family: var(--sans);
  font-size: 15px;
  font-weight: 500;
  color: var(--text);
  margin: 0 0 0.25rem 0;
}

.two-fa-info p {
  margin: 0;
  font-size: 13px;
  color: var(--text-mute);
}
</style>
