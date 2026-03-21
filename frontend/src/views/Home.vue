<template>
  <div class="home-page">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="logo">
        <span class="logo-icon">🍵</span>
        <span class="logo-text">Tea Cloud</span>
      </div>
      <el-divider style="margin: 8px 0" />
      <PeerList />
    </aside>

    <!-- Main content -->
    <main class="main-content">
      <!-- Upload zone -->
      <el-card class="upload-card" shadow="never">
        <FileUpload />
      </el-card>

      <!-- File explorer -->
      <el-card class="explorer-card" shadow="never">
        <FileExplorer @preview="openPreview" />
      </el-card>
    </main>

    <!-- Preview dialog -->
    <FilePreview :item="previewItem" @close="previewItem = null" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import FileExplorer from '@/components/FileExplorer.vue'
import FileUpload from '@/components/FileUpload.vue'
import FilePreview from '@/components/FilePreview.vue'
import PeerList from '@/components/PeerList.vue'
import { type FileInfo } from '@/services/api'
import { useFileStore } from '@/stores/file'

const store = useFileStore()
const previewItem = ref<FileInfo | null>(null)

function openPreview(item: FileInfo) {
  previewItem.value = item
}

onMounted(() => {
  store.loadDir('.')
})
</script>

<style scoped lang="scss">
.home-page {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: var(--el-bg-color-page);
}

.sidebar {
  width: 220px;
  flex-shrink: 0;
  background: var(--el-bg-color);
  border-right: 1px solid var(--el-border-color-light);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
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
  color: var(--el-text-color-primary);
}

.main-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.upload-card {
  flex-shrink: 0;
}

.explorer-card {
  flex: 1;
  min-height: 0;
}
</style>
