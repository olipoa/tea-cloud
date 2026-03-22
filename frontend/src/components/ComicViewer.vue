<template>
  <Teleport to="body">
    <div
      ref="viewerRef"
      class="comic-viewer"
      tabindex="-1"
      @keydown.esc="emit('close')"
      @mousemove="showControls"
      @touchstart.passive="showControls"
    >
      <Swiper
        class="comic-swiper"
        :modules="swiperModules"
        :initial-slide="initialIndex"
        :keyboard="{ enabled: true, onlyInViewport: false }"
        :mousewheel="{ forceToAxis: true }"
        :zoom="{ maxRatio: 2.5 }"
        :autoplay="autoplayProp"
        :slides-per-view="1"
        :grab-cursor="true"
        :space-between="0"
        @swiper="onSwiper"
        @slideChange="onSlideChange"
        @autoplayStop="onAutoplayStop"
        @touchStart="onUserInteract"
        @keyPress="onUserInteract"
      >
        <SwiperSlide v-for="file in files" :key="file.path">
          <div class="swiper-zoom-container">
            <img
              :src="fileApi.rawUrl(file.path)"
              :alt="file.name"
              draggable="false"
              loading="lazy"
            />
          </div>
        </SwiperSlide>
      </Swiper>

      <Transition name="controls-fade">
        <div
          v-show="controlsVisible"
          class="controls-overlay"
          @click.stop
          @mousemove.stop
        >
          <!-- Top bar -->
          <div class="controls-top">
            <button
              class="ctrl-btn close-btn"
              @click="emit('close')"
              title="关闭 (Esc)"
            >
              ✕
            </button>
            <span class="page-counter"
              >{{ currentIndex + 1 }} / {{ files.length }}</span
            >
            <div class="spacer" />
          </div>

          <!-- Bottom bar -->
          <div class="controls-bottom">
            <div class="bottom-inner">
              <button
                class="ctrl-btn autoplay-btn"
                :title="autoplayEnabled ? '暂停自动翻页' : '开始自动翻页'"
                @click="toggleAutoplay"
              >
                {{ autoplayEnabled ? "⏸" : "▶" }}
              </button>
              <div class="speed-wrapper">
                <label class="speed-label">间隔 {{ autoplaySeconds }}s</label>
                <input
                  type="range"
                  class="speed-slider"
                  v-model.number="autoplaySeconds"
                  min="1"
                  max="10"
                  step="1"
                  @change="onSpeedChange"
                  @click.stop
                />
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { type FileInfo, fileApi } from "@/services/api";
import type { Swiper as SwiperClass } from "swiper";
import "swiper/css";
import "swiper/css/zoom";
import { Autoplay, Keyboard, Mousewheel, Zoom } from "swiper/modules";
import { Swiper, SwiperSlide } from "swiper/vue";
import { onBeforeUnmount, onMounted, ref } from "vue";

const props = defineProps<{
  files: FileInfo[];
  initialIndex: number;
}>();

const emit = defineEmits<{ (e: "close"): void }>();

const swiperModules = [Keyboard, Mousewheel, Zoom, Autoplay];

// 将 autoplay 参数固定，通过实例方法控制 start/stop
const autoplayProp = {
  delay: 6000,
  disableOnInteraction: true,
  stopOnLastSlide: false,
};

const viewerRef = ref<HTMLDivElement | null>(null);
const swiperInstance = ref<SwiperClass | null>(null);
const currentIndex = ref(props.initialIndex);

// ─── Controls auto-hide ──────────────────────────────────────────────────────
const controlsVisible = ref(true);
let hideTimer: ReturnType<typeof setTimeout> | null = null;

function showControls() {
  controlsVisible.value = true;
  if (hideTimer) clearTimeout(hideTimer);
  hideTimer = setTimeout(() => {
    controlsVisible.value = false;
  }, 3000);
}

// ─── Autoplay ────────────────────────────────────────────────────────────────
const autoplayEnabled = ref(false);
const autoplaySeconds = ref(6);

