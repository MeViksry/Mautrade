<script setup lang="ts">
defineProps<{
  title: string
  value: string | number
  unit?: string
  trend?: number // positive or negative percentage
}>()
</script>

<template>
  <div class="stat-card">
    <div class="stat-card__label">
      {{ title }}
    </div>
    <div class="stat-card__value-wrapper">
      <div class="stat-card__value">
        {{ value }}<span
          v-if="unit"
          class="stat-card__unit"
        >{{ unit }}</span>
      </div>
      <div
        v-if="trend !== undefined"
        class="stat-card__trend"
        :class="trend >= 0 ? 'text-green-500' : 'text-red-500'"
      >
        <UIcon
          :name="trend >= 0 ? 'lucide:trending-up' : 'lucide:trending-down'"
          class="w-3 h-3 mr-1"
        />
        {{ Math.abs(trend) }}%
      </div>
    </div>
  </div>
</template>

<style scoped>
.stat-card {
  background: var(--charcoal);
  border: 1px solid var(--line);
  padding: 2rem;
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
  transition: border-color 300ms var(--ease-quiet);
}
.stat-card:hover {
  border-color: var(--accent);
}
.stat-card__label {
  font-family: var(--mono);
  font-size: 10px;
  letter-spacing: 0.25em;
  text-transform: uppercase;
  color: var(--text-mute);
}
.stat-card__value-wrapper {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
}
.stat-card__value {
  font-family: 'Oswald', sans-serif;
  font-size: 2.5rem;
  font-weight: 300;
  line-height: 1;
  color: var(--text);
  display: flex;
  align-items: baseline;
}
.stat-card__unit {
  font-size: 0.9rem;
  color: var(--text-mute);
  margin-left: 0.3rem;
  font-weight: 300;
}
.stat-card__trend {
  display: flex;
  align-items: center;
  font-family: var(--mono);
  font-size: 11px;
}
</style>
