import {
  ClearAllTrash,
  ListTrash,
  RestoreFromTrash,
} from "../../wailsjs/go/wailsapp/TrashBindings";
import type { TrashEntry } from "@/types/domain";

export const trashApi = {
  list: (): Promise<TrashEntry[]> => ListTrash() as Promise<TrashEntry[]>,
  restore: (id: string): Promise<TrashEntry> =>
    RestoreFromTrash(id) as Promise<TrashEntry>,
  clear: (): Promise<void> => ClearAllTrash(),
};
