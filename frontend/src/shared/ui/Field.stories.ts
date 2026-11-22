import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import Checkbox from "./Checkbox.vue";
import Field from "./Field.vue";
import Input from "./Input.vue";
import Select from "./Select.vue";
import Textarea from "./Textarea.vue";
import Toggle from "./Toggle.vue";

const meta = {
    title: "UI/Fields/Showcase",
    component: Field,
    tags: ["autodocs"],
    parameters: { layout: "padded" },
} satisfies Meta<typeof Field>;
export default meta;
type Story = StoryObj<typeof meta>;

const opts = [
    { label: "Breakfast", value: "b" },
    { label: "Lunch", value: "l" },
    { label: "Dinner", value: "d" },
];

/** Every control in default, hint, invalid, disabled and read-only. Below lg the font is 16px (no iOS zoom). */
export const AllStates: Story = {
    args: { label: "Meal name" },
    render: () => ({
        components: { Field, Input, Textarea, Select, Checkbox, Toggle },
        setup: () => ({ opts, t: ref("Oats"), n: ref(80), e: ref(""), d: ref("Locked"), ro: ref("Read only"), sel: ref(""), sel2: ref("l"), ta: ref("A note\nover two lines"), c1: ref(false), c2: ref(true), c3: ref<boolean | "indeterminate">("indeterminate"), tg: ref(true) }),
        template: `
        <div class="grid max-w-3xl gap-6 lg:grid-cols-2">
            <Field label="Meal name" hint="Shown on the day view" required><Input v-model="t" /></Field>
            <Field label="Weight"><Input v-model="n" type="number" inputmode="decimal"><template #suffix>kg</template></Input></Field>
            <Field label="Price"><Input v-model="n" type="number" inputmode="decimal"><template #prefix>$</template></Input></Field>
            <Field label="Username" error="That username is taken"><Input v-model="e" placeholder="Pick one" /></Field>
            <Field label="Disabled"><Input v-model="d" disabled /></Field>
            <Field label="Read only"><Input v-model="ro" readonly /></Field>
            <Field label="Meal type"><Select v-model="sel" :options="opts" placeholder="Choose…" /></Field>
            <Field label="Meal type (set)"><Select v-model="sel2" :options="opts" /></Field>
            <Field label="Meal type (invalid)" error="Pick a meal"><Select v-model="sel" :options="opts" placeholder="Choose…" /></Field>
            <Field label="Meal type (disabled)"><Select v-model="sel2" :options="opts" disabled /></Field>
            <Field label="Notes" hint="Grows with the text"><Textarea v-model="ta" /></Field>
            <Field label="Notes (invalid)" error="Too long"><Textarea v-model="ta" /></Field>
            <div>
                <Checkbox v-model="c1" label="Unchecked" />
                <Checkbox v-model="c2" label="Checked" />
                <Checkbox v-model="c3" label="Indeterminate" />
                <Checkbox :model-value="true" label="Disabled" disabled />
                <Toggle v-model="tg" label="Toggle" />
            </div>
        </div>`,
    }),
};
