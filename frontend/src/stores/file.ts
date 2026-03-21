import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { fileApi, type FileInfo } from '@/services/api'

export const useFileStore = defineStore('file', () => {
  const currentPath = ref('.')
  const items = ref<FileInfo[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const viewMode = ref<'list' | 'grid'>('list')

  // Breadcrumb segments derived from currentPath
  const breadcrumbs = computed(() => {
    if (currentPath.value === '.') return [{ label: '根目录', path: '.' }]
    const parts = currentPath.value.split('/')
    const crumbs: { label: string; path: string }[] = [{ label: '根目录', path: '.' }]
    let acc = ''
    for (const part of parts) {
      if (!part || part === '.') continue
      acc = acc ? `${acc}/${part}` : part
      crumbs.push({ label: part, path: acc })
    }
    return crumbs
  })

  async function loadDir(path = '.') {
    loading.value = true
    error.value = null
    try {
      const result = await fileApi.list(path)
      items.value = result.sort((a, b) => {
        // Directories first, then files
        if (a.isDir && !b.isDir) return -1
        if (!a.isDir && b.isDir) return 1
        return a.name.localeCompare(b.name)
      })
      currentPath.value = path
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : '加载失败'
    } finally {
      loading.value = false
    }
  }

  function navigateTo(path: string) {
    loadDir(path)
  }

  function navigateUp() {
    if (currentPath.value === '.') return
    const parts = currentPath.value.split('/')
    parts.pop()
    const parent = parts.join('/') || '.'
    loadDir(parent)
  }

  async function deleteItem(path: string) {
    await fileApi.delete(path)
    await loadDir(currentPath.value)
  }

  async function createDir(name: string) {
    const newPath = currentPath.value === '.' ? name : `${currentPath.value}/${name}`
    await fileApi.mkdir(newPath)
    await loadDir(currentPath.value)
  }

  async function refresh() {
    await loadDir(currentPath.value)
  }

  return {
    currentPath,
    items,
    loading,
    error,
    viewMode,
    breadcrumbs,
    loadDir,
    navigateTo,
    navigateUp,
    deleteItem,
    createDir,
    refresh,
  }
})
