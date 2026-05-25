import {
  EventsOn,
  EventsOff,
} from "../../wailsjs/runtime/runtime";
import type {
  ConnectionState,
  KnownHost,
  TransferProgress,
  SyncConflict,
  SyncStats,
} from "@/types/domain";

export type EventPayload = {
  "connection:state": ConnectionState;
  "hostkey:prompt": KnownHost;
  "hostkey:mismatch": KnownHost;
  "transfer:started": TransferProgress;
  "transfer:progress": TransferProgress;
  "transfer:done": TransferProgress;
  "sync:status": SyncStats;
  "sync:conflict": SyncConflict;
  "trash:updated": Record<string, never>;
};

export function on<K extends keyof EventPayload>(
  name: K,
  handler: (payload: EventPayload[K]) => void,
): () => void {
  EventsOn(name, handler as (...data: unknown[]) => void);
  return () => EventsOff(name);
}
