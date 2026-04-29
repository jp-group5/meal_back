<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed, ref, watch } from 'vue'

import { getMealsByDate } from '@/api/meal'
import AIDecisionDashboard from '@/components/ai/AIDecisionDashboard.vue'
import GoogleCalendar from '@/components/calendar/GoogleCalendar.vue'
import { useDateStore } from '@/stores/date'
import type { Meal } from '@/types'

const dateStore = useDateStore()
const { selectedDate, selectedMonth } = storeToRefs(dateStore)

const meals = ref<Meal[]>([])
const mealsLoading = ref(false)
const mealsMessage = ref('')

const hasMeals = computed(() => meals.value.length > 0)

watch(
  selectedDate,
  (date) => {
    void loadMeals(date)
  },
  { immediate: true },
)

async function loadMeals(date: string) {
  mealsLoading.value = true
  mealsMessage.value = ''

  try {
    const response = await getMealsByDate(date)
    meals.value = response.data ?? []
  } catch (error) {
    mealsMessage.value = 'Backend is unavailable. Showing local sample meal records.'
    meals.value = createFallbackMeals(date)
  } finally {
    mealsLoading.value = false
  }
}

function handleRecommendationAccepted(payload: {
  recommendationId: string
  name: string
  date: string
  calories: number
}) {
  if (payload.date !== selectedDate.value) {
    return
  }

  meals.value = [
    {
      id: `ai-${payload.recommendationId}-${Date.now()}`,
      date: payload.date,
      type: 'dinner',
      content: payload.name,
      calories: payload.calories,
    },
    ...meals.value,
  ]
}

function createFallbackMeals(date: string): Meal[] {
  return [
    {
      id: `${date}-breakfast`,
      date,
      type: 'breakfast',
      content: 'Greek Yogurt + Berries',
      calories: 320,
    },
    {
      id: `${date}-lunch`,
      date,
      type: 'lunch',
      content: 'Chicken Brown Rice Bowl',
      calories: 560,
    },
  ]
}
</script>

<template>
  <section class="page">
    <div class="page-header">
      <div>
        <p class="eyebrow">Unified Workspace</p>
        <h2>Meal Records + AI Recommendation Workspace</h2>
      </div>

      <div class="toolbar">
        <input
          :value="selectedDate"
          type="date"
          aria-label="Select date"
          @input="dateStore.setSelectedDate(($event.target as HTMLInputElement).value)"
        />
        <button type="button" @click="dateStore.goToday">Today</button>
      </div>
    </div>

    <div class="content-grid">
      <GoogleCalendar />
    </div>

    <div class="content-grid">
      <section class="panel">
        <p class="label">Current month</p>
        <strong>{{ selectedMonth }}</strong>
      </section>

      <section class="panel">
        <p class="label">Current date</p>
        <strong>{{ selectedDate }}</strong>
      </section>
    </div>

    <div class="workspace-grid">
      <section class="panel">
        <div class="section-header">
          <div>
            <p class="label">Meals</p>
            <h3>Daily meal records</h3>
          </div>
          <button
            type="button"
            class="button-outline"
            :disabled="mealsLoading"
            @click="loadMeals(selectedDate)"
          >
            Refresh records
          </button>
        </div>

        <p v-if="mealsMessage" class="info-message">
          {{ mealsMessage }}
        </p>

        <p v-if="mealsLoading" class="subtle-text">Loading meal records...</p>

        <ul v-else-if="hasMeals" class="meal-list">
          <li v-for="meal in meals" :key="meal.id" class="meal-list-item">
            <div>
              <p class="meal-type">{{ meal.type }}</p>
              <p class="meal-content">{{ meal.content }}</p>
            </div>
            <strong class="meal-calories">{{ meal.calories ?? '--' }} kcal</strong>
          </li>
        </ul>

        <p v-else class="subtle-text">No meal records yet.</p>
      </section>

      <AIDecisionDashboard :selected-date="selectedDate" @accepted="handleRecommendationAccepted" />
    </div>
  </section>
</template>

<style scoped>
.workspace-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 1rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  gap: 0.8rem;
  align-items: flex-start;
  margin-bottom: 0.8rem;
}

.section-header h3 {
  margin: 0;
}

.button-outline {
  border-color: #d8dee6;
  background: #ffffff;
  color: #263241;
}

.button-outline:hover {
  background: #f3f6f9;
}

.info-message {
  margin: 0 0 0.8rem;
  border: 1px solid #efd8b2;
  border-radius: 8px;
  background: #fff8eb;
  color: #845f23;
  padding: 0.6rem 0.75rem;
  font-size: 0.86rem;
}

.subtle-text {
  margin: 0;
  color: #5c6a78;
}

.meal-list {
  display: grid;
  gap: 0.7rem;
  list-style: none;
  margin: 0;
  padding: 0;
}

.meal-list-item {
  border: 1px solid #d8dee6;
  border-radius: 8px;
  background: #ffffff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.8rem 0.9rem;
}

.meal-type {
  margin: 0;
  color: #1f6f63;
  text-transform: capitalize;
  font-size: 0.74rem;
  font-weight: 700;
}

.meal-content {
  margin: 0.22rem 0 0;
}

.meal-calories {
  color: #263241;
  font-size: 0.95rem;
  white-space: nowrap;
}

@media (max-width: 960px) {
  .workspace-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .section-header {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
