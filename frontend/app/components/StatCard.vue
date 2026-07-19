<script setup lang="ts">
defineProps<{
  title: string
  value: string | number
  unit?: string
  trend?: number // positive or negative percentage
  actionLabel?: string
  actionIcon?: string
}>()

const emit = defineEmits<{
  action: []
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
    <button
      v-if="actionLabel"
      class="stat-card__action"
      type="button"
      @click.stop="emit('action')"
    >
      <UIcon
        v-if="actionIcon"
        :name="actionIcon"
        class="stat-card__action-icon"
      />
      <span>{{ actionLabel }}</span>
    </button>
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
.stat-card:hover:not(:has(.stat-card__action:hover)) {
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
.stat-card__action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  align-self: flex-start;
  margin-top: 0.35rem;
  padding: 0.45rem 0.75rem;
  border: 1px solid var(--accent);
  background: var(--accent);
  color: #000;
  font-family: var(--mono);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  cursor: pointer;
  transition: background 220ms var(--ease-quiet), border-color 220ms var(--ease-quiet), transform 220ms var(--ease-quiet);
}
.stat-card__action:hover {
  background: #ff7324;
  border-color: #ff7324;
  transform: translateY(-1px);
}
.stat-card__action-icon {
  width: 14px;
  height: 14px;
}

@media (max-width: 640px) {
  .stat-card {
    padding: 0.75rem;
    gap: 0.35rem;
    min-width: 0;
    overflow: hidden;
  }

  .stat-card__label {
    font-size: 7px;
    letter-spacing: 0.15em;
  }

  .stat-card__value-wrapper {
    flex-wrap: wrap;
    gap: 0.1rem 0.4rem;
  }

  .stat-card__value {
    font-size: 1.3rem;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .stat-card__unit {
    font-size: 0.55rem;
    margin-left: 0.2rem;
  }

  .stat-card__trend {
    font-size: 8px;
  }

  .stat-card__action {
    margin-top: 0.1rem;
    padding: 0.25rem 0.5rem;
    font-size: 7px;
    gap: 0.25rem;
  }

  .stat-card__action-icon {
    width: 10px;
    height: 10px;
  }
}

/* Extra-small devices (320px and below) */
@media (max-width: 380px) {
  .stat-card {
    padding: 0.6rem;
    gap: 0.25rem;
  }

  .stat-card__label {
    font-size: 6.5px;
    letter-spacing: 0.1em;
  }

  .stat-card__value {
    font-size: 1.15rem;
  }

  .stat-card__unit {
    font-size: 0.5rem;
  }

  .stat-card__trend {
    font-size: 7.5px;
  }

  .stat-card__action {
    padding: 0.2rem 0.4rem;
    font-size: 6.5px;
  }
}
</style>
