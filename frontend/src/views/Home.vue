<template>
  <div class="home-page">
    <!-- Top bar -->
    <header class="topbar">
      <!-- Left: Tea Cloud logo + node selector -->
      <div class="topbar-left">
        <span class="topbar-logo">🍵</span>
        <span class="topbar-title">Tea Cloud</span>
        <t-select
          v-model="selectedNodeValue"
          :options="nodeOptions"
          :loading="nodeLoading"
          size="small"
          style="width: 180px; margin-left: 12px"
          placeholder="选择节点"
          @change="onNodeChange"
        />
      </div>

      <!-- Right: transfer badge + dark mode toggle -->
      <div class="topbar-right">
        <t-badge :count="transferStore.runningCount" :offset="[-4, 4]">
          <t-button
            variant="text"
            shape="square"
            @click="$router.push('/transfer')"
            title="传输列表"
          >
            <template #icon><t-icon name="swap" /></template>
          </t-button>
        </t-badge>
        <t-switch :value="isDark" size="small" @change="toggleDark" />
      </div>
    </header>

    <!-- Main content -->
    <main
      class="main-content"
      :class="{ 'has-player': audioStore.playlist.length > 0 }"
    >
      <!-- Upload trigger row -->
      <div class="upload-row">
        <FileUpload />
      </div>
      <!-- File explorer -->
      <FileExplorer @preview="openPreview" />
    </main>

    <!-- Preview dialog (non-audio/image) -->
    <FilePreview :item="previewItem" @close="previewItem = null" />

    <!-- Fullscreen comic viewer -->
    <ComicViewer
      v-if="comicOpen"
      :files="comicFiles"
      :initial-index="comicIndex"
      @close="comicOpen = false"
    />

    <!-- Player shell (audio only — bottom bar) -->
    <Transition name="player-slide">
      <div
        v-if="audioStore.playlist.length > 0 && !audioStore.isVideo"
        class="player-shell"
      >
        <AudioPlayer />
      </div>
    </Transition>

    <!-- Video player (full-screen overlay, higher z-index) -->
    <VideoPlayer v-if="audioStore.playlist.length > 0 && audioStore.isVideo" />

    <!-- Audio playlist drawer -->
    <AudioPlaylist />
  </div>
</template>

<script setup lang="ts">
import AudioPlayer from "@/components/AudioPlayer.vue";
import AudioPlaylist from "@/components/AudioPlaylist.vue";
import ComicViewer from "@/components/ComicViewer.vue";
import FileExplorer from "@/components/FileExplorer.vue";
import FilePreview from "@/components/FilePreview.vue";
import FileUpload from "@/components/FileUpload.vue";
import VideoPlayer from "@/components/VideoPlayer.vue";
import { useTheme } from "@/composables/useTheme";
import {
  type FileInfo,
  peerApi,
  type PeerInfo,
  type SelfInfo,
  setBaseUrl,
} from "@/services/api";
import { useMediaPlayerStore } from "@/stores/audioPlayer";
import { useFileStore } from "@/stores/file";
import { useTransferStore } from "@/stores/transfer";
import { getCategory } from "@/utils/fileUtils";
import { computed, onBeforeUnmount, onMounted, ref } from "vue";

const store = useFileStore();
const audioStore = useMediaPlayerStore();
const transferStore = useTransferStore();
const { isDark, toggleDark } = useTheme();

const previewItem = ref<FileInfo | null>(null);
const comicOpen = ref(false);
const comicFiles = ref<FileInfo[]>([]);
const comicIndex = ref(0);

// ---- Node selector ----
const peers = ref<PeerInfo[]>([]);
const selfInfo = ref<SelfInfo | null>(null);
const nodeLoading = ref(false);
const selectedNodeValue = ref<string>("self");

let nodeTimer: ReturnType<typeof setInterval> | null = null;

const nodeOptions = computed(() => {
  const opts: { label: string; value: string }[] = [];
  if (selfInfo.value) {
    opts.push({
      label: `${selfInfo.value.name} (本机)`,
      value: "self",
    });
  }
  for (const p of peers.value) {
    opts.push({
      label: `${p.name} (${p.addrV4})`,
      value: p.url,
    });
  }
  return opts;
});

async function refreshNodes() {
  nodeLoading.value = true;
  try {
    const [p, s] = await Promise.all([peerApi.list(), peerApi.self()]);
    selfInfo.value = s;
    peers.value = p.filter((peer) => !s.ips.includes(peer.addrV4));
  } catch {
    /* ignore */
  } finally {
    nodeLoading.value = false;
  }
}

function onNodeChange(val: string) {
  if (val === "self") {
    setBaseUrl("");
  } else {
    setBaseUrl(val);
  }
  store.loadDir(".");
}

// ---- Preview ----
function openPreview(item: FileInfo, siblings?: FileInfo[]) {
  const sourceItems = siblings ?? store.sortedItems;
  const cat = getCategory(item.ext);
  if (cat === "audio" || cat === "video") {
    const mediaFiles = sourceItems.filter(
      (f) => !f.isDir && getCategory(f.ext) === cat,
    );
    const startIndex = mediaFiles.findIndex((f) => f.path === item.path);
    audioStore.setPlaylist(mediaFiles, startIndex >= 0 ? startIndex : 0);
  } else if (cat === "image") {
    const imageFiles = sourceItems.filter(
      (f) => !f.isDir && getCategory(f.ext) === "image",
    );
    const startIndex = imageFiles.findIndex((f) => f.path === item.path);
    comicFiles.value = imageFiles;
    comicIndex.value = startIndex >= 0 ? startIndex : 0;
    comicOpen.value = true;
  } else {
    previewItem.value = item;
  }
}

onMounted(() => {
  store.loadDir(".");
  refreshNodes();
  nodeTimer = setInterval(refreshNodes, 30000);
});

onBeforeUnmount(() => {
  if (nodeTimer) clearInterval(nodeTimer);
});
</script>

<style scoped lang="scss">
$topbar-h: 52px;
$breakpoint: 768px;

.home-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  background: var(--td-bg-color-page);
}

/* ====== TOP BAR ====== */
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: $topbar-h;
  padding: 0 16px;
  background: var(--td-bg-color-container);
  border-bottom: 1px solid var(--td-component-stroke);
  flex-shrink: 0;
  z-index: 100;
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
  flex: 1;
}

.topbar-logo {
  font-size: 1.4rem;
  flex-shrink: 0;
}

.topbar-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--td-text-color-primary);
  flex-shrink: 0;
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
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

.upload-row {
  flex-shrink: 0;
}

/* ====== PLAYER SHELL ====== */
.player-shell {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 400;
  background: var(--td-bg-color-container);
  border-top: 1px solid var(--td-component-stroke);
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

/* ====== RESPONSIVE ====== */
@media (max-width: #{$breakpoint}) {
  .topbar-title {
    display: none;
  }

  .main-content {
    padding: 12px;
  }
}
</style>
