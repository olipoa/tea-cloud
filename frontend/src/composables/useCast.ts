import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { castApi, type DLNARenderer } from '@/services/api'
import { useMediaPlayerStore } from '@/stores/audioPlayer'

export function useCast() {
  const store = useMediaPlayerStore()
  const castPopoverVisible = ref(false)
  const castLoading = ref(false)
  const castDevices = ref<DLNARenderer[]>([])

  async function openCastPopover() {
    castPopoverVisible.value = true
    castLoading.value = true
    castDevices.value = []
    try {
      castDevices.value = await castApi.devices()
    } catch {
      ElMessage.error('扫描 DLNA 设备失败')
      castPopoverVisible.value = false
    } finally {
      castLoading.value = false
    }
  }

  async function castTo(device: DLNARenderer) {
    castPopoverVisible.value = false
    const track = store.currentTrack
    if (!track) return
    try {
      await castApi.send(device.location, track.path)
      ElMessage.success(`已投屏到 ${device.name}`)
    } catch {
      ElMessage.error(`投屏到 ${device.name} 失败，请确认设备在线且支持 DLNA`)
    }
  }

  return { castPopoverVisible, castLoading, castDevices, openCastPopover, castTo }
}
