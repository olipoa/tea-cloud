<template>
  <div class="file-list">
    <div class="list-header">
      <span class="col-name">名称</span>
      <span class="col-date">修改时间</span>
    </div>

    <FileContextMenu
      v-for="item in items"
      :key="item.path"
      :item="item"
      @download="emit('download', item)"
      @rename="emit('rename', item)"
      @copy="emit('copy', item)"
      @move="emit('move', item)"
      @detail="emit('detail', item)"
      @delete="emit('delete', item)"
      @preview="emit('preview', item)"
    >
      <div
        class="list-row"
        :class="{ selected: selected?.path === item.path }"
        @click="emit('open', item)"
      >
        <span class="col-name">
          <t-icon
            :name="fileIconName(item.ext, item.isDir)"
            :class="['item-icon', fileIconColor(item.ext, item.isDir)]"
          />
          <span class="file-name" :title="item.name">{{ item.name }}</span>
        </span>
        <span class="col-date muted">{{ formatDate(item.modTime) }}</span>
      </div>
    </FileContextMenu>

    <div v-if="items.length === 0" class="list-empty">
      <t-icon name="folder-open" size="24px" />
      <span>空文件夹</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { FileInfo } from "@/services/api";
import { fileIconColor, fileIconName } from "@/utils/fileIcons";
import { formatDate } from "@/utils/fileUtils";
import FileContextMenu from "./FileContextMenu.vue";

defineProps<{ items: FileInfo[]; selected: FileInfo | null }>();

const emit = defineEmits<{
  (e: "open", item: FileInfo): void;
  (e: "preview", item: FileInfo): void;
  (e: "download", item: FileInfo): void;
  (e: "rename", item: FileInfo): void;
  (e: "copy", item: FileInfo): void;
  (e: "move", item: FileInfo): void;
  (e: "detail", item: FileInfo): void;
  (e: "delete", item: FileInfo): void;
}>();
</script>

<style scoped lang="scss">
.file-list {
  width: 100%;
}

.list-header {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  font-size: 12px;
  color: var(--td-text-color-placeholder);
  border-bottom: 1px solid var(--td-component-stroke);
}

.list-row {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  border-bottom: 1px solid var(--td-component-stroke);
  cursor: pointer;
  transition: background 0.12s;
  user-select: none;

  &:last-child {
    border-bottom: none;
  }
  &:hover {
    background: var(--td-bg-color-container-hover);
  }
  &.selected {
    background: var(--td-brand-color-light);
  }
}

.col-name {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.item-icon {
  flex-shrink: 0;
  &.color-folder {
    color: #f6a623;
  }
  &.color-video {
    color: #e5534b;
  }
  &.color-audio {
    color: #a371f7;
  }
  &.color-image {
    color: #3fb950;
  }
  &.color-pdf {
    color: #f85149;
  }
  &.color-text {
    color: #58a6ff;
  }
  &.color-archive {
    color: #d29922;
  }
  &.color-other {
    color: #8b949e;
  }
}

.file-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 14px;
  color: var(--td-text-color-primary);
}

.col-date {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--td-text-color-secondary);
  min-width: 120px;
  text-align: right;
}

.list-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 48px 0;
  color: var(--td-text-color-placeholder);
  font-size: 14px;
}
</style>
