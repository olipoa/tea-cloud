<template>
  <div class="transfer-task-item">
    <div class="task-icon">
      <t-icon
        :name="task.type === 'upload' ? 'upload' : 'download'"
        :style="{ color: statusColor }"
      />
    </div>
    <div class="task-body">
      <div class="task-name" :title="task.name">{{ task.name }}</div>
      <t-progress
        v-if="task.status === 'running' || task.status === 'pending'"
        :percentage="Math.round(task.progress)"
        size="small"
        :color="'var(--td-brand-color)'"
        style="margin: 4px 0"
      />
      <div class="task-meta">
        <span :class="['status-label', `status-${task.status}`]">{{
          statusLabel
        }}</span>
        <span v-if="task.size > 0" class="task-size">{{
          formatSize(task.size)
        }}</span>
        <span v-if="task.error" class="task-error" :title="task.error">{{
          task.error
        }}</span>
      </div>
    </div>
    <div class="task-actions">
      <t-button
        v-if="task.status === 'failed'"
        variant="text"
        size="small"
        @click="transferStore.retryTask(task.id)"
        title="重试"
      >
        <template #icon><t-icon name="refresh" /></template>
      </t-button>
      <t-button
        variant="text"
        size="small"
        @click="transferStore.removeTask(task.id)"
        title="移除"
      >
        <template #icon><t-icon name="close" /></template>
      </t-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useTransferStore, type TransferTask } from "@/stores/transfer";
import { formatSize } from "@/utils/fileUtils";
import { computed } from "vue";

const props = defineProps<{ task: TransferTask }>();
const transferStore = useTransferStore();

const statusLabel = computed(() => {
  switch (props.task.status) {
    case "pending":
      return "等待中";
    case "running":
      return "进行中";
    case "done":
      return "已完成";
    case "failed":
      return "失败";
    default:
      return "";
  }
});

const statusColor = computed(() => {
  switch (props.task.status) {
    case "done":
      return "var(--td-success-color)";
    case "failed":
      return "var(--td-error-color)";
    case "running":
      return "var(--td-brand-color)";
    default:
      return "var(--td-text-color-secondary)";
  }
});
</script>

<style scoped lang="scss">
.transfer-task-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 0;
  border-bottom: 1px solid var(--td-component-stroke);
  &:last-child {
    border-bottom: none;
  }
}
.task-icon {
  padding-top: 2px;
  font-size: 18px;
}
.task-body {
  flex: 1;
  min-width: 0;
}
.task-name {
  font-size: 14px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.task-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}
.task-size {
  color: var(--td-text-color-secondary);
}
.task-error {
  color: var(--td-error-color);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 180px;
}
.status-label {
  font-size: 12px;
  &.status-pending {
    color: var(--td-text-color-secondary);
  }
  &.status-running {
    color: var(--td-brand-color);
  }
  &.status-done {
    color: var(--td-success-color);
  }
  &.status-failed {
    color: var(--td-error-color);
  }
}
.task-actions {
  flex-shrink: 0;
}
</style>
