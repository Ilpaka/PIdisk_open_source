import { useEffect } from "react";

export interface HotkeyHandlers {
  onRename?: () => void;
  onDelete?: () => void;
  onSelectAll?: () => void;
  onEscape?: () => void;
  onBack?: () => void;
  onRefresh?: () => void;
}

export function useHotkeys(handlers: HotkeyHandlers): void {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement | null;
      if (target && ["INPUT", "TEXTAREA"].includes(target.tagName)) return;

      const mod = e.metaKey || e.ctrlKey;
      switch (e.key) {
        case "F2":
          handlers.onRename?.();
          break;
        case "Delete":
        case "Backspace":
          if (e.key === "Backspace" && !mod) {
            handlers.onBack?.();
          } else {
            handlers.onDelete?.();
          }
          break;
        case "Escape":
          handlers.onEscape?.();
          break;
        case "a":
        case "A":
          if (mod) {
            e.preventDefault();
            handlers.onSelectAll?.();
          }
          break;
        case "F5":
          handlers.onRefresh?.();
          break;
        case "r":
        case "R":
          if (mod) {
            e.preventDefault();
            handlers.onRefresh?.();
          }
          break;
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [handlers]);
}
