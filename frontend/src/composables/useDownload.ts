import { fileApi } from '@/services/api'

export function useDownload() {
  function download(path: string, filename: string) {
    const a = document.createElement('a')
    a.href = fileApi.downloadUrl(path)
    a.download = filename
    a.click()
  }

  return { download }
}
