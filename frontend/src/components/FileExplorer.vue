<template>
  <div class="file-explorer">
    <!-- Toolbar -->
    <div class="toolbar">
      <div class="breadcrumb-row">
        <button
          class="icon-btn"
          :disabled="store.currentPath === '.'"
          @click="store.navigateUp()"
          title="返回上级"
        >
          <svg viewBox="0 0 24 24" fill="currentColor" width="18" height="18"><path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/></svg>
        </button>
        <n-breadcrumb class="breadcrumb">
          <n-breadcrumb-item
            v-for="crumb in store.breadcrumbs"
            :key="crumb.path"
            @click="store.navigateTo(crumb.path)"
          >
            {{ crumb.label }}
          </n-breadcrumb-item>
        </n-breadcrumb>
      </div>

      <div class="actions">
        <button class="icon-btn" @click="store.refresh()" :disabled="store.loading" title="刷新">
          <svg viewBox="0 0 24 24" fill="currentColor" width="18" height="18" :class="{ spinning: store.loading }">
            <path d="M17.65 6.35A7.958 7.958 0 0 0 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08A5.99 5.99 0 0 1 12 18c-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/>
          </svg>
        </button>
        <n-button size="small" @click="showMkdir = true">
          <template #icon>
            <svg viewBox="0 0 24 24" fill="currentColor" width="16" height="16"><path d="M20 6h-8l-2-2H4c-1.11 0-2 .89-2 2v12c0 1.11.89 2 2 2h16c1.11 0 2-.89 2-2V8c0-1.11-.89-2-2-2zm-1 8h-3v3h-2v-3h-3v-2h3V9h2v3h3v2z"/></svg>
          </template>
          新建文件夹
        </n-button>
        <n-button-group size="small">
          <n-button :type="store.viewMode === 'list' ? 'primary' : 'default'" @click="store.viewMode = 'list'" title="列表视图">
            <svg viewBox="0 0 24 24" fill="currentColor" width="16" height="16"><path d="M3 13h2v-2H3zm0 4h2v-2H3zm0-8h2V7H3zm4 4h14v-2H7zm0 4h14v-2H7zM7 7v2h14V7z"/></svg>
          </n-button>
          <n-button :type="store.viewMode === 'grid' ? 'primary' : 'default'" @click="store.viewMode = 'grid'" title="网格视图">
            <svg viewBox="0 0 24 24" fill="currentColor" width="16" height="16"><path d="M3 3v8h8V3H3zm6 6H5V5h4v4zm-6 4v8h8v-8H3zm6 6H5v-4h4v4zm4-16v8h8V3h-8zm6 6h-4V5h4v4zm-6 4v8h8v-8h-8zm6 6h-4v-4h4v4z"/></svg>
          </n-button>
        </n-button-group>
      </div>
    </div>

    <!-- Error -->
    <n-alert v-if="store.error" type="error" :title="store.error" closable @after-leave="store.error = null" style="margin-bottom:12px" />

    <!-- Loading spin -->
    <n-spin :show="store.loading" style="min-height: 80px;">
      <!-- Grid view -->
      <div v-if="store.viewMode === 'grid'" class="grid-view">
        <div
          v-for="item in store.items"
          :key="item.path"
          class="grid-item"
          @dblclick="handleOpen(item)"
          @click="handleSingleClick(item)"
          :class="{ selected: selected?.path === item.path }"
        >
          <div class="grid-icon">
            <span class="icon-large">{{ item.isDir ? '📁' : fileEmoji(item.ext) }}</span>
          </div>
          <div class="grid-name" :title="item.name">{{ item.name }}</div>
          <div class="grid-size" v-if="!item.isDir">{{ formatSize(item.size) }}</div>
          <div class="grid-actions">
            <button v-if="!item.isDir" class="icon-btn-sm" title="下载" @click.stop="download(item)">⬇</button>
            <n-popconfirm @positive-click="deleteItem(item)">
              <template #trigger>
                <button class="icon-btn-sm danger" title="删除" @click.stop>🗑</button>
              </template>
              确认删除吗？
            </n-popconfirm>
          </div>
        </div>
        <div v-if="!store.loading && store.items.length === 0" class="empty-dir">空文件夹</div>
      </div>

      <!-- List view -->
      <div v-else class="list-view">
        <div class="list-header">
          <span class="col-name">名称</span>
          <span class="col-size">大小</span>
          <span class="col-date">修改时间</span>
          <span class="col-actions">操作</span>
        </div>
        <div
          v-for="item in store.items"
          :key="item.path"
          class="list-row"
          :class="{ selected: selected?.path === item.path }"
          @dblclick="handleOpen(item)"
          @click="handleSingleClick(item)"
        >
          <span class="col-name">
            <span class="file-icon">{{ item.isDir ? '📁' : fileEmoji(item.ext) }}</span>
            <span class="file-name" :title="item.name">{{ item.name }}</span>
          </span>
          <span class="col-size muted">{{ item.isDir ? '—' : formatSize(item.size) }}</span>
          <span class="col-date muted">{{ formatDate(item.modTime) }}</span>
          <span class="col-actions" @click.stop>
            <button v-if="!item.isDir" class="icon-btn-sm" title="下载" @click="download(item)">⬇</button>
            <button v-if="!item.isDir && canPreview(item.ext)" class="icon-btn-sm" title="预览" @click="$emit('preview', item)">👁</button>
            <n-popconfirm @positive-click="deleteItem(item)">
              <template #trigger>
                <button class="icon-btn-sm danger" title="删除">🗑</button>
              </template>
              确认删除吗？
            </n-popconfirm>
          </span>
        </div>
        <div v-if="!store.loading && store.items.length === 0" class="empty-dir">空文件夹</div>
      </div>
    </n-spin>

    <!-- Mkdir dialog -->
    <n-modal v-model:show="showMkdir" preset="dialog" title="新建文件夹" @after-leave="mkdirName = ''">
      <n-input v-model:value="mkdirName" placeholder="文件夹名称" @keyup.enter="doMkdir" autofocus />
      <template #action>
        <n-button @click="showMkdir = false">取消</n-button>
        <n-button type="primary" @click="doMkdir" :disabled="!mkdirName.trim()">创建</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { NButton, NButtonGroup, NBreadcrumb, NBreadcrumbItem, NAlert, NSpin, NModal, NInput, NPopconfirm } from 'naive-ui'
