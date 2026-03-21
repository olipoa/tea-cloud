<template>
  <div class="pdf-viewer">
    <div class="pdf-toolbar">
      <el-button :icon="ArrowLeft" size="small" :disabled="page <= 1" @click="page--" />
      <span class="page-info">{{ page }} / {{ totalPages }}</span>
      <el-button :icon="ArrowRight" size="small" :disabled="page >= totalPages" @click="page++" />
      <el-button :icon="ZoomIn" size="small" @click="scale = Math.min(3, scale + 0.2)" />
      <el-button :icon="ZoomOut" size="small" @click="scale = Math.max(0.3, scale - 0.2)" />
    </div>

    <div class="pdf-container" v-loading="loading">
      <canvas ref="canvasRef" class="pdf-canvas" />
      <div v-if="error" class="error-msg">PDF 加载失败</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { ArrowLeft, ArrowRight, ZoomIn, ZoomOut } from '@element-plus/icons-vue'

const props = defineProps<{ url: string }>()

const canvasRef = ref<HTMLCanvasElement | null>(null)
const page = ref(1)
const totalPages = ref(0)
const scale = ref(1.5)
const loading = ref(false)
const error = ref(false)

// eslint-disable-next-line @typescript-eslint/no-explicit-any
let pdfDoc: any = null

async function loadPdf() {
  loading.value = true
  error.value = false
  try {
    // Lazy-load pdfjs to avoid bloating the initial bundle
    const pdfjsLib = await import('pdfjs-dist')
    // Use local worker so it works offline on LAN
    pdfjsLib.GlobalWorkerOptions.workerSrc = new URL(
      'pdfjs-dist/build/pdf.worker.mjs',
      import.meta.url,
    ).toString()
    pdfDoc = await pdfjsLib.getDocument(props.url).promise
    totalPages.value = pdfDoc.numPages
    page.value = 1
    await renderPage(1)
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}

async function renderPage(num: number) {
  if (!pdfDoc || !canvasRef.value) return
  loading.value = true
  try {
    const p = await pdfDoc.getPage(num)
    const viewport = p.getViewport({ scale: scale.value })
    const canvas = canvasRef.value
    const ctx = canvas.getContext('2d')!
    canvas.height = viewport.height
    canvas.width = viewport.width
    await p.render({ canvasContext: ctx, viewport }).promise
  } finally {
    loading.value = false
  }
}

watch(page, val => renderPage(val))
watch(scale, () => renderPage(page.value))
watch(() => props.url, loadPdf)
onMounted(loadPdf)
</script>

<style scoped lang="scss">
.pdf-viewer {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.pdf-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--el-fill-color-light);
  border-bottom: 1px solid var(--el-border-color-light);
  width: 100%;
  box-sizing: border-box;
}

.page-info {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  min-width: 60px;
  text-align: center;
}

.pdf-container {
  overflow: auto;
  max-height: 75vh;
  width: 100%;
  display: flex;
  justify-content: center;
  padding: 16px;
  box-sizing: border-box;
}

.pdf-canvas {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
}

.error-msg {
  color: var(--el-color-danger);
  padding: 32px;
}
</style>
