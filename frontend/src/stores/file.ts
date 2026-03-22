import { fileApi, type FileInfo } from "@/services/api";
import { getCategory } from "@/utils/fileUtils";
import { defineStore } from "pinia";
import { computed, ref } from "vue";

export type SortField = "modTime" | "name" | "type" | "size";
export type SortOrder = "asc" | "desc";

export const useFileStore = defineStore("file", () => {
  const currentPath = ref(".");
  const items = ref<FileInfo[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  // Default: grid on ≥768px, list on mobile
  const viewMode = ref<"list" | "grid">(
    window.innerWidth >= 768 ? "grid" : "list",
  );

  // Sort & filter
  const sortField = ref<SortField>("name");
  const sortOrder = ref<SortOrder>("asc");
  const filterType = ref<string>("all");

  // Search
  const searchKeyword = ref("");
  const searchResults = ref<FileInfo[]>([]);
  const isSearching = ref(false);
  const searchLoading = ref(false);

  // Breadcrumb segments derived from currentPath
  const breadcrumbs = computed(() => {
    if (currentPath.value === ".") return [{ label: "根目录", path: "." }];
    const parts = currentPath.value.split("/");
    const crumbs: { label: string; path: string }[] = [
      { label: "根目录", path: "." },
    ];
    let acc = "";
    for (const part of parts) {
      if (!part || part === ".") continue;
      acc = acc ? `${acc}/${part}` : part;
      crumbs.push({ label: part, path: acc });
    }
    return crumbs;
  });

  // Sorted & filtered items
  const sortedItems = computed(() => {
    let list = isSearching.value ? searchResults.value : [...items.value];

    // Filter by type
    if (filterType.value !== "all") {
      list = list.filter((item) => {
        if (item.isDir) return filterType.value === "folder";
        return getCategory(item.ext) === filterType.value;
      });
    }

    // Sort: always directories first (unless type filter is active)
    list.sort((a, b) => {
      if (filterType.value === "all") {
        if (a.isDir && !b.isDir) return -1;
        if (!a.isDir && b.isDir) return 1;
      }
      let cmp = 0;
      switch (sortField.value) {
        case "name":
          cmp = a.name.localeCompare(b.name);
          break;
        case "size":
          cmp = a.size - b.size;
          break;
        case "type":
          cmp = (a.ext || "").localeCompare(b.ext || "");
          break;
        case "modTime":
          cmp = a.modTime - b.modTime;
          break;
      }
      return sortOrder.value === "asc" ? cmp : -cmp;
    });
    return list;
  });

  async function loadDir(path = ".") {
    loading.value = true;
    error.value = null;
    isSearching.value = false;
    searchKeyword.value = "";
    try {
      items.value = await fileApi.list(path);
      currentPath.value = path;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "加载失败";
    } finally {
      loading.value = false;
    }
  }

  async function search(keyword: string, path?: string) {
    if (!keyword.trim()) {
      isSearching.value = false;
      searchKeyword.value = "";
      return;
    }
    searchKeyword.value = keyword;
    isSearching.value = true;
    searchLoading.value = true;
    try {
      searchResults.value = await fileApi.search(
        path ?? currentPath.value,
        keyword,
      );
    } catch {
      searchResults.value = [];
    } finally {
      searchLoading.value = false;
    }
  }

  function clearSearch() {
    isSearching.value = false;
    searchKeyword.value = "";
    searchResults.value = [];
  }

  function navigateTo(path: string) {
    loadDir(path);
  }

  function navigateUp() {
    if (currentPath.value === ".") return;
    const parts = currentPath.value.split("/");
    parts.pop();
    const parent = parts.join("/") || ".";
    loadDir(parent);
  }

  async function deleteItem(path: string) {
    await fileApi.delete(path);
    await loadDir(currentPath.value);
  }

  async function createDir(name: string) {
    const newPath =
      currentPath.value === "." ? name : `${currentPath.value}/${name}`;
    await fileApi.mkdir(newPath);
    await loadDir(currentPath.value);
  }

  async function refresh() {
    if (isSearching.value && searchKeyword.value) {
      await search(searchKeyword.value);
    } else {
      await loadDir(currentPath.value);
    }
  }

  return {
    currentPath,
    items,
    loading,
    error,
    viewMode,
    sortField,
    sortOrder,
    filterType,
    searchKeyword,
    searchResults,
    isSearching,
    searchLoading,
    sortedItems,
    breadcrumbs,
    loadDir,
    search,
    clearSearch,
    navigateTo,
    navigateUp,
    deleteItem,
    createDir,
    refresh,
  };
});
