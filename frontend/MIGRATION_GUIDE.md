# Migration Guide: Nuxt 3 → Vue 3 + Vite + Vue Router

## Overview
This guide documents the migration from Nuxt 3 to a plain Vue 3 + Vite + Vue Router setup.

## Key Changes

### 1. Project Structure
- **Before**: `app/` directory with Nuxt conventions
- **After**: `src/` directory with standard Vue structure

### 2. Routing
- **Before**: File-based routing with `pages/` directory
- **After**: Explicit route definitions in `src/router/index.ts`

### 3. Configuration
- **Before**: `useRuntimeConfig()` for environment variables
- **After**: `import.meta.env.VITE_*` with `src/config/env.ts` wrapper

### 4. Components
- **Before**: `NuxtLink` component
- **After**: `RouterLink` from vue-router

### 5. Data Fetching
- **Before**: `useFetch()` and `$fetch()` from Nuxt
- **After**: Custom `useAPIGet()` using axios

### 6. Auto-imports
- **Before**: Nuxt auto-imports
- **After**: `unplugin-auto-import` for Vue composables

## Migration Steps

### Step 1: Install Dependencies
```bash
cd frontend
npm install
```

### Step 2: Create Environment File
Create `.env` file:
```
VITE_API_BASE=http://192.168.4.64:8080
```

### Step 3: Move Files
All files from `app/` need to be moved to `src/`:
- `app/components/` → `src/components/`
- `app/pages/` → `src/pages/`
- `app/composables/` → `src/composables/`
- `app/utils/` → `src/utils/`
- `app/types/` → `src/types/`

### Step 4: Replace NuxtLink
Find and replace in all components:
- `<NuxtLink :to="...">` → `<RouterLink :to="...">`
- `NuxtLink` → `RouterLink` in imports (if any)

### Step 5: Update Imports
- `~/` alias still works (configured in vite.config.ts)
- Remove any Nuxt-specific imports

### Step 6: Update Error Handling
- Remove `error.vue` or convert to a route component
- Update error handling in router

## Files to Update

### Components with NuxtLink:
1. `src/components/SideBar.vue` - Replace `NuxtLink` with `RouterLink`
2. `src/components/today/Gym.vue` - Replace `NuxtLink` with `RouterLink`
3. `src/error.vue` (if keeping) - Replace `NuxtLink` with `RouterLink`

### Files Already Updated:
- ✅ `vite.config.ts` - Vite configuration
- ✅ `src/main.ts` - App entry point
- ✅ `src/App.vue` - Root component (RouterView instead of NuxtPage)
- ✅ `src/router/index.ts` - Route definitions
- ✅ `src/config/env.ts` - Environment config
- ✅ `src/utils/axios.ts` - Updated to use env config
- ✅ `src/composables/useApiFetch.ts` - Updated to use axios

## Running the App

```bash
npm run dev
```

The app will run on `http://localhost:3000` (or the port Vite assigns).

## Build

```bash
npm run build
```

Output will be in `dist/` directory.

