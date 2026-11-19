import { inject, provide, type InjectionKey } from "vue";

// Field provides the control id, the aria-describedby ids and the invalid flag, so Input,
// Textarea and Select wire up their label and error without the caller passing ids around.
export type FieldContext = { id: string; describedBy: () => string | undefined; invalid: () => boolean };
const key: InjectionKey<FieldContext> = Symbol("ui-field");

export function provideField(ctx: FieldContext) {
    provide(key, ctx);
}
export function useField(): FieldContext | undefined {
    return inject(key, undefined);
}
