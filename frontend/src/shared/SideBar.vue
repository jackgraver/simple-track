<script setup lang="ts">
import Menu from 'primevue/menu';
import { Home, Dumbbell, Utensils, Settings, BarChart, List } from "lucide-vue-next";
import { useRoute } from "vue-router";
import type { MenuItem } from 'primevue/menuitem';

const route = useRoute();

const iconMap: Record<string, any> = {
    Home,
    Utensils,
    Dumbbell,
    Settings,
    BarChart,
    List,
};

const menuItems: MenuItem[] = [
    { label: "Home", route: "/", iconName: "Home" },
    { label: "Diet", route: "/diet", iconName: "Utensils" },
    { label: "Gym", route: "/gym", iconName: "Dumbbell" },
    { label: "Settings", route: "/settings", iconName: "Settings" },
    { label: "Progression", route: "/progression", iconName: "BarChart" },
    { label: "Manage Plans", route: "/manageplans", iconName: "List" },
];
</script>

<template>
    <nav class="sidebar">
        <Menu :model="menuItems" class="sidebar-menu">
            <template #item="{ item, props }">
                <router-link v-if="(item as any).route" v-slot="{ href, navigate }" :to="(item as any).route" custom>
                    <a
                        v-ripple
                        :href="href"
                        v-bind="props.action"
                        @click="navigate"
                        class="menu-item-link"
                        :class="{ 'p-highlight': route.path === (item as any).route }"
                    >
                        <component :is="iconMap[(item as any).iconName]" class="menu-icon" />
                        <span class="menu-label">{{ item.label }}</span>
                    </a>
                </router-link>
            </template>
        </Menu>
    </nav>
</template>

<style scoped>
.sidebar {
    background: transparent;
}

:deep(.sidebar-menu) {
    width: 100%;
    border: none;
    background: transparent !important;
}

:deep(.sidebar-menu .p-menu-root) {
    background: transparent !important;
}

:deep(.sidebar-menu .p-menu-list) {
    background: transparent !important;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

:deep(.sidebar-menu .p-menuitem) {
    margin: 0;
    background: transparent;
}

:deep(.sidebar-menu .p-menuitem-link) {
    background: transparent;
}

:deep(.sidebar-menu .p-menuitem-content) {
    background: transparent;
}

.menu-item-link {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem;
    border-radius: 6px;
    text-decoration: none;
    color: inherit;
    transition: all 0.2s;
}

.menu-icon {
    width: 1.25rem;
    height: 1.25rem;
    flex-shrink: 0;
}

.menu-label {
    font-size: 0.875rem;
}

@media (max-width: 767px) {
    :deep(.sidebar-menu .p-menu-list) {
        flex-direction: row;
        justify-content: center;
        flex-wrap: wrap;
        gap: 0.5rem;
    }

    .menu-label {
        display: none;
    }

    .menu-icon {
        width: 1.5rem;
        height: 1.5rem;
    }
}
</style>
