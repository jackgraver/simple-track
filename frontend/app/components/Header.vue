<script lang="ts" setup>
import { useRouter } from "vue-router";

const router = useRouter();
</script>

<template>
    <header>
        <div class="logo">
            <img src="/favicon.ico" alt="logo" />
            <h1>Simple Tracker</h1>
        </div>
        <div class="nav">
            <h2
                @click="router.push('/')"
                :class="router.currentRoute.value.path === '/' ? 'active' : ''"
            >
                Home
            </h2>
            <h2
                @click="router.push('/mealplan')"
                :class="
                    router.currentRoute.value.path === '/mealplan'
                        ? 'active'
                        : ''
                "
            >
                Meal Plan
            </h2>
            <h2
                @click="router.push('/gym')"
                :class="
                    router.currentRoute.value.path === '/gym' ? 'active' : ''
                "
            >
                Workouts
            </h2>
        </div>
    </header>
</template>

<style scoped>
header {
    background-color: black;
    position: sticky;
    top: 0;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 16px;
    z-index: 1000;
}

.logo {
    display: flex;
    align-items: center;
    gap: 8px;
}

.logo img {
    width: 32px;
    height: 32px;
}

.logo h1 {
    margin: 0;
    font-size: 1.2rem;
    color: white;
}

.nav {
    display: flex;
    gap: 16px;
}

.nav h2 {
    cursor: pointer;
    margin: 0;
    color: lightgray;
    transition: color 0.2s;
}

.nav h2:hover {
    text-decoration: underline;
    color: white;
}

.nav h2.active {
    font-weight: bold;
    color: white;
    text-decoration: underline;
}
</style>
