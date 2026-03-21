<template>
  <div class="image-viewer" @click="handleBgClick">
    <div class="image-toolbar">
      <button class="tb-btn" @click.stop="zoom(0.2)" title="放大">🔍+</button>
      <button class="tb-btn" @click.stop="zoom(-0.2)" title="缩小">🔍-</button>
      <button class="tb-btn" @click.stop="rotate(90)" title="旋转">🔄</button>
      <button class="tb-btn" @click.stop="scale = 1; angle = 0" title="重置">重置</button>
    </div>
    <div class="image-container" ref="containerRef">
      <img
        :src="url"
        :style="imgStyle"
        class="preview-img"
        draggable="false"
        @error="error = true"
        @load="error = false"
      />
      <div v-if="error" class="error-msg">图片加载失败</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{ url: string }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const scale = ref(1)
const angle = ref(0)
const error = ref(false)
const containerRef = ref<HTMLDivElement | null>(null)

const imgStyle = computed(() => ({
  transform: `scale(${scale.value}) rotate(${angle.value}deg)`,
  transition: 'transform 0.2s',
  maxWidth: '100%',
  maxHeight: '75vh',
  display: 'block',
  margin: 'auto',
}))

function zoom(delta: number) {
  scale.value = Math.min(5, Math.max(0.1, scale.value + delta))
}

function rotate(deg: number) {
  angle.value = (angle.value + deg) % 360
}

function handleBgClick(e: MouseEvent) {
  if (e.target === e.currentTarget) emit('close')
}
</script>

<style scoped lang="scss">
.image-viewer {
  background: rgba(0,0,0,0.04);
  border-radius: 6px;
  overflow: hidden;
}

.image-toolbar {
  padding: 8px 12px;
  background: #f7f7f7;
  border-bottom: 1px solid #eee;
  display: flex;
  gap: 6px;
}

.tb-btn {
  padding: 4px 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: #fff;
  cursor: pointer;
  font-size: 13px;
  &:hover { background: #f0f0f0; }
}

.image-container {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: auto;
  max-height: 75vh;
  padding: 16px;
}

.preview-img { object-fit: contain; user-select: none; }
.error-msg { color: #d03050; padding: 32px; }
</style>
