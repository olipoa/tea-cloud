# Tea Cloud - 局域网云盘

局域网文件共享应用，每台电脑运行同一个程序后都是一个节点，通过 mDNS 自动发现对方，用浏览器即可浏览、预览、上传、下载文件。

## 功能

- 📁 **文件浏览** — 列表/网格双视图，支持文件夹导航
- ⬆️ **文件上传** — 拖拽上传，支持多文件，带进度条
- ⬇️ **文件下载** — 一键下载
- 🎬 **视频播放** — 浏览器原生播放 mp4/webm/mkv 等，支持拖拽进度条（HTTP Range）
- 🎵 **音频播放** — mp3/flac/wav/ogg 等
- 🖼️ **图片预览** — jpg/png/gif/webp/svg 等，支持缩放旋转
- 📄 **PDF 预览** — 在线渲染，支持翻页缩放
- 📝 **文本预览** — txt/md/json/代码文件等
- 🌐 **节点发现** — mDNS 自动发现局域网内其他 Tea Cloud 节点，一键切换浏览
- 💾 **单二进制** — 前端嵌入后端，无需额外部署

## 快速开始

### 直接运行

```
.\tea-cloud.exe
```

启动后在浏览器访问终端输出的地址，默认 `http://localhost:8080`。

共享目录默认为 `%USERPROFILE%\TeaCloud`（Windows）或 `~/tea-cloud`（macOS/Linux）。

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `TEA_SHARE_DIR` | 共享目录路径 | `~/TeaCloud` |
| `TEA_NODE_NAME` | 节点显示名称 | 主机名 |

示例：
```powershell
$env:TEA_SHARE_DIR = "D:\共享文件"; .\tea-cloud.exe
```

## 构建

需要：Go 1.22+，Node.js 18+

```powershell
.\build.ps1
```

或手动：
```powershell
# 构建前端
cd frontend
npm install
npm run build

# 构建后端（会嵌入前端静态资源）
cd ..\backend
go build -o ..\tea-cloud.exe .
```

## 开发调试

```powershell
# 前端热更新（proxy 到后端 8080）
cd frontend
npm run dev   # 访问 http://localhost:5173

# 后端
cd backend
go run .
```

## 技术栈

- **后端**：Go + Gin + grandcat/zeroconf（mDNS）
- **前端**：Vue 3 + TypeScript + Vite + Element Plus + pdfjs-dist

## API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/files?path=` | 列出目录内容 |
| GET | `/api/files/download?path=` | 下载/流式播放文件（Range 支持） |
| POST | `/api/files/upload?path=` | 上传文件（multipart） |
| DELETE | `/api/files?path=` | 删除文件/目录 |
| POST | `/api/dirs?path=` | 创建目录 |
| GET | `/api/peers` | 扫描局域网节点（~2秒） |
| GET | `/api/self` | 本机节点信息 |
| GET | `/raw/*path` | 直接访问原始文件（用于预览） |
