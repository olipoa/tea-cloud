<template>
  <Transition name="player-slide">
    <div v-if="store.playlist.length > 0" class="media-player-bar">

      <!-- ArtPlayer container (visible for both audio & video) -->
      <div ref="artRef" class="art-container" :class="{ 'audio-mode': !store.isVideo, 'video-mode': store.isVideo && store.showVideo }" />

      <!-- Bottom control bar (always visible) -->
      <div class="bar-body">
        <!-- Left: track info -->
        <div class="track-info">
          <span class="media-emoji">{{ store.isVideo ? '🎬' : '🎵' }}</span>
          <div class="track-text">
            <span class="track-name" :title="store.currentTrack?.name">{{ store.currentTrack?.name }}</span>
            <span class="time-text">{{ formatTime(store.currentTime) }} / {{ formatTime(store.duration) }}</span>
          </div>
        </div>

        <!-- Center: prev / play-pause / next -->
        <div class="controls">
          <button class="ctrl-btn" :disabled="!store.hasPrev" @click="store.prev()" title="上一首">
            <SkipBack />
          </button>
          <button class="ctrl-btn play-btn" @click="togglePlayPause" title="播放/暂停">
            <svg v-if="store.isPlaying" viewBox="0 0 24 24" fill="currentColor"><path d="M6 19h4V5H6zm8-14v14h4V5z"/></svg>
            <svg v-else viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>
          </button>
          <button class="ctrl-btn" :disabled="!store.hasNext" @click="store.next()" title="下一首">
            <SkipForward />
          </button>
        </div>

        <!-- Right: cast + video expand + playlist -->
        <div class="bar-right">
          <!-- Cast popover -->
          <n-popover v-model:show="castPopoverVisible" trigger="manual" placement="top-end" :width="220">
            <template #trigger>
              <button class="ctrl-btn cast-btn" @click.stop="openCastPopover" title="投屏">
                <CastIcon />
              </button>
            </template>
            <div class="cast-popover">
              <div class="cast-popover-title">选择投屏设备</div>
              <div v-if="castLoading" class="cast-popover-msg">正在扫描...</div>
              <div v-else-if="castDevices.length === 0" class="cast-popover-msg">未发现 DLNA 设备</div>
              <ul v-else class="cast-device-list">
                <li
                  v-for="d in castDevices"
                  :key="d.location"
                  class="cast-device-item"
                  :title="d.name"
                  @click="castTo(d)"
                >
                  <span class="cast-device-icon">📺</span>
                  <span class="cast-device-name">{{ d.name }}</span>
                </li>
              </ul>
            </div>
          </n-popover>

          <button v-if="store.isVideo" class="ctrl-btn" @click="store.toggleVideo()" :title="store.showVideo ? '收起视频' : '展开视频'">
            <svg viewBox="0 0 24 24" fill="currentColor"><path v-if="store.showVideo" d="M5 16h3v3h2v-5H5zm3-8H5v2h5V5H8zm6 11h2v-3h3v-2h-5zm2-11V5h-2v5h5V8z"/><path v-else d="M7 14H5v5h5v-2H7v-3zm-2-4h2V7h3V5H5v5zm12 7h-3v2h5v-5h-2v3zM14 5v2h3v3h2V5h-5z"/></svg>
          </button>

          <span class="playlist-count">{{ store.currentIndex + 1 }} / {{ store.playlist.length }}</span>
          <button class="ctrl-btn" :class="{ active: store.showPlaylist }" @click="store.togglePlaylist()" title="播放列表">
            <svg viewBox="0 0 24 24" fill="currentColor"><path d="M3 13h2v-2H3zm0 4h2v-2H3zm0-8h2V7H3zm4 4h14v-2H7zm0 4h14v-2H7zM7 7v2h14V7z"/></svg>
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, computed, watch, onBeforeUnmount, nextTick, h } from 'vue'
import { NPopover } from 'naive-ui'
import Artplayer from 'artplayer'
import { useMediaPlayerStore } from '@/stores/audioPlayer'
import { fileApi } from '@/services/api'
import { useCast } from '@/composables/useCast'
import { formatTime } from '@/utils/fileUtils'
import { SkipBack, SkipForward, CastIcon } from '@/components/icons/index'

const store = useMediaPlayerStore()
const { castPopoverVisible, castLoading, castDevices, openCastPopover, castTo } = useCast()

const artRef = ref<HTMLDivElement | null>(null)
let art: Artplayer | null = null

const currentUrl = computed(() =>
  store.currentTrack ? fileApi.rawUrl(store.currentTrack.path) : '',
)

function getMimeType(ext: string): string {
  const map: Record<string, string> = {
    mp4: 'video/mp4', webm: 'video/webm', ogg: 'video/ogg', mkv: 'video/x-matroska',
    avi: 'video/x-msvideo', mov: 'video/quicktime',
    mp3: 'audio/mpeg', flac: 'audio/flac', wav: 'audio/wav', aac: 'audio/aac',
    m4a: 'audio/mp4', opus: 'audio/ogg',
  }
  return map[ext.toLowerCase()] ?? ''
}

