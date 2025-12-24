# Client Data Layer Architecture (Nuxt 3 + Axios + TanStack Query)

## Goal

Separate data fetching, server-state management, and UI concerns so each layer has a single responsibility, predictable behavior, and minimal coupling.

## Folder Structure

```
app/
  lib/
    api.ts                    # Axios instance & configuration
  api/
    workouts.ts               # Workout API functions
    exercises.ts              # Exercise API functions
    diet.ts                   # Diet API functions
  queries/
    keys.ts                   # Centralized query keys
    workouts/
      useWorkouts.ts          # Workout queries
      useWorkoutLogs.ts       # Workout log queries
      useCreateWorkout.ts     # Workout mutations
    exercises/
      useExercises.ts         # Exercise queries
    diet/
      useDietLogs.ts          # Diet queries
  composables/
    useAuth.ts                # Auth state management
  plugins/
    vue-query.client.ts       # TanStack Query setup
  types/
    workout.ts                # Workout types
    diet.ts                   # Diet types
    api.ts                    # API response types
```

## Folder Responsibilities

### `lib/api.ts`

**Purpose:** Low-level infrastructure

- Contains shared libraries and configuration
- Axios instance lives here
- No domain logic
- No application state

**Example contents:**
- Axios base URL configuration
- Request/response interceptors
- Auth token injection
- Error handling
- Timeouts
- Request/response transformers

**Nuxt-specific notes:**
- Uses `useRuntimeConfig()` for base URL
- Can use `useCookie()` or `useState()` for token storage
- Interceptors handle token refresh logic

### `api/`

**Purpose:** Raw HTTP request functions

- One function = one HTTP request
- Directly calls Axios instance
- Returns parsed response data
- No caching
- No retries (handled by Axios interceptors)
- No state
- No side effects

These functions represent the API contract, not application behavior.

**Example:**
```typescript
// api/workouts.ts
import { apiClient } from '~/lib/api';
import type { WorkoutLog, PaginatedResponse } from '~/types/api';

export async function getWorkoutLogs(params: {
  page?: number;
  pageSize?: number;
  offset?: number;
}): Promise<PaginatedResponse<WorkoutLog>> {
  const response = await apiClient.get('/workout/logs', { params });
  return response.data;
}

export async function getWorkoutLogById(id: number): Promise<WorkoutLog> {
  const response = await apiClient.get(`/workout/logs/${id}`);
  return response.data.workout_log;
}

export async function createWorkoutLog(data: Partial<WorkoutLog>): Promise<WorkoutLog> {
  const response = await apiClient.post('/workout/logs', data);
  return response.data.workout_log;
}
```

### `queries/`

**Purpose:** Server-state management

- Wraps API functions using TanStack Query
- Owns:
  - Caching
  - Loading/error states
  - Refetching
  - Invalidation
  - Mutations
  - Optimistic updates

This is where application data behavior lives.

**Each query or mutation:**
- Has a stable query key (from `keys.ts`)
- Uses exactly one API function
- Encapsulates invalidation logic
- Handles error states appropriately

**Example:**
```typescript
// queries/workouts/useWorkoutLogs.ts
import { useQuery } from '@tanstack/vue-query';
import { getWorkoutLogs } from '~/api/workouts';
import { workoutKeys } from '~/queries/keys';

export function useWorkoutLogs(params: {
  page?: number;
  pageSize?: number;
  offset?: number;
}) {
  return useQuery({
    queryKey: workoutKeys.logs.list(params),
    queryFn: () => getWorkoutLogs(params),
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}
```

### `queries/keys.ts`

**Purpose:** Centralized query keys

- All query keys are defined here
- Prevents key duplication and typos
- Enables safe cache invalidation
- Keys are hierarchical and composable

**No queries or mutations live here—only key definitions.**

**Example:**
```typescript
// queries/keys.ts
export const workoutKeys = {
  all: ['workouts'] as const,
  lists: () => [...workoutKeys.all, 'list'] as const,
  list: (filters?: Record<string, any>) => 
    [...workoutKeys.lists(), filters] as const,
  details: () => [...workoutKeys.all, 'detail'] as const,
  detail: (id: number) => [...workoutKeys.details(), id] as const,
  logs: {
    all: [...workoutKeys.all, 'logs'] as const,
    lists: () => [...workoutKeys.logs.all, 'list'] as const,
    list: (params?: { page?: number; pageSize?: number; offset?: number }) =>
      [...workoutKeys.logs.lists(), params] as const,
    detail: (id: number) => [...workoutKeys.logs.all, 'detail', id] as const,
  },
};
```

