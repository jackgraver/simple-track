<script setup lang="ts">
import { useGroceryItems } from "~/api/tracking/queries"; 
import addTextList from "./dialog/useListDialog.ts"

import onInputKeydown from "./useItemInput.ts";

// API call for Grocery List
const {data, isError, isPending } = useGroceryItems();

const temp = ref("tmp");
function test(e) {
    console.log('res', e.key);
    temp.value = e.key;
}

</script>

<template>
    <div class="flex w-full min-w-0 flex-col gap-4">
        <div class="flex w-full">
            <h1 class="m-0 text-lg font-semibold text-textPrimary flex-1">Grocery List - {{ temp }}</h1>
            <button
                class="bg-gray-800 p-1"
                @click="addTextList">
            Add List
            </button>
        </div>
        <div v-if="isError" class="text-sm text-(--color-cf-red)">
            {{ "Failed to load grocery list" }}
        </div>
        <div v-else-if="isPending" class="text-sm text-textSecondary">
            Loading…
        </div>
        <section class="flex flex-col">
            <div class="flex flex-col">
                <div
                    v-for="item in displayItems"
                    :key="item.ID"
                    class="grocery-row-collapse grid transition-[grid-template-rows] duration-500 ease-in-out"
                    :class="
                        isRemoving(item.ID)
                            ? 'grid-rows-[0fr]'
                            : 'grid-rows-[1fr]'
                    "
                >
                    <div class="min-h-0 overflow-hidden">
                        <div
                            class="grocery-row flex items-center gap-3 py-1.5 text-sm transition-opacity duration-500 ease-out"
                            :class="{
                                'grocery-row--completing': isCompleting(item.ID),
                                'opacity-0': isRemoving(item.ID),
                            }"
                        >
                            <input
                                type="checkbox"
                    class="m-0 size-4 shrink-0 cursor-pointer"
                                :aria-label="`Complete ${item.name}`"
                                :checked="isCompleting(item.ID)"
                                @change="onItemCheck(item.ID, $event)"
                            />
                            <span
                                class="grocery-row__label min-w-0 flex-1 truncate"
                                >{{ item.name }}</span
                            >
                        </div>
                    </div>
                </div>
            </div>
            <div class="flex items-center gap-3 py-1.5">
                <input
                    type="checkbox"
                    class="m-0 size-4 shrink-0 pointer-events-none opacity-30"
                    disabled
                    aria-hidden="true"
                />
                <div class="grocery-input-wrap relative min-w-0 flex-1">
                    <div
                        aria-hidden="true"
                        class="grocery-input-mirror pointer-events-none absolute inset-0 overflow-hidden whitespace-pre"
                    >
                        <span class="text-textPrimary">{{ newItem }}</span
                        ><span v-if="showGhost" class="text-textSecondary/45">{{
                            ghostSuffix
                        }}</span>
                    </div>
                    <input
                        ref="newItemInput"
                        v-model="newItem"
                        type="text"
                        autocomplete="off"
                        aria-label="Add grocery item"
                        placeholder="Add item…"
                        class="relative w-full text-base min-w-0 outline-none p-0 m-0 border-none bg-transparent caret-[#dadada] leading-5"
                        @focus="inputFocused = true"
                        @blur="onInputBlur"
                        @input="onInputInput"
                        @keydown="test"
                    />
                </div>
            </div>
            </section>
    </div>
</template>

