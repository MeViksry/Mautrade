<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

definePageMeta({
  layout: 'dashboard'
})

const seoTitle = 'Settings | Mautrade Dashboard'
const seoDescription = 'Update Mautrade profile information, profile photo, timezone, notification preferences, and account security.'

useSeoMeta({
  title: seoTitle,
  description: seoDescription,
  ogTitle: seoTitle,
  ogDescription: seoDescription,
  twitterTitle: seoTitle,
  twitterDescription: seoDescription
})

const activeTab = ref('profile')

const profileForm = ref({
  fullName: 'User Account',
  email: 'user@mautrade.com',
  timezone: 'Asia/Jakarta'
})

const profilePhoto = ref<string | null>(null)
const profilePhotoInput = ref<HTMLInputElement | null>(null)
const timezoneSearch = ref('')
const timezoneDropdownOpen = ref(false)
const timezoneSelectRef = ref<HTMLElement | null>(null)

const notifSettings = ref({
  emailNewTrade: true,
  emailDailyReport: false,
  telegramAlerts: true
})

const openProfilePhotoPicker = () => {
  profilePhotoInput.value?.click()
}

const handleProfilePhotoChange = (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  const reader = new FileReader()
  reader.onload = () => {
    profilePhoto.value = typeof reader.result === 'string' ? reader.result : null
  }
  reader.readAsDataURL(file)
}

const deleteProfilePhoto = () => {
  profilePhoto.value = null

  if (profilePhotoInput.value) {
    profilePhotoInput.value.value = ''
  }
}

const fallbackTimezones = [
  'Africa/Cairo',
  'Africa/Johannesburg',
  'America/Argentina/Buenos_Aires',
  'America/Chicago',
  'America/Los_Angeles',
  'America/Mexico_City',
  'America/New_York',
  'America/Sao_Paulo',
  'Asia/Dubai',
  'Asia/Hong_Kong',
  'Asia/Jakarta',
  'Asia/Kolkata',
  'Asia/Seoul',
  'Asia/Shanghai',
  'Asia/Singapore',
  'Asia/Tokyo',
  'Australia/Sydney',
  'Europe/Amsterdam',
  'Europe/Berlin',
  'Europe/London',
  'Europe/Madrid',
  'Europe/Paris',
  'Europe/Rome',
  'Pacific/Auckland'
]

const browserTimezones = () => {
  const intlWithTimezones = Intl as typeof Intl & {
    supportedValuesOf?: (key: 'timeZone') => string[]
  }

  return intlWithTimezones.supportedValuesOf?.('timeZone') ?? fallbackTimezones
}

const formatTimezoneName = (timezone: string) => {
  return timezone.replaceAll('_', ' ')
}

const formatTimezoneOffset = (timezone: string) => {
  const timeZoneName = new Intl.DateTimeFormat('en-US', {
    timeZone: timezone,
    timeZoneName: 'shortOffset'
  }).formatToParts(new Date()).find(part => part.type === 'timeZoneName')?.value ?? 'GMT'

  if (timeZoneName === 'GMT') return 'UTC+00:00'

  const offset = timeZoneName.replace('GMT', '')
  const sign = offset.startsWith('-') ? '-' : '+'
  const [hours = '0', minutes = '0'] = offset.replace(/[+-]/, '').split(':')

  return `UTC${sign}${hours.padStart(2, '0')}:${minutes.padStart(2, '0')}`
}

const timezoneOptions = computed(() => {
  return browserTimezones().map((timezone) => {
    const offset = formatTimezoneOffset(timezone)
    const label = formatTimezoneName(timezone)

    return {
      value: timezone,
      label,
      offset,
      searchText: `${timezone} ${label} ${offset}`.toLowerCase()
    }
  })
})

const selectedTimezone = computed(() => {
  return timezoneOptions.value.find(timezone => timezone.value === profileForm.value.timezone)
})

const timezoneSearchTerm = computed(() => timezoneSearch.value.trim().toLowerCase())

