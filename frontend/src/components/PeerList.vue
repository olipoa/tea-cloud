<template>
  <div class="peer-list">
    <div class="peer-header">
      <span class="label">局域网节点</span>
      <el-button :icon="Refresh" size="small" :loading="loading" @click="refresh" link />
    </div>

    <div v-if="loading && peers.length === 0" class="scanning">
      <el-icon class="is-loading"><Loading /></el-icon> 扫描中…
    </div>

    <div v-else-if="peers.length === 0" class="empty">
      未发现其他节点
    </div>

    <el-scrollbar max-height="240px" v-else>
      <div
        v-for="peer in peers"
        :key="peer.url"
        class="peer-item"
        :class="{ active: peer.url === currentUrl }"
        @click="switchTo(peer)"
        :title="`${peer.name}\n${peer.url}`"
      >
        <el-icon><Monitor /></el-icon>
        <div class="peer-info">
          <div class="peer-name">{{ peer.name }}</div>
          <div class="peer-addr">{{ peer.addrV4 || peer.host }}:{{ peer.port }}</div>
        </div>
        <el-icon v-if="peer.url === currentUrl" class="active-icon"><Check /></el-icon>
      </div>
    </el-scrollbar>

    <!-- This device -->
    <div class="self-section" v-if="self">
      <el-divider content-position="left">本机</el-divider>
      <div class="peer-item" :class="{ active: !currentUrl }" @click="switchToSelf">
        <el-icon><House /></el-icon>
        <div class="peer-info">
          <div class="peer-name">{{ self.name }}</div>
          <div class="peer-addr">{{ self.ips[0] || 'localhost' }}:{{ self.port }}</div>
        </div>
        <el-icon v-if="!currentUrl" class="active-icon"><Check /></el-icon>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { Refresh, Monitor, Check, House, Loading } from '@element-plus/icons-vue'
import { peerApi, setBaseUrl, getBaseUrl, type PeerInfo, type SelfInfo } from '@/services/api'
import { useFileStore } from '@/stores/file'

const peers = ref<PeerInfo[]>([])
const self = ref<SelfInfo | null>(null)
const loading = ref(false)
const currentUrl = ref(getBaseUrl())

const store = useFileStore()

let timer: ReturnType<typeof setInterval> | null = null

async function refresh() {
  loading.value = true
  try {
    const [p, s] = await Promise.all([peerApi.list(), peerApi.self()])
    // Filter out ourselves from peers list
    peers.value = p.filter(peer => {
      if (!s) return true
      return !s.ips.includes(peer.addrV4)
    })
    self.value = s
  } catch {
    // ignore
  } finally {
    loading.value = false
  }
}

function switchTo(peer: PeerInfo) {
  setBaseUrl(peer.url)
  currentUrl.value = peer.url
  store.loadDir('.')
}

function switchToSelf() {
  setBaseUrl('')
  currentUrl.value = ''
  store.loadDir('.')
}

onMounted(() => {
  refresh()
  timer = setInterval(refresh, 30000)
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped lang="scss">
.peer-list {
  padding: 8px 0;
}

.peer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 12px 8px;
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.scanning, .empty {
  padding: 12px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  display: flex;
  align-items: center;
  gap: 6px;
}

.peer-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  cursor: pointer;
  border-radius: 6px;
  margin: 2px 4px;
  transition: background 0.15s;

  &:hover {
    background: var(--el-fill-color-light);
  }

  &.active {
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
  }
}

.peer-info {
  flex: 1;
  min-width: 0;
}

.peer-name {
  font-size: 13px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.peer-addr {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

.active-icon {
  color: var(--el-color-primary);
  font-size: 14px;
}

.self-section {
  margin-top: 4px;
}
</style>
