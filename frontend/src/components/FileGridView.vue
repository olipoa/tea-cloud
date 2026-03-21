<template>
  <div class="file-grid">
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
        class="grid-item"
        :class="{ selected: selected?.path === item.path }"
        @click="emit('select', item)"
        @dblclick="emit('open', item)"
      >
        <!-- Thumbnail or icon -->
        <div class="grid-thumb-wrap">
          <img
            v-if="shouldShowThumb(item)"
            class="grid-thumb"
            loading="lazy"
            :src="thumbUrl(item)"
            @error="onThumbError(item.path)"
            alt=""
          />
          <t-icon
            v-else
            :name="fileIconName(item.ext, item.isDir)"
            :class="fileIconColor(item.ext, item.isDir)"
            size="36px"
          />
        </div>

        <div class="grid-name" :title="item.name">{{ item.name }}</div>
        <div class="grid-date muted">{{ formatDate(item.modTime) }}</div>
      </div>
    </FileContextMenu>

    <div v-if="items.length === 0" class="grid-empty">
      <t-icon name="folder-open" size="32px" />
      <span>空文件夹</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { FileInfo } from "@/services/api";
import { fileApi } from "@/services/api";
import { fileIconColor, fileIconName } from "@/utils/fileIcons";
import { formatDate, getCategory } from "@/utils/fileUtils";
import { reactive } from "vue";
import FileContextMenu from "./FileContextMenu.vue";

defineProps<{ items: FileInfo[]; selected: FileInfo | null }>();

const emit = defineEmits<{
  (e: "open", item: FileInfo): void;
  (e: "select", item: FileInfo): void;
  (e: "preview", item: FileInfo): void;
  (e: "download", item: FileInfo): void;
  (e: "rename", item: FileInfo): void;
  (e: "copy", item: FileInfo): void;
  (e: "move", item: FileInfo): void;
  (e: "detail", item: FileInfo): void;
  (e: "delete", item: FileInfo): void;
}>();

const failedThumbs = reactive(new Set<string>());

function shouldShowThumb(item: FileInfo): boolean {
  if (item.isDir) return false;
  if (failedThumbs.has(item.path)) return false;
  const cat = getCategory(item.ext);
  return cat === "video" || cat === "image";
}

function thumbUrl(item: FileInfo): string {
  return getCategory(item.ext) === "video"
    ? fileApi.thumbnailUrl(item.path)
    : fileApi.rawUrl(item.path);
}

function onThumbError(path: string) {
  failedThumbs.add(path);
}
</script>

<style scoped lang="scss">
.file-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(110px, 1fr));
  gap: 8px;
  padding: 4px 0;
}

.grid-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px 6px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s;
  user-select: none;

  &:hover {
    background: var(--td-bg-color-container-hover);
  }
  &.selected {
    background: var(--td-brand-color-light);
  }
}

.grid-thumb-wrap {
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  overflow: hidden;
  flex-shrink: 0;

  :deep(.color-folder) {
    color: #f6a623;
  }
  :deep(.color-video) {
    color: #e5534b;
  }
  :deep(.color-audio) {
    color: #a371f7;
  }
  :deep(.color-image) {
    color: #3fb950;
  }
  :deep(.color-pdf) {
    color: #f85149;
  }
  :deep(.color-text) {
    color: #58a6ff;
  }
  :deep(.color-archive) {
    color: #d29922;
  }
  :deep(.color-other) {
    color: #8b949e;
  }
}

.grid-thumb {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.grid-name {
  font-size: 12px;
  text-align: center;
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--td-text-color-primary);
}

.grid-date {
  font-size: 10px;
  color: var(--td-text-color-placeholder);
}

.grid-empty {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 48px 0;
  color: var(--td-text-color-placeholder);
  font-size: 14px;
}
</style>
