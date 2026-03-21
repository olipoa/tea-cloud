<template>
  <t-drawer
    v-model:visible="store.showPlaylist"
    :footer="false"
    header="播放列表"
    size="320px"
    placement="right"
  >
    <div class="playlist">
      <div
        v-for="(track, idx) in store.playlist"
        :key="track.path"
        class="playlist-item"
        :class="{ active: idx === store.currentIndex }"
        @click="store.playAt(idx)"
      >
        <div class="item-icon">
          <t-icon :name="getCategory(track.ext) === 'video' ? 'film' : 'music'" />
        </div>
        <div class="item-info">
          <span class="item-name" :title="track.name">{{ track.name }}</span>
          <span class="item-size">{{ formatSize(track.size) }}</span>
        </div>
        <div v-if="idx === store.currentIndex" class="item-indicator" />
      </div>
      <div v-if="store.playlist.length === 0" class="empty-hint">播放列表为空</div>
    </div>
  </t-drawer>
</template>

<script setup lang="ts">
import { useMediaPlayerStore } from '@/stores/audioPlayer'
import { formatSize, getCategory } from '@/utils/fileUtils'

const store = useMediaPlayerStore()
</script>

<style scoped lang="scss">
.playlist {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 4px 0;
}
.playlist-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  position: relative;
  transition: background 0.15s;
  &:hover { background: var(--td-bg-color-secondarycontainer); }
  &.active {
    background: var(--td-brand-color-light);
    border-left: 3px solid var(--td-brand-color);
    padding-left: 9px;
  }
}
.item-icon { flex-shrink: 0; color: var(--td-text-color-secondary); }
.item-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.item-name { font-size: 14px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.item-size { font-size: 11px; color: var(--td-text-color-secondary); }
.item-indicator { width: 6px; height: 6px; border-radius: 50%; background: var(--td-brand-color); flex-shrink: 0; }
.empty-hint { text-align: center; color: var(--td-text-color-placeholder); padding: 32px 0; font-size: 13px; }
</style>
