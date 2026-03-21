<template>
  <div class="file-explorer">
    <!-- Search bar -->
    <div class="search-bar">
      <t-input
        v-model="searchInput"
        placeholder="搜索当前目录（递归）…"
        clearable
        @enter="doSearch"
        @clear="cancelSearch"
      >
        <template #prefix-icon><t-icon name="search" /></template>
      </t-input>
      <t-button v-if="store.isSearching" variant="text" @click="cancelSearch"
        >取消</t-button
      >
    </div>

    <!-- Filter bar -->
    <div class="filter-bar">
      <t-select
        v-model="store.sortField"
        size="small"
        style="width: 110px"
        :options="sortOptions"
      />
      <t-button
        variant="text"
        shape="square"
        size="small"
        @click="toggleSortOrder"
        :title="store.sortOrder === 'desc' ? '降序' : '升序'"
      >
        <template #icon
          ><t-icon
            :name="store.sortOrder === 'desc' ? 'arrow-down' : 'arrow-up'"
        /></template>
      </t-button>
      <t-select
        v-model="store.filterType"
        size="small"
        style="width: 90px"
        :options="typeOptions"
      />
      <div style="flex: 1" />
      <div class="btn-group">
        <t-button
          :variant="store.viewMode === 'grid' ? 'base' : 'outline'"
          @click="store.viewMode = 'grid'"
          title="网格视图"
        >
          <template #icon><t-icon name="view-module" /></template>
        </t-button>
        <t-button
          :variant="store.viewMode === 'list' ? 'base' : 'outline'"
          @click="store.viewMode = 'list'"
          title="列表视图"
        >
          <template #icon><t-icon name="view-list" /></template>
        </t-button>
      </div>
    </div>

    <!-- Breadcrumb row -->
    <div class="breadcrumb-row">
      <t-button
        variant="text"
        shape="square"
        size="small"
        :disabled="store.currentPath === '.'"
        @click="store.navigateUp()"
        title="返回上级"
      >
        <template #icon><t-icon name="chevron-left" /></template>
      </t-button>
      <t-breadcrumb class="breadcrumb">
        <t-breadcrumb-item
          v-for="crumb in store.breadcrumbs"
          :key="crumb.path"
          @click="store.navigateTo(crumb.path)"
          style="cursor: pointer"
        >
          {{ crumb.label }}
        </t-breadcrumb-item>
      </t-breadcrumb>
      <t-button
        variant="text"
        shape="square"
        size="small"
        @click="store.refresh()"
        :loading="store.loading"
        title="刷新"
      >
        <template #icon><t-icon name="refresh" /></template>
      </t-button>
      <t-button
        size="small"
        variant="text"
        @click="showMkdir = true"
        title="新建文件夹"
      >
        <template #icon><t-icon name="folder-add" /></template>
      </t-button>
    </div>

    <!-- Error -->
    <t-alert
      v-if="store.error"
      theme="error"
      :message="store.error"
      style="margin-bottom: 8px"
    />

    <!-- Search status -->
    <div v-if="store.searchLoading" class="search-hint">
      <t-loading size="small" /><span
        >正在搜索 "{{ store.searchKeyword }}"…</span
      >
    </div>
    <div
      v-else-if="store.isSearching && !store.searchLoading"
      class="search-hint"
    >
      找到 {{ store.sortedItems.length }} 个结果（关键词: "{{
        store.searchKeyword
      }}"）
    </div>

    <!-- Content -->
    <t-loading
      :loading="store.loading && !store.isSearching"
      style="min-height: 120px"
    >
      <FileGridView
        v-if="store.viewMode === 'grid'"
        :items="store.sortedItems"
        :selected="selected"
        @open="handleOpen"
        @select="selected = $event"
        @preview="emit('preview', $event)"
        @download="handleDownload"
        @rename="startRename"
        @copy="startCopy"
        @move="startMove"
        @detail="openDetail"
        @delete="confirmDelete"
      />
      <FileListView
        v-else
        :items="store.sortedItems"
        :selected="selected"
        @open="handleOpen"
        @preview="emit('preview', $event)"
        @download="handleDownload"
        @rename="startRename"
        @copy="startCopy"
        @move="startMove"
        @detail="openDetail"
        @delete="confirmDelete"
      />
    </t-loading>
  </div>

  <!-- Rename dialog -->
  <t-dialog
    v-model:visible="renameVisible"
    header="重命名"
    confirm-btn="确认"
    cancel-btn="取消"
    :on-confirm="doRename"
  >
    <t-input v-model="renameName" autofocus @keyup.enter="doRename" />
  </t-dialog>

  <!-- Mkdir dialog -->
  <t-dialog
    v-model:visible="showMkdir"
    header="新建文件夹"
    confirm-btn="创建"
    cancel-btn="取消"
    :on-confirm="doMkdir"
  >
    <t-input
      v-model="mkdirName"
      placeholder="文件夹名称"
      autofocus
      @keyup.enter="doMkdir"
    />
  </t-dialog>

  <!-- Delete confirm -->
  <t-dialog
    v-model:visible="deleteVisible"
    header="确认删除"
    theme="danger"
    confirm-btn="删除"
    cancel-btn="取消"
    :on-confirm="doDelete"
  >
    确定要删除 <strong>{{ actionTarget?.name }}</strong> 吗？此操作不可恢复。
  </t-dialog>

  <!-- Folder picker -->
  <FolderPicker
    v-model="pickerVisible"
    :title="pickerMode === 'copy' ? '选择复制目标文件夹' : '选择移动目标文件夹'"
    :exclude-path="actionTarget?.isDir ? actionTarget.path : undefined"
    @select="onPickerSelect"
  />

  <!-- Detail drawer -->
  <FileDetailDrawer v-model="detailVisible" :item="actionTarget" />

  <!-- Panel stack (folder drill-down overlay) -->
  <template v-for="(panel, i) in panelStack" :key="panel.path + i">
    <FilePanel
      :visible="i === panelStack.length - 1"
      :root-path="panel.path"
      :view-mode="store.viewMode"
      @close="panelStack.splice(i)"
      @preview="emit('preview', $event)"
    />
  </template>
