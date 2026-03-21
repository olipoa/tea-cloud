<template>
  <t-dialog
    v-model:visible="visible"
    :title="item?.name"
    style="max-width: 92vw; width: 900px"
    :on-close="() => emit('close')"
  >
    <template v-if="item">
      <PdfViewer v-if="category === 'pdf'" :url="rawUrl" />
      <TextViewer v-else-if="category === 'text'" :url="rawUrl" />
    </template>

    <template #footer>
      <div class="preview-footer">
        <span class="file-meta" v-if="item">
          {{ formatSize(item.size) }} · {{ formatDate(item.modTime) }}
        </span>
        <div class="footer-actions">
          <t-button @click="visible = false">关闭</t-button>
          <t-button theme="primary" @click="download">
            <template #icon><span>⬇</span></template>
            下载
          </t-button>
        </div>
      </div>
    </template>
  </t-dialog>
</template>

<script setup lang="ts">
import { useDownload } from "@/composables/useDownload";
import { type FileInfo, fileApi } from "@/services/api";
import { formatDate, formatSize, getCategory } from "@/utils/fileUtils";

import { computed, ref, watch } from "vue";
import PdfViewer from "./PdfViewer.vue";
import TextViewer from "./TextViewer.vue";

const props = defineProps<{ item: FileInfo | null }>();
const emit = defineEmits<{ (e: "close"): void }>();

const visible = ref(false);

watch(
  () => props.item,
  (val) => {
    visible.value = !!val;
  },
);

const category = computed(() =>
  props.item ? getCategory(props.item.ext) : "other",
);
const rawUrl = computed(() =>
  props.item ? fileApi.rawUrl(props.item.path) : "",
);

function download() {
  if (!props.item) return;
  const { download: dl } = useDownload();
  dl(props.item.path, props.item.name);
}
</script>

<style scoped lang="scss">
.preview-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.file-meta {
  font-size: 12px;
  color: #aaa;
}

.footer-actions {
  display: flex;
  gap: 8px;
}
</style>
