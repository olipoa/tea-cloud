<template>
  <div class="aplayer-wrapper">
    <APlayer
      ref="aplayerRef"
      :audio="aplayerList"
      :autoplay="true"
      loop="all"
      order="list"
      theme="#18a058"
      :listFolded="store.playlist.length <= 1"
      @play="() => store.setIsPlaying(true)"
      @pause="() => store.setIsPlaying(false)"
      @timeupdate="onAPlayerTimeupdate"
      @durationchange="onAPlayerDurationChange"
      @listswitch="onAPlayerListSwitch"
    />
  </div>
</template>

<script setup lang="ts">
import { useCast } from "@/composables/useCast";
import { fileApi } from "@/services/api";
import { useMediaPlayerStore } from "@/stores/audioPlayer";
import APlayer from "@worstone/vue-aplayer";
import "@worstone/vue-aplayer/dist/style.css";
import { computed, nextTick, ref, watch } from "vue";

const store = useMediaPlayerStore();
const {
  castPopoverVisible,
  castLoading,
  castDevices,
  openCastPopover,
  castTo,
} = useCast();
const aplayerRef = ref<any>(null);
let aplayerSwitching = false;

const aplayerList = computed(() =>
  store.playlist.map((f) => ({
    name: f.name,
    url: fileApi.rawUrl(f.path),
    artist: "",
  })),
);

function onAPlayerListSwitch(idx: number) {
  aplayerSwitching = true;
  store.playAt(idx);
  nextTick(() => {
    aplayerSwitching = false;
  });
}

function onAPlayerTimeupdate(e: Event) {
  const audio = e.target as HTMLAudioElement;
  store.setCurrentTime(audio.currentTime);
}

function onAPlayerDurationChange(e: Event) {
  const audio = e.target as HTMLAudioElement;
  if (audio.duration && isFinite(audio.duration)) {
    store.setDuration(audio.duration);
  }
}

// Sync store.currentIndex → APlayer（如侧边栏切换曲目时由外部修改索引）
watch(
  () => store.currentIndex,
  (idx) => {
    if (aplayerSwitching) return;
    aplayerRef.value?.switchList(idx);
  },
);

// Sync store.isPlaying → APlayer
watch(
  () => store.isPlaying,
  (playing) => {
    if (!aplayerRef.value) return;
    if (playing) aplayerRef.value.play();
    else aplayerRef.value.pause();
  },
);
</script>

<style scoped lang="scss">
.aplayer-wrapper {
  :deep(.aplayer) {
    margin: 0;
    border-radius: 0;
    box-shadow: none;
  }
}

.player-extras {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;
  padding: 0 12px;
  height: 36px;
  border-top: 1px solid #efeff5;
  background: #fff;
}

.bar-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border: none;
  background: transparent;
  border-radius: 50%;
  cursor: pointer;
  color: #555;
  padding: 0;
  transition:
    background 0.15s,
    color 0.15s;

  &:hover {
    background: rgba(0, 0, 0, 0.07);
    color: #18a058;
  }

  &.active {
    color: #18a058;
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