</template>

<script setup lang="ts">
import type { FileInfo } from "@/services/api";
import { fileApi } from "@/services/api";
import { useFileStore } from "@/stores/file";
import { useTransferStore } from "@/stores/transfer";
import { getCategory } from "@/utils/fileUtils";
import { MessagePlugin } from "tdesign-vue-next";
import { ref } from "vue";
import FileDetailDrawer from "./FileDetailDrawer.vue";
import FileGridView from "./FileGridView.vue";
import FileListView from "./FileListView.vue";
import FilePanel from "./FilePanel.vue";
import FolderPicker from "./FolderPicker.vue";

const emit = defineEmits<{ (e: "preview", item: FileInfo): void }>();

const store = useFileStore();
const transferStore = useTransferStore();

const searchInput = ref("");
function doSearch() {
  if (searchInput.value.trim()) store.search(searchInput.value.trim());
}
function cancelSearch() {
  store.clearSearch();
  searchInput.value = "";
}

const sortOptions = [
  { label: "修改时间", value: "modTime" },
  { label: "文件名", value: "name" },
  { label: "文件类型", value: "type" },
  { label: "文件大小", value: "size" },
];
const typeOptions = [
  { label: "全部", value: "all" },
  { label: "文件夹", value: "folder" },
  { label: "视频", value: "video" },
  { label: "音频", value: "audio" },
  { label: "图片", value: "image" },
  { label: "PDF", value: "pdf" },
  { label: "文本", value: "text" },
  { label: "压缩包", value: "archive" },
];

function toggleSortOrder() {
  store.sortOrder = store.sortOrder === "desc" ? "asc" : "desc";
}

const selected = ref<FileInfo | null>(null);

function handleOpen(item: FileInfo) {
  if (item.isDir) {
    panelStack.value.push({ path: item.path });
  } else if (
    ["video", "audio", "image", "pdf", "text"].includes(getCategory(item.ext))
  ) {
    emit("preview", item);
  } else {
    transferStore.addDownload(item.path, item.name, item.size);
  }
}
function handleDownload(item: FileInfo) {
  if (item.isDir) transferStore.addFolderDownload(item.path);
  else transferStore.addDownload(item.path, item.name, item.size);
}

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
    store.refresh();
  } catch {
    MessagePlugin.error("重命名失败");
  }
}

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
    MessagePlugin.error(pickerMode.value === "copy" ? "复制失败" : "移动失败");
  }
}

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

const detailVisible = ref(false);
function openDetail(item: FileInfo) {
  actionTarget.value = item;
  detailVisible.value = true;
}

const panelStack = ref<{ path: string }[]>([]);
</script>

<style scoped lang="scss">
.file-explorer {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  height: 100%;
  overflow: hidden;
}
.search-bar {
  display: flex;
  gap: 8px;
  align-items: center;
}
.filter-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
.breadcrumb-row {
  display: flex;
  align-items: center;
  gap: 4px;
  border-bottom: 1px solid var(--td-component-stroke);
  padding-bottom: 6px;
}
.breadcrumb {
  flex: 1;
  min-width: 0;
}
.search-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--td-text-color-secondary);
  padding: 4px;
}
</style>
