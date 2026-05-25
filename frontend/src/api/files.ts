import {
  ClearTrash,
  DiskUsage,
  Join,
  Mkdir,
  Move,
  ReadDir,
  Remove,
  Rename,
} from "../../wailsjs/go/wailsapp/FileBindings";
import type { DiskUsage as DiskUsageT, Listing } from "@/types/domain";

export interface RemoveResult {
  trashed: boolean;
  originalPath?: string;
  trashedPath?: string;
  isDir: boolean;
  size: number;
}

export const filesApi = {
  readDir: (p: string): Promise<Listing> => ReadDir(p) as Promise<Listing>,
  mkdir: (parent: string, name: string): Promise<string> => Mkdir(parent, name),
  move: (src: string, dst: string): Promise<void> => Move(src, dst),
  rename: (cwd: string, oldName: string, newName: string): Promise<string> =>
    Rename(cwd, oldName, newName),
  remove: (target: string): Promise<RemoveResult> =>
    Remove(target) as Promise<RemoveResult>,
  clearTrash: (): Promise<void> => ClearTrash(),
  diskUsage: (p: string): Promise<DiskUsageT> =>
    DiskUsage(p) as Promise<DiskUsageT>,
  join: (parent: string, child: string): Promise<string> => Join(parent, child),
};
