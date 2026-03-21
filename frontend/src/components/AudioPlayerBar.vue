<template>
  <Transition name="player-slide">
    <div v-if="store.playlist.length > 0" class="media-player-bar" :class="{ 'has-video': store.isVideo && store.showVideo }">

      <!-- Inline video display -->
      <div v-if="store.isVideo && store.showVideo" ref="videoAreaRef" class="video-area">
        <div class="video-overlay-header">
          <span class="video-title">{{ store.currentTrack?.name }}</span>
          <div class="video-header-actions">
            <el-button size="small" circle :icon="CastIcon" class="overlay-btn" @click="startCasting" title="投屏" />
            <el-button size="small" circle :icon="FullScreen" class="overlay-btn" @click="toggleNativeFullscreen" title="全屏" />
            <el-button size="small" circle :icon="Close" class="overlay-btn" @click="store.toggleVideo()" title="最小化" />
          </div>
        </div>
        <video
          ref="mediaRef"
          :src="currentUrl"
          preload="metadata"
          playsinline
          class="video-el"
          @timeupdate="onTimeUpdate"
          @loadedmetadata="onLoadedMetadata"
          @ended="onEnded"
          @play="store.setIsPlaying(true)"
          @pause="store.setIsPlaying(false)"
          @error="onError"
        />
      </div>

      <!-- Hidden audio element (audio mode, or video minimised) -->
      <audio
        v-else
        ref="mediaRef"
        :src="currentUrl"
        preload="metadata"
        @timeupdate="onTimeUpdate"
        @loadedmetadata="onLoadedMetadata"
        @ended="onEnded"
        @play="store.setIsPlaying(true)"
        @pause="store.setIsPlaying(false)"
        @error="onError"
      />

      <!-- Progress bar -->
      <div class="progress-bar" @click="onProgressClick" @mousemove="onProgressHover" @mouseleave="hoverTime = null">
        <div class="progress-fill" :style="{ width: progressPercent + '%' }" />
        <div v-if="hoverTime !== null" class="progress-tooltip" :style="{ left: hoverPercent + '%' }">
          {{ formatTime(hoverTime) }}
        </div>
      </div>

      <!-- Controls row -->
      <div class="bar-body">
        <!-- Left: track info -->
        <div class="track-info">
          <el-icon class="media-icon">
            <VideoCamera v-if="store.isVideo" />
            <Headset v-else />
          </el-icon>
          <div class="track-text">
            <span class="track-name" :title="store.currentTrack?.name">{{ store.currentTrack?.name }}</span>
            <span class="time-text">{{ formatTime(store.currentTime) }} / {{ formatTime(store.duration) }}</span>
          </div>
        </div>

        <!-- Center: playback controls -->
        <div class="controls">
          <el-button :icon="SkipBack" circle size="small" :disabled="!store.hasPrev" @click="store.prev()" class="ctrl-btn" />
          <el-button :icon="store.isPlaying ? VideoPause : VideoPlay" circle @click="store.togglePlay()" type="primary" class="ctrl-btn play-btn" />
          <el-button :icon="SkipForward" circle size="small" :disabled="!store.hasNext" @click="store.next()" class="ctrl-btn" />
        </div>

        <!-- Right: expand video + playlist -->
        <div class="bar-right">
          <el-button
            v-if="store.isVideo && !store.showVideo"
            :icon="FullScreen"
            circle
            size="small"
            @click="store.toggleVideo()"
            class="ctrl-btn"
            title="展开视频"
          />
          <span class="playlist-count">{{ store.currentIndex + 1 }} / {{ store.playlist.length }}</span>
          <el-button
            :icon="List"
            circle
            size="small"
            @click="store.togglePlaylist()"
            :type="store.showPlaylist ? 'primary' : ''"
            class="ctrl-btn"
            title="播放列表"
          />
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { Headset, VideoCamera, VideoPlay, VideoPause, List, Close, FullScreen } from '@element-plus/icons-vue'
import { useMediaPlayerStore } from '@/stores/audioPlayer'
import { fileApi } from '@/services/api'
import { defineComponent, h } from 'vue'
import { ElMessage } from 'element-plus'

const SkipBack = defineComponent({
  render: () => h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'currentColor' }, [
    h('path', { d: 'M6 6h2v12H6zm3.5 6 8.5 6V6z' }),
  ]),
})

const SkipForward = defineComponent({
  render: () => h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'currentColor' }, [
    h('path', { d: 'M6 18l8.5-6L6 6v12zm2.5-6 5.5 3.9V8.1L8.5 12zM16 6h2v12h-2z' }),
  ]),
})

const CastIcon = defineComponent({
  render: () => h('svg', { xmlns: 'http://www.w3.org/2000/svg', viewBox: '0 0 24 24', fill: 'currentColor' }, [
    h('path', { d: 'M1 18v3h3c0-1.66-1.34-3-3-3zm0-4v2c2.76 0 5 2.24 5 5h2c0-3.87-3.13-7-7-7zm18-7H5v1.63c3.96 1.28 7.09 4.41 8.37 8.37H19V7zM1 10v2c4.97 0 9 4.03 9 9h2c0-6.08-4.93-11-11-11zm20-7H3c-1.1 0-2 .9-2 2v3h2V5h18v14h-7v2h7c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2z' }),
  ]),
})

const store = useMediaPlayerStore()
const mediaRef = ref<HTMLMediaElement | null>(null)
const videoAreaRef = ref<HTMLElement | null>(null)
const hoverTime = ref<number | null>(null)
const hoverPercent = ref(0)

function startCasting() {
  const el = mediaRef.value as any
  if (!el) return
  if (typeof el.webkitShowPlaybackTargetPicker === 'function') {
    el.webkitShowPlaybackTargetPicker()
  } else if (el.remote && typeof el.remote.prompt === 'function') {
    el.remote.prompt().catch(() => {})
  } else {
    ElMessage.warning('当前浏览器不支持投屏功能')
  }
}

