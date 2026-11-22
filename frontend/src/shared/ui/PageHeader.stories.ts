import type { Meta, StoryObj } from "@storybook/vue3-vite";
import Button from "./Button.vue";
import Card from "./Card.vue";
import PageHeader from "./PageHeader.vue";
import Section from "./Section.vue";

const meta = {
    title: "UI/PageHeader and Section",
    component: PageHeader,
    tags: ["autodocs"],
    parameters: { layout: "padded" },
    args: { title: "Water sizes & goal" },
} satisfies Meta<typeof PageHeader>;
export default meta;
type Story = StoryObj<typeof meta>;

export const TitleOnly: Story = {};
export const WithActionsAndMeta: Story = {
    render: (args) => ({
        components: { PageHeader, Button },
        setup: () => ({ args }),
        template: `<PageHeader v-bind="args"><template #meta>Updated today</template><template #actions><Button variant="ghost" size="sm">Back</Button><Button variant="primary" size="sm">Add size</Button></template></PageHeader>`,
    }),
};

/** A page: header, then `gap-6` between sections. Below lg the actions stack under the title. */
export const PageWithSections: Story = {
    render: (args) => ({
        components: { PageHeader, Section, Card, Button },
        setup: () => ({ args }),
        template: `
        <div class="mx-auto flex max-w-3xl flex-col gap-6">
            <PageHeader v-bind="args"><template #actions><Button variant="ghost" size="sm">Back to diet</Button></template></PageHeader>
            <Section title="Display"><Card><p class="m-0 text-sm text-fg-muted">Section content goes in a Card, a List or a form.</p></Card></Section>
            <Section title="Saved sizes"><template #actions><Button size="sm" variant="secondary">Add</Button></template><Card><p class="m-0 text-sm text-fg-muted">A section with an action.</p></Card></Section>
        </div>`,
    }),
};

/** `inline`: one icon action (a ⋯ menu) stays beside the title on mobile too. */
export const InlineIconAction: Story = {
    args: { title: "Push", inline: true },
    render: (args) => ({
        components: { PageHeader, Button },
        setup: () => ({ args }),
        template: `<PageHeader v-bind="args"><template #meta>Push Pull Legs · 4 exercises</template><template #actions><Button variant="ghost" size="sm" aria-label="Routine actions">⋯</Button></template></PageHeader>`,
    }),
};
