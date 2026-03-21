<template>
  <div class="file-explorer">
    <!-- Toolbar -->
    <div class="toolbar">
      <div class="breadcrumb-row">
        <el-button
          :disabled="store.currentPath === '.'"
          @click="store.navigateUp()"
          :icon="ArrowLeft"
          circle
          size="small"
        />
        <el-breadcrumb separator="/" class="breadcrumb">
          <el-breadcrumb-item
            v-for="crumb in store.breadcrumbs"
            :key="crumb.path"
            @click="store.navigateTo(crumb.path)"
            class="crumb-item"
          >
            {{ crumb.label }}
          </el-breadcrumb-item>
        </el-breadcrumb>
      </div>

      <div class="actions">
        <el-button :icon="Refresh" size="small" @click="store.refresh()" :loading="store.loading" />
        <el-button :icon="FolderAdd" size="small" @click="showMkdir = true">新建文件夹</el-button>
        <el-radio-group v-model="store.viewMode" size="small">
          <el-radio-button label="list"><el-icon><List /></el-icon></el-radio-button>
          <el-radio-button label="grid"><el-icon><Grid /></el-icon></el-radio-button>
        </el-radio-group>
      </div>
    </div>

    <!-- Error -->
    <el-alert v-if="store.error" :title="store.error" type="error" show-icon closable @close="store.error = null" style="margin-bottom:12px" />

    <!-- Grid view -->
    <div v-if="store.viewMode === 'grid'" class="grid-view" v-loading="store.loading">
      <div
        v-for="item in store.items"
        :key="item.path"
        class="grid-item"
        @dblclick="handleOpen(item)"
        @click="handleSingleClick(item)"
        :class="{ selected: selected?.path === item.path }"
      >
        <div class="grid-icon">
          <span v-if="item.isDir" class="icon-large">📁</span>
          <span v-else class="icon-large">{{ fileEmoji(item.ext) }}</span>
        </div>
        <div class="grid-name" :title="item.name">{{ item.name }}</div>
        <div class="grid-size" v-if="!item.isDir">{{ formatSize(item.size) }}</div>
        <div class="grid-actions">
          <el-button v-if="!item.isDir" size="small" link :icon="Download" @click.stop="download(item)" />
          <el-popconfirm title="确认删除？" @confirm="deleteItem(item)" @click.stop>
            <template #reference>
              <el-button size="small" link :icon="Delete" type="danger" @click.stop />
            </template>
          </el-popconfirm>
        </div>
      </div>
      <el-empty v-if="!store.loading && store.items.length === 0" description="空文件夹" />
    </div>

    <!-- List view -->
    <el-table
      v-else
      :data="store.items"
      v-loading="store.loading"
      @row-dblclick="handleOpen"
      @row-click="handleSingleClick"
      style="width: 100%"
      :row-class-name="rowClass"
      stripe
    >
      <el-table-column label="名称" min-width="300">
        <template #default="{ row }">
          <div class="name-cell">
            <span class="file-icon">{{ row.isDir ? '📁' : fileEmoji(row.ext) }}</span>
            <span class="file-name">{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="大小" width="120" align="right">
        <template #default="{ row }">
          <span v-if="!row.isDir" class="muted">{{ formatSize(row.size) }}</span>
          <span v-else class="muted">—</span>
        </template>
      </el-table-column>
      <el-table-column label="修改时间" width="180">
        <template #default="{ row }">
          <span class="muted">{{ formatDate(row.modTime) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120" align="center">
        <template #default="{ row }">
          <el-button v-if="!row.isDir" :icon="Download" size="small" link @click.stop="download(row)" title="下载" />
          <el-button v-if="!row.isDir && canPreview(row.ext)" :icon="View" size="small" link @click.stop="$emit('preview', row)" title="预览" />
          <el-popconfirm title="确认删除？" @confirm="deleteItem(row)">
            <template #reference>
              <el-button :icon="Delete" size="small" link type="danger" @click.stop title="删除" />
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <!-- Mkdir dialog -->
    <el-dialog v-model="showMkdir" title="新建文件夹" width="380px" @close="mkdirName = ''">
      <el-input v-model="mkdirName" placeholder="文件夹名称" @keyup.enter="doMkdir" autofocus />
      <template #footer>
        <el-button @click="showMkdir = false">取消</el-button>
        <el-button type="primary" @click="doMkdir" :disabled="!mkdirName.trim()">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ArrowLeft, Refresh, FolderAdd, List, Grid, Download, Delete, View } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useFileStore } from '@/stores/file'
import { fileApi } from '@/services/api'
import { type FileInfo } from '@/services/api'
import { formatSize, formatDate, canPreview, getCategory } from '@/utils/fileUtils'
import { useDownload } from '@/composables/useDownload'

const emit = defineEmits<{
  (e: 'preview', item: FileInfo): void
}>()

const store = useFileStore()
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
    ElMessage.success(`已删除 ${item.name}`)
  } catch {
    ElMessage.error('删除失败')
  }
}

async function doMkdir() {
  const name = mkdirName.value.trim()
  if (!name) return
  try {
    await store.createDir(name)
    ElMessage.success(`文件夹 "${name}" 已创建`)
    showMkdir.value = false
    mkdirName.value = ''
  } catch {
    ElMessage.error('创建文件夹失败')
  }
}

function rowClass({ row }: { row: FileInfo }) {
  return selected.value?.path === row.path ? 'row-selected' : ''
}

function fileEmoji(ext: string): string {
  const cat = getCategory(ext)
  const map: Record<string, string> = {
    video: '🎬',
    audio: '🎵',
    image: '🖼️',
    pdf: '📄',
    text: '📝',
    archive: '🗜️',
    other: '📦',
  }
  return map[cat] ?? '📦'
}
</script>

<style scoped lang="scss">
.file-explorer {
  width: 100%;
}

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
  text-overflow: ellipsis;
  white-space: nowrap;
}

.crumb-item {
  cursor: pointer;
  &:hover :deep(.el-breadcrumb__inner) {
    color: var(--el-color-primary);
  }
}

.actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

/* Grid view */
.grid-view {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 12px;
  min-height: 120px;
}

.grid-item {
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  padding: 12px 8px 8px;
  text-align: center;
  cursor: pointer;
  position: relative;
  transition: border-color 0.2s, box-shadow 0.2s;

  &:hover {
    border-color: var(--el-color-primary);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  }

  &.selected {
    border-color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);
  }
}

.grid-icon {
  margin-bottom: 6px;
}

.icon-large {
  font-size: 2.5rem;
  line-height: 1;
}

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

.grid-size {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

.grid-actions {
  position: absolute;
  top: 4px;
  right: 4px;
  opacity: 0;
  transition: opacity 0.2s;

  .grid-item:hover & {
    opacity: 1;
  }
}

/* List view */
.name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-icon {
  font-size: 1.2rem;
  flex-shrink: 0;
}

.file-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.muted {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

:deep(.row-selected) {
  background-color: var(--el-color-primary-light-9) !important;
}
</style>
