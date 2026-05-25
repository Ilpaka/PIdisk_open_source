import { create } from "zustand";
import type { ThemeMode } from "@/theme";

interface SettingsState {
  themeMode: ThemeMode;
  setThemeMode: (mode: ThemeMode) => void;
}

const storageKey = "pidisk.themeMode";

const detectSystemTheme = (): ThemeMode =>
  typeof window !== "undefined" &&
  window.matchMedia &&
  window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";

const initialMode = ((): ThemeMode => {
  if (typeof window === "undefined") return "light";
  const stored = window.localStorage?.getItem(storageKey);
  if (stored === "light" || stored === "dark") return stored;
  return detectSystemTheme();
})();

export const useSettingsStore = create<SettingsState>((set) => ({
  themeMode: initialMode,
  setThemeMode: (themeMode) => {
    set({ themeMode });
    try {
      window.localStorage.setItem(storageKey, themeMode);
    } catch {
      // ignore persistence failure
    }
  },
}));