import { useMessage } from 'naive-ui'
import { useFileStore } from '@/stores/file'
import { fileApi } from '@/services/api'
import { type FileInfo } from '@/services/api'
import { formatSize, formatDate, canPreview, getCategory } from '@/utils/fileUtils'
import { useDownload } from '@/composables/useDownload'

const emit = defineEmits<{
  (e: 'preview', item: FileInfo): void
}>()

const store = useFileStore()
const message = useMessage()
const showMkdir = ref(false)
const mkdirName = ref('')
const selected = ref<FileInfo | null>(null)

function handleOpen(item: FileInfo) {
  if (item.isDir) {
    store.navigateTo(item.path)
  } else if (canPreview(item.ext)) {
    emit('preview', item)
  } else {
    download(item)
  }
}

function handleSingleClick(item: FileInfo) {
  selected.value = item
}

function download(item: FileInfo) {
  const { download: dl } = useDownload()
  dl(item.path, item.name)
}

async function deleteItem(item: FileInfo) {
  try {
    await store.deleteItem(item.path)
    message.success(`已删除 ${item.name}`)
  } catch {
    message.error('删除失败')
  }
}

async function doMkdir() {
  const name = mkdirName.value.trim()
  if (!name) return
  try {
    await store.createDir(name)
    message.success(`文件夹 "${name}" 已创建`)
    showMkdir.value = false
    mkdirName.value = ''
  } catch {
    message.error('创建文件夹失败')
  }
}

function fileEmoji(ext: string): string {
  const cat = getCategory(ext)
  const map: Record<string, string> = {
    video: '🎬', audio: '🎵', image: '🖼️', pdf: '📄',
    text: '📝', archive: '🗜️', other: '📦',
  }
  return map[cat] ?? '📦'
}
</script>

<style scoped lang="scss">
.file-explorer { width: 100%; }

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  flex-wrap: wrap;
  gap: 8px;
}

.breadcrumb-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.breadcrumb {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  flex-wrap: wrap;
}

.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border: none;
  background: transparent;
  border-radius: 4px;
  cursor: pointer;
  color: #555;
  &:hover { background: #f0f0f0; color: #18a058; }
  &:disabled { opacity: 0.4; cursor: not-allowed; }
}

.icon-btn-sm {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  border-radius: 3px;
  cursor: pointer;
  font-size: 13px;
  &:hover { background: #f0f0f0; }
  &.danger:hover { background: #fff0f0; }
}

.spinning { animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* Grid view */
.grid-view {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 12px;
  min-height: 80px;
}

.grid-item {
  border: 1px solid #efeff5;
  border-radius: 8px;
  padding: 12px 8px 8px;
  text-align: center;
  cursor: pointer;
  position: relative;
  transition: border-color 0.2s, box-shadow 0.2s;

  &:hover { border-color: #18a058; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
  &.selected { border-color: #18a058; background: #e8f5ee; }
}

.grid-icon { margin-bottom: 6px; }
.icon-large { font-size: 2.5rem; line-height: 1; }

.grid-name {
  font-size: 12px;
  word-break: break-all;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  line-height: 1.4;
  margin-bottom: 4px;
}

.grid-size { font-size: 11px; color: #aaa; }

.grid-actions {
  position: absolute;
  top: 4px;
  right: 4px;
  opacity: 0;
  transition: opacity 0.2s;
  display: flex;
  gap: 2px;
  .grid-item:hover & { opacity: 1; }
}

/* List view */
.list-view { width: 100%; }

.list-header {
  display: grid;
  grid-template-columns: 1fr 100px 160px 100px;
  padding: 6px 10px;
  font-size: 12px;
  color: #aaa;
  font-weight: 600;
  border-bottom: 1px solid #f0f0f0;
}

.list-row {
  display: grid;
  grid-template-columns: 1fr 100px 160px 100px;
  padding: 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  align-items: center;
  transition: background 0.15s;
  min-height: 44px;

  &:hover { background: #f5f5f7; }
  &.selected { background: #e8f5ee; }
}

.col-name {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.file-icon { font-size: 1.2rem; flex-shrink: 0; }
.file-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
}

.col-size, .col-date {
  font-size: 12px;
}

.col-actions {
  display: flex;
  align-items: center;
  gap: 2px;
}

.muted { color: #aaa; }

.empty-dir {
  text-align: center;
  color: #aaa;
  padding: 32px;
  font-size: 13px;
}

/* Mobile: hide date column */
@media (max-width: 600px) {
  .list-header { grid-template-columns: 1fr 80px 80px; }
  .list-row { grid-template-columns: 1fr 80px 80px; }
  .col-date { display: none; }
}
</style>
