import axios from "axios";

const http = axios.create({
  baseURL: "/",
  timeout: 0, // no timeout for large file uploads
});

export interface FileInfo {
  name: string;
  path: string;
  size: number;
  isDir: boolean;
  modTime: number;
  mime: string;
  ext: string;
}

export interface PeerInfo {
  name: string;
  host: string;
  port: number;
  addrV4: string;
  url: string;
}

export interface SelfInfo {
  name: string;
  port: number;
  ips: string[];
}

// Current base URL (default: same origin)
let _baseUrl = "";

export function setBaseUrl(url: string) {
  _baseUrl = url.replace(/\/$/, "");
  http.defaults.baseURL = _baseUrl + "/";
}

export function getBaseUrl(): string {
  return _baseUrl;
}

export const fileApi = {
  /** List directory contents */
  list(path = "."): Promise<FileInfo[]> {
    return http.get("/api/files", { params: { path } }).then((r) => r.data);
  },

  /** Get download URL for a file */
  downloadUrl(path: string): string {
    return `${_baseUrl}/api/files/download?path=${encodeURIComponent(path)}&download=1`;
  },

  /** Get streaming/preview URL (used for video/audio/image) */
  rawUrl(path: string): string {
    return `${_baseUrl}/raw/${encodeURIComponent(path)}`;
  },

  /** Get thumbnail URL for a video file (requires ffmpeg on server) */
  thumbnailUrl(path: string): string {
    return `${_baseUrl}/api/files/thumbnail?path=${encodeURIComponent(path)}`;
  },

  /** Delete a file or folder */
  delete(path: string): Promise<void> {
    return http
      .delete("/api/files", { params: { path } })
      .then(() => undefined);
  },

  /** Create a directory */
  mkdir(path: string): Promise<void> {
    return http
      .post("/api/dirs", null, { params: { path } })
      .then(() => undefined);
  },
};

export const peerApi = {
  /** Discover peers on the LAN */
  list(): Promise<PeerInfo[]> {
    return http.get("/api/peers").then((r) => r.data);
  },

  /** Get this node's info */
  self(): Promise<SelfInfo> {
    return http.get("/api/self").then((r) => r.data);
  },
};

export interface DLNARenderer {
  name: string;
  location: string;
  protocol: "dlna" | "airplay";
}

export const castApi = {
  /** Discover AirPlay / DLNA cast devices on the LAN */
  devices(): Promise<DLNARenderer[]> {
    return http.get("/api/cast/devices").then((r) => r.data);
  },

  /** Cast the given file path to a renderer */
  send(location: string, filePath: string): Promise<void> {
    return http
      .post("/api/cast", { location, path: filePath })
      .then(() => undefined);
  },
};

export interface UploadInitResponse {
  uploadId: string;
  uploadedChunks: number[];
}

export const uploadApi = {
  /** Initialize a resumable upload session */
  init(
    saveDir: string,
    filename: string,
    totalSize: number,
    totalChunks: number,
  ): Promise<UploadInitResponse> {
    return http
      .post("/api/uploads", { saveDir, filename, totalSize, totalChunks })
      .then((r) => r.data);
  },

  /** Get upload session status (list of already-uploaded chunk indices) */
  status(uploadId: string): Promise<UploadInitResponse> {
    return http.get(`/api/uploads/${uploadId}`).then((r) => r.data);
  },

  /** Upload a single chunk */
  chunk(
    uploadId: string,
    index: number,
    data: Blob,
    onProgress?: (percent: number) => void,
  ): Promise<void> {
    return http
      .put(`/api/uploads/${uploadId}/chunk`, data, {
        params: { index },
        headers: { "Content-Type": "application/octet-stream" },
        onUploadProgress: (e) => {
          if (onProgress && e.total) {
            onProgress(Math.round((e.loaded * 100) / e.total));
          }
        },
      })
      .then(() => undefined);
  },

  /** Finalize upload — server merges all chunks and returns file info */
  complete(uploadId: string): Promise<FileInfo> {
    return http.post(`/api/uploads/${uploadId}/complete`).then((r) => r.data);
  },
};
