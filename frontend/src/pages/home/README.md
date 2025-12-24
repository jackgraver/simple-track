# Home Page Architecture

This directory follows the architecture pattern with page-specific `api/` and `queries/` directories.

## Structure

```
home/
├── api/
│   ├── workouts.ts      # Workout API functions
│   └── diet.ts          # Diet API functions
├── queries/
│   ├── keys.ts          # Query key definitions
│   ├── useWorkoutLogsPrevious.ts  # Workout query composable
│   ├── useDietLogsToday.ts       # Diet query composable
│   └── useMealMutations.ts      # Meal mutation composables
├── components/
│   ├── Gym.vue          # Uses useWorkoutLogsPrevious
│   ├── Meal.vue         # Uses useDietLogsToday + mutations
│   └── ...
└── home.vue
```

## API Functions

### `api/workouts.ts`
- `getWorkoutLogsPrevious(offset: number)` - Fetches workout logs with previous exercises

### `api/diet.ts`
- `getDietLogsToday(offset: number)` - Fetches diet logs for a specific day
- `logPlannedMeal(mealId: number)` - Logs a planned meal
- `deleteLoggedMeal(mealId: number, dayId: number)` - Deletes a logged meal
- `editLoggedMeal(meal: Meal, oldMealId: number)` - Edits a logged meal

## Query Composables

### `queries/useWorkoutLogsPrevious.ts`
- Returns workout logs with previous exercises
- Used by `Gym.vue` component
- Automatically caches and refetches based on `offset`

### `queries/useDietLogsToday.ts`
- Returns diet logs for a specific day
- Used by `Meal.vue` component
- Automatically caches and refetches based on `offset`

### `queries/useMealMutations.ts`
- `useLogPlannedMeal(offset)` - Mutation for logging planned meals
- `useDeleteLoggedMeal(offset)` - Mutation for deleting logged meals
- `useEditLoggedMeal(offset)` - Mutation for editing logged meals
- All mutations automatically invalidate the diet query cache on success

## Query Keys

All query keys are defined in `queries/keys.ts`:
- `homeKeys.workouts.previous(offset)` - For workout logs
- `homeKeys.diet.today(offset)` - For diet logs

## Component Usage

### Gym.vue
```typescript
import { useWorkoutLogsPrevious } from "../queries/useWorkoutLogsPrevious";

const { data, isPending: pending, error } = useWorkoutLogsPrevious(props.dateOffset);
```

### Meal.vue
```typescript
import { useDietLogsToday } from "../queries/useDietLogsToday";
import { useLogPlannedMeal, useDeleteLoggedMeal, useEditLoggedMeal } from "../queries/useMealMutations";

const { data, isPending: pending, error } = useDietLogsToday(props.dateOffset);
const logPlannedMealMutation = useLogPlannedMeal(props.dateOffset);
// ... use mutations
```

## Benefits

1. **Separation of Concerns**: API functions, queries, and components are clearly separated
2. **Type Safety**: Full TypeScript support throughout
3. **Automatic Caching**: TanStack Query handles caching automatically
4. **Cache Invalidation**: Mutations automatically invalidate related queries
5. **Reusability**: API functions can be reused, queries can be composed
6. **Testability**: Each layer can be tested independently

