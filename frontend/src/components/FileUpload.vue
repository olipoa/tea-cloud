<!-- FileUpload: drag-drop trigger, delegates to transferStore -->
<template>
  <div
    class="upload-zone"
    :class="{ dragover: isDragover }"
    @dragover.prevent="isDragover = true"
    @dragleave.prevent="isDragover = false"
    @drop.prevent="handleDrop"
  >
    <input
      ref="fileInput"
      type="file"
      multiple
      style="display: none"
      @change="onFilesChange"
    />
    <input
      ref="folderInput"
      type="file"
      webkitdirectory
      multiple
      style="display: none"
      @change="onFolderChange"
    />
    <t-space size="small" align="center">
      <t-button size="small" variant="outline" @click="fileInput?.click()">
        <template #icon><t-icon name="upload" /></template>
        上传文件
      </t-button>
      <t-button size="small" variant="outline" @click="folderInput?.click()">
        <template #icon><t-icon name="folder-open" /></template>
        上传文件夹
      </t-button>
      <span v-if="isDragover" class="drag-hint">松开即可上传</span>
    </t-space>
  </div>
</template>

<script setup lang="ts">
import { useFileStore } from "@/stores/file";
import { useTransferStore } from "@/stores/transfer";
import { MessagePlugin } from "tdesign-vue-next";
import { ref } from "vue";

const fileStore = useFileStore();
const transferStore = useTransferStore();

const fileInput = ref<HTMLInputElement | null>(null);
const folderInput = ref<HTMLInputElement | null>(null);
const isDragover = ref(false);

function onFilesChange(e: Event) {
  const files = (e.target as HTMLInputElement).files;
  if (!files?.length) return;
  const path = fileStore.currentPath;
  for (const file of Array.from(files)) transferStore.addUpload(file, path);
  (e.target as HTMLInputElement).value = "";
  MessagePlugin.success(`已添加 ${files.length} 个上传任务`);
}

function onFolderChange(e: Event) {
  const files = (e.target as HTMLInputElement).files;
  if (!files?.length) return;
  transferStore.addFolderUpload(Array.from(files), fileStore.currentPath);
  (e.target as HTMLInputElement).value = "";
  MessagePlugin.success(`已添加文件夹上传任务（${files.length} 个文件）`);
}

function handleDrop(e: DragEvent) {
  isDragover.value = false;
  const items = e.dataTransfer?.items;
  if (!items?.length) return;
  const path = fileStore.currentPath;
  const files: File[] = [];
  for (let i = 0; i < items.length; i++) {
    if (items[i].kind === "file") {
      const f = items[i].getAsFile();
      if (f) files.push(f);
    }
  }
  if (files.length) {
    for (const f of files) transferStore.addUpload(f, path);
    MessagePlugin.success(`已添加 ${files.length} 个上传任务`);
  }
}
</script>

<style scoped lang="scss">
.upload-zone {
  padding: 8px 12px;
  border: 1.5px dashed var(--td-component-stroke);
  border-radius: var(--td-radius-default);
  transition: border-color 0.2s;
  &.dragover {
    border-color: var(--td-brand-color);
    background: var(--td-brand-color-light);
  }
}
.drag-hint {
  font-size: 12px;
  color: var(--td-brand-color);
}
</style>
