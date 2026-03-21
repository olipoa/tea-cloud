<template>
  <div
    class="upload-zone"
    :class="{ dragover: isDragover, uploading: isUploading }"
    @dragover.prevent="isDragover = true"
    @dragleave.prevent="isDragover = false"
    @drop.prevent="handleDrop"
    @click="triggerFileInput"
  >
    <input ref="fileInputRef" type="file" multiple style="display:none" @change="handleInputChange" />

    <div v-if="!isUploading" class="upload-hint">
      <div class="upload-icon">📤</div>
      <div>拖拽文件到此处，或<span class="link">点击选择文件</span></div>
      <div class="sub-hint">支持多文件同时上传</div>
    </div>

    <div v-else class="upload-progress">
      <div v-for="job in jobs" :key="job.name" class="job">
        <div class="job-name">{{ job.name }}</div>
        <n-progress
          type="line"
          :percentage="job.percent"
          :status="job.status === 'success' ? 'success' : job.status === 'exception' ? 'error' : 'default'"
          :show-indicator="true"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { NProgress } from 'naive-ui'
import { useMessage } from 'naive-ui'
import { fileApi } from '@/services/api'
import { useFileStore } from '@/stores/file'

interface UploadJob {
  name: string
  percent: number
  status: '' | 'success' | 'exception'
}

const store = useFileStore()
const message = useMessage()
const isDragover = ref(false)
const isUploading = ref(false)
const jobs = ref<UploadJob[]>([])
const fileInputRef = ref<HTMLInputElement | null>(null)

function triggerFileInput() {
  if (!isUploading.value) fileInputRef.value?.click()
}

function handleInputChange(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files) {
    uploadFiles(Array.from(input.files))
    input.value = ''
  }
}

function handleDrop(e: DragEvent) {
  isDragover.value = false
  if (e.dataTransfer?.files) uploadFiles(Array.from(e.dataTransfer.files))
}

async function uploadFiles(files: File[]) {
  if (files.length === 0) return
  isUploading.value = true
  jobs.value = files.map(f => ({ name: f.name, percent: 0, status: '' as const }))
  try {
    await fileApi.upload(store.currentPath, files, percent => {
      for (const job of jobs.value) job.percent = percent
    })
    for (const job of jobs.value) { job.percent = 100; job.status = 'success' }
    message.success(`成功上传 ${files.length} 个文件`)
    await store.refresh()
  } catch {
    for (const job of jobs.value) { if (job.percent < 100) job.status = 'exception' }
    message.error('上传失败，请重试')
  } finally {
    setTimeout(() => { isUploading.value = false; jobs.value = [] }, 1500)
  }
}
</script>

<style scoped lang="scss">
.upload-zone {
  border: 2px dashed #d9d9d9;
  border-radius: 8px;
  padding: 24px 16px;
  text-align: center;
  cursor: pointer;
  transition: border-color 0.2s, background 0.2s;
  min-height: 100px;
  display: flex;
  align-items: center;
  justify-content: center;

  &:hover, &.dragover {
    border-color: #18a058;
    background: #e8f5ee;
  }
  &.uploading {
    cursor: default;
    width: 100%;
  }
}

.upload-hint {
  color: #888;
  user-select: none;
}

.upload-icon {
  font-size: 2rem;
  margin-bottom: 8px;
}

.link { color: #18a058; font-weight: 500; }
.sub-hint { font-size: 12px; margin-top: 4px; color: #bbb; }

.upload-progress { width: 100%; max-width: 480px; }
.job { margin-bottom: 8px; text-align: left; }
.job-name {
  font-size: 12px;
  color: #555;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
