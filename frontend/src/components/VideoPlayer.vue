<template>
  <div class="video-wrapper">
    <div ref="artRef" class="art-container" />

    <!-- Cast device popover anchored to bottom-right area -->
    <n-popover
      v-model:show="castPopoverVisible"
      trigger="manual"
      placement="top-end"
      :width="220"
    >
      <template #trigger>
        <span class="cast-anchor" />
      </template>
      <div class="cast-popover">
        <div class="cast-popover-title">选择投屏设备</div>
        <div v-if="castLoading" class="cast-popover-msg">正在扫描...</div>
        <div v-else-if="castDevices.length === 0" class="cast-popover-msg">
          未发现 DLNA 设备
        </div>
        <ul v-else class="cast-device-list">
          <li
            v-for="d in castDevices"
            :key="d.location"
            class="cast-device-item"
            :title="d.name"
            @click="castTo(d)"
          >
            <span>📺</span>
            <span class="cast-device-name">{{ d.name }}</span>
          </li>
        </ul>
      </div>
    </n-popover>
  </div>
</template>

<script setup lang="ts">
import { useCast } from "@/composables/useCast";
import { fileApi } from "@/services/api";
import { useMediaPlayerStore } from "@/stores/audioPlayer";
import Artplayer from "artplayer";
import { NPopover } from "naive-ui";
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";

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
    playbackRate: false,
    aspectRatio: false,
    setting: false,
    hotkey: true,
    screenshot: false,
    lock: true,
    isLive: false,
    theme: "#18a058",
    controls: [
      {
        name: "prev",
        position: "right",
        html: `<svg viewBox="0 0 24 24" width="22" height="22" fill="#fff"><path d="M6 6h2v12H6zm3.5 6 8.5 6V6z"/></svg>`,
        tooltip: "上一首",
        click: () => {
          if (store.hasPrev) store.prev();
        },
      },
      {
        name: "next",
        position: "right",
        html: `<svg viewBox="0 0 24 24" width="22" height="22" fill="#fff"><path d="M6 18l8.5-6L6 6v12zM16 6v12h2V6h-2z"/></svg>`,
        tooltip: "下一首",
        click: () => {
          if (store.hasNext) store.next();
        },
      },
      {
        name: "cast",
        position: "right",
        html: `<svg viewBox="0 0 24 24" width="22" height="22" fill="#fff"><path d="M1 18v3h3c0-1.66-1.34-3-3-3zm0-4v2c2.76 0 5 2.24 5 5h2c0-3.87-3.13-7-7-7zm18-7H5v1.63c3.96 1.28 7.09 4.41 8.37 8.37H19V7zM1 10v2c4.97 0 9 4.03 9 9h2c0-6.08-4.93-11-11-11zm20-7H3c-1.1 0-2 .9-2 2v3h2V5h18v14h-7v2h7c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2z"/></svg>`,
        tooltip: "投屏",
        click: () => openCastPopover(),
      },
      {
        name: "playlist",
        position: "right",
        html: `<svg viewBox="0 0 24 24" width="22" height="22" fill="#fff"><path d="M3 13h2v-2H3zm0 4h2v-2H3zm0-8h2V7H3zm4 4h14v-2H7zm0 4h14v-2H7zM7 7v2h14V7z"/></svg>`,
        tooltip: "播放列表",
        click: () => store.togglePlaylist(),
      },
      {
        name: "close-player",
        position: "right",
        html: `<svg viewBox="0 0 24 24" width="22" height="22" fill="#fff"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>`,
        tooltip: "关闭播放器",
        click: () => store.setPlaylist([], 0),
      },
    ],
  });

  art.on("ready", () => {
    store.setDuration(art!.duration);
    if (store.currentTime > 0) art!.currentTime = store.currentTime;
  });
  art.on("video:timeupdate", () => {
    store.setCurrentTime(art!.currentTime);
  });
  art.on("video:durationchange", () => {
    store.setDuration(art!.duration);
  });
  art.on("play", () => store.setIsPlaying(true));
  art.on("pause", () => store.setIsPlaying(false));
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

// 切换曲目 URL
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

// 播放/暂停同步
watch(
  () => store.isPlaying,
  (playing) => {
    if (!art) return;
    if (playing) art.play().catch?.(() => {});
    else art.pause();
  },
);

// artRef 出现/消失时创建/销毁
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

// 首次加载播放列表时
watch(
  () => store.playlist.length,
  async (len) => {
    if (len > 0 && artRef.value && !art) {
      await nextTick();
      createArtPlayer();
    }
  },
);

function togglePlayPause() {
  if (!art) return;
  if (art.playing) art.pause();
  else art.play().catch(() => {});
}

defineExpose({ togglePlayPause });

onBeforeUnmount(() => {
  if (art) {
    art.destroy();
    art = null;
  }
});
</script>

<style scoped lang="scss">
$topbar-h: 48px;
$breakpoint: 768px;

.video-wrapper {
  position: relative;
}

.art-container {
  width: 100%;
  background: #000;
  height: 100vh;
}

/* Invisible anchor element for cast popover positioning */
.cast-anchor {
  position: absolute;
  bottom: 60px;
  right: 130px;
  width: 1px;
  height: 1px;
  pointer-events: none;
}

@media (max-width: #{$breakpoint}) {
  .art-container {
    height: calc(100vh - #{$topbar-h});
  }
}

/* Cast popover content */
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

.cast-device-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
