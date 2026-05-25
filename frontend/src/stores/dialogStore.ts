import { create } from "zustand";

export interface ConfirmOptions {
  title?: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  destructive?: boolean;
}

export interface InputOptions {
  title?: string;
  message?: string;
  label?: string;
  defaultValue?: string;
  placeholder?: string;
  confirmText?: string;
  cancelText?: string;
}

interface ConfirmState extends ConfirmOptions {
  open: boolean;
  resolver: ((value: boolean) => void) | null;
}

interface InputState extends InputOptions {
  open: boolean;
  resolver: ((value: string | null) => void) | null;
}

interface DialogState {
  confirm: ConfirmState;
  input: InputState;
  askConfirm: (opts: ConfirmOptions) => Promise<boolean>;
  resolveConfirm: (value: boolean) => void;
  askInput: (opts: InputOptions) => Promise<string | null>;
  resolveInput: (value: string | null) => void;
}

const emptyConfirm: ConfirmState = {
  open: false,
  message: "",
  resolver: null,
};
const emptyInput: InputState = {
  open: false,
  resolver: null,
};

export const useDialogStore = create<DialogState>((set, get) => ({
  confirm: emptyConfirm,
  input: emptyInput,

  askConfirm: (opts) =>
    new Promise<boolean>((resolve) => {
      set({
        confirm: {
          open: true,
          title: opts.title ?? "Confirm",
          message: opts.message,
          confirmText: opts.confirmText ?? "OK",
          cancelText: opts.cancelText ?? "Cancel",
          destructive: opts.destructive ?? false,
          resolver: resolve,
        },
      });
    }),

  resolveConfirm: (value) => {
    const r = get().confirm.resolver;
    set({ confirm: emptyConfirm });
    r?.(value);
  },

  askInput: (opts) =>
    new Promise<string | null>((resolve) => {
      set({
        input: {
          open: true,
          title: opts.title ?? "Enter value",
          message: opts.message,
          label: opts.label ?? "Value",
          defaultValue: opts.defaultValue ?? "",
          placeholder: opts.placeholder,
          confirmText: opts.confirmText ?? "OK",
          cancelText: opts.cancelText ?? "Cancel",
          resolver: resolve,
        },
      });
    }),

  resolveInput: (value) => {
    const r = get().input.resolver;
    set({ input: emptyInput });
    r?.(value);
  },
}));

// Convenience helpers, so call sites read as cleanly as window.confirm/prompt did.
export const confirmDialog = (opts: ConfirmOptions) =>
  useDialogStore.getState().askConfirm(opts);

export const promptDialog = (opts: InputOptions) =>
  useDialogStore.getState().askInput(opts);
