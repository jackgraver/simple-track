# Log Meal Page Architecture

This directory follows the architecture pattern with page-specific `api/` and `queries/` directories.

## Structure

```
logmeal/
├── api/
│   └── meals.ts              # Meal API functions
├── queries/
│   ├── keys.ts               # Query key definitions
│   ├── useMeal.ts            # Meal query composable
│   ├── useDietLogsToday.ts   # Today's diet logs query
│   └── useMealMutations.ts   # Meal mutation composables
└── logmeal.vue               # Main component
```

## API Functions

### `api/meals.ts`
- `getMealById(id: number)` - Fetches a meal by ID
- `getDietLogsToday()` - Fetches today's diet logs
- `createMeal(meal: Meal, log: boolean)` - Creates a new meal
- `logEditedMeal(meal: Meal)` - Logs an edited meal
- `updateLoggedMeal(meal: Meal, oldMealId: number)` - Updates a logged meal

## Query Composables

### `queries/useMeal.ts`
- Returns meal data by ID
- Automatically disabled when ID is null or 0
- Reactive to route query parameter changes

### `queries/useDietLogsToday.ts`
- Returns today's diet logs with totals
- Used for displaying macro bars

### `queries/useMealMutations.ts`
- `useCreateMeal()` - Mutation for creating meals
- `useLogEditedMeal()` - Mutation for logging edited meals
- `useUpdateLoggedMeal()` - Mutation for updating logged meals
- All mutations automatically invalidate related queries and navigate on success

## Query Keys

All query keys are defined in `queries/keys.ts`:
- `logmealKeys.meals.detail(id)` - For meal queries
- `logmealKeys.diet.today()` - For today's diet logs

## Component Usage

The component uses:
- `useMeal(mealId)` - Reactive meal query based on route query param
- `useDietLogsToday()` - Today's diet logs
- Mutation composables for all write operations

## Benefits

1. **Separation of Concerns**: API functions, queries, and components are clearly separated
2. **Type Safety**: Full TypeScript support throughout
3. **Automatic Caching**: TanStack Query handles caching automatically
4. **Cache Invalidation**: Mutations automatically invalidate related queries (including home page)
5. **Reusability**: API functions can be reused, queries can be composed
6. **Testability**: Each layer can be tested independently

