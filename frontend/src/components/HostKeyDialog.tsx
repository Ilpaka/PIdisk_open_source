import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  Typography,
} from "@mui/material";
import { useConnectionStore } from "@/stores/connectionStore";

export default function HostKeyDialog() {
  const prompt = useConnectionStore((s) => s.hostKeyPrompt);
  const resolve = useConnectionStore((s) => s.resolveHostKey);
  if (!prompt) return null;

  return (
    <Dialog open onClose={() => resolve(false)} maxWidth="sm" fullWidth>
      <DialogTitle>Verify host key</DialogTitle>
      <DialogContent dividers>
        <Stack spacing={2}>
          <Alert severity="warning">
            The server presented a key PIdisk has not seen before.
            Confirm the fingerprint out of band before trusting it.
          </Alert>
          <Stack direction="row" spacing={4}>
            <Typography color="text.secondary">Host</Typography>
            <Typography>{prompt.host}</Typography>
          </Stack>
          <Stack direction="row" spacing={4}>
            <Typography color="text.secondary">Type</Typography>
            <Typography>{prompt.keyType}</Typography>
          </Stack>
          <Stack direction="row" spacing={4}>
            <Typography color="text.secondary">Fingerprint</Typography>
            <Typography sx={{ fontFamily: "monospace" }}>
              {prompt.fingerprint}
            </Typography>
          </Stack>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={() => resolve(false)}>Reject</Button>
        <Button variant="contained" onClick={() => resolve(true)}>
          Trust and connect
        </Button>
      </DialogActions>
    </Dialog>
  );
}
