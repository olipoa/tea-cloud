<template>
  <div class="audio-player">
    <div class="track-info">
      <el-icon class="music-icon"><Headset /></el-icon>
      <span class="track-name">{{ filename }}</span>
    </div>
    <audio
      ref="audioRef"
      :src="url"
      controls
      preload="metadata"
      class="audio-el"
      @error="error = true"
    />
    <div v-if="error" class="error-msg">无法播放该音频格式</div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { Headset } from '@element-plus/icons-vue'

const props = defineProps<{ url: string; filename: string }>()
const audioRef = ref<HTMLAudioElement | null>(null)
const error = ref(false)

watch(() => props.url, () => {
  error.value = false
  audioRef.value?.load()
})

onBeforeUnmount(() => {
  audioRef.value?.pause()
})
</script>

<style scoped lang="scss">
.audio-player {
  padding: 24px;
  background: var(--el-fill-color-light);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.track-info {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 15px;
  font-weight: 500;
}

.music-icon {
  font-size: 2rem;
  color: var(--el-color-primary);
}

.track-name {
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.audio-el {
  width: 100%;
  max-width: 480px;
}

.error-msg {
  color: var(--el-color-danger);
  font-size: 13px;
}
</style>
