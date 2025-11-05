<script setup lang="ts">
import { Home, Dumbbell, Utensils, Settings } from "lucide-vue-next";
import { useRoute } from "vue-router";

const route = useRoute();

const links = [
    { name: "Home", path: "/", icon: Home },
    { name: "Diet", path: "/diet", icon: Utensils },
    { name: "Gym", path: "/gym", icon: Dumbbell },
    { name: "Settings", path: "/settings", icon: Settings },
];
</script>

<template>
    <nav class="sidebar">
        <ul>
            <li
                v-for="link in links"
                :key="link.path"
                :class="{ active: route.path === link.path }"
            >
                <NuxtLink :to="link.path" class="nav-link">
                    <component :is="link.icon" />
                </NuxtLink>
            </li>
        </ul>
    </nav>
</template>

<style scoped>
.sidebar {
    padding: 1rem 1rem 1rem 0;
}

.sidebar ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

@media (max-width: 767px) {
    .sidebar ul {
        flex-direction: row;
        justify-content: center;
    }

    .sidebar li {
        margin-bottom: 0;
    }
}

.nav-link {
    display: flex;
    align-items: center;
    justify-content: center;
    color: #ccc;
    text-decoration: none;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    transition:
        background-color 0.2s ease,
        color 0.2s ease;
}

.nav-link:hover {
    background-color: #333;
    color: #fff;
}

.active .nav-link {
    background-color: #444;
    color: #fff;
}
</style>
