<template>
  <!-- Panel overlay: slides in from right -->
  <Transition name="panel-slide">
    <div v-if="visible" class="file-panel">
      <!-- Panel header with breadcrumb -->
      <div class="panel-header">
        <t-button variant="text" shape="square" @click="emit('close')">
          <template #icon><t-icon name="chevron-left" /></template>
        </t-button>
        <t-breadcrumb class="panel-breadcrumb">
          <t-breadcrumb-item
            v-for="crumb in breadcrumbs"
            :key="crumb.path"
            @click="navigateTo(crumb.path)"
            style="cursor: pointer"
          >
            {{ crumb.label }}
          </t-breadcrumb-item>
        </t-breadcrumb>
        <t-button
          variant="text"
          shape="square"
          :loading="loading"
          @click="reload"
          title="刷新"
        >
          <template #icon><t-icon name="refresh" /></template>
        </t-button>
      </div>

      <!-- File content -->
      <div ref="scrollEl" class="panel-content" @scroll="onScroll">
        <t-loading :loading="loading" style="min-height: 80px">
          <FileGridView
            v-if="viewMode === 'grid'"
            :items="sortedItems"
            :selected="selected"
            @open="handleOpen"
            @select="selected = $event"
            @preview="(item: FileInfo) => emit('preview', item, sortedItems)"
            @download="handleDownload"
            @rename="startRename"
            @copy="startCopy"
            @move="startMove"
            @detail="openDetail"
            @delete="confirmDelete"
          />
          <FileListView
            v-else
            :items="sortedItems"
            :selected="selected"
            @open="handleOpen"
            @preview="(item: FileInfo) => emit('preview', item, sortedItems)"
            @download="handleDownload"
            @rename="startRename"
            @copy="startCopy"
            @move="startMove"
            @detail="openDetail"
            @delete="confirmDelete"
          />
        </t-loading>
      </div>
    </div>
  </Transition>

  <!-- Rename dialog -->
  <t-dialog
    v-model:visible="renameVisible"
    header="重命名"
    confirm-btn="确认"
    cancel-btn="取消"
    @confirm="doRename"
    @close="renameVisible = false"
  >
    <t-input v-model="renameName" autofocus @keyup.enter="doRename" />
  </t-dialog>

  <!-- Delete confirm -->
  <t-dialog
    v-model:visible="deleteVisible"
    header="确认删除"
    theme="danger"
    confirm-btn="删除"
    cancel-btn="取消"
    @confirm="doDelete"
  >
    确定要删除 <strong>{{ actionTarget?.name }}</strong> 吗？此操作不可恢复。
  </t-dialog>

  <!-- Folder picker for copy/move -->
  <FolderPicker
    v-model="pickerVisible"
    :title="pickerMode === 'copy' ? '选择复制目标文件夹' : '选择移动目标文件夹'"
    :exclude-path="actionTarget?.isDir ? actionTarget.path : undefined"
    @select="onPickerSelect"
  />

  <!-- File detail drawer -->
  <FileDetailDrawer v-model="detailVisible" :item="actionTarget" />
</template>

<script setup lang="ts">
import type { FileInfo } from "@/services/api";
import { fileApi } from "@/services/api";
import { type SortField, type SortOrder } from "@/stores/file";
import { useTransferStore } from "@/stores/transfer";
import { getCategory } from "@/utils/fileUtils";
import { MessagePlugin } from "tdesign-vue-next";
import { computed, nextTick, ref, watch } from "vue";
import FileDetailDrawer from "./FileDetailDrawer.vue";
import FileGridView from "./FileGridView.vue";
import FileListView from "./FileListView.vue";
import FolderPicker from "./FolderPicker.vue";

const props = defineProps<{
  visible: boolean;
  rootPath: string;
  viewMode: "grid" | "list";
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "preview", item: FileInfo, siblings: FileInfo[]): void;
}>();

const transferStore = useTransferStore();

const currentPath = ref(props.rootPath);
const items = ref<FileInfo[]>([]);
const loading = ref(false);
const selected = ref<FileInfo | null>(null);

// Sort state (local to panel)
const sortField = ref<SortField>("modTime");
const sortOrder = ref<SortOrder>("desc");

