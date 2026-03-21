<template>
  <div class="audio-player">
    <!-- Hidden audio element -->
    <audio
      ref="audioEl"
      :src="currentUrl"
      :autoplay="store.isPlaying"
      @timeupdate="onTimeUpdate"
      @durationchange="onDurationChange"
      @ended="onEnded"
      @play="store.setIsPlaying(true)"
      @pause="store.setIsPlaying(false)"
    />

    <div class="player-inner">
      <!-- Track info -->
      <div class="track-info">
        <t-icon name="music" class="track-icon" />
        <span class="track-name" :title="store.currentTrack?.name">
          {{ store.currentTrack?.name ?? '未选择音频' }}
        </span>
      </div>

      <!-- Controls -->
      <div class="controls">
        <t-button variant="text" shape="square" size="small" :disabled="!store.hasPrev" @click="store.prev()">
          <template #icon><t-icon name="skip-previous-filled" /></template>
        </t-button>
        <t-button variant="text" shape="square" @click="togglePlay">
          <template #icon><t-icon :name="store.isPlaying ? 'pause-circle-filled' : 'play-circle-filled'" class="play-icon" /></template>
        </t-button>
        <t-button variant="text" shape="square" size="small" :disabled="!store.hasNext" @click="store.next()">
          <template #icon><t-icon name="skip-next-filled" /></template>
        </t-button>
      </div>

      <!-- Progress -->
      <div class="progress-row">
        <span class="time">{{ formatTime(store.currentTime) }}</span>
        <t-slider
          v-model="seekValue"
          :min="0"
          :max="100"
          size="small"
          :label="false"
          @change="onSeek"
          style="flex:1"
        />
        <span class="time">{{ formatTime(store.duration) }}</span>
      </div>

      <!-- Right: playlist toggle -->
      <div class="right-actions">
        <t-button variant="text" shape="square" size="small" @click="store.togglePlaylist()" title="播放列表">
          <template #icon><t-icon name="playlist" /></template>
        </t-button>
        <t-button variant="text" shape="square" size="small" @click="store.setPlaylist([], 0)" title="关闭">
          <template #icon><t-icon name="close" /></template>
        </t-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { fileApi } from '@/services/api'
import { useMediaPlayerStore } from '@/stores/audioPlayer'

const store = useMediaPlayerStore()
const audioEl = ref<HTMLAudioElement | null>(null)

const currentUrl = computed(() =>
  store.currentTrack ? fileApi.rawUrl(store.currentTrack.path) : ''
)

const seekValue = computed({
  get: () => store.duration > 0 ? (store.currentTime / store.duration) * 100 : 0,
  set: () => {},
})

function formatTime(s: number): string {
  if (!s || isNaN(s)) return '0:00'
  const m = Math.floor(s / 60)
  const ss = Math.floor(s % 60).toString().padStart(2, '0')
  return `${m}:${ss}`
}

function togglePlay() {
  if (!audioEl.value) return
  if (store.isPlaying) audioEl.value.pause()
  else audioEl.value.play().catch(() => {})
}

function onTimeUpdate() {
  if (audioEl.value) store.setCurrentTime(audioEl.value.currentTime)
}
function onDurationChange() {
  if (audioEl.value) store.setDuration(audioEl.value.duration)
}
function onEnded() {
  if (store.hasNext) store.next()
  else { store.setIsPlaying(false); store.setCurrentTime(0) }
}
function onSeek(val: number) {
  if (!audioEl.value || !store.duration) return
  const t = (val / 100) * store.duration
  audioEl.value.currentTime = t
  store.seek(t)
}

// Sync play/pause from store
watch(() => store.isPlaying, (playing) => {
  if (!audioEl.value) return
  if (playing) audioEl.value.play().catch(() => {})
  else audioEl.value.pause()
})

// Seek when store.currentTime set externally  
watch(() => store.currentTime, (t) => {
  if (audioEl.value && Math.abs(audioEl.value.currentTime - t) > 1) {
    audioEl.value.currentTime = t
  }
})
</script>

<style scoped lang="scss">
.audio-player {
  padding: 8px 16px;
}
.player-inner {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.track-info {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 120px;
  flex: 1;
  overflow: hidden;
}
.track-icon { color: var(--td-brand-color); flex-shrink: 0; }
.track-name {
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.controls {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}
.play-icon { font-size: 28px !important; }
.progress-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 2;
  min-width: 160px;
}
.time { font-size: 11px; color: var(--td-text-color-secondary); white-space: nowrap; }
.right-actions { display: flex; align-items: center; gap: 4px; flex-shrink: 0; }
</style>
