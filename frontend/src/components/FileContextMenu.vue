<template>
  <!-- PC: t-dropdown context menu via right-click -->
  <t-dropdown
    v-if="!isMobile"
    :options="menuOptions"
    trigger="context-menu"
    @click="onMenuClick"
  >
    <slot />
  </t-dropdown>

  <!-- Mobile: wrap slot, emit long-press via drawer -->
  <template v-else>
    <div
      @touchstart="onTouchStart"
      @touchend="onTouchEnd"
      @touchmove="onTouchMove"
      @touchcancel="onTouchEnd"
    >
      <slot />
    </div>

    <!-- Bottom action sheet (drawer) -->
    <t-drawer
      v-model:visible="sheetVisible"
      placement="bottom"
      :footer="false"
      :header="props.item?.name ?? ''"
      size="auto"
    >
      <div class="action-list">
        <div
          v-for="it in sheetItems"
          :key="it.value"
          class="action-item"
          :class="{ danger: it.value === 'delete' }"
          @click="onSheetSelect(it.value)"
        >
          {{ it.label }}
        </div>
      </div>
    </t-drawer>
  </template>
</template>

<script setup lang="ts">
import type { FileInfo } from "@/services/api";
import { getCategory } from "@/utils/fileUtils";
import type { DropdownOption } from "tdesign-vue-next";
import { computed, ref } from "vue";

const props = defineProps<{
  item: FileInfo | null;
}>();

const emit = defineEmits<{
  (e: "download"): void;
  (e: "rename"): void;
  (e: "copy"): void;
  (e: "move"): void;
  (e: "detail"): void;
  (e: "delete"): void;
  (e: "preview"): void;
}>();

const isMobile = window.innerWidth < 768;

// Whether preview is supported
const canPreviewFile = computed(() => {
  if (!props.item || props.item.isDir) return false;
  const cat = getCategory(props.item.ext);
  return ["video", "audio", "image", "pdf", "text"].includes(cat);
});

const menuOptions = computed<DropdownOption[]>(() => {
  const opts: DropdownOption[] = [];
  if (canPreviewFile.value)
    opts.push({ content: "预览", value: "preview", prefixIcon: () => null });
  opts.push(
    { content: "下载", value: "download", prefixIcon: () => null },
    { content: "重命名", value: "rename", prefixIcon: () => null },
    { content: "复制到...", value: "copy", prefixIcon: () => null },
    { content: "移动到...", value: "move", prefixIcon: () => null },
    { content: "文件详情", value: "detail", prefixIcon: () => null },
    { divider: true, value: "divider" } as DropdownOption,
    {
      content: "删除",
      value: "delete",
      theme: "error",
      prefixIcon: () => null,
    },
  );
  return opts;
});

const sheetItems = computed(() => {
  const items: { label: string; value: string }[] = [];
  if (canPreviewFile.value) items.push({ label: "预览", value: "preview" });
  items.push(
    { label: "下载", value: "download" },
    { label: "重命名", value: "rename" },
    { label: "复制到...", value: "copy" },
    { label: "移动到...", value: "move" },
    { label: "文件详情", value: "detail" },
    { label: "删除", value: "delete" },
  );
  return items;
});

const sheetVisible = ref(false);
let longPressTimer: ReturnType<typeof setTimeout> | null = null;
let touchMoved = false;

function onTouchStart() {
  touchMoved = false;
  longPressTimer = setTimeout(() => {
    if (!touchMoved) sheetVisible.value = true;
  }, 500);
}

function onTouchEnd() {
  if (longPressTimer) clearTimeout(longPressTimer);
}

function onTouchMove() {
  touchMoved = true;
  if (longPressTimer) clearTimeout(longPressTimer);
}

function onMenuClick(data: { value: string }) {
  dispatchAction(data.value);
}

function onSheetSelect(value: string) {
  sheetVisible.value = false;
  dispatchAction(value);
}

function dispatchAction(action: string) {
  const map: Record<string, () => void> = {
    download: () => emit("download"),
    rename: () => emit("rename"),
    copy: () => emit("copy"),
    move: () => emit("move"),
    detail: () => emit("detail"),
    delete: () => emit("delete"),
    preview: () => emit("preview"),
  };
  map[action]?.();
}
</script>

<style scoped lang="scss">
.action-list {
  display: flex;
  flex-direction: column;
  padding: 8px 0 16px;
}
.action-item {
  padding: 14px 20px;
  font-size: 16px;
  cursor: pointer;
  &:hover {
    background: var(--td-bg-color-secondarycontainer);
  }
  &.danger {
    color: var(--td-error-color);
  }
}
</style>
