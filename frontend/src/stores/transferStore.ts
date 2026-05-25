import { create } from "zustand";
import type { TransferProgress } from "@/types/domain";

interface TransferState {
  active: Record<string, TransferProgress>;
  finished: TransferProgress[];
  upsert: (progress: TransferProgress) => void;
  complete: (progress: TransferProgress) => void;
  remove: (id: string) => void;
  clearFinished: () => void;
}

export const useTransferStore = create<TransferState>((set) => ({
  active: {},
  finished: [],

  upsert(progress) {
    set((state) => ({
      active: { ...state.active, [progress.id]: progress },
    }));
  },

  complete(progress) {
    set((state) => {
      const { [progress.id]: _, ...rest } = state.active;
      const trimmed = [progress, ...state.finished].slice(0, 20);
      return { active: rest, finished: trimmed };
    });
  },

  remove(id) {
    set((state) => {
      const { [id]: _, ...rest } = state.active;
      return { active: rest };
    });
  },

  clearFinished() {
    set({ finished: [] });
  },
}));
