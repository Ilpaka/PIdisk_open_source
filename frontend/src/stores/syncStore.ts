import { create } from "zustand";
import type { SyncFolder, SyncStats } from "@/types/domain";
import { syncApi } from "@/api/sync";

const emptyStats: SyncStats = {
  isRunning: false,
  lastSyncTime: "",
  syncedFolders: [],
  errors: [],
  filesSynced: 0,
  bytesSynced: 0,
  conflicts: [],
};

interface SyncState {
  folders: SyncFolder[];
  stats: SyncStats;
  loading: boolean;
  refreshFolders: () => Promise<void>;
  setStats: (s: SyncStats) => void;
  add: (folder: SyncFolder) => Promise<void>;
  remove: (name: string) => Promise<void>;
  toggle: (name: string, enabled: boolean) => Promise<void>;
  start: () => Promise<void>;
  stop: () => Promise<void>;
  refreshStatus: () => Promise<void>;
}

export const useSyncStore = create<SyncState>((set, get) => ({
  folders: [],
  stats: emptyStats,
  loading: false,

  async refreshFolders() {
    set({ loading: true });
    try {
      const folders = await syncApi.list();
      set({ folders, loading: false });
    } catch {
      set({ loading: false });
    }
  },

  setStats: (stats) => set({ stats }),

  async add(folder) {
    await syncApi.add(folder);
    await get().refreshFolders();
  },

  async remove(name) {
    await syncApi.remove(name);
    await get().refreshFolders();
  },

  async toggle(name, enabled) {
    await syncApi.toggle(name, enabled);
    await get().refreshFolders();
  },

  async start() {
    await syncApi.start();
    await get().refreshStatus();
  },

  async stop() {
    await syncApi.stop();
    await get().refreshStatus();
  },

  async refreshStatus() {
    const stats = await syncApi.status();
    set({ stats });
  },
}));
