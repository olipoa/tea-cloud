import type { FileInfo } from "@/services/api";
import { fileApi } from "@/services/api";
import { useFileStore } from "@/stores/file";
import { MessagePlugin } from "tdesign-vue-next";
import { ref } from "vue";

export function useFileActions() {
  const store = useFileStore();

  // Shared target for all actions
  const actionTarget = ref<FileInfo | null>(null);

  // ── Rename ─────────────────────────────────────────────────────────────────
  const renameVisible = ref(false);
  const renameName = ref("");

  function startRename(item: FileInfo) {
    actionTarget.value = item;
    renameName.value = item.name;
    renameVisible.value = true;
  }

  async function doRename() {
    if (!actionTarget.value || !renameName.value.trim()) return;
    try {
      await fileApi.rename(actionTarget.value.path, renameName.value.trim());
      MessagePlugin.success("重命名成功");
      renameVisible.value = false;
      store.refresh();
    } catch {
      MessagePlugin.error("重命名失败");
    }
  }

  // ── Mkdir ──────────────────────────────────────────────────────────────────
  const showMkdir = ref(false);
  const mkdirName = ref("");

  async function doMkdir() {
    const name = mkdirName.value.trim();
    if (!name) return;
    try {
      await store.createDir(name);
      MessagePlugin.success("文件夹创建成功");
      showMkdir.value = false;
      mkdirName.value = "";
    } catch {
      MessagePlugin.error("创建失败");
    }
  }

  // ── Copy / Move ────────────────────────────────────────────────────────────
  const pickerVisible = ref(false);
  const pickerMode = ref<"copy" | "move">("copy");

  function startCopy(item: FileInfo) {
    actionTarget.value = item;
    pickerMode.value = "copy";
    pickerVisible.value = true;
  }

  function startMove(item: FileInfo) {
    actionTarget.value = item;
    pickerMode.value = "move";
    pickerVisible.value = true;
  }

  async function onPickerSelect(destPath: string) {
    if (!actionTarget.value) return;
    try {
      if (pickerMode.value === "copy") {
        await fileApi.copy(actionTarget.value.path, destPath);
        MessagePlugin.success("复制成功");
      } else {
        await fileApi.move(actionTarget.value.path, destPath);
        MessagePlugin.success("移动成功");
      }
      store.refresh();
    } catch {
      MessagePlugin.error(
        pickerMode.value === "copy" ? "复制失败" : "移动失败",
      );
    }
  }

  // ── Delete ─────────────────────────────────────────────────────────────────
  const deleteVisible = ref(false);

  function confirmDelete(item: FileInfo) {
    actionTarget.value = item;
    deleteVisible.value = true;
  }

  async function doDelete() {
    if (!actionTarget.value) return;
    try {
      await store.deleteItem(actionTarget.value.path);
      MessagePlugin.success("删除成功");
      deleteVisible.value = false;
    } catch {
      MessagePlugin.error("删除失败");
    }
  }

  // ── Detail drawer ──────────────────────────────────────────────────────────
  const detailVisible = ref(false);

  function openDetail(item: FileInfo) {
    actionTarget.value = item;
    detailVisible.value = true;
  }

  return {
    actionTarget,
    // rename
    renameVisible,
    renameName,
    startRename,
    doRename,
    // mkdir
    showMkdir,
    mkdirName,
    doMkdir,
    // copy / move
    pickerVisible,
    pickerMode,
    startCopy,
    startMove,
    onPickerSelect,
    // delete
    deleteVisible,
    confirmDelete,
    doDelete,
    // detail
    detailVisible,
    openDetail,
  };
}
