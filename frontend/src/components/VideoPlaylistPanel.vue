<template>
  <Transition name="playlist-slide">
    <div v-if="visible" class="playlist-panel">
      <div class="playlist-header">
        <span class="playlist-title"
          >播放列表（{{ store.playlist.length }}）</span
        >
        <t-button
          variant="text"
          shape="square"
          size="small"
          @click="emit('close')"
        >
          <template #icon><t-icon name="close" /></template>
        </t-button>
      </div>
      <div class="playlist-body">
        <div
          v-for="(item, i) in store.playlist"
          :key="item.path"
          class="playlist-item"
          :class="{ active: i === store.currentIndex }"
          @click="store.playAt(i)"
        >
          <t-icon
            :name="i === store.currentIndex ? 'play-circle' : 'file-icon'"
            class="item-icon"
          />
          <span class="item-name" :title="item.name">{{ item.name }}</span>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { useMediaPlayerStore } from "@/stores/audioPlayer";

defineProps<{ visible: boolean }>();
const emit = defineEmits<{ (e: "close"): void }>();

const store = useMediaPlayerStore();
</script>

<style scoped lang="scss">
.playlist-panel {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: 260px;
  background: rgba(0, 0, 0, 0.85);
  z-index: 100;
  display: flex;
  flex-direction: column;
  backdrop-filter: blur(4px);
}
.playlist-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  flex-shrink: 0;
}
.playlist-title {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
}
.playlist-body {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}
.playlist-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  cursor: pointer;
  transition: background 0.15s;
  color: rgba(255, 255, 255, 0.75);
  &:hover {
    background: rgba(255, 255, 255, 0.08);
  }
  &.active {
    color: #18a058;
    background: rgba(24, 160, 88, 0.1);
  }
}
.item-icon {
  font-size: 16px;
  flex-shrink: 0;
}
.item-name {
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.playlist-slide-enter-active,
.playlist-slide-leave-active {
  transition: transform 0.25s ease;
}
.playlist-slide-enter-from,
.playlist-slide-leave-to {
  transform: translateX(100%);
}
</style>
