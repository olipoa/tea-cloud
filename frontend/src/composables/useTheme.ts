import { onMounted, onUnmounted, ref, watch } from "vue";

type ThemeMode = "light" | "dark" | "system";

const STORAGE_KEY = "tea-cloud-theme";

const theme = ref<ThemeMode>(
  (localStorage.getItem(STORAGE_KEY) as ThemeMode) || "system",
);

function applyTheme(mode: ThemeMode) {
  const prefersDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
  const isDark = mode === "dark" || (mode === "system" && prefersDark);
  document.documentElement.setAttribute("theme-mode", isDark ? "dark" : "");
}

applyTheme(theme.value);

export function useTheme() {
  function setTheme(mode: ThemeMode) {
    theme.value = mode;
    localStorage.setItem(STORAGE_KEY, mode);
    applyTheme(mode);
  }

  function toggleDark() {
    const current =
      theme.value === "dark" ||
      (theme.value === "system" &&
        window.matchMedia("(prefers-color-scheme: dark)").matches);
    setTheme(current ? "light" : "dark");
  }

  const isDark = ref(false);

  function updateIsDark() {
    const prefersDark = window.matchMedia(
      "(prefers-color-scheme: dark)",
    ).matches;
    isDark.value =
      theme.value === "dark" || (theme.value === "system" && prefersDark);
  }

  updateIsDark();

  const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");

  onMounted(() => {
    mediaQuery.addEventListener("change", () => {
      updateIsDark();
      if (theme.value === "system") applyTheme("system");
    });
  });

  onUnmounted(() => {
    mediaQuery.removeEventListener("change", updateIsDark);
  });

  watch(theme, updateIsDark);

  return { theme, isDark, setTheme, toggleDark };
}
