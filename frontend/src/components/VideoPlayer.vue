<template>
  <Teleport to="body">
    <div
      class="video-overlay"
      @mousemove="showControls"
      @touchstart.passive="showControls"
    >
      <div ref="artRef" class="art-container" />

      <Transition name="ctrl-fade">
        <div v-show="ctrlVisible" class="top-bar">
          <div class="top-left">
            <button
              class="top-btn"
              @click="store.setPlaylist([], 0)"
              title="关闭"
            >
              <svg viewBox="0 0 24 24" width="22" height="22" fill="#fff">
                <path
                  d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"
                />
              </svg>
            </button>
            <button class="top-btn" @click="openCastPopover" title="投屏">
              <svg viewBox="0 0 24 24" width="22" height="22" fill="#fff">
                <path
                  d="M1 18v3h3c0-1.66-1.34-3-3-3zm0-4v2c2.76 0 5 2.24 5 5h2c0-3.87-3.13-7-7-7zm18-7H5v1.63c3.96 1.28 7.09 4.41 8.37 8.37H19V7zM1 10v2c4.97 0 9 4.03 9 9h2c0-6.08-4.93-11-11-11zm20-7H3c-1.1 0-2 .9-2 2v3h2V5h18v14h-7v2h7c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2z"
                />
              </svg>
            </button>
          </div>
          <div class="top-right">
            <button
              class="top-btn"
              :class="{ disabled: !store.hasPrev }"
              @click="store.hasPrev && store.prev()"
              title="上一个"
            >
              <svg viewBox="0 0 24 24" width="22" height="22" fill="#fff">
                <path d="M6 6h2v12H6zm3.5 6 8.5 6V6z" />
              </svg>
            </button>
            <button
              class="top-btn"
              :class="{ disabled: !store.hasNext }"
              @click="store.hasNext && store.next()"
              title="下一个"
            >
              <svg viewBox="0 0 24 24" width="22" height="22" fill="#fff">
                <path d="M6 18l8.5-6L6 6v12zM16 6v12h2V6h-2z" />
              </svg>
            </button>
            <button
              class="top-btn"
              @click="store.togglePlaylist()"
              title="播放列表"
            >
              <svg viewBox="0 0 24 24" width="22" height="22" fill="#fff">
                <path
                  d="M3 13h2v-2H3zm0 4h2v-2H3zm0-8h2V7H3zm4 4h14v-2H7zm0 4h14v-2H7zM7 7v2h14V7z"
                />
              </svg>
            </button>
          </div>
        </div>
      </Transition>

      <!-- Cast device popup -->
      <div v-show="castPopoverVisible" class="cast-popup">
        <div class="cast-popup-title">选择投屏设备</div>
        <div v-if="castLoading" class="cast-popup-msg">正在扫描...</div>
        <div v-else-if="castDevices.length === 0" class="cast-popup-msg">
          未发现 DLNA 设备
        </div>
        <ul v-else class="cast-device-list">
          <li
            v-for="d in castDevices"
            :key="d.location"
            class="cast-device-item"
            @click="
              castTo(d);
              castPopoverVisible = false;
            "
          >
            <t-icon name="tv" />
            <span class="cast-device-name">{{ d.name }}</span>
          </li>
        </ul>
        <div class="cast-popup-close" @click="castPopoverVisible = false">
          关闭
        </div>
      </div>

      <!-- Playlist side panel -->
      <VideoPlaylistPanel
        :visible="store.showPlaylist"
        @close="store.togglePlaylist()"
      />
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { useCast } from "@/composables/useCast";
import { fileApi } from "@/services/api";
import { useMediaPlayerStore } from "@/stores/audioPlayer";
import Artplayer from "artplayer";
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";
import VideoPlaylistPanel from "./VideoPlaylistPanel.vue";

const store = useMediaPlayerStore();
const {
  castPopoverVisible,
  castLoading,
  castDevices,
  openCastPopover,
  castTo,
} = useCast();
const artRef = ref<HTMLDivElement | null>(null);
let art: Artplayer | null = null;

const ctrlVisible = ref(true);
let ctrlHideTimer: ReturnType<typeof setTimeout> | null = null;

function showControls() {
  ctrlVisible.value = true;
  if (ctrlHideTimer) clearTimeout(ctrlHideTimer);
  ctrlHideTimer = setTimeout(() => {
    ctrlVisible.value = false;
  }, 3000);
}

const currentUrl = computed(() =>
  store.currentTrack ? fileApi.rawUrl(store.currentTrack.path) : "",
);

