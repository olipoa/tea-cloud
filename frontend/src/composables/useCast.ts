import { castApi, type DLNARenderer } from "@/services/api";
import { useMediaPlayerStore } from "@/stores/audioPlayer";
import { useMessage } from "naive-ui";
import { ref } from "vue";

export function useCast() {
  const store = useMediaPlayerStore();
  const message = useMessage();
  const castPopoverVisible = ref(false);
  const castLoading = ref(false);
  const castDevices = ref<DLNARenderer[]>([]);

  async function openCastPopover() {
    castPopoverVisible.value = true;
    castLoading.value = true;
    castDevices.value = [];
    try {
      castDevices.value = await castApi.devices();
    } catch {
      message.error("扫描 DLNA 设备失败");
      castPopoverVisible.value = false;
    } finally {
      castLoading.value = false;
    }
  }

  async function castTo(device: DLNARenderer) {
    castPopoverVisible.value = false;
    const track = store.currentTrack;
    if (!track) return;
    try {
      await castApi.send(device.location, track.path);
      message.success(`已投屏到 ${device.name}`);
      store.setIsPlaying(false);
    } catch {
      message.error(`投屏到 ${device.name} 失败，请确认设备在线且支持 DLNA`);
    }
  }

  return {
    castPopoverVisible,
    castLoading,
    castDevices,
    openCastPopover,
    castTo,
  };
}
