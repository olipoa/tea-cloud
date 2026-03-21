<template>
  <div class="peer-list">
    <div class="peer-header">
      <span class="label">局域网节点</span>
      <button class="refresh-btn" :disabled="loading" @click="refresh" title="刷新">
        <svg viewBox="0 0 24 24" fill="currentColor" width="16" height="16" :class="{ spinning: loading }">
          <path d="M17.65 6.35A7.958 7.958 0 0 0 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08A5.99 5.99 0 0 1 12 18c-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
        </svg>
      </button>
    </div>

    <div v-if="loading && peers.length === 0" class="scanning">扫描中…</div>
    <div v-else-if="peers.length === 0" class="empty">未发现其他节点</div>

    <n-scrollbar style="max-height: 240px" v-else>
      <div
        v-for="peer in peers"
        :key="peer.url"
        class="peer-item"
        :class="{ active: peer.url === currentUrl }"
        @click="switchTo(peer)"
        :title="`${peer.name}\n${peer.url}`"
      >
        <span class="peer-icon">🖥️</span>
        <div class="peer-info">
          <div class="peer-name">{{ peer.name }}</div>
          <div class="peer-addr">{{ peer.addrV4 || peer.host }}:{{ peer.port }}</div>
        </div>
        <span v-if="peer.url === currentUrl" class="active-dot">✓</span>
      </div>
    </n-scrollbar>

    <!-- This device -->
    <div class="self-section" v-if="self">
      <n-divider title-placement="left" style="margin: 8px 0; font-size: 11px;">本机</n-divider>
      <div class="peer-item" :class="{ active: !currentUrl }" @click="switchToSelf">
        <span class="peer-icon">🏠</span>
        <div class="peer-info">
          <div class="peer-name">{{ self.name }}</div>
          <div class="peer-addr">{{ self.ips[0] || 'localhost' }}:{{ self.port }}</div>
        </div>
        <span v-if="!currentUrl" class="active-dot">✓</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { NScrollbar, NDivider } from 'naive-ui'
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
  color: #888;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.refresh-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border: none;
  background: transparent;
  border-radius: 4px;
  cursor: pointer;
  color: #888;
  padding: 0;
  &:hover { background: #f0f0f0; color: #18a058; }
  &:disabled { opacity: 0.4; cursor: not-allowed; }
}

.spinning { animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.scanning, .empty {
  padding: 8px 16px;
  font-size: 13px;
  color: #888;
}

.peer-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  cursor: pointer;
  border-radius: 6px;
  margin: 2px 6px;
  transition: background 0.15s;

  &:hover { background: #f5f5f7; }
  &.active { background: #e8f5ee; }
}

.peer-icon { font-size: 1.1rem; flex-shrink: 0; }

.peer-info {
  flex: 1;
  min-width: 0;
}

.peer-name {
  font-size: 13px;
  font-weight: 500;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.active .peer-name { color: #18a058; }

.peer-addr {
  font-size: 11px;
  color: #aaa;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.active-dot {
  font-size: 12px;
  color: #18a058;
  font-weight: 700;
  flex-shrink: 0;
}

.self-section {
  padding: 0 6px;
}
</style>
