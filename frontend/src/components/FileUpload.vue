<template>
  <div
    class="upload-zone"
    :class="{ dragover: isDragover, uploading: isUploading }"
    @dragover.prevent="isDragover = true"
    @dragleave.prevent="isDragover = false"
    @drop.prevent="handleDrop"
    @click="triggerFileInput"
  >
    <input
      ref="fileInputRef"
      type="file"
      multiple
      style="display: none"
      @change="handleInputChange"
    />

    <div v-if="!isUploading" class="upload-hint">
      <div class="upload-icon">📤</div>
      <div>拖拽文件到此处，或<span class="link">点击选择文件</span></div>
      <div class="sub-hint">支持多文件，自动断点续传</div>
      <div class="concurrent-row" @click.stop>
        <span class="concurrent-label">并发数：</span>
        <n-select
          v-model:value="concurrentLimit"
          :options="concurrentOptions"
          size="small"
          style="width: 70px"
        />
      </div>
    </div>

    <div v-else class="upload-progress">
      <div v-for="job in jobs" :key="job.name" class="job">
        <div class="job-name">{{ job.name }}</div>
        <n-progress
          type="line"
          :percentage="job.percent"
          :status="
            job.status === 'success'
              ? 'success'
              : job.status === 'exception'
                ? 'error'
                : 'default'
          "
          :show-indicator="true"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { uploadApi } from "@/services/api";
import { useFileStore } from "@/stores/file";
import { NProgress, NSelect, useMessage } from "naive-ui";
import { ref } from "vue";

const CHUNK_SIZE = 5 * 1024 * 1024; // 5 MB per chunk
const MAX_RETRIES = 3;

interface UploadJob {
  name: string;
  percent: number;
  status: "" | "success" | "exception";
}

const store = useFileStore();
const message = useMessage();
const isDragover = ref(false);
const isUploading = ref(false);
const jobs = ref<UploadJob[]>([]);
const fileInputRef = ref<HTMLInputElement | null>(null);
const concurrentLimit = ref(3);
const concurrentOptions = [1, 2, 3, 4, 5].map((n) => ({
  label: String(n),
  value: n,
}));

function lsKey(file: File): string {
  return `tea-cloud-up:${file.name}:${file.size}`;
}

function triggerFileInput() {
  if (!isUploading.value) fileInputRef.value?.click();
}

function handleInputChange(e: Event) {
  const input = e.target as HTMLInputElement;
  if (input.files) {
    uploadFiles(Array.from(input.files));
    input.value = "";
  }
}

function handleDrop(e: DragEvent) {
  isDragover.value = false;
  if (e.dataTransfer?.files) uploadFiles(Array.from(e.dataTransfer.files));
}

async function uploadOneFile(file: File, jobIndex: number): Promise<boolean> {
  const totalChunks = Math.max(1, Math.ceil(file.size / CHUNK_SIZE));
  const key = lsKey(file);

  try {
    let uploadId = localStorage.getItem(key) ?? "";
    let uploadedChunks: number[] = [];

    if (uploadId) {
      try {
        const status = await uploadApi.status(uploadId);
        uploadedChunks = status.uploadedChunks ?? [];
      } catch {
        // Session expired or not found — start fresh
        uploadId = "";
      }
    }

    if (!uploadId) {
      const init = await uploadApi.init(
        store.currentPath,
        file.name,
        file.size,
        totalChunks,
      );
      uploadId = init.uploadId;
      uploadedChunks = init.uploadedChunks ?? [];
      localStorage.setItem(key, uploadId);
    }

    const uploadedSet = new Set(uploadedChunks);
    const pending = Array.from({ length: totalChunks }, (_, i) => i).filter(
      (i) => !uploadedSet.has(i),
    );

    let doneCount = uploadedChunks.length;
    jobs.value[jobIndex].percent = Math.round((doneCount / totalChunks) * 100);

    for (const chunkIndex of pending) {
      const start = chunkIndex * CHUNK_SIZE;
      const blob = file.slice(start, start + CHUNK_SIZE);

      let uploaded = false;
      for (let attempt = 0; attempt < MAX_RETRIES; attempt++) {
        try {
          await uploadApi.chunk(uploadId, chunkIndex, blob);
          uploaded = true;
          break;
        } catch {
          if (attempt === MAX_RETRIES - 1)
            throw new Error(
              `chunk ${chunkIndex} failed after ${MAX_RETRIES} retries`,
            );
          await new Promise((r) => setTimeout(r, 800 * (attempt + 1)));
        }
      }

      if (uploaded) {
        doneCount++;
        jobs.value[jobIndex].percent = Math.round(
          (doneCount / totalChunks) * 100,
        );
      }
    }

    await uploadApi.complete(uploadId);
    localStorage.removeItem(key);
    return true;
  } catch {
    return false;
  }
}

async function uploadFiles(files: File[]) {
  if (files.length === 0) return;
  isUploading.value = true;
  jobs.value = files.map((f) => ({
    name: f.name,
    percent: 0,
    status: "" as const,
  }));

  let running = 0;
  let idx = 0;
  let successCount = 0;

  await new Promise<void>((resolve) => {
    function startNext() {
      while (running < concurrentLimit.value && idx < files.length) {
        const i = idx++;
        running++;
        uploadOneFile(files[i], i)
          .then((ok) => {
            if (ok) {
              jobs.value[i].percent = 100;
              jobs.value[i].status = "success";
              successCount++;
            } else {
              jobs.value[i].status = "exception";
            }
          })
          .finally(() => {
            running--;
            if (idx < files.length) {
              startNext();
            } else if (running === 0) {
              resolve();
            }
          });
      }
    }
    startNext();
  });

  if (successCount === files.length) {
    message.success(`成功上传 ${files.length} 个文件`);
  } else {
    message.warning(
      `${successCount}/${files.length} 个文件上传成功，${files.length - successCount} 个失败`,
    );
  }
  await store.refresh();
  setTimeout(() => {
    isUploading.value = false;
    jobs.value = [];
  }, 1500);
}
</script>

<style scoped lang="scss">
.upload-zone {
  border: 2px dashed #d9d9d9;
  border-radius: 8px;
  padding: 24px 16px;
  text-align: center;
  cursor: pointer;
  transition:
    border-color 0.2s,
    background 0.2s;
  min-height: 100px;
  display: flex;
  align-items: center;
  justify-content: center;

  &:hover,
  &.dragover {
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

.link {
  color: #18a058;
  font-weight: 500;
}
.sub-hint {
  font-size: 12px;
  margin-top: 4px;
  color: #bbb;
}

.concurrent-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 10px;
  justify-content: center;
}

.concurrent-label {
  font-size: 12px;
  color: #888;
}

.upload-progress {
  width: 100%;
  max-width: 480px;
}
.job {
  margin-bottom: 8px;
  text-align: left;
}
.job-name {
  font-size: 12px;
  color: #555;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