### `types/`

**Purpose:** Shared TypeScript types

- DTOs (Data Transfer Objects)
- API response shapes
- Domain models
- Pagination types

**No logic**

**Example:**
```typescript
// types/api.ts
export interface PaginatedResponse<T> {
  [key: string]: T[]; // Dynamic key based on resource name
  pagination: {
    total: number;
    page: number;
    pageSize: number;
    totalPages: number;
    hasNext: boolean;
    hasPrev: boolean;
  };
}

export interface ApiError {
  error: string;
  message?: string;
  statusCode?: number;
}
```

## Rules (Important)

### Axios Rules

- ✅ Axios functions must not store state
- ✅ Axios functions must not know about TanStack Query
- ✅ Axios functions must not retry, cache, or invalidate
- ✅ Axios functions must be callable independently
- ✅ Axios functions must handle auth tokens via interceptors

**Axios answers:**
> "How do I talk to the server?"

### TanStack Query Rules

- ✅ All server data must go through TanStack Query
- ✅ Components must not call Axios directly
- ✅ Components must not call API functions directly
- ✅ Cache invalidation happens only here
- ✅ Queries are read-only
- ✅ Mutations handle writes and side effects
- ✅ Query keys must come from `keys.ts`

**TanStack Query answers:**
> "How long do I keep this data, and when does it change?"

### Component Rules

- ✅ Components consume queries and mutations
- ✅ Components do not manage loading or error state manually (use query states)
- ✅ Components do not fetch data imperatively (use queries/mutations)
- ✅ Components do not know API URLs
- ✅ Components do not import from `api/` or `lib/api`

**Components answer:**
> "How do I display and interact with data?"

### State Management Rule

- ✅ TanStack Query = server state
- ✅ Pinia (if used) = client/UI state only (modals, form state, etc.)
- ✅ `useState()` = component-level reactive state
- ✅ `useCookie()` = persistent client state (auth tokens)
- ❌ Never mirror server data into Pinia or local state

## Authentication

### Token Storage

**Use Nuxt's `useCookie()` for token persistence:**

```typescript
// composables/useAuth.ts
export const useAuth = () => {
  const token = useCookie<string | null>('auth_token', {
    default: () => null,
    secure: true,
    sameSite: 'strict',
    maxAge: 60 * 60 * 24 * 7, // 7 days
  });

  const setToken = (newToken: string) => {
    token.value = newToken;
  };

  const clearToken = () => {
    token.value = null;
  };

  const isAuthenticated = computed(() => !!token.value);

  return {
    token: readonly(token),
    setToken,
    clearToken,
    isAuthenticated,
  };
};
```

### Token Injection in Axios

**Add request interceptor to inject token:**

```typescript
// lib/api.ts
import axios from 'axios';
import { useAuth } from '~/composables/useAuth';

export function createAxiosInstance() {
  const config = useRuntimeConfig();
  const baseURL = config.public.apiBase || 'http://localhost:8080';

  const instance = axios.create({
    baseURL,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  // Request interceptor: inject auth token
  instance.interceptors.request.use(
    (config) => {
      const { token } = useAuth();
      if (token.value) {
        config.headers.Authorization = `Bearer ${token.value}`;
      }
      return config;
    },
    (error) => Promise.reject(error)
  );

  // Response interceptor: handle auth errors
  instance.interceptors.response.use(
    (response) => response,
    async (error) => {
      const { clearToken } = useAuth();
      
      // Handle 401 Unauthorized
      if (error.response?.status === 401) {
        clearToken();
        // Optionally redirect to login
        await navigateTo('/login');
      }

      // Handle token refresh (if applicable)
      if (error.response?.status === 403 && error.config && !error.config._retry) {
        error.config._retry = true;
        // Implement token refresh logic here
        // const newToken = await refreshToken();
        // error.config.headers.Authorization = `Bearer ${newToken}`;
        // return instance(error.config);
      }

      return Promise.reject(error);
    }
  );

  return instance;
}

export const apiClient = createAxiosInstance();
```

### Auth in Queries

**Queries automatically use token via Axios interceptor:**

```typescript
// No special handling needed in queries
// The token is automatically injected by the Axios interceptor

export function useWorkoutLogs(params: WorkoutLogParams) {
  return useQuery({
    queryKey: workoutKeys.logs.list(params),
    queryFn: () => getWorkoutLogs(params), // Token injected automatically
  });
}
```

## Pagination

### Standard Pagination

**For paginated endpoints with page/pageSize:**

