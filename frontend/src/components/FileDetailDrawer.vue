<template>
  <t-drawer
    v-model:visible="visible"
    :header="item?.name ?? '文件详情'"
    size="medium"
    placement="right"
    :footer="false"
  >
    <div v-if="item" class="detail-body">
      <!-- Icon -->
      <div class="detail-icon-row">
        <t-icon
          :name="fileIconName(item.ext, item.isDir)"
          :class="fileIconColor(item.ext, item.isDir)"
          size="48px"
        />
      </div>

      <t-divider />

      <t-descriptions :column="1" bordered>
        <t-descriptions-item label="名称">{{ item.name }}</t-descriptions-item>
        <t-descriptions-item label="类型">
          {{ item.isDir ? "文件夹" : item.mime || item.ext || "未知" }}
        </t-descriptions-item>
        <t-descriptions-item v-if="!item.isDir" label="大小">
          {{ formatSize(item.size) }}
        </t-descriptions-item>
        <t-descriptions-item
          v-if="item.isDir && childFiles !== null"
          label="包含文件"
        >
          {{ childFiles }} 个
        </t-descriptions-item>
        <t-descriptions-item
          v-if="item.isDir && childDirs !== null"
          label="包含文件夹"
        >
          {{ childDirs }} 个
        </t-descriptions-item>
        <t-descriptions-item label="修改时间">
          {{ formatDate(item.modTime) }}
        </t-descriptions-item>
        <t-descriptions-item label="相对路径">
          <span class="mono">{{ item.path }}</span>
        </t-descriptions-item>
      </t-descriptions>
    </div>
  </t-drawer>
</template>

<script setup lang="ts">
import { type FileInfo, fileApi } from "@/services/api";
import { fileIconColor, fileIconName } from "@/utils/fileIcons";
import { formatDate, formatSize } from "@/utils/fileUtils";
import { computed, ref, watch } from "vue";

const props = defineProps<{
  modelValue: boolean;
  item: FileInfo | null;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", val: boolean): void;
}>();

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit("update:modelValue", v),
});

const childFiles = ref<number | null>(null);
const childDirs = ref<number | null>(null);

watch(
  () => props.item,
  async (item) => {
    childFiles.value = null;
    childDirs.value = null;
    if (item?.isDir) {
      try {
        const children = await fileApi.list(item.path);
        childFiles.value = children.filter((c) => !c.isDir).length;
        childDirs.value = children.filter((c) => c.isDir).length;
      } catch {
        /* ignore */
      }
    }
  },
);
</script>

<style scoped lang="scss">
.detail-body {
  padding: 8px 0;
}

.detail-icon-row {
  display: flex;
  justify-content: center;
  padding: 16px 0;

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

.mono {
  font-family: monospace;
  font-size: 12px;
  word-break: break-all;
}
</style>
