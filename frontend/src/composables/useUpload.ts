import { useFileStore } from "@/stores/file";
import { useTransferStore } from "@/stores/transfer";
import { MessagePlugin } from "tdesign-vue-next";
import { ref } from "vue";

export function useUpload() {
  const fileStore = useFileStore();
  const transferStore = useTransferStore();

  const fileInput = ref<HTMLInputElement | null>(null);
  const folderInput = ref<HTMLInputElement | null>(null);
  const isDragover = ref(false);

  function triggerFileUpload() {
    fileInput.value?.click();
  }

  function triggerFolderUpload() {
    folderInput.value?.click();
  }

  function _addFiles(files: File[]) {
    const path = fileStore.currentPath;
    // Group files by webkitRelativePath prefix to detect folder uploads
    const hasFolderStructure = files.some((f) => f.webkitRelativePath);
    if (hasFolderStructure) {
      transferStore.addFolderUpload(files, path);
      MessagePlugin.success(`已添加文件夹上传任务（${files.length} 个文件）`);
    } else {
      for (const file of files) {
        transferStore.addUpload(file, path);
      }
      MessagePlugin.success(`已添加 ${files.length} 个上传任务`);
    }
  }

  function onFilesChange(e: Event) {
    const files = (e.target as HTMLInputElement).files;
    if (!files?.length) return;
    _addFiles(Array.from(files));
    (e.target as HTMLInputElement).value = "";
  }

  function handleDrop(e: DragEvent) {
    isDragover.value = false;
    const items = e.dataTransfer?.items;
    if (!items?.length) return;
    const files: File[] = [];
    for (let i = 0; i < items.length; i++) {
      if (items[i].kind === "file") {
        const f = items[i].getAsFile();
        if (f) files.push(f);
      }
    }
    if (files.length) _addFiles(files);
  }

  function handlePaste(e: ClipboardEvent) {
    const items = e.clipboardData?.items;
    if (!items?.length) return;
    const files: File[] = [];
    for (let i = 0; i < items.length; i++) {
      if (items[i].kind === "file") {
        const f = items[i].getAsFile();
        if (f) files.push(f);
      }
    }
    if (files.length) _addFiles(files);
  }

  return {
    fileInput,
    folderInput,
    isDragover,
    triggerFileUpload,
    triggerFolderUpload,
    onFilesChange,
    handleDrop,
    handlePaste,
  };
}
