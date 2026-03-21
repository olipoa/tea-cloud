<template>
  <n-modal
    v-model:show="visible"
    :title="item?.name"
    preset="card"
    style="max-width: 92vw; width: 900px;"
    :auto-focus="false"
    @after-leave="emit('close')"
  >
    <template v-if="item">
      <ImageViewer v-if="category === 'image'" :url="rawUrl" @close="visible = false" />
      <PdfViewer v-else-if="category === 'pdf'" :url="rawUrl" />
      <TextViewer v-else-if="category === 'text'" :url="rawUrl" />
    </template>

    <template #footer>
      <div class="preview-footer">
        <span class="file-meta" v-if="item">
          {{ formatSize(item.size) }} · {{ formatDate(item.modTime) }}
        </span>
        <div class="footer-actions">
          <n-button @click="visible = false">关闭</n-button>
          <n-button type="primary" @click="download">
            <template #icon><span>⬇</span></template>
            下载
          </n-button>
        </div>
      </div>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NModal, NButton } from 'naive-ui'
import { type FileInfo, fileApi } from '@/services/api'
import { getCategory, formatSize, formatDate } from '@/utils/fileUtils'
import { useDownload } from '@/composables/useDownload'
import ImageViewer from './ImageViewer.vue'
import PdfViewer from './PdfViewer.vue'
import TextViewer from './TextViewer.vue'

const props = defineProps<{ item: FileInfo | null }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const visible = ref(false)

watch(() => props.item, (val) => { visible.value = !!val })

const category = computed(() => props.item ? getCategory(props.item.ext) : 'other')
const rawUrl = computed(() => props.item ? fileApi.rawUrl(props.item.path) : '')

function download() {
  if (!props.item) return
  const { download: dl } = useDownload()
  dl(props.item.path, props.item.name)
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
