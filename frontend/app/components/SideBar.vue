<script setup lang="ts">
import { Home, Dumbbell, Utensils } from "lucide-vue-next";
import { useRoute } from "vue-router";

const route = useRoute();

const links = [
    { name: "Home", path: "/", icon: Home },
    { name: "Gym", path: "/gym", icon: Dumbbell },
    { name: "Diet", path: "/diet", icon: Utensils },
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
                    <!-- <span>{{ link.name }}</span> -->
                </NuxtLink>
            </li>
        </ul>
    </nav>
</template>

<style scoped>
.sidebar {
    position: fixed;
    top: 0;
    left: 0;
    height: 100vh;
    display: flex;
    flex-direction: column;
    justify-content: flex-start;
    padding-top: 1rem;
    z-index: 1000;
}

.sidebar ul {
    list-style: none;
    margin: 0;
    padding: 0;
}

.sidebar li {
    margin-bottom: 1rem;
}

.nav-link {
    display: flex;
    align-items: center;
    gap: 0.75rem;
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