```typescript
// api/workouts.ts
export async function getWorkoutLogsPaginated(params: {
  page: number;
  pageSize: number;
}): Promise<PaginatedResponse<WorkoutLog>> {
  const response = await apiClient.get('/workout/logs', {
    params: {
      page: params.page,
      pageSize: params.pageSize,
    },
  });
  return response.data;
}

// queries/workouts/useWorkoutLogsPaginated.ts
import { useQuery } from '@tanstack/vue-query';
import { getWorkoutLogsPaginated } from '~/api/workouts';
import { workoutKeys } from '~/queries/keys';

export function useWorkoutLogsPaginated(params: {
  page: number;
  pageSize: number;
}) {
  return useQuery({
    queryKey: workoutKeys.logs.list(params),
    queryFn: () => getWorkoutLogsPaginated(params),
    keepPreviousData: true, // Smooth transitions between pages
  });
}

// Component usage
const page = ref(1);
const pageSize = ref(10);
const { data, isPending } = useWorkoutLogsPaginated({
  page: page.value,
  pageSize: pageSize.value,
});

const pagination = computed(() => data.value?.pagination);
```

### Infinite Queries

**For infinite scroll or "load more" patterns:**

```typescript
// api/exercises.ts
export async function getExercises(params: {
  page: number;
  pageSize: number;
}): Promise<PaginatedResponse<Exercise>> {
  const response = await apiClient.get('/exercises', {
    params: {
      page: params.page,
      pageSize: params.pageSize,
    },
  });
  return response.data;
}

// queries/exercises/useExercisesInfinite.ts
import { useInfiniteQuery } from '@tanstack/vue-query';
import { getExercises } from '~/api/exercises';
import { exerciseKeys } from '~/queries/keys';

export function useExercisesInfinite(pageSize: number = 20) {
  return useInfiniteQuery({
    queryKey: exerciseKeys.infinite(pageSize),
    queryFn: ({ pageParam = 1 }) => getExercises({ page: pageParam, pageSize }),
    getNextPageParam: (lastPage) => {
      const { pagination } = lastPage;
      return pagination.hasNext ? pagination.page + 1 : undefined;
    },
    getPreviousPageParam: (firstPage) => {
      const { pagination } = firstPage;
      return pagination.hasPrev ? pagination.page - 1 : undefined;
    },
    initialPageParam: 1,
  });
}

// Component usage
const {
  data,
  fetchNextPage,
  hasNextPage,
  isFetchingNextPage,
  isPending,
} = useExercisesInfinite(20);

const exercises = computed(() => 
  data.value?.pages.flatMap(page => page.exercises) ?? []
);

// Load more button
<button 
  @click="fetchNextPage()" 
  :disabled="!hasNextPage || isFetchingNextPage"
>
  {{ isFetchingNextPage ? 'Loading...' : 'Load More' }}
</button>
```

### Infinite Query Keys

**Add infinite query keys to `keys.ts`:**

```typescript
// queries/keys.ts
export const exerciseKeys = {
  all: ['exercises'] as const,
  lists: () => [...exerciseKeys.all, 'list'] as const,
  list: (params?: Record<string, any>) => 
    [...exerciseKeys.lists(), params] as const,
  infinite: (pageSize: number) => 
    [...exerciseKeys.all, 'infinite', pageSize] as const,
  details: () => [...exerciseKeys.all, 'detail'] as const,
  detail: (id: number) => [...exerciseKeys.details(), id] as const,
};
```

## Mutations

### Creating Mutations

**Mutations handle writes and side effects:**

```typescript
// queries/workouts/useCreateWorkoutLog.ts
import { useMutation, useQueryClient } from '@tanstack/vue-query';
import { createWorkoutLog } from '~/api/workouts';
import { workoutKeys } from '~/queries/keys';

export function useCreateWorkoutLog() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Partial<WorkoutLog>) => createWorkoutLog(data),
    onSuccess: () => {
      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: workoutKeys.logs.all });
    },
  });
}

// Component usage
const createMutation = useCreateWorkoutLog();

const handleCreate = async () => {
  try {
    await createMutation.mutateAsync({
      date: new Date().toISOString(),
      workout_plan_id: 1,
    });
    // Success handled by onSuccess
  } catch (error) {
    // Error handling
    console.error('Failed to create workout log:', error);
  }
};
```

### Optimistic Updates

**For better UX, update cache optimistically:**

