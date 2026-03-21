import axios from 'axios'

const http = axios.create({
  baseURL: '/',
  timeout: 0, // no timeout for large file uploads
})

export interface FileInfo {
  name: string
  path: string
  size: number
  isDir: boolean
  modTime: number
  mime: string
  ext: string
}

export interface PeerInfo {
  name: string
  host: string
  port: number
  addrV4: string
  url: string
}

export interface SelfInfo {
  name: string
  port: number
  ips: string[]
}

// Current base URL (default: same origin)
let _baseUrl = ''

export function setBaseUrl(url: string) {
  _baseUrl = url.replace(/\/$/, '')
  http.defaults.baseURL = _baseUrl + '/'
}

export function getBaseUrl(): string {
  return _baseUrl
}

export const fileApi = {
  /** List directory contents */
  list(path = '.'): Promise<FileInfo[]> {
    return http.get('/api/files', { params: { path } }).then(r => r.data)
  },

  /** Get single file info */
  info(path: string): Promise<FileInfo> {
    return http.get('/api/files/info', { params: { path } }).then(r => r.data)
  },

  /** Get download URL for a file */
  downloadUrl(path: string): string {
    return `${_baseUrl}/api/files/download?path=${encodeURIComponent(path)}&download=1`
  },

  /** Get streaming/preview URL (used for video/audio/image) */
  rawUrl(path: string): string {
    return `${_baseUrl}/raw/${encodeURIComponent(path)}`
  },

  /** Upload files to a directory */
  upload(
    dir: string,
    files: File[],
    onProgress?: (percent: number) => void,
  ): Promise<{ uploaded: FileInfo[] }> {
    const form = new FormData()
    for (const f of files) {
      form.append('file', f)
    }
    return http
      .post('/api/files/upload', form, {
        params: { path: dir },
        headers: { 'Content-Type': 'multipart/form-data' },
        onUploadProgress: e => {
          if (onProgress && e.total) {
            onProgress(Math.round((e.loaded * 100) / e.total))
          }
        },
      })
      .then(r => r.data)
  },

  /** Delete a file or folder */
  delete(path: string): Promise<void> {
    return http.delete('/api/files', { params: { path } }).then(() => undefined)
  },

  /** Create a directory */
  mkdir(path: string): Promise<void> {
    return http.post('/api/dirs', null, { params: { path } }).then(() => undefined)
  },
}

export const peerApi = {
  /** Discover peers on the LAN */
  list(): Promise<PeerInfo[]> {
    return http.get('/api/peers').then(r => r.data)
  },

  /** Get this node's info */
  self(): Promise<SelfInfo> {
    return http.get('/api/self').then(r => r.data)
  },
}