const filteredTimezones = computed(() => {
  if (!timezoneSearchTerm.value) return timezoneOptions.value

  return timezoneOptions.value.filter(timezone => timezone.searchText.includes(timezoneSearchTerm.value))
})

const openTimezoneDropdown = () => {
  timezoneDropdownOpen.value = true
}

const selectTimezone = (timezone: string) => {
  profileForm.value.timezone = timezone
  timezoneSearch.value = ''
  timezoneDropdownOpen.value = false
}

const loading = ref(true)

const handleSettingsClickOutside = (event: MouseEvent) => {
  if (!timezoneSelectRef.value?.contains(event.target as Node)) {
    timezoneDropdownOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleSettingsClickOutside)

  // Simulate loading delay for skeleton shimmer effect
  setTimeout(() => {
    loading.value = false
  }, 400)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleSettingsClickOutside)
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
      <div class="settings-container skeleton-container">
        <div class="settings-sidebar">
          <div class="skeleton-bone skeleton-sidebar-item" />
          <div class="skeleton-bone skeleton-sidebar-item" />
          <div class="skeleton-bone skeleton-sidebar-item" />
        </div>
        <div class="settings-content skeleton-content">
          <div class="skeleton-bone skeleton-pane-title" />
          <div class="skeleton-bone skeleton-pane-subtitle" />
          <div class="skeleton-form">
            <div class="skeleton-bone skeleton-input" />
            <div class="skeleton-bone skeleton-input" />
            <div class="skeleton-bone skeleton-input" />
          </div>
        </div>
      </div>
    </div>

    <template v-else>
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
              <div class="profile-photo-field">
                <div class="profile-photo-preview">
                  <img
                    v-if="profilePhoto"
                    :src="profilePhoto"
                    alt="Profile photo"
                  >
                  <UIcon
                    v-else
                    name="lucide:user"
                    class="profile-photo-preview__icon"
                  />
                </div>

                <div class="profile-photo-actions">
                  <input
                    ref="profilePhotoInput"
                    class="profile-photo-input"
                    type="file"
                    accept="image/*"
                    @change="handleProfilePhotoChange"
                  >
                  <button
                    class="btn-photo"
                    type="button"
                    @click="openProfilePhotoPicker"
                  >
                    <UIcon name="lucide:image-plus" />
                    <span>Upload Photo</span>
                  </button>
                  <button
                    class="btn-photo btn-photo--danger"
                    type="button"
                    :disabled="!profilePhoto"
                    @click="deleteProfilePhoto"
                  >
                    <UIcon name="lucide:trash-2" />
                    <span>Delete Photo</span>
                  </button>
                </div>
              </div>

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
                <div
                  ref="timezoneSelectRef"
                  class="timezone-select"
                >
                  <button
                    class="timezone-select__trigger"
                    type="button"
                    @click="timezoneDropdownOpen = !timezoneDropdownOpen"
                  >
                    <span>
                      <strong>{{ selectedTimezone?.offset }}</strong>
                      {{ selectedTimezone?.label }}
                    </span>
                    <UIcon name="lucide:chevrons-up-down" />
                  </button>

                  <div
                    v-if="timezoneDropdownOpen"
                    class="timezone-select__dropdown"
                  >
                    <div class="timezone-select__search">
                      <UIcon name="lucide:search" />
                      <input
                        v-model="timezoneSearch"
                        type="text"
                        placeholder="Search timezone"
                        autocomplete="off"
                        @focus="openTimezoneDropdown"
                      >
                    </div>

                    <div class="timezone-select__list">
                      <button
                        v-for="timezone in filteredTimezones"
                        :key="timezone.value"
                        class="timezone-option"
                        :class="{ 'is-selected': timezone.value === profileForm.timezone }"
                        type="button"
                        @click="selectTimezone(timezone.value)"
                      >
                        <span>{{ timezone.label }}</span>
                        <strong>{{ timezone.offset }}</strong>
                      </button>
                      <div
                        v-if="filteredTimezones.length === 0"
                        class="timezone-empty"
                      >
                        No timezone found.
                      </div>
                    </div>
                  </div>
                </div>
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

