import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { type FileInfo } from '@/services/api'
import { getCategory } from '@/utils/fileUtils'

export type MediaType = 'audio' | 'video'

export const useMediaPlayerStore = defineStore('mediaPlayer', () => {
  const playlist = ref<FileInfo[]>([])
  const currentIndex = ref(0)
  const isPlaying = ref(false)
  const currentTime = ref(0)
  const duration = ref(0)
  const showPlaylist = ref(false)
  const showVideo = ref(true)

  const currentTrack = computed(() =>
    playlist.value.length > 0 ? playlist.value[currentIndex.value] : null,
  )

  const mediaType = computed<MediaType>(() =>
    currentTrack.value && getCategory(currentTrack.value.ext) === 'video' ? 'video' : 'audio',
  )

  const isVideo = computed(() => mediaType.value === 'video')

  const hasNext = computed(() => currentIndex.value < playlist.value.length - 1)
  const hasPrev = computed(() => currentIndex.value > 0)

  function setPlaylist(files: FileInfo[], startIndex = 0) {
    playlist.value = files
    currentIndex.value = startIndex
    isPlaying.value = true
    currentTime.value = 0
    duration.value = 0
    showVideo.value = true
  }

  function next() {
    if (hasNext.value) {
      currentIndex.value++
      currentTime.value = 0
      isPlaying.value = true
    }
  }

  function prev() {
    if (hasPrev.value) {
      currentIndex.value--
      currentTime.value = 0
      isPlaying.value = true
    }
  }

  function playAt(index: number) {
    if (index >= 0 && index < playlist.value.length) {
      currentIndex.value = index
      currentTime.value = 0
      isPlaying.value = true
    }
  }

  function togglePlay() {
    isPlaying.value = !isPlaying.value
  }

  function seek(time: number) {
    currentTime.value = time
  }

  function togglePlaylist() {
    showPlaylist.value = !showPlaylist.value
  }

  function toggleVideo() {
    showVideo.value = !showVideo.value
  }

  function setCurrentTime(time: number) {
    currentTime.value = time
  }

  function setDuration(dur: number) {
    duration.value = dur
  }

  function setIsPlaying(val: boolean) {
    isPlaying.value = val
  }

  return {
    playlist,
    currentIndex,
    isPlaying,
    currentTime,
    duration,
    showPlaylist,
    showVideo,
    currentTrack,
    mediaType,
    isVideo,
    hasNext,
    hasPrev,
    setPlaylist,
    next,
    prev,
    playAt,
    togglePlay,
    seek,
    togglePlaylist,
    toggleVideo,
    setCurrentTime,
    setDuration,
    setIsPlaying,
  }
})

// backwards-compat alias
export { useMediaPlayerStore as useAudioPlayerStore }
