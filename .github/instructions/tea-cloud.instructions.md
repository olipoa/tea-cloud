---
applyTo: "**"
---

# Tea Cloud 项目编码规范

## 项目简介

Tea Cloud 是一个局域网文件共享与媒体投屏应用：

- **前端**：Vue 3 + TypeScript + Vite + Pinia + Naive UI + ArtPlayer
- **后端**：Go + Gin，按 handlers / services / middleware 三层组织
- 构建时前端输出到 `backend/static/`，后端以嵌入方式提供前端资源

---

## 前端规范（`frontend/src/**`）

### Vue 组件

- 统一使用 **Composition API + `<script setup lang="ts">`**，禁用 Options API
- 文件结构顺序：`<template>` → `<script setup>` → `<style scoped>`
- 组件文件名使用 **PascalCase**（如 `AudioPlayerBar.vue`）
- 样式必须加 `scoped` 且使用 **SCSS**：`<style scoped lang="scss">`
- 穿透子组件样式用 `:deep(.selector)`

### TypeScript

- 严格模式（`strict: true`），路径别名 `@/` 指向 `src/`
- 类型/接口名使用 **PascalCase**（如 `FileInfo`、`DLNARenderer`）
- 变量、函数名使用 **camelCase**
- 导入类型时用 `import { type Xxx } from '...'`

### Pinia Store

- 统一使用 **Composition API 风格**的 `defineStore`
- 命名规则：`export const use[Name]Store = defineStore('[name]', () => { ... })`
- 状态全部用 `ref()`，派生值用 `computed()`
- 文件名使用 camelCase，位于 `src/stores/`

```typescript
// 正确示例
export const useFileStore = defineStore("file", () => {
  const currentPath = ref(".");
  const items = ref<FileInfo[]>([]);
  const breadcrumbs = computed(() => {
    /* ... */
  });
  async function loadDir(path: string) {
    /* ... */
  }
  return { currentPath, items, breadcrumbs, loadDir };
});
```

### API 服务（`src/services/api.ts`）

- Axios 实例，无超时限制（支持大文件上传）
- API 按业务分组，导出命名对象：`fileApi`、`peerApi`、`castApi`
- 方法以动词开头：`list()`、`upload()`、`delete()`、`mkdir()`
- 返回 URL 的方法直接返回字符串：`rawUrl(path)`、`downloadUrl(path)`
- 使用 `.then(r => r.data)` 提取响应数据

```typescript
export const fileApi = {
  /** 列出目录内容 */
  list(path: string): Promise<FileInfo[]> {
    return http.get("/api/files", { params: { path } }).then((r) => r.data);
  },
};
```

### Composables（`src/composables/`）

- 文件名和函数名均以 `use` 开头（如 `useCast.ts`）
- 返回对象，包含响应式状态和方法
- 可访问 Store 和 Naive UI 工具函数

### 异步错误处理（统一模式）

```typescript
async function operation() {
  try {
    const result = await apiCall();
    message.success("操作成功");
    return result;
  } catch (e) {
    message.error("操作失败");
  } finally {
    // 清理 loading 状态等
  }
}
```

### 样式约定

- 全局样式：`src/styles/main.scss`
- 主题色：`#18a058`（绿）
- SCSS 变量：`$breakpoint: 768px` 等
- 响应式使用 media query

---

## 后端规范（`backend/**`）

- Web 框架：**Gin**，三层组织：`handlers/`（HTTP 层）→ `services/`（业务逻辑）→ `middleware/`
- 日志格式：`log.Printf("message: %v", value)`，错误日志加 `[WARN]`/`[ERROR]` 前缀
- 配置统一在 `config/config.go` 加载

---

## 文件命名规范

| 类型            | 约定             | 示例                      |
| --------------- | ---------------- | ------------------------- |
| Vue 组件        | PascalCase       | `FileExplorer.vue`        |
| TypeScript 脚本 | camelCase        | `audioPlayer.ts`          |
| 目录            | 小写             | `composables/`, `stores/` |
| Go 文件         | snake_case       | `cast_service.go`         |
| 类型/接口       | PascalCase       | `FileInfo`, `PeerInfo`    |
| Store           | `use[Name]Store` | `useMediaPlayerStore`     |
| Composable      | `use[Name]`      | `useCast`                 |
| API 对象        | `[name]Api`      | `fileApi`, `castApi`      |

---

## 构建与开发

- 开发服务器：`pnpm dev`（端口 5173，`/api` 和 `/raw` 代理到 `:8080`）
- 类型检查 + 构建：`pnpm build`（先 `vue-tsc --noEmit` 后 `vite build`）
- 构建输出：`backend/static/`（`emptyOutDir: true`）
- 包管理器：**pnpm**