.profile-photo-field {
  display: flex;
  align-items: center;
  gap: 1.25rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px dashed var(--line);
}

.profile-photo-preview {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  width: 88px;
  height: 88px;
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: 50%;
  background: var(--charcoal);
}

.profile-photo-preview img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.profile-photo-preview__icon {
  width: 34px;
  height: 34px;
  color: var(--text-mute);
}

.profile-photo-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.profile-photo-input {
  display: none;
}

.btn-photo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  min-height: 40px;
  padding: 0 0.85rem;
  border: 1px solid var(--accent);
  background: var(--accent);
  color: #000;
  font-family: var(--mono);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  cursor: pointer;
  transition: background 220ms var(--ease-quiet), border-color 220ms var(--ease-quiet), color 220ms var(--ease-quiet);
}

.btn-photo:hover:not(:disabled) {
  background: #ff7324;
  border-color: #ff7324;
}

.btn-photo--danger {
  background: transparent;
  color: #ef4444;
  border-color: rgba(239, 68, 68, 0.35);
}

.btn-photo--danger:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.08);
  border-color: #ef4444;
  color: #ff6b6b;
}

.btn-photo:disabled {
  cursor: not-allowed;
  opacity: 0.35;
}

.btn-photo svg {
  width: 14px;
  height: 14px;
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

.timezone-select {
  position: relative;
}

.timezone-select__trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  width: 100%;
  min-height: 46px;
  border: 1px solid var(--line);
  background: var(--bg);
  color: var(--text);
  padding: 0.8rem 1rem;
  text-align: left;
  transition: border-color 0.2s ease;
}

.timezone-select__trigger:hover,
.timezone-select__trigger:focus {
  border-color: var(--accent);
}

.timezone-select__trigger span {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  min-width: 0;
  font-size: 14px;
}

.timezone-select__trigger strong {
  flex: 0 0 auto;
  font-family: var(--mono);
  font-size: 11px;
  color: var(--accent);
  font-weight: 700;
}

.timezone-select__trigger svg {
  flex: 0 0 auto;
  width: 16px;
  height: 16px;
  color: var(--text-mute);
}

.timezone-select__dropdown {
  position: absolute;
  z-index: 20;
  top: calc(100% + 0.5rem);
  left: 0;
  right: 0;
  overflow: hidden;
  border: 1px solid var(--line);
  background: var(--bg-elevated);
  box-shadow: 0 20px 48px rgba(0, 0, 0, 0.35);
}

.timezone-select__search {
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr);
  align-items: center;
  gap: 0.65rem;
  padding: 0.75rem;
  border-bottom: 1px solid var(--line);
  background: var(--charcoal);
}

.timezone-select__search svg {
  width: 16px;
  height: 16px;
  color: var(--text-mute);
}

.timezone-select__search input {
  width: 100%;
  min-width: 0;
  border: none;
  outline: none;
  background: transparent;
  color: var(--text);
  font-family: var(--mono);
  font-size: 12px;
}

.timezone-select__list {
  display: flex;
  flex-direction: column;
  max-height: 280px;
  overflow-y: auto;
  scrollbar-color: var(--accent) var(--charcoal);
  scrollbar-width: thin;
}

.timezone-select__list::-webkit-scrollbar {
  width: 10px;
}

.timezone-select__list::-webkit-scrollbar-track {
  background: var(--charcoal);
  border-left: 1px solid var(--line);
}

.timezone-select__list::-webkit-scrollbar-thumb {
  background: var(--accent);
  border: 2px solid var(--charcoal);
  border-radius: 999px;
}

.timezone-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  width: 100%;
  padding: 0.85rem 1rem;
  border: none;
  border-bottom: 1px solid var(--line);
  background: var(--bg-elevated);
  color: var(--text);
  text-align: left;
  transition: background 220ms var(--ease-quiet), color 220ms var(--ease-quiet);
}

