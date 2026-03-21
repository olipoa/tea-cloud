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
import { fileApi } from "@/services/api";
import { useMediaPlayerStore } from "@/stores/audioPlayer";
import APlayer from "@worstone/vue-aplayer";
import "@worstone/vue-aplayer/dist/style.css";
import { computed, nextTick, ref, watch } from "vue";

const store = useMediaPlayerStore();
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
</style>