function getMimeType(ext: string): string {
  const map: Record<string, string> = {
    mp4: "video/mp4",
    webm: "video/webm",
    ogg: "video/ogg",
    mkv: "video/x-matroska",
    avi: "video/x-msvideo",
    mov: "video/quicktime",
    mp3: "audio/mpeg",
    flac: "audio/flac",
    wav: "audio/wav",
    aac: "audio/aac",
    m4a: "audio/mp4",
    opus: "audio/ogg",
  };
  return map[ext.toLowerCase()] ?? "";
}

function createArtPlayer() {
  if (!artRef.value || !currentUrl.value) return;
  if (art) {
    art.destroy();
    art = null;
  }

  const ext = store.currentTrack?.ext ?? "";
  const track = store.currentTrack;
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
    playbackRate: true,
    aspectRatio: false,
    setting: false,
    hotkey: true,
    screenshot: false,
    lock: true,
    isLive: false,
    theme: "#18a058",
  });

  art.on("ready", () => {
    store.setDuration(art!.duration);
    if (store.currentTime > 0) art!.currentTime = store.currentTime;
    showControls();
  });
  art.on("video:timeupdate", () => {
    store.setCurrentTime(art!.currentTime);
  });
  art.on("video:durationchange", () => {
    store.setDuration(art!.duration);
  });
  art.on("play", () => {
    store.setIsPlaying(true);
    showControls();
  });
  art.on("pause", () => {
    store.setIsPlaying(false);
    showControls();
  });
  art.on("video:ended", () => {
    if (store.hasNext) store.next();
    else {
      store.setIsPlaying(false);
      store.setCurrentTime(0);
    }
  });
  art.on("error", () => {
    if (store.hasNext) store.next();
  });
}

watch(currentUrl, async (url) => {
  if (!url) return;
  if (art) {
    art.switchUrl(url);
    if (store.isPlaying) {
      await nextTick();
      art.play().catch(() => {});
    }
  } else {
    await nextTick();
    createArtPlayer();
  }
});

watch(
  () => store.isPlaying,
  (playing) => {
    if (!art) return;
    if (playing) art.play().catch?.(() => {});
    else art.pause();
  },
);

watch(artRef, async (el) => {
  if (!el) {
    if (art) {
      art.destroy();
      art = null;
    }
  } else if (currentUrl.value) {
    await nextTick();
    createArtPlayer();
  }
});

watch(
  () => store.playlist.length,
  async (len) => {
    if (len > 0 && artRef.value && !art) {
      await nextTick();
      createArtPlayer();
    }
  },
);

onBeforeUnmount(() => {
  if (ctrlHideTimer) clearTimeout(ctrlHideTimer);
  if (art) {
    art.destroy();
    art = null;
  }
});
</script>

<style scoped lang="scss">
.video-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: #000;
}
.art-container {
  width: 100%;
  height: 100%;
}
.top-bar {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: linear-gradient(
    to bottom,
    rgba(0, 0, 0, 0.6) 0%,
    transparent 100%
  );
  pointer-events: none;
}
.top-left,
.top-right {
  display: flex;
  align-items: center;
  gap: 4px;
  pointer-events: auto;
}
.top-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  padding: 0;
  background: transparent;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  opacity: 0.85;
  transition:
    opacity 0.15s,
    background 0.15s;
  &:hover {
    opacity: 1;
    background: rgba(255, 255, 255, 0.15);
  }
  &.disabled {
    opacity: 0.3;
    cursor: default;
    &:hover {
      background: transparent;
    }
  }
}
.ctrl-fade-enter-active,
.ctrl-fade-leave-active {
  transition: opacity 0.3s;
}
.ctrl-fade-enter-from,
.ctrl-fade-leave-to {
  opacity: 0;
}
.cast-popup {
  position: absolute;
  top: 56px;
  left: 12px;
  width: 220px;
  background: rgba(20, 20, 20, 0.95);
  border-radius: 8px;
  padding: 12px;
  z-index: 20;
  color: #fff;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
}
.cast-popup-title {
  font-size: 12px;
  font-weight: 600;
  color: #aaa;
  margin-bottom: 8px;
}
.cast-popup-msg {
  font-size: 13px;
  color: #aaa;
  text-align: center;
  padding: 8px 0;
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
  &:hover {
    background: rgba(255, 255, 255, 0.1);
  }
}
.cast-device-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.cast-popup-close {
  margin-top: 8px;
  font-size: 12px;
  color: #888;
  text-align: center;
  cursor: pointer;
  &:hover {
    color: #fff;
  }
}
</style>