.timezone-option:hover,
.timezone-option.is-selected {
  background: rgba(255, 90, 0, 0.1);
  color: var(--accent);
}

.timezone-option span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.timezone-option strong {
  flex: 0 0 auto;
  font-family: var(--mono);
  font-size: 11px;
  color: inherit;
}

.timezone-empty {
  padding: 1rem;
  font-family: var(--mono);
  font-size: 11px;
  color: var(--text-mute);
  text-transform: uppercase;
  letter-spacing: 0.08em;
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

/* ─── Skeleton Loading ─── */
@keyframes shimmer {
  0% { background-position: -400px 0; }
  100% { background-position: 400px 0; }
}

.skeleton-loading {
  animation: skeletonFadeIn 0.4s ease-out;
  width: 100%;
  max-width: 100%;
  overflow: hidden;
  min-width: 0;
}

@keyframes skeletonFadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

.skeleton-bone {
  background: linear-gradient(
    90deg,
    rgba(255, 255, 255, 0.04) 0%,
    rgba(255, 255, 255, 0.08) 20%,
    rgba(255, 138, 76, 0.12) 40%,
    rgba(255, 138, 76, 0.18) 50%,
    rgba(255, 138, 76, 0.12) 60%,
    rgba(255, 255, 255, 0.08) 80%,
    rgba(255, 255, 255, 0.04) 100%
  );
  background-size: 800px 100%;
  animation: shimmer 1.8s ease-in-out infinite;
  border-radius: 4px;
}

.skeleton-page-header {
  margin-bottom: 1rem;
}

.skeleton-title { width: 180px; height: 38px; }

.skeleton-sidebar-item {
  width: 140px;
  height: 20px;
  margin: 1rem 2rem;
  border-radius: 4px;
}

.skeleton-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.skeleton-pane-title { width: 220px; height: 24px; }
.skeleton-pane-subtitle { width: 340px; height: 16px; }

.skeleton-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  margin-top: 1rem;
}

.skeleton-input {
  width: 100%;
  height: 72px;
  border-radius: 4px;
}

@media (max-width: 768px) {
  .settings-container {
    display: flex;
    flex-direction: column;
    gap: 0;
  }
  .settings-sidebar {
    width: 100%;
    flex-direction: row;
    flex-wrap: wrap;
    justify-content: center;
    border-right: none;
    border-bottom: 1px solid var(--line);
    padding: 1rem;
    gap: 0.5rem;
  }
  .settings-tab {
    flex: 1 1 auto;
    justify-content: center;
    padding: 0.75rem 1rem;
    border-left: none;
    border-bottom: 3px solid transparent;
  }
  .settings-tab.active {
    border-left-color: transparent;
    border-bottom-color: var(--accent);
  }
  .skeleton-sidebar-item {
    margin: 1rem 1.5rem;
  }
}

@media (max-width: 640px) {
  .dashboard-page { gap: 0.75rem; }
  .page-header { margin-bottom: 0.5rem; }
  .page-title { font-size: 1.3rem; }

  .skeleton-page-header { margin-bottom: 0.5rem; }
  .skeleton-title { width: 150px; height: 22px; }
  .skeleton-pane-subtitle { width: 240px; }

  .settings-content {
    padding: 1.25rem;
  }
  .profile-photo-field {
    flex-direction: column;
    align-items: flex-start;
    gap: 1.5rem;
  }
  .profile-photo-actions {
    flex-direction: column;
    width: 100%;
  }
  .btn-photo {
    width: 100%;
    justify-content: center;
  }
  .two-fa-section {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  .btn-primary { width: 100%; }
}

@media (max-width: 380px) {
  .dashboard-page { gap: 0.5rem; }
  .page-title { font-size: 1.15rem; }
  .settings-content { padding: 1rem; }
}
</style>
