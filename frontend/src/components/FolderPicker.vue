<template>
  <t-dialog
    v-model:visible="visible"
    :title="title"
    width="480px"
    @confirm="onConfirm"
    @close="onClose"
    confirm-btn="确认"
    cancel-btn="取消"
    :confirm-btn-props="{ disabled: !selectedPath }"
  >
    <div class="folder-picker">
      <!-- Breadcrumb navigation -->
      <t-breadcrumb class="picker-breadcrumb">
        <t-breadcrumb-item
          v-for="crumb in breadcrumbs"
          :key="crumb.path"
          @click="navigateTo(crumb.path)"
          style="cursor: pointer"
        >
          {{ crumb.label }}
        </t-breadcrumb-item>
      </t-breadcrumb>

      <!-- Folder list -->
      <div class="folder-list" v-if="!loading">
        <!-- Current folder as selectable option -->
        <div
          class="folder-item"
          :class="{ selected: selectedPath === currentPath }"
          @click="selectedPath = currentPath"
          @dblclick="selectedPath = currentPath"
        >
          <t-icon name="folder" class="item-icon" />
          <span>（当前目录）</span>
          <t-icon
            v-if="selectedPath === currentPath"
            name="check"
            class="check-icon"
          />
        </div>

        <div
          v-for="item in folders"
          :key="item.path"
          class="folder-item"
          :class="{ selected: selectedPath === item.path }"
          @click="selectedPath = item.path"
          @dblclick="navigateTo(item.path)"
        >
          <t-icon name="folder" class="item-icon" />
          <span class="item-name" :title="item.name">{{ item.name }}</span>
          <t-icon
            name="chevron-right"
            class="arrow-icon"
            @click.stop="navigateTo(item.path)"
          />
          <t-icon
            v-if="selectedPath === item.path"
            name="check"
            class="check-icon"
          />
        </div>

        <div v-if="folders.length === 0" class="empty-hint">
          <t-icon name="folder-open" />&nbsp;无子文件夹
        </div>
      </div>
      <div v-else class="loading-hint">
        <t-loading size="small" />
      </div>
    </div>
  </t-dialog>
</template>

<script setup lang="ts">
import { fileApi, type FileInfo } from "@/services/api";
import { computed, ref, watch } from "vue";

const props = defineProps<{
  modelValue: boolean;
  title?: string;
  excludePath?: string; // exclude this path from selection (e.g., the source path)
}>();

const emit = defineEmits<{
  (e: "update:modelValue", val: boolean): void;
  (e: "select", path: string): void;
}>();

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit("update:modelValue", v),
});

const currentPath = ref(".");
const items = ref<FileInfo[]>([]);
const loading = ref(false);
const selectedPath = ref<string | null>(null);

const folders = computed(() =>
  items.value.filter((i) => i.isDir && i.path !== props.excludePath),
);

const breadcrumbs = computed(() => {
  if (currentPath.value === ".") return [{ label: "根目录", path: "." }];
  const parts = currentPath.value.split("/");
  const crumbs: { label: string; path: string }[] = [
    { label: "根目录", path: "." },
  ];
  let acc = "";
  for (const part of parts) {
    if (!part || part === ".") continue;
    acc = acc ? `${acc}/${part}` : part;
    crumbs.push({ label: part, path: acc });
  }
  return crumbs;
});

async function navigateTo(path: string) {
  loading.value = true;
  currentPath.value = path;
  selectedPath.value = path;
  try {
    items.value = await fileApi.list(path);
  } catch {
    items.value = [];
  } finally {
    loading.value = false;
  }
}

watch(visible, (v) => {
  if (v) {
    currentPath.value = ".";
    selectedPath.value = ".";
    navigateTo(".");
  }
});

function onConfirm() {
  if (selectedPath.value) {
    emit("select", selectedPath.value);
    emit("update:modelValue", false);
  }
}

function onClose() {
  emit("update:modelValue", false);
}
</script>

<style scoped lang="scss">
.folder-picker {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 200px;
}

.picker-breadcrumb {
  flex-shrink: 0;
}

.folder-list {
  flex: 1;
  overflow-y: auto;
  max-height: 320px;
  border: 1px solid var(--td-component-stroke);
  border-radius: 6px;
}

.folder-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  cursor: pointer;
  border-bottom: 1px solid var(--td-component-stroke);
  transition: background 0.15s;

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

.item-icon {
  color: var(--td-warning-color);
  flex-shrink: 0;
}

.item-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.arrow-icon {
  flex-shrink: 0;
  color: var(--td-text-color-placeholder);
  &:hover {
    color: var(--td-brand-color);
  }
}

.check-icon {
  flex-shrink: 0;
  color: var(--td-success-color);
}

.empty-hint,
.loading-hint {
  padding: 24px;
  text-align: center;
  color: var(--td-text-color-placeholder);
  font-size: 13px;
}
</style>
