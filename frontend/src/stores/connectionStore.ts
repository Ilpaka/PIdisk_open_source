import { create } from "zustand";
import type { ConnectionState, KnownHost } from "@/types/domain";
import { connectionApi } from "@/api/connection";

interface ConnectionStore {
  connected: boolean;
  lastError: string | null;
  hostKeyPrompt: KnownHost | null;
  setState: (state: ConnectionState) => void;
  showHostKeyPrompt: (entry: KnownHost) => void;
  resolveHostKey: (accept: boolean) => Promise<void>;
  reset: () => void;
}

export const useConnectionStore = create<ConnectionStore>((set, get) => ({
  connected: false,
  lastError: null,
  hostKeyPrompt: null,
  setState: (state) =>
    set({
      connected: state.connected,
      lastError: state.lastError ?? null,
    }),
  showHostKeyPrompt: (entry) => set({ hostKeyPrompt: entry }),
  async resolveHostKey(accept) {
    const prompt = get().hostKeyPrompt;
    if (!prompt) return;
    await connectionApi.confirmHostKey(prompt.fingerprint, accept);
    set({ hostKeyPrompt: null });
  },
  reset: () =>
    set({ connected: false, lastError: null, hostKeyPrompt: null }),
}));