function toggleNativeFullscreen() {
  const el = videoAreaRef.value
  if (!el) return
  if (!document.fullscreenElement) {
    el.requestFullscreen().catch(() => {})
  } else {
    document.exitFullscreen().catch(() => {})
  }
}

const currentUrl = computed(() =>
  store.currentTrack ? fileApi.rawUrl(store.currentTrack.path) : '',
)

const progressPercent = computed(() =>
  store.duration > 0 ? (store.currentTime / store.duration) * 100 : 0,
)

// Track changes — reload + autoplay
watch(currentUrl, async () => {
  await nextTick()
  if (!currentUrl.value || !mediaRef.value) return
  mediaRef.value.load()
  if (store.isPlaying) {
    try { await mediaRef.value.play() } catch (_) {}
  }
})

// Play/pause from store
watch(() => store.isPlaying, async (playing) => {
  if (!mediaRef.value) return
  if (playing) {
    try { await mediaRef.value.play() } catch (_) {}
  } else {
    mediaRef.value.pause()
  }
})

// Seek from store
watch(() => store.currentTime, (time) => {
  if (mediaRef.value && Math.abs(mediaRef.value.currentTime - time) > 0.5) {
    mediaRef.value.currentTime = time
  }
})

// Re-attach when switching video ↔ audio element
watch(() => store.showVideo, async () => {
  await nextTick()
  if (!mediaRef.value || !currentUrl.value) return
  mediaRef.value.load()
  mediaRef.value.currentTime = store.currentTime
  if (store.isPlaying) {
    try { await mediaRef.value.play() } catch (_) {}
  }
})

// Initial start
watch(() => store.playlist.length, async (len) => {
  if (len > 0 && store.isPlaying) {
    await nextTick()
    mediaRef.value?.load()
    mediaRef.value?.play().catch(() => {})
  }
}, { immediate: true })

function onTimeUpdate() {
  if (mediaRef.value) store.setCurrentTime(mediaRef.value.currentTime)
}

function onLoadedMetadata() {
  if (mediaRef.value) {
    store.setDuration(mediaRef.value.duration)
    if (store.isPlaying) mediaRef.value.play().catch(() => {})
  }
}

function onEnded() {
  if (store.hasNext) {
    store.next()
  } else {
    store.setIsPlaying(false)
    store.setCurrentTime(0)
  }
}

function onError() {
  if (store.hasNext) store.next()
}

function onProgressClick(e: MouseEvent) {
  const bar = e.currentTarget as HTMLElement
  const newTime = (e.offsetX / bar.clientWidth) * store.duration
  store.seek(newTime)
  if (mediaRef.value) mediaRef.value.currentTime = newTime
}

function onProgressHover(e: MouseEvent) {
  const bar = e.currentTarget as HTMLElement
  const ratio = e.offsetX / bar.clientWidth
  hoverPercent.value = ratio * 100
  hoverTime.value = ratio * store.duration
}

function formatTime(secs: number): string {
  if (!secs || isNaN(secs)) return '0:00'
  const h = Math.floor(secs / 3600)
  const m = Math.floor((secs % 3600) / 60)
  const s = Math.floor(secs % 60)
  if (h > 0) return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  return `${m}:${s.toString().padStart(2, '0')}`
}
</script>

<style scoped lang="scss">
.media-player-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 200;
  background: var(--el-bg-color);
  border-top: 1px solid var(--el-border-color-light);
  box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.08);
  display: flex;
  flex-direction: column;

  &.has-video {
    top: 0;
  }


}

/* Video area */
.video-area {
  background: #000;
  position: relative;
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
}

.video-overlay-header {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: linear-gradient(to bottom, rgba(0,0,0,0.65), transparent);
  z-index: 1;
  opacity: 0;
  transition: opacity 0.2s;
}

.video-area:hover .video-overlay-header {
  opacity: 1;
}

.video-title {
  color: #fff;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  margin-right: 8px;
}

.video-header-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.overlay-btn {
  color: #fff !important;
  background: rgba(255,255,255,0.18) !important;
  border: none !important;
}

.video-el {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: contain;
}

/* Progress bar */
.progress-bar {
  width: 100%;
  height: 4px;
  background: var(--el-fill-color);
  cursor: pointer;
  position: relative;
  flex-shrink: 0;
  transition: height 0.15s;

  &:hover {
    height: 6px;
  }
}

.progress-fill {
  height: 100%;
  background: var(--el-color-primary);
  border-radius: 0 2px 2px 0;
  transition: width 0.25s linear;
  pointer-events: none;
}

.progress-tooltip {
  position: absolute;
  top: -24px;
  transform: translateX(-50%);
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 3px;
  pointer-events: none;
  white-space: nowrap;
}

/* Controls row */
.bar-body {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  gap: 12px;
  height: 56px;
}

.track-info {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1;
}

.media-icon {
  font-size: 1.6rem;
  color: var(--el-color-primary);
  flex-shrink: 0;
}

.track-text {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.track-name {
  font-size: 13px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 240px;
  color: var(--el-text-color-primary);
}

.time-text {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  margin-top: 1px;
}

.controls {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.play-btn {
  width: 40px !important;
  height: 40px !important;
  font-size: 1.1rem;
}

.ctrl-btn {
  border: none;
}

.bar-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  justify-content: flex-end;
}

.playlist-count {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  white-space: nowrap;
}

/* Slide transition */
.player-slide-enter-active,
.player-slide-leave-active {
  transition: transform 0.3s ease;
}

.player-slide-enter-from,
.player-slide-leave-to {
  transform: translateY(100%);
}
</style>
