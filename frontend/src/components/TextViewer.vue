<template>
  <div class="text-viewer">
    <t-loading :loading="loading" style="min-height: 120px">
      <pre v-if="content !== null" class="code-content">{{ content }}</pre>
      <div v-else-if="!loading" class="error-msg">文本加载失败</div>
    </t-loading>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue";

const props = defineProps<{ url: string }>();

const content = ref<string | null>(null);
const loading = ref(false);

async function load() {
  loading.value = true;
  content.value = null;
  try {
    const res = await fetch(props.url);
    if (!res.ok) throw new Error("fetch failed");
    const blob = await res.blob();
    if (blob.size > 2 * 1024 * 1024) {
      content.value =
        `[文件过大 (${(blob.size / 1024 / 1024).toFixed(1)} MB)，仅显示前 2 MB]\n\n` +
        (await blob.slice(0, 2 * 1024 * 1024).text());
    } else {
      content.value = await blob.text();
    }
  } catch {
    content.value = null;
  } finally {
    loading.value = false;
  }
}

watch(() => props.url, load);
onMounted(load);
</script>

<style scoped lang="scss">
.text-viewer {
  background: var(--td-bg-color-page);
  border-radius: 6px;
  overflow: hidden;
}

.code-content {
  margin: 0;
  padding: 20px;
  font-family: "Cascadia Code", "Fira Code", "Consolas", monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  overflow: auto;
  max-height: 75vh;
  color: var(--td-text-color-primary);
  background: var(--td-bg-color-page);
}

.error-msg {
  color: var(--td-error-color);
  padding: 32px;
  text-align: center;
}
</style>
