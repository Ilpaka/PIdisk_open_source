import { create } from "zustand";

export type Severity = "success" | "info" | "warning" | "error";

export interface SnackbarMessage {
  id: number;
  severity: Severity;
  text: string;
  duration: number;
}

interface SnackbarState {
  current: SnackbarMessage | null;
  queue: SnackbarMessage[];
  push: (severity: Severity, text: string, duration?: number) => void;
  dismiss: () => void;
}

const defaultDuration: Record<Severity, number> = {
  success: 3000,
  info: 3000,
  warning: 5000,
  error: 6000,
};

let counter = 0;

export const useSnackbarStore = create<SnackbarState>((set, get) => ({
  current: null,
  queue: [],
  push(severity, text, duration) {
    counter += 1;
    const msg: SnackbarMessage = {
      id: counter,
      severity,
      text,
      duration: duration ?? defaultDuration[severity],
    };
    const { current, queue } = get();
    if (current) {
      set({ queue: [...queue, msg] });
    } else {
      set({ current: msg });
    }
  },
  dismiss() {
    const { queue } = get();
    if (queue.length === 0) {
      set({ current: null });
      return;
    }
    const [next, ...rest] = queue;
    set({ current: next, queue: rest });
  },
}));
