import { dialogManager } from "~/composables/dialog/useDialog";
import ImportWeightDialog from "./ImportWeightDialog.vue";

/** Opens the "paste Renpho CSV" dialog. Resolves true if rows were imported. */
export function openImportWeightDialog() {
    return dialogManager.custom<boolean>({
        title: "Import from Renpho",
        component: ImportWeightDialog,
        wide: true,
    });
}
