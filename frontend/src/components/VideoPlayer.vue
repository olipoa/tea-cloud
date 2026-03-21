<template>
  <div class="video-player">
    <video
      ref="videoRef"
      :src="url"
      controls
      preload="metadata"
      playsinline
      @error="onError"
      class="video-el"
    >
      您的浏览器不支持视频播放。
    </video>
    <div v-if="error" class="error-msg">
      <el-icon><WarningFilled /></el-icon>
      无法播放该视频格式，请下载后在本地播放。
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onBeforeUnmount } from 'vue'
import { WarningFilled } from '@element-plus/icons-vue'

const props = defineProps<{ url: string }>()
const videoRef = ref<HTMLVideoElement | null>(null)
const error = ref(false)

watch(() => props.url, () => {
  error.value = false
  videoRef.value?.load()
})

function onError() {
  error.value = true
}

onBeforeUnmount(() => {
  if (videoRef.value) {
    videoRef.value.pause()
    videoRef.value.src = ''
  }
})
</script>

<style scoped lang="scss">
.video-player {
  width: 100%;
  background: #000;
  border-radius: 6px;
  overflow: hidden;
  position: relative;
}

.video-el {
  width: 100%;
  max-height: 70vh;
  display: block;
}

.error-msg {
  color: #fff;
  background: rgba(0,0,0,0.8);
  padding: 16px;
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}
</style>
