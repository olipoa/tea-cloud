// File type detection utilities

export type FileCategory = 'video' | 'audio' | 'image' | 'pdf' | 'text' | 'archive' | 'other'

const VIDEO_EXTS = new Set(['.mp4', '.mkv', '.webm', '.avi', '.mov', '.m4v', '.flv', '.wmv', '.ts', '.m2ts'])
const AUDIO_EXTS = new Set(['.mp3', '.flac', '.wav', '.ogg', '.aac', '.m4a', '.wma', '.opus'])
const IMAGE_EXTS = new Set(['.jpg', '.jpeg', '.png', '.gif', '.webp', '.svg', '.bmp', '.ico', '.avif'])
const TEXT_EXTS = new Set(['.txt', '.md', '.json', '.xml', '.csv', '.log', '.yaml', '.yml', '.toml', '.ini', '.conf', '.html', '.css', '.js', '.ts', '.vue', '.go', '.py', '.rs', '.java', '.cpp', '.c', '.h', '.sh', '.bat'])
const ARCHIVE_EXTS = new Set(['.zip', '.7z', '.tar', '.gz', '.rar', '.bz2', '.xz'])

export function getCategory(ext: string): FileCategory {
  const e = ext.toLowerCase()
  if (VIDEO_EXTS.has(e)) return 'video'
  if (AUDIO_EXTS.has(e)) return 'audio'
  if (IMAGE_EXTS.has(e)) return 'image'
  if (e === '.pdf') return 'pdf'
  if (TEXT_EXTS.has(e)) return 'text'
  if (ARCHIVE_EXTS.has(e)) return 'archive'
  return 'other'
}

export function canPreview(ext: string): boolean {
  const cat = getCategory(ext)
  return cat === 'video' || cat === 'audio' || cat === 'image' || cat === 'pdf' || cat === 'text'
}

export function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return `${(bytes / Math.pow(1024, i)).toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}

export function formatDate(ms: number): string {
  return new Date(ms).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