function toggleAutoplay() {
  if (!swiperInstance.value) return;
  if (autoplayEnabled.value) {
    swiperInstance.value.autoplay.stop();
    autoplayEnabled.value = false;
  } else {
    // 更新 delay 后启动
    (swiperInstance.value.params.autoplay as Record<string, unknown>).delay =
      autoplaySeconds.value * 1000;
    swiperInstance.value.autoplay.start();
    autoplayEnabled.value = true;
  }
  showControls();
}

function onSpeedChange() {
  if (!swiperInstance.value) return;
  (swiperInstance.value.params.autoplay as Record<string, unknown>).delay =
    autoplaySeconds.value * 1000;
  if (autoplayEnabled.value) {
    swiperInstance.value.autoplay.stop();
    swiperInstance.value.autoplay.start();
  }
}

// ─── Swiper event handlers ───────────────────────────────────────────────────
function onSwiper(swiper: SwiperClass) {
  swiperInstance.value = swiper;
}

function onSlideChange() {
  currentIndex.value = swiperInstance.value?.activeIndex ?? 0;
  // 每次换页重置缩放
  swiperInstance.value?.zoom.out();
}

function onAutoplayStop() {
  // disableOnInteraction 触发后同步按钮状态
  autoplayEnabled.value = false;
}

function onUserInteract() {
  // 手动操作时显示控制栏
  showControls();
}

// ─── Lifecycle ───────────────────────────────────────────────────────────────
onMounted(() => {
  viewerRef.value?.focus();
  document.body.style.overflow = "hidden";
  showControls();
});

onBeforeUnmount(() => {
  document.body.style.overflow = "";
  if (hideTimer) clearTimeout(hideTimer);
});
</script>

<style scoped lang="scss">
.comic-viewer {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: #000;
  outline: none;
  display: flex;
  align-items: center;
  justify-content: center;
}

// ─── Swiper ──────────────────────────────────────────────────────────────────
.comic-swiper {
  width: 100%;
  height: 100%;

  :deep(.swiper-slide) {
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
  }

  :deep(.swiper-zoom-container) {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;

    img {
      max-width: 100%;
      max-height: 100vh;
      object-fit: contain;
      user-select: none;
      -webkit-user-drag: none;
    }
  }
}

// ─── Controls overlay ────────────────────────────────────────────────────────
.controls-overlay {
  position: absolute;
  inset: 0;
  z-index: 10;
  pointer-events: none;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.controls-top,
.controls-bottom {
  pointer-events: all;
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: linear-gradient(to bottom, rgba(0, 0, 0, 0.55), transparent);
}

.controls-bottom {
  background: linear-gradient(to top, rgba(0, 0, 0, 0.55), transparent);
  justify-content: center;
}

.controls-top {
  gap: 12px;
}

.spacer {
  width: 36px;
  flex-shrink: 0;
}

.page-counter {
  flex: 1;
  text-align: center;
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  letter-spacing: 0.5px;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8);
}

.ctrl-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
  font-size: 16px;
  cursor: pointer;
  backdrop-filter: blur(4px);
  transition: background 0.15s;
  flex-shrink: 0;

  &:hover {
    background: rgba(255, 255, 255, 0.3);
  }
}

.bottom-inner {
  display: flex;
  align-items: center;
  gap: 16px;
}

.autoplay-btn {
  font-size: 14px;
}

.speed-wrapper {
  display: flex;
  align-items: center;
  gap: 10px;
}

.speed-label {
  color: rgba(255, 255, 255, 0.85);
  font-size: 13px;
  white-space: nowrap;
  min-width: 52px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
}

.speed-slider {
  width: 120px;
  accent-color: #18a058;
  cursor: pointer;
}

// ─── Transitions ─────────────────────────────────────────────────────────────
.controls-fade-enter-active,
.controls-fade-leave-active {
  transition: opacity 0.3s ease;
}

.controls-fade-enter-from,
.controls-fade-leave-to {
  opacity: 0;
}
</style>
