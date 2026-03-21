<template>
  <div class="video-wrapper">
    <div ref="artRef" class="art-container" />

    <!-- Cast device popup (replaces n-popover) -->
    <div v-show="castPopoverVisible" class="cast-popup">
      <div class="cast-popup-title">选择投屏设备</div>
      <div v-if="castLoading" class="cast-popup-msg">正在扫描...</div>
      <div v-else-if="castDevices.length === 0" class="cast-popup-msg">未发现 DLNA 设备</div>
      <ul v-else class="cast-device-list">
        <li v-for="d in castDevices" :key="d.location" class="cast-device-item" @click="castTo(d); castPopoverVisible = false">
          <t-icon name="tv" />
          <span class="cast-device-name">{{ d.name }}</span>
        </li>
      </ul>
      <div class="cast-popup-close" @click="castPopoverVisible = false">关闭</div>
    </div>

    <!-- Playlist side panel -->
    <VideoPlaylistPanel :visible="store.showPlaylist" @close="store.togglePlaylist()" />
  </div>
</template>

<script setup lang="ts">
import { useCast } from '@/composables/useCast'
import { fileApi } from '@/services/api'
import { useMediaPlayerStore } from '@/stores/audioPlayer'
import { ICON_CLOSE, ICON_CAST, ICON_PREV, ICON_NEXT, ICON_PLAYLIST } from '@/utils/playerIcons'
import Artplayer from 'artplayer'
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import VideoPlaylistPanel from './VideoPlaylistPanel.vue'

const store = useMediaPlayerStore()
const { castPopoverVisible, castLoading, castDevices, openCastPopover, castTo } = useCast()
const artRef = ref<HTMLDivElement | null>(null)
let art: Artplayer | null = null

const currentUrl = computed(() =>
  store.currentTrack ? fileApi.rawUrl(store.currentTrack.path) : ''
)

function getMimeType(ext: string): string {
  const map: Record<string, string> = {
    mp4: 'video/mp4', webm: 'video/webm', ogg: 'video/ogg',
    mkv: 'video/x-matroska', avi: 'video/x-msvideo', mov: 'video/quicktime',
    mp3: 'audio/mpeg', flac: 'audio/flac', wav: 'audio/wav',
    aac: 'audio/aac', m4a: 'audio/mp4', opus: 'audio/ogg',
  }
  return map[ext.toLowerCase()] ?? ''
}

function prevDisabled() { return !store.hasPrev }
function nextDisabled() { return !store.hasNext }

function createArtPlayer() {
  if (!artRef.value || !currentUrl.value) return
  if (art) { art.destroy(); art = null }

  const ext = store.currentTrack?.ext ?? ''
  const track = store.currentTrack
  art = new Artplayer({
    container: artRef.value,
    url: currentUrl.value,
    type: getMimeType(ext),
    volume: 1,
    autoplay: store.isPlaying,
    poster: track ? fileApi.thumbnailUrl(track.path) : undefined,
    pip: false,
    fullscreen: true,
    fullscreenWeb: false,
    playbackRate: false,
    aspectRatio: false,
    setting: false,
    hotkey: true,
    screenshot: false,
    lock: true,
    isLive: false,
    theme: '#18a058',
    controls: [
      // Top-left: close (negative index = left side)
      {
        name: 'close-player',
        position: 'top',
        index: -20,
        html: ICON_CLOSE,
        tooltip: '关闭播放器',
        click: () => store.setPlaylist([], 0),
      },
      // Top-left: cast
      {
        name: 'cast',
        position: 'top',
        index: -10,
        html: ICON_CAST,
        tooltip: '投屏',
        click: () => openCastPopover(),
      },
      // Top-right: prev
      {
        name: 'prev',
        position: 'top',
        index: 10,
        html: ICON_PREV,
        tooltip: '上一个',
        click() { if (store.hasPrev) store.prev() },
      },
      // Top-right: next
      {
        name: 'next',
        position: 'top',
        index: 20,
        html: ICON_NEXT,
        tooltip: '下一个',
        click() { if (store.hasNext) store.next() },
      },
      // Top-right: playlist
      {
        name: 'playlist',
        position: 'top',
        index: 30,
        html: ICON_PLAYLIST,
        tooltip: '播放列表',
        click: () => store.togglePlaylist(),
      },
    ],
  })

  art.on('ready', () => {
    store.setDuration(art!.duration)
    if (store.currentTime > 0) art!.currentTime = store.currentTime
  })
  art.on('video:timeupdate', () => { store.setCurrentTime(art!.currentTime) })
  art.on('video:durationchange', () => { store.setDuration(art!.duration) })
  art.on('play', () => store.setIsPlaying(true))
  art.on('pause', () => store.setIsPlaying(false))
  art.on('video:ended', () => {
    if (store.hasNext) store.next()
    else { store.setIsPlaying(false); store.setCurrentTime(0) }
  })
  art.on('error', () => { if (store.hasNext) store.next() })

  // Keep prev/next button disabled state in sync
  watch([() => store.hasPrev, () => store.hasNext], () => {
    if (!art) return
    const prevEl = art.controls['prev']
    const nextEl = art.controls['next']
    if (prevEl) prevEl.style.opacity = store.hasPrev ? '1' : '0.3'
    if (nextEl) nextEl.style.opacity = store.hasNext ? '1' : '0.3'
  })
}

watch(currentUrl, async (url) => {
  if (!url) return
  if (art) {
    art.switchUrl(url)
    if (store.isPlaying) { await nextTick(); art.play().catch(() => {}) }
  } else {
    await nextTick()
    createArtPlayer()
  }
})

watch(() => store.isPlaying, (playing) => {
  if (!art) return
  if (playing) art.play().catch?.(() => {})
  else art.pause()
})

watch(artRef, async (el) => {
  if (!el) { if (art) { art.destroy(); art = null } }
  else if (currentUrl.value) { await nextTick(); createArtPlayer() }
})

watch(() => store.playlist.length, async (len) => {
  if (len > 0 && artRef.value && !art) { await nextTick(); createArtPlayer() }
})

function togglePlayPause() {
  if (!art) return
  if (art.playing) art.pause()
  else art.play().catch(() => {})
}

defineExpose({ togglePlayPause })

onBeforeUnmount(() => { if (art) { art.destroy(); art = null } })
</script>

<style scoped lang="scss">
$topbar-h: 48px;
$breakpoint: 768px;

.video-wrapper { position: relative; }
.art-container { width: 100%; background: #000; height: 100vh; }

@media (max-width: #{$breakpoint}) {
  .art-container { height: calc(100vh - #{$topbar-h}); }
}

.cast-popup {
  position: absolute;
  bottom: 80px;
  right: 16px;
  width: 220px;
  background: rgba(20, 20, 20, 0.95);
  border-radius: 8px;
  padding: 12px;
  z-index: 200;
  color: #fff;
  box-shadow: 0 4px 16px rgba(0,0,0,0.4);
}
.cast-popup-title { font-size: 12px; font-weight: 600; color: #aaa; margin-bottom: 8px; }
.cast-popup-msg { font-size: 13px; color: #aaa; text-align: center; padding: 8px 0; }
.cast-device-list { list-style: none; margin: 0; padding: 0; }
.cast-device-item {
  display: flex; align-items: center; gap: 8px; padding: 8px 6px;
  font-size: 13px; border-radius: 4px; cursor: pointer;
  &:hover { background: rgba(255,255,255,0.1); }
}
.cast-device-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.cast-popup-close { margin-top: 8px; font-size: 12px; color: #888; text-align: center; cursor: pointer; &:hover { color: #fff; } }
</style>