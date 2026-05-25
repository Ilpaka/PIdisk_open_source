import { useEffect, useState } from "react";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Stack,
  TextField,
} from "@mui/material";
import { useDialogStore } from "@/stores/dialogStore";

/**
 * Renders the global confirm / input dialogs. Mounted once at the app root.
 * window.confirm and window.prompt are disabled inside the Wails WebKit
 * webview, so any call site that needs that flow goes through dialogStore.
 */
export default function DialogHost() {
  const confirm = useDialogStore((s) => s.confirm);
  const resolveConfirm = useDialogStore((s) => s.resolveConfirm);
  const input = useDialogStore((s) => s.input);
  const resolveInput = useDialogStore((s) => s.resolveInput);

  return (
    <>
      <ConfirmDialog
        open={confirm.open}
        title={confirm.title ?? "Confirm"}
        message={confirm.message}
        confirmText={confirm.confirmText ?? "OK"}
        cancelText={confirm.cancelText ?? "Cancel"}
        destructive={Boolean(confirm.destructive)}
        onAnswer={resolveConfirm}
      />
      <InputDialog
        open={input.open}
        title={input.title ?? "Enter value"}
        message={input.message}
        label={input.label ?? "Value"}
        defaultValue={input.defaultValue ?? ""}
        placeholder={input.placeholder}
        confirmText={input.confirmText ?? "OK"}
        cancelText={input.cancelText ?? "Cancel"}
        onAnswer={resolveInput}
      />
    </>
  );
}

interface ConfirmProps {
  open: boolean;
  title: string;
  message: string;
  confirmText: string;
  cancelText: string;
  destructive: boolean;
  onAnswer: (value: boolean) => void;
}

function ConfirmDialog({
  open,
  title,
  message,
  confirmText,
  cancelText,
  destructive,
  onAnswer,
}: ConfirmProps) {
  return (
    <Dialog open={open} onClose={() => onAnswer(false)} maxWidth="xs" fullWidth>
      <DialogTitle>{title}</DialogTitle>
      <DialogContent>
        <DialogContentText sx={{ whiteSpace: "pre-line" }}>
          {message}
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button onClick={() => onAnswer(false)}>{cancelText}</Button>
        <Button
          variant="contained"
          color={destructive ? "error" : "primary"}
          onClick={() => onAnswer(true)}
          autoFocus
        >
          {confirmText}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

interface InputProps {
  open: boolean;
  title: string;
  message?: string;
  label: string;
  defaultValue: string;
  placeholder?: string;
  confirmText: string;
  cancelText: string;
  onAnswer: (value: string | null) => void;
}

function InputDialog({
  open,
  title,
  message,
  label,
  defaultValue,
  placeholder,
  confirmText,
  cancelText,
  onAnswer,
}: InputProps) {
  const [value, setValue] = useState(defaultValue);

  useEffect(() => {
    if (open) setValue(defaultValue);
  }, [open, defaultValue]);

  const submit = () => {
    onAnswer(value);
  };

  return (
    <Dialog open={open} onClose={() => onAnswer(null)} maxWidth="xs" fullWidth>
      <DialogTitle>{title}</DialogTitle>
      <DialogContent>
        <Stack spacing={2}>
          {message ? (
            <DialogContentText sx={{ whiteSpace: "pre-line" }}>
              {message}
            </DialogContentText>
          ) : null}
          <TextField
            autoFocus
            fullWidth
            label={label}
            value={value}
            onChange={(e) => setValue(e.target.value)}
            placeholder={placeholder}
            inputProps={{
              autoCapitalize: "none",
              autoCorrect: "off",
              spellCheck: false,
            }}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                submit();
              }
            }}
          />
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={() => onAnswer(null)}>{cancelText}</Button>
        <Button variant="contained" onClick={submit}>
          {confirmText}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
