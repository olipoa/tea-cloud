<template>
  <div class="home-page">
    <!-- Desktop sidebar (>768px) -->
    <aside class="sidebar" :class="{ open: drawerOpen }">
      <div class="sidebar-inner">
        <div class="logo">
          <span class="logo-icon">🍵</span>
          <span class="logo-text">Tea Cloud</span>
        </div>
        <n-divider style="margin: 8px 0" />
        <PeerList />
      </div>
    </aside>

    <!-- Mobile overlay -->
    <div v-if="drawerOpen" class="drawer-overlay" @click="drawerOpen = false" />

    <!-- Main content -->
    <div class="main-wrapper">
      <!-- Mobile top bar -->
      <header class="topbar">
        <button
          class="hamburger"
          @click="drawerOpen = !drawerOpen"
          aria-label="菜单"
        >
          <svg viewBox="0 0 24 24" fill="currentColor" width="24" height="24">
            <path d="M3 6h18v2H3V6zm0 5h18v2H3v-2zm0 5h18v2H3v-2z" />
          </svg>
        </button>
        <div class="topbar-logo">
          <span>🍵</span>
          <span class="topbar-title">Tea Cloud</span>
        </div>
      </header>

      <main
        class="main-content"
        :class="{ 'has-player': audioStore.playlist.length > 0 }"
      >
        <n-card class="upload-card" :bordered="false">
          <FileUpload />
        </n-card>
        <n-card class="explorer-card" :bordered="false">
          <FileExplorer @preview="openPreview" />
        </n-card>
      </main>
    </div>

    <!-- Preview dialog (non-audio) -->
    <FilePreview :item="previewItem" @close="previewItem = null" />

    <!-- Player shell (persistent bottom bar) -->
    <Transition name="player-slide">
      <div v-if="audioStore.playlist.length > 0" class="player-shell">
        <AudioPlayer v-if="!audioStore.isVideo" />
        <VideoPlayer v-else />
      </div>
    </Transition>

    <!-- Audio playlist drawer -->
    <AudioPlaylist />
  </div>
</template>

<script setup lang="ts">
import AudioPlayer from "@/components/AudioPlayer.vue";
import AudioPlaylist from "@/components/AudioPlaylist.vue";
import FileExplorer from "@/components/FileExplorer.vue";
import FilePreview from "@/components/FilePreview.vue";
import FileUpload from "@/components/FileUpload.vue";
import PeerList from "@/components/PeerList.vue";
import VideoPlayer from "@/components/VideoPlayer.vue";
import { type FileInfo } from "@/services/api";
import { useMediaPlayerStore } from "@/stores/audioPlayer";
import { useFileStore } from "@/stores/file";
import { getCategory } from "@/utils/fileUtils";
import { NCard, NDivider } from "naive-ui";
import { onMounted, ref } from "vue";

const store = useFileStore();
const audioStore = useMediaPlayerStore();
const previewItem = ref<FileInfo | null>(null);
const drawerOpen = ref(false);

function openPreview(item: FileInfo) {
  const cat = getCategory(item.ext);
  if (cat === "audio" || cat === "video") {
    const mediaFiles = store.items.filter(
      (f) => !f.isDir && getCategory(f.ext) === cat,
    );
    const startIndex = mediaFiles.findIndex((f) => f.path === item.path);
    audioStore.setPlaylist(mediaFiles, startIndex >= 0 ? startIndex : 0);
  } else {
    previewItem.value = item;
  }
}

onMounted(() => {
  store.loadDir(".");
});
</script>

<style scoped lang="scss">
$sidebar-w: 220px;
$topbar-h: 48px;
$breakpoint: 768px;

.home-page {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: #f5f5f7;
}

/* ====== SIDEBAR ====== */
.sidebar {
  width: $sidebar-w;
  flex-shrink: 0;
  background: #fff;
  border-right: 1px solid #efeff5;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  z-index: 300;
  transition: transform 0.25s ease;
}

.sidebar-inner {
  padding: 12px 0;
}

.logo {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 16px 8px;
}

.logo-icon {
  font-size: 1.6rem;
}
.logo-text {
  font-size: 18px;
  font-weight: 700;
  color: #333;
}

/* ====== MAIN WRAPPER ====== */
.main-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

/* ====== TOP BAR (hidden on desktop) ====== */
.topbar {
  display: none;
  align-items: center;
  gap: 12px;
  height: $topbar-h;
  padding: 0 16px;
  background: #fff;
  border-bottom: 1px solid #efeff5;
  flex-shrink: 0;
  z-index: 10;
}

.hamburger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  color: #333;
  &:hover {
    background: #f0f0f0;
  }
}

.topbar-logo {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 700;
  color: #333;
}

.topbar-title {
  font-size: 16px;
}

/* ====== MAIN CONTENT ====== */
.main-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
  transition: padding-bottom 0.3s ease;

  &.has-player {
    padding-bottom: 120px;
  }
}

/* ====== PLAYER SHELL ====== */
.player-shell {
  position: fixed;
  bottom: 0;
  left: $sidebar-w;
  right: 0;
  z-index: 200;
  background: #fff;
  border-top: 1px solid #efeff5;
  box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.08);
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

.upload-card {
  flex-shrink: 0;
}

/* ====== MOBILE OVERLAY ====== */
.drawer-overlay {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  z-index: 299;
}

/* ====== RESPONSIVE ====== */
@media (max-width: #{$breakpoint}) {
  .sidebar {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    transform: translateX(-100%);
    box-shadow: 4px 0 16px rgba(0, 0, 0, 0.15);

    &.open {
      transform: translateX(0);
    }
  }

  .drawer-overlay {
    display: block;
  }

  .topbar {
    display: flex;
    position: relative;
    z-index: 250; // above media-player-bar (200) so hamburger always accessible
  }

  .player-shell {
    left: 0; // full width on mobile
  }

  .main-content {
    padding: 12px;
  }
}
</style>
