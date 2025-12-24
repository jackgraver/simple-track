# Next Steps for Migration

## ✅ Completed

1. ✅ Created Vite configuration (`vite.config.ts`)
2. ✅ Created `index.html` entry point
3. ✅ Set up Vue Router with all routes (`src/router/index.ts`)
4. ✅ Created environment config (`src/config/env.ts`)
5. ✅ Updated `package.json` with Vite dependencies
6. ✅ Created `main.ts` entry point
7. ✅ Created `App.vue` (replaced NuxtPage with RouterView)
8. ✅ Updated `axios.ts` to use environment variables
9. ✅ Updated `useApiFetch.ts` to use axios instead of useFetch/$fetch
10. ✅ Set up auto-imports in vite.config.ts
11. ✅ Migrated Vue Query plugin setup to main.ts
12. ✅ Updated `SideBar.vue` with RouterLink

## 🔄 Remaining Tasks

### 1. Install Dependencies
```bash
cd frontend
npm install
```

### 2. Create Environment File
Create `.env` file in `frontend/`:
```
VITE_API_BASE=http://192.168.4.64:8080
```

### 3. Move Files from `app/` to `src/`

You can either:
- **Option A**: Use the migration script (Linux/Mac):
  ```bash
  chmod +x migrate.sh
  ./migrate.sh
  ```

- **Option B**: Manually copy files:
  - `app/components/` → `src/components/`
  - `app/pages/` → `src/pages/`
  - `app/composables/` → `src/composables/` (already created in src)
  - `app/utils/` → `src/utils/` (already created in src)
  - `app/types/` → `src/types/`

### 4. Update Components with NuxtLink

Files that need `NuxtLink` → `RouterLink` replacement:

1. **`src/components/today/Gym.vue`** (after moving)
   - Replace: `<NuxtLink :to="'liveworkout'">` 
   - With: `<RouterLink :to="'/liveworkout'">`

2. **`src/error.vue`** (if you keep it, or create as route component)
   - Replace: `<NuxtLink to="/">`
   - With: `<RouterLink to="/">`

### 5. Update Imports

After moving files, check for any remaining Nuxt-specific imports:
- Remove `#app` imports
- Remove `useRuntimeConfig` (already replaced)
- Ensure all `~/` imports work (configured in vite.config.ts)

### 6. Test the Application

```bash
npm run dev
```

## File Structure After Migration

```
frontend/
├── src/
│   ├── components/
│   │   ├── SideBar.vue (✅ updated)
│   │   ├── today/
│   │   │   └── Gym.vue (needs RouterLink update)
│   │   └── ... (other components)
│   ├── pages/
│   │   ├── index.vue
│   │   ├── diet.vue
│   │   ├── gym.vue
│   │   └── ... (other pages)
│   ├── composables/
│   │   ├── useApiFetch.ts (✅ updated)
│   │   └── ... (other composables)
│   ├── utils/
│   │   ├── axios.ts (✅ updated)
│   │   └── dateUtil.ts
│   ├── types/
│   │   ├── diet.ts
│   │   └── workout.ts
│   ├── config/
│   │   └── env.ts (✅ created)
│   ├── router/
│   │   └── index.ts (✅ created)
│   ├── App.vue (✅ created)
│   └── main.ts (✅ created)
├── public/
│   ├── favicon.ico
│   └── robots.txt
├── index.html (✅ created)
├── vite.config.ts (✅ created)
├── tsconfig.json (✅ updated)
└── package.json (✅ updated)
```

## Quick Reference: Replacements

| Nuxt 3 | Vue 3 + Vite |
|--------|--------------|
| `NuxtLink` | `RouterLink` |
| `NuxtPage` | `RouterView` |
| `useRuntimeConfig()` | `import.meta.env.VITE_*` or `config` from `~/config/env` |
| `useFetch()` | `useAPIGet()` (custom composable) |
| `$fetch()` | `apiClient` (axios instance) |
| File-based routing | Explicit routes in `src/router/index.ts` |
| Auto-imports | `unplugin-auto-import` (configured) |

## Troubleshooting

### If components aren't auto-imported:
- Check `vite.config.ts` - Components plugin is configured
- Ensure components are in `src/components/`
- Restart dev server

### If routes don't work:
- Check `src/router/index.ts` - all routes are defined
- Ensure `RouterView` is in `App.vue`
- Check browser console for errors

### If environment variables don't work:
- Ensure `.env` file exists
- Variables must start with `VITE_`
- Restart dev server after changing `.env`

