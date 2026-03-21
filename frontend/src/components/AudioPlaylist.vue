<template>
  <el-drawer
    v-model="store.showPlaylist"
    title="播放列表"
    direction="rtl"
    size="320px"
    :append-to-body="true"
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
          <el-icon v-if="idx === store.currentIndex && store.isPlaying" class="playing-icon">
            <VideoCamera v-if="getCategory(track.ext) === 'video'" />
            <Headset v-else />
          </el-icon>
          <el-icon v-else class="idle-icon">
            <VideoCamera v-if="getCategory(track.ext) === 'video'" />
            <Headset v-else />
          </el-icon>
        </div>
        <div class="item-info">
          <span class="item-name" :title="track.name">{{ track.name }}</span>
          <span class="item-size">{{ formatSize(track.size) }}</span>
        </div>
        <div v-if="idx === store.currentIndex" class="item-indicator" />
      </div>

      <el-empty v-if="store.playlist.length === 0" description="播放列表为空" />
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { Headset, VideoCamera } from '@element-plus/icons-vue'
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

  &:hover {
    background: var(--el-fill-color-light);
  }

  &.active {
    background: var(--el-color-primary-light-9);
    border-left: 3px solid var(--el-color-primary);
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

.playing-icon {
  color: var(--el-color-primary);
  animation: pulse 1.2s ease-in-out infinite;
}

.idle-icon {
  color: var(--el-text-color-secondary);
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
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.active .item-name {
  color: var(--el-color-primary);
}

.item-size {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}
</style>
