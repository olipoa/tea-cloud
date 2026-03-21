<template>
  <div class="transfer-view">
    <!-- Header -->
    <div class="transfer-header">
      <t-button
        variant="text"
        shape="square"
        @click="router.push('/')"
        title="返回"
      >
        <template #icon><t-icon name="chevron-left" /></template>
      </t-button>
      <span class="transfer-title">传输列表</span>
    </div>

    <!-- Type tabs: Downloads / Uploads -->
    <t-tabs v-model="typeTab" class="type-tabs">
      <t-tab-panel value="download" label="下载">
        <div class="tab-content">
          <!-- Concurrency setting -->
          <div class="concurrency-row">
            <span class="concurrency-label">最大并发数</span>
            <t-input-number
              v-model="transferStore.maxDownloadConcurrency"
              :min="1"
              :max="10"
              size="small"
              style="width: 80px"
            />
            <t-button
              size="small"
              variant="text"
              @click="transferStore.clearCompleted()"
              :disabled="!hasCompleted('download')"
            >
              清除已完成
            </t-button>
          </div>

          <!-- Status sub-tabs -->
          <t-tabs v-model="downloadSubTab" size="small">
            <t-tab-panel
              v-for="sub in subTabs"
              :key="sub.value"
              :value="sub.value"
              :label="subTabLabel(sub, 'download')"
            >
              <div class="task-list">
                <div
                  v-if="filteredTasks('download', sub.value).length === 0"
                  class="empty-hint"
                >
                  暂无任务
                </div>
                <TransferTaskItem
                  v-for="task in filteredTasks('download', sub.value)"
                  :key="task.id"
                  :task="task"
                />
              </div>
            </t-tab-panel>
          </t-tabs>
        </div>
      </t-tab-panel>

      <t-tab-panel value="upload" label="上传">
        <div class="tab-content">
          <!-- Concurrency setting -->
          <div class="concurrency-row">
            <span class="concurrency-label">最大并发数</span>
            <t-input-number
              v-model="transferStore.maxUploadConcurrency"
              :min="1"
              :max="10"
              size="small"
              style="width: 80px"
            />
            <t-button
              size="small"
              variant="text"
              @click="transferStore.clearCompleted()"
              :disabled="!hasCompleted('upload')"
            >
              清除已完成
            </t-button>
          </div>

          <!-- Status sub-tabs -->
          <t-tabs v-model="uploadSubTab" size="small">
            <t-tab-panel
              v-for="sub in subTabs"
              :key="sub.value"
              :value="sub.value"
              :label="subTabLabel(sub, 'upload')"
            >
              <div class="task-list">
                <div
                  v-if="filteredTasks('upload', sub.value).length === 0"
                  class="empty-hint"
                >
                  暂无任务
                </div>
                <TransferTaskItem
                  v-for="task in filteredTasks('upload', sub.value)"
                  :key="task.id"
                  :task="task"
                />
              </div>
            </t-tab-panel>
          </t-tabs>
        </div>
      </t-tab-panel>
    </t-tabs>
  </div>
</template>

<script setup lang="ts">
import TransferTaskItem from "@/components/TransferTaskItem.vue";
import { useTransferStore, type TaskType } from "@/stores/transfer";
import { ref } from "vue";
import { useRouter } from "vue-router";

const router = useRouter();
const transferStore = useTransferStore();

const typeTab = ref<"download" | "upload">("download");
const downloadSubTab = ref("all");
const uploadSubTab = ref("all");

const subTabs = [
  { value: "all", label: "全部" },
  { value: "running", label: "进行中" },
  { value: "done", label: "已完成" },
  { value: "failed", label: "任务失败" },
];

function filteredTasks(type: TaskType, status: string) {
  return transferStore.tasks.filter((t) => {
    if (t.type !== type) return false;
    if (status === "all") return true;
    if (status === "running")
      return t.status === "running" || t.status === "pending";
    return t.status === status;
  });
}

function subTabLabel(sub: { value: string; label: string }, type: TaskType) {
  const count = filteredTasks(type, sub.value).length;
  return count > 0 ? `${sub.label} (${count})` : sub.label;
}

function hasCompleted(type: TaskType) {
  return transferStore.tasks.some(
    (t) => t.type === type && t.status === "done",
  );
}
</script>

<style scoped lang="scss">
.transfer-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--td-bg-color-page);
}
.transfer-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--td-component-stroke);
  background: var(--td-bg-color-container);
  flex-shrink: 0;
}
.transfer-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}
.type-tabs {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.tab-content {
  display: flex;
  flex-direction: column;
  padding: 12px 16px;
}
.concurrency-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}
.concurrency-label {
  font-size: 13px;
  color: var(--td-text-color-secondary);
  white-space: nowrap;
}
.task-list {
  padding: 8px 0;
  max-height: calc(100vh - 220px);
  overflow-y: auto;
}
.empty-hint {
  text-align: center;
  color: var(--td-text-color-placeholder);
  padding: 32px 0;
  font-size: 14px;
}
</style>
