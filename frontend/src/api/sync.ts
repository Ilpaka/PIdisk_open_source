import {
  AddSyncFolder,
  GetSyncStatus,
  ListSyncFolders,
  RemoveSyncFolder,
  StartSync,
  StopSync,
  ToggleSyncFolder,
} from "../../wailsjs/go/wailsapp/SyncBindings";
import type { SyncFolder, SyncStats } from "@/types/domain";

export const syncApi = {
  list: (): Promise<SyncFolder[]> => ListSyncFolders() as Promise<SyncFolder[]>,
  add: (folder: SyncFolder): Promise<SyncFolder> =>
    AddSyncFolder(folder as Parameters<typeof AddSyncFolder>[0]) as Promise<SyncFolder>,
  remove: (name: string): Promise<void> => RemoveSyncFolder(name),
  toggle: (name: string, enabled: boolean): Promise<SyncFolder> =>
    ToggleSyncFolder(name, enabled) as Promise<SyncFolder>,
  start: (): Promise<void> => StartSync(),
  stop: (): Promise<void> => StopSync(),
  status: (): Promise<SyncStats> => GetSyncStatus() as Promise<SyncStats>,
};
