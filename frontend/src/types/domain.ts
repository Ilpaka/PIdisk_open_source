export type ProfileID = string;

export interface Profile {
  id: ProfileID;
  name: string;
  host: string;
  port: number;
  username: string;
  privateKeyPath: string;
  authMethod: string;
  rootDir: string;
  trashDir: string;
  localSyncDir: string;
  createdAt: string;
  lastUsedAt: string;
}

export interface ProfileInput {
  name: string;
  host: string;
  port: number;
  username: string;
  privateKeyPath: string;
  passphrase?: string;
  rootDir: string;
  trashDir: string;
  localSyncDir?: string;
}

export interface FileEntry {
  name: string;
  path: string;
  isDir: boolean;
  size: number;
  modified: string;
  mode: number;
}

export interface Listing {
  path: string;
  entries: FileEntry[];
}

export interface DiskUsage {
  path: string;
  used: number;
  free: number;
  total: number;
  percent: number;
  raw?: string;
}

export interface ConnectionState {
  profileId: ProfileID;
  connected: boolean;
  lastError?: string;
  lastPing: string;
}

export interface KnownHost {
  host: string;
  port: number;
  keyType: string;
  fingerprint: string;
  publicKey?: string;
  addedAt: string;
}

export interface TransferProgress {
  id: string;
  kind: "upload" | "download" | "download_dir";
  name: string;
  localPath?: string;
  remotePath?: string;
  bytesDone: number;
  bytesTotal: number;
  speed: number;
  state: "running" | "done" | "error" | "cancelled";
  startedAt: string;
  completedAt?: string;
  error?: string;
}

export interface SyncFolder {
  name: string;
  localPath: string;
  remotePath: string;
  enabled: boolean;
  lastSync: string;
  direction: "both" | "to_remote" | "to_local";
}

export interface SyncStats {
  isRunning: boolean;
  lastSyncTime: string;
  syncedFolders: string[];
  errors: string[];
  filesSynced: number;
  bytesSynced: number;
  conflicts: SyncConflict[];
}

export interface SyncConflict {
  folder: string;
  path: string;
  localMtime: string;
  remoteMtime: string;
  resolution: string;
}

export interface TrashEntry {
  id: string;
  originalPath: string;
  trashedPath: string;
  deletedAt: string;
  profileId: ProfileID;
  isDir: boolean;
  size: number;
}