function createArtPlayer() {
  if (!artRef.value || !currentUrl.value) return
  if (art) { art.destroy(); art = null }

  const ext = store.currentTrack?.ext ?? ''
  art = new Artplayer({
    container: artRef.value,
    url: currentUrl.value,
    type: getMimeType(ext),
    volume: 1,
    autoplay: store.isPlaying,
    pip: false,
    fullscreen: store.isVideo,
    fullscreenWeb: false,
    playbackRate: false,
    aspectRatio: false,
    setting: false,
    hotkey: true,
    screenshot: false,
    lock: true,
    isLive: false,
    theme: '#18a058',
  })

  art.on('ready', () => {
    store.setDuration(art!.duration)
    if (store.currentTime > 0) art!.currentTime = store.currentTime
  })

  art.on('video:timeupdate', () => {
    store.setCurrentTime(art!.currentTime)
  })

  art.on('video:durationchange', () => {
    store.setDuration(art!.duration)
  })

  art.on('play', () => store.setIsPlaying(true))
  art.on('pause', () => store.setIsPlaying(false))

  art.on('video:ended', () => {
    if (store.hasNext) {
      store.next()
    } else {
      store.setIsPlaying(false)
      store.setCurrentTime(0)
    }
  })

  art.on('error', () => {
    if (store.hasNext) store.next()
  })
}

// Switch track
watch(currentUrl, async (url) => {
  if (!url) return
  if (art) {
    art.switchUrl(url)
    if (store.isPlaying) {
      await nextTick()
      art.play().catch(() => {})
    }
  } else {
    await nextTick()
    createArtPlayer()
  }
})

// Play/pause from store
watch(() => store.isPlaying, (playing) => {
  if (!art) return
  if (playing) art.play().catch?.(() => {})
  else art.pause()
})

// When artRef becomes available (initially)
watch(artRef, async (el) => {
  if (el && currentUrl.value) {
    await nextTick()
    createArtPlayer()
  }
})

// When playlist first gets items
watch(() => store.playlist.length, async (len) => {
  if (len > 0 && artRef.value && !art) {
    await nextTick()
    createArtPlayer()
  }
})

function togglePlayPause() {
  if (!art) return
  if (art.playing) art.pause()
  else art.play().catch(() => {})
}

onBeforeUnmount(() => {
  if (art) { art.destroy(); art = null }
})
</script>

<style scoped lang="scss">
$sidebar-w: 220px;
$topbar-h: 48px;
$bar-body-h: 56px;
$breakpoint: 768px;

.media-player-bar {
  position: fixed;
  bottom: 0;
  left: $sidebar-w;  // desktop: start after sidebar
  right: 0;
  z-index: 200;
  background: #fff;
  border-top: 1px solid #efeff5;
  box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.08);
  display: flex;
  flex-direction: column;
}

/* ArtPlayer container */
.art-container {
  width: 100%;
  background: #000;

  /* Audio mode: collapsed to 0 height — ArtPlayer renders a mini audio bar natively */
  &.audio-mode {
    height: 48px;
    :deep(.art-video-player) {
      background: #1a1a1a;
    }
  }

  /* Video mode: fills from top of viewport to bar-body */
  &.video-mode {
    height: calc(100vh - #{$bar-body-h});
  }

  /* Hidden when video is minimised */
  &:not(.audio-mode):not(.video-mode) {
    height: 0;
    overflow: hidden;
  }
}

/* Controls row */
.bar-body {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  gap: 12px;
  height: 56px;
  flex-shrink: 0;
}

.track-info {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  flex: 1;
}

.media-emoji {
  font-size: 1.4rem;
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
  max-width: 200px;
  color: #333;
}

.time-text {
  font-size: 11px;
  color: #888;
  margin-top: 1px;
}

.controls {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.ctrl-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  border-radius: 50%;
  cursor: pointer;
  color: #555;
  padding: 0;
  transition: background 0.15s, color 0.15s;

  &:hover {
    background: rgba(0,0,0,0.07);
    color: #18a058;
  }

  &:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }

  &.active {
    color: #18a058;
  }

  svg {
    width: 20px;
    height: 20px;
  }
}

.play-btn {
  width: 40px;
  height: 40px;
  background: #18a058;
  color: #fff;

  &:hover {
    background: #0e7a43;
    color: #fff;
  }

  svg {
    width: 22px;
    height: 22px;
  }
}

.cast-btn svg {
  width: 18px;
  height: 18px;
}

.bar-right {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
  justify-content: flex-end;
}

.playlist-count {
  font-size: 12px;
  color: #888;
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

/* ====== MOBILE ====== */
@media (max-width: #{$breakpoint}) {
  .media-player-bar {
    left: 0;  // full width on mobile
    top: $topbar-h;  // don't cover the topbar (hamburger)
  }

  .art-container.video-mode {
    // subtract topbar height so video never covers hamburger
    height: calc(100vh - #{$topbar-h} - #{$bar-body-h});
  }
}

/* Cast popover */
.cast-popover {
  padding: 2px 0;
}

.cast-popover-title {
  font-size: 12px;
  font-weight: 600;
  color: #888;
  padding: 0 4px 6px;
  border-bottom: 1px solid #eee;
  margin-bottom: 4px;
}

.cast-popover-msg {
  font-size: 13px;
  color: #888;
  padding: 8px 4px;
  text-align: center;
}

.cast-device-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.cast-device-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 6px;
  font-size: 13px;
  border-radius: 4px;
  cursor: pointer;
  color: #333;

  &:hover {
    background: #f5f5f7;
  }
}

.cast-device-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.cast-device-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 160px;
}
</style>
