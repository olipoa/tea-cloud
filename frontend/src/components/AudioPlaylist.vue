<template>
  <n-drawer v-model:show="store.showPlaylist" :width="320" placement="right">
    <n-drawer-content title="播放列表" :native-scrollbar="false">
      <div class="playlist">
        <div
          v-for="(track, idx) in store.playlist"
          :key="track.path"
          class="playlist-item"
          :class="{ active: idx === store.currentIndex }"
          @click="store.playAt(idx)"
        >
          <div class="item-icon">
            <span v-if="idx === store.currentIndex && store.isPlaying" class="playing-icon">
              {{ getCategory(track.ext) === 'video' ? '🎬' : '🎵' }}
            </span>
            <span v-else class="idle-icon">
              {{ getCategory(track.ext) === 'video' ? '🎬' : '🎵' }}
            </span>
          </div>
          <div class="item-info">
            <span class="item-name" :title="track.name">{{ track.name }}</span>
            <span class="item-size">{{ formatSize(track.size) }}</span>
          </div>
          <div v-if="idx === store.currentIndex" class="item-indicator" />
        </div>

        <div v-if="store.playlist.length === 0" class="empty-hint">播放列表为空</div>
      </div>
    </n-drawer-content>
  </n-drawer>
</template>

<script setup lang="ts">
import { NDrawer, NDrawerContent } from 'naive-ui'
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

  &:hover { background: #f5f5f7; }

  &.active {
    background: #e8f5ee;
    border-left: 3px solid #18a058;
    padding-left: 9px;
  }
}

.item-icon {
  flex-shrink: 0;
  font-size: 1.2rem;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.playing-icon { animation: pulse 1.2s ease-in-out infinite; }
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.item-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.item-name {
  font-size: 13px;
  font-weight: 500;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.active .item-name { color: #18a058; }

.item-size {
  font-size: 11px;
  color: #aaa;
}

.item-indicator {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #18a058;
  flex-shrink: 0;
}

.empty-hint {
  text-align: center;
  color: #aaa;
  padding: 32px 16px;
  font-size: 13px;
}
</style>
