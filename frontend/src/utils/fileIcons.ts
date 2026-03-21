import { getCategory } from "@/utils/fileUtils";

/** Returns a TDesign icon name for a file type */
export function fileIconName(ext: string, isDir: boolean): string {
  if (isDir) return "folder";
  const cat = getCategory(ext);
  const map: Record<string, string> = {
    video: "play-circle",
    audio: "music",
    image: "image",
    pdf: "file-pdf",
    text: "file-code",
    archive: "file-zip",
    other: "file",
  };
  return map[cat] ?? "file";
}

/** Returns a color class for the icon */
export function fileIconColor(ext: string, isDir: boolean): string {
  if (isDir) return "color-folder";
  const cat = getCategory(ext);
  const map: Record<string, string> = {
    video: "color-video",
    audio: "color-audio",
    image: "color-image",
    pdf: "color-pdf",
    text: "color-text",
    archive: "color-archive",
  };
  return map[cat] ?? "color-other";
}