const sortedItems = computed(() => {
  const list = [...items.value];
  list.sort((a, b) => {
    if (a.isDir && !b.isDir) return -1;
    if (!a.isDir && b.isDir) return 1;
    let cmp = 0;
    switch (sortField.value) {
      case "name":
        cmp = a.name.localeCompare(b.name);
        break;
      case "size":
        cmp = a.size - b.size;
        break;
      case "type":
        cmp = (a.ext || "").localeCompare(b.ext || "");
        break;
      case "modTime":
        cmp = a.modTime - b.modTime;
        break;
    }
    return sortOrder.value === "asc" ? cmp : -cmp;
  });
  return list;
});

const breadcrumbs = computed(() => {
  const root = props.rootPath;
  const cur = currentPath.value;
  if (cur === root || cur === ".")
    return [
      { label: root === "." ? "根目录" : root.split("/").pop()!, path: root },
    ];
  const rel = cur.startsWith(root + "/") ? cur.slice(root.length + 1) : cur;
  const parts = rel.split("/");
  const crumbs: { label: string; path: string }[] = [
    { label: root === "." ? "根目录" : root.split("/").pop()!, path: root },
  ];
  let acc = root;
  for (const part of parts) {
    if (!part) continue;
    acc = acc === "." ? part : `${acc}/${part}`;
    crumbs.push({ label: part, path: acc });
  }
  return crumbs;
});

// Scroll position memory
const scrollEl = ref<HTMLElement | null>(null);
const scrollPositions = new Map<string, number>();

function onScroll() {
  if (scrollEl.value)
    scrollPositions.set(currentPath.value, scrollEl.value.scrollTop);
}

async function loadDir(path: string) {
  loading.value = true;
  try {
    items.value = await fileApi.list(path);
    currentPath.value = path;
    await nextTick();
    if (scrollEl.value) {
      scrollEl.value.scrollTop = scrollPositions.get(path) ?? 0;
    }
  } catch {
    items.value = [];
  } finally {
    loading.value = false;
  }
}

function navigateTo(path: string) {
  if (scrollEl.value)
    scrollPositions.set(currentPath.value, scrollEl.value.scrollTop);
  loadDir(path);
}

function reload() {
  loadDir(currentPath.value);
}

// Open file/folder
function handleOpen(item: FileInfo) {
  if (item.isDir) {
    navigateTo(item.path);
  } else if (
    ["video", "audio", "image", "pdf", "text"].includes(getCategory(item.ext))
  ) {
    emit("preview", item, sortedItems.value);
  } else {
    transferStore.addDownload(item.path, item.name, item.size);
  }
}

function handleDownload(item: FileInfo) {
  if (item.isDir) {
    transferStore.addFolderDownload(item.path);
  } else {
    transferStore.addDownload(item.path, item.name, item.size);
  }
}

// Rename
const renameVisible = ref(false);
const renameName = ref("");
const actionTarget = ref<FileInfo | null>(null);

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
    reload();
  } catch {
    MessagePlugin.error("重命名失败");
  }
}

// Copy / Move
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
    reload();
  } catch {
    MessagePlugin.error(pickerMode.value === "copy" ? "复制失败" : "移动失败");
  }
}

// Delete
const deleteVisible = ref(false);

function confirmDelete(item: FileInfo) {
  actionTarget.value = item;
  deleteVisible.value = true;
}

async function doDelete() {
  if (!actionTarget.value) return;
  try {
    await fileApi.delete(actionTarget.value.path);
    MessagePlugin.success("删除成功");
    deleteVisible.value = false;
    reload();
  } catch {
    MessagePlugin.error("删除失败");
  }
}

// Detail
const detailVisible = ref(false);

function openDetail(item: FileInfo) {
  actionTarget.value = item;
  detailVisible.value = true;
}

// Load on open
watch(
  () => props.visible,
  (v) => {
    if (v) loadDir(props.rootPath);
  },
  { immediate: true },
);
</script>

<style scoped lang="scss">
.file-panel {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: 100%;
  max-width: 100vw;
  background: var(--td-bg-color-page);
  z-index: 300;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--td-component-stroke);
  background: var(--td-bg-color-container);
  flex-shrink: 0;
}

.panel-breadcrumb {
  flex: 1;
  min-width: 0;
}

.panel-content {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

/* Slide animation from right */
.panel-slide-enter-active,
.panel-slide-leave-active {
  transition: transform 0.28s cubic-bezier(0.4, 0, 0.2, 1);
}
.panel-slide-enter-from,
.panel-slide-leave-to {
  transform: translateX(100%);
}
</style>