```typescript
// queries/workouts/useUpdateWorkoutLog.ts
import { useMutation, useQueryClient } from '@tanstack/vue-query';
import { updateWorkoutLog } from '~/api/workouts';
import { workoutKeys } from '~/queries/keys';
import type { WorkoutLog } from '~/types/workout';

export function useUpdateWorkoutLog() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<WorkoutLog> }) =>
      updateWorkoutLog(id, data),
    onMutate: async ({ id, data }) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ 
        queryKey: workoutKeys.logs.detail(id) 
      });

      // Snapshot previous value
      const previousLog = queryClient.getQueryData<WorkoutLog>(
        workoutKeys.logs.detail(id)
      );

      // Optimistically update
      queryClient.setQueryData<WorkoutLog>(
        workoutKeys.logs.detail(id),
        (old) => old ? { ...old, ...data } : undefined
      );

      return { previousLog };
    },
    onError: (err, variables, context) => {
      // Rollback on error
      if (context?.previousLog) {
        queryClient.setQueryData(
          workoutKeys.logs.detail(variables.id),
          context.previousLog
        );
      }
    },
    onSettled: (data, error, variables) => {
      // Refetch to ensure consistency
      queryClient.invalidateQueries({ 
        queryKey: workoutKeys.logs.detail(variables.id) 
      });
    },
  });
}
```

## Data Flow (End to End)

1. **Component** calls a query or mutation composable
2. **Query/Mutation** calls an API function with query key
3. **API function** calls Axios instance
4. **Axios interceptor** injects auth token
5. **Axios** makes HTTP request
6. **Response** flows back through Axios
7. **Response interceptor** handles errors/auth
8. **TanStack Query** caches response
9. **Component** reactively updates via query state

**No layer skips another.**

## Mental Model

- **Axios** fetches (HTTP layer)
- **TanStack Query** owns the data (state management)
- **Components** render (UI layer)
- **State** lives where it belongs

## Migration Guide

### Current State Analysis

**✅ Already Implemented:**
- TanStack Vue Query plugin (`plugins/vue-query.client.ts`)
- Axios instance (`utils/axios.ts`)
- Example implementation in `Gym.vue`

**🔄 Needs Migration:**
- `composables/useApiFetch.ts` → Replace with TanStack Query
- Direct API calls in components → Move to `api/` functions
- Components calling `useAPIGet` → Use query composables

### Migration Steps

1. **Create API functions** in `api/` directory
2. **Create query keys** in `queries/keys.ts`
3. **Create query composables** in `queries/` directory
4. **Update components** to use query composables
5. **Add auth composable** and update Axios instance
6. **Remove old `useApiFetch.ts`** once migration complete

### Example Migration

**Before:**
```typescript
// Component
const { data, pending } = useAPIGet<WorkoutLog>(
  `workout/logs/previous?offset=${props.dateOffset}`
);
```

**After:**
```typescript
// api/workouts.ts
export async function getWorkoutLogsPrevious(offset: number) {
  const response = await apiClient.get(`/workout/logs/previous`, {
    params: { offset },
  });
  return response.data;
}

// queries/workouts/useWorkoutLogsPrevious.ts
export function useWorkoutLogsPrevious(offset: number) {
  return useQuery({
    queryKey: workoutKeys.logs.previous(offset),
    queryFn: () => getWorkoutLogsPrevious(offset),
  });
}

// Component
const { data, isPending } = useWorkoutLogsPrevious(props.dateOffset);
```

## Best Practices

1. **Always use query keys from `keys.ts`** - Never hardcode keys
2. **One API function per endpoint** - Keep functions focused
3. **Invalidate related queries** - When mutations succeed
4. **Use `keepPreviousData`** - For pagination smoothness
5. **Handle errors gracefully** - In mutations and queries
6. **Type everything** - Use TypeScript types from `types/`
7. **Test API functions independently** - They're pure functions
8. **Use optimistic updates** - For better UX on mutations
9. **Cache appropriately** - Set `staleTime` based on data freshness needs
10. **Compose queries** - Build complex queries from simple ones

## Nuxt-Specific Considerations

1. **SSR Compatibility**: TanStack Query works with SSR, but ensure API calls are client-only when needed
2. **Runtime Config**: Use `useRuntimeConfig()` for API base URL
3. **Auto-imports**: Nuxt auto-imports from `composables/`, so composables are available globally
4. **Plugins**: TanStack Query plugin must be `.client.ts` for client-side only
5. **State Hydration**: TanStack Query handles SSR hydration automatically
6. **Cookie Storage**: Use `useCookie()` for persistent auth tokens
7. **Navigation**: Use `navigateTo()` for auth redirects (not `router.push()`)

