import { dialogManager } from "~/composables/dialog/useDialog";
import AddListDialog from "./AddListDialog.vue";

/** Opens the "paste a list" dialog. Resolves true if items were added. */
export function openAddListDialog() {
    return dialogManager.custom<boolean>({
        title: "Add from text",
        component: AddListDialog,
        wide: true,
    });
}
