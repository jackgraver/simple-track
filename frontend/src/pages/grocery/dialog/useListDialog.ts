import { dialogManager } from "~/composables/dialog/useDialog";
import AddListDialog from "./AddListDialog.vue";

// List Dialog Handler function 
export default async function addTextList() {
    const res = await dialogManager.custom<boolean>({
        title: "Add to List",
        component: AddListDialog,
    });
    console.log('dialog finished', res);
    return res;
}
