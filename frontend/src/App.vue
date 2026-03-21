<template>
  <t-config-provider :global-config="globalConfig">
    <router-view />
  </t-config-provider>
</template>

<script setup lang="ts">
import { useTheme } from "@/composables/useTheme";
import zhCN from "tdesign-vue-next/es/locale/zh_CN";
import { computed } from "vue";

useTheme();

const globalConfig = computed(() => ({
  locale: zhCN,
}));
</script>

<style lang="scss"></style>
