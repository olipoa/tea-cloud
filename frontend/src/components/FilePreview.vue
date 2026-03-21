<template>
  <el-dialog
    v-model="visible"
    :title="item?.name"
    width="90%"
    :max-width="'900px'"
    destroy-on-close
    class="preview-dialog"
    @closed="emit('close')"
  >
    <template v-if="item">
      <VideoPlayer v-if="category === 'video'" :url="rawUrl" />
      <AudioPlayer v-else-if="category === 'audio'" :url="rawUrl" :filename="item.name" />
      <ImageViewer v-else-if="category === 'image'" :url="rawUrl" @close="visible = false" />
      <PdfViewer v-else-if="category === 'pdf'" :url="rawUrl" />
      <TextViewer v-else-if="category === 'text'" :url="rawUrl" />
    </template>

    <template #footer>
      <div class="preview-footer">
        <span class="file-meta" v-if="item">
          {{ formatSize(item.size) }} · {{ formatDate(item.modTime) }}
        </span>
        <div class="footer-actions">
          <el-button @click="visible = false">关闭</el-button>
          <el-button type="primary" :icon="Download" @click="download">下载</el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Download } from '@element-plus/icons-vue'
import { type FileInfo, fileApi } from '@/services/api'
import { getCategory, formatSize, formatDate } from '@/utils/fileUtils'
import VideoPlayer from './VideoPlayer.vue'
import AudioPlayer from './AudioPlayer.vue'
import ImageViewer from './ImageViewer.vue'
import PdfViewer from './PdfViewer.vue'
import TextViewer from './TextViewer.vue'

const props = defineProps<{ item: FileInfo | null }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const visible = ref(false)

watch(() => props.item, (val) => {
  visible.value = !!val
})

const category = computed(() => props.item ? getCategory(props.item.ext) : 'other')
const rawUrl = computed(() => props.item ? fileApi.rawUrl(props.item.path) : '')

function download() {
  if (!props.item) return
  const a = document.createElement('a')
  a.href = fileApi.downloadUrl(props.item.path)
  a.download = props.item.name
  a.click()
}
</script>

<style lang="scss">
.preview-dialog {
  .el-dialog__body {
    padding: 0;
    overflow: hidden;
  }
}
</style>

<style scoped lang="scss">
.preview-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.file-meta {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.footer-actions {
  display: flex;
  gap: 8px;
}
</style>
