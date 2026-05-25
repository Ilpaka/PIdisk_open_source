import { create } from "zustand";
import type { FileEntry } from "@/types/domain";
import { filesApi } from "@/api/files";

interface FilesState {
  cwd: string;
  entries: FileEntry[];
  selection: Set<string>;
  loading: boolean;
  error: string | null;
  setCwd: (path: string) => Promise<void>;
  refresh: () => Promise<void>;
  toggleSelect: (path: string) => void;
  selectOnly: (path: string) => void;
  selectAll: () => void;
  clearSelection: () => void;
}

export const useFilesStore = create<FilesState>((set, get) => ({
  cwd: "",
  entries: [],
  selection: new Set<string>(),
  loading: false,
  error: null,

  async setCwd(p) {
    set({ cwd: p, loading: true, error: null });
    try {
      const listing = await filesApi.readDir(p);
      set({
        cwd: listing.path,
        entries: listing.entries ?? [],
        selection: new Set(),
        loading: false,
      });
    } catch (err) {
      set({ error: String(err), loading: false, entries: [] });
    }
  },

  async refresh() {
    const cwd = get().cwd;
    if (!cwd) return;
    await get().setCwd(cwd);
  },

  toggleSelect(p) {
    const next = new Set(get().selection);
    if (next.has(p)) next.delete(p);
    else next.add(p);
    set({ selection: next });
  },

  selectOnly(p) {
    set({ selection: new Set([p]) });
  },

  selectAll() {
    set({ selection: new Set(get().entries.map((e) => e.path)) });
  },

  clearSelection() {
    set({ selection: new Set() });
  },
}));
