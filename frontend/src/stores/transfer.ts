import { useDownload } from "@/composables/useDownload";
import { fileApi, uploadApi } from "@/services/api";
import { defineStore } from "pinia";
import { computed, ref } from "vue";

export type TaskStatus = "pending" | "running" | "done" | "failed";
export type TaskType = "upload" | "download";

export interface TransferTask {
  id: string;
  name: string;
  type: TaskType;
  status: TaskStatus;
  progress: number; // 0-100
  size: number;
  path: string; // server-relative path (upload dest or download src)
  error?: string;
  createdAt: number;
  file?: File; // only for upload tasks
}

const CHUNK_SIZE = 5 * 1024 * 1024;

function uid(): string {
  return Math.random().toString(36).slice(2) + Date.now().toString(36);
}

export const useTransferStore = defineStore("transfer", () => {
  const tasks = ref<TransferTask[]>([]);
  const maxUploadConcurrency = ref(3);
  const maxDownloadConcurrency = ref(3);

  let uploadRunning = 0;
  let downloadRunning = 0;
  const uploadQueue: TransferTask[] = [];
  const downloadQueue: TransferTask[] = [];

  const uploadTasks = computed(() =>
    tasks.value.filter((t) => t.type === "upload"),
  );
  const downloadTasks = computed(() =>
    tasks.value.filter((t) => t.type === "download"),
  );
  const runningCount = computed(
    () =>
      tasks.value.filter(
        (t) => t.status === "running" || t.status === "pending",
      ).length,
  );

  function addUpload(file: File, destPath: string) {
    const task: TransferTask = {
      id: uid(),
      name: file.name,
      type: "upload",
      status: "pending",
      progress: 0,
      size: file.size,
      path: destPath,
      createdAt: Date.now(),
      file,
    };
    tasks.value.unshift(task);
    uploadQueue.push(task);
    drainUploadQueue();
  }

  function addFolderUpload(files: File[], destBasePath: string) {
    for (const file of files) {
      const relPath = (file as any).webkitRelativePath as string;
      // e.g., "folderName/subdir/file.txt"
      const parts = relPath.split("/");
      parts.pop(); // remove filename
      const dirPath =
        parts.length > 0
          ? destBasePath === "."
            ? parts.join("/")
            : `${destBasePath}/${parts.join("/")}`
          : destBasePath;
      addUpload(file, dirPath);
    }
  }

  function addDownload(path: string, name: string, size = 0) {
    const task: TransferTask = {
      id: uid(),
      name,
      type: "download",
      status: "pending",
      progress: 0,
      size,
      path,
      createdAt: Date.now(),
    };
    tasks.value.unshift(task);
    downloadQueue.push(task);
    drainDownloadQueue();
  }

  async function addFolderDownload(path: string) {
    // Recursively collect all files under path
    async function collectFiles(p: string) {
      const items = await fileApi.list(p);
      for (const item of items) {
        if (item.isDir) {
          await collectFiles(item.path);
        } else {
          addDownload(item.path, item.name, item.size);
        }
      }
    }
    await collectFiles(path);
  }

  function retryTask(id: string) {
    const task = tasks.value.find((t) => t.id === id);
    if (!task || task.status !== "failed") return;
    task.status = "pending";
    task.progress = 0;
    task.error = undefined;
    if (task.type === "upload") {
      uploadQueue.push(task);
      drainUploadQueue();
    } else {
      downloadQueue.push(task);
      drainDownloadQueue();
    }
  }

  function removeTask(id: string) {
    const idx = tasks.value.findIndex((t) => t.id === id);
    if (idx !== -1) tasks.value.splice(idx, 1);
  }

  function clearCompleted() {
    tasks.value = tasks.value.filter((t) => t.status !== "done");
  }

  function drainUploadQueue() {
    while (
      uploadRunning < maxUploadConcurrency.value &&
      uploadQueue.length > 0
    ) {
      const task = uploadQueue.shift()!;
      if (task.status !== "pending") continue;
      uploadRunning++;
      task.status = "running";
      runUploadTask(task).finally(() => {
        uploadRunning--;
        drainUploadQueue();
      });
    }
  }

  function drainDownloadQueue() {
    while (
      downloadRunning < maxDownloadConcurrency.value &&
      downloadQueue.length > 0
    ) {
      const task = downloadQueue.shift()!;
      if (task.status !== "pending") continue;
      downloadRunning++;
      task.status = "running";
      runDownloadTask(task).finally(() => {
        downloadRunning--;
        drainDownloadQueue();
      });
    }
  }

  async function runUploadTask(task: TransferTask) {
    const file = task.file;
    if (!file) {
      task.status = "failed";
      task.error = "文件对象丢失";
      return;
    }
    try {
      // Ensure destination directory exists
      if (task.path && task.path !== ".") {
        await fileApi.mkdir(task.path);
      }
      const totalChunks = Math.ceil(file.size / CHUNK_SIZE);
      const session = await uploadApi.init(
        task.path,
        file.name,
        file.size,
        totalChunks,
      );
      const uploadedSet = new Set(session.uploadedChunks);

      for (let i = 0; i < totalChunks; i++) {
        if (uploadedSet.has(i)) continue;
        const start = i * CHUNK_SIZE;
        const end = Math.min(start + CHUNK_SIZE, file.size);
        const chunk = file.slice(start, end);
        await uploadApi.chunk(session.uploadId, i, chunk, (p) => {
          task.progress = Math.round(((i + p / 100) / totalChunks) * 100);
        });
        task.progress = Math.round(((i + 1) / totalChunks) * 100);
      }

      await uploadApi.complete(session.uploadId);
      task.progress = 100;
      task.status = "done";
    } catch (e: unknown) {
      task.status = "failed";
      task.error = e instanceof Error ? e.message : "上传失败";
    }
  }

  async function runDownloadTask(task: TransferTask) {
    try {
      const { download } = useDownload();
      download(task.path, task.name);
      task.progress = 100;
      task.status = "done";
    } catch (e: unknown) {
      task.status = "failed";
      task.error = e instanceof Error ? e.message : "下载失败";
    }
  }

  return {
    tasks,
    maxUploadConcurrency,
    maxDownloadConcurrency,
    uploadTasks,
    downloadTasks,
    runningCount,
    addUpload,
    addFolderUpload,
    addDownload,
    addFolderDownload,
    retryTask,
    removeTask,
    clearCompleted,
  };
});
