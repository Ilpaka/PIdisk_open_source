import { FormEvent, useState } from "react";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import type { Profile } from "@/types/domain";

interface Props {
  profile: Profile | null;
  onSubmit: (passphrase: string) => Promise<void>;
  onCancel: () => void;
}

export default function PassphrasePrompt({ profile, onSubmit, onCancel }: Props) {
  const [value, setValue] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!profile) return null;

  const handle = async (e: FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      await onSubmit(value);
      setValue("");
    } catch (err) {
      setError(String(err));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open onClose={onCancel} maxWidth="xs" fullWidth>
      <DialogTitle>Unlock "{profile.name}"</DialogTitle>
      <form onSubmit={handle}>
        <DialogContent dividers>
          <Stack spacing={2}>
            <Typography variant="body2" color="text.secondary">
              The private key is encrypted. Enter its passphrase to connect.
            </Typography>
            <TextField
              autoFocus
              type="password"
              label="Passphrase"
              value={value}
              onChange={(e) => setValue(e.target.value)}
              fullWidth
              inputProps={{ autoCapitalize: "none", autoCorrect: "off", spellCheck: false }}
            />
            {error ? (
              <Typography color="error" variant="body2">
                {error}
              </Typography>
            ) : null}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={onCancel}>Cancel</Button>
          <Button type="submit" variant="contained" disabled={submitting}>
            {submitting ? "Connecting..." : "Connect"}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
}
