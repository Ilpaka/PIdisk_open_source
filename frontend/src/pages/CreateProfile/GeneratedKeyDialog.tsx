import {
  Alert,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Stack,
  Tooltip,
  Typography,
} from "@mui/material";
import ContentCopyIcon from "@mui/icons-material/ContentCopy";

interface Props {
  open: boolean;
  publicKey: string;
  keyPath: string;
  username: string;
  host: string;
  onClose: () => void;
}

export default function GeneratedKeyDialog({
  open,
  publicKey,
  keyPath,
  username,
  host,
  onClose,
}: Props) {
  const oneLiner = publicKey
    ? `mkdir -p ~/.ssh && echo '${publicKey.trim()}' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys`
    : "";

  const sshCopy = username && host
    ? `ssh-copy-id -i ${keyPath}.pub ${username}@${host}`
    : "";

  const copy = (text: string) => {
    void navigator.clipboard.writeText(text);
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>Profile created. One thing left.</DialogTitle>
      <DialogContent dividers>
        <Stack spacing={2}>
          <Alert severity="info">
            PIdisk generated a fresh SSH key for this profile and saved its
            passphrase to your macOS Keychain. To let PIdisk connect, install
            the public key below on the server (once).
          </Alert>

          <Box>
            <Typography variant="subtitle2" gutterBottom>
              Public key
            </Typography>
            <Box
              sx={{
                position: "relative",
                p: 2,
                pr: 6,
                bgcolor: "action.hover",
                borderRadius: 1,
                fontFamily: "monospace",
                fontSize: 12,
                whiteSpace: "pre-wrap",
                wordBreak: "break-all",
              }}
            >
              {publicKey}
              <Tooltip title="Copy">
                <IconButton
                  size="small"
                  sx={{ position: "absolute", top: 8, right: 8 }}
                  onClick={() => copy(publicKey)}
                >
                  <ContentCopyIcon fontSize="small" />
                </IconButton>
              </Tooltip>
            </Box>
            <Typography variant="caption" color="text.secondary">
              Saved at {keyPath} (with a matching .pub file).
            </Typography>
          </Box>

          {sshCopy ? (
            <Box>
              <Typography variant="subtitle2" gutterBottom>
                Option 1: install via ssh-copy-id
              </Typography>
              <CodeBlock text={sshCopy} onCopy={copy} />
              <Typography variant="caption" color="text.secondary">
                Requires password access to the server one time. After that
                PIdisk will use the key.
              </Typography>
            </Box>
          ) : null}

          {oneLiner ? (
            <Box>
              <Typography variant="subtitle2" gutterBottom>
                Option 2: paste on the server manually
              </Typography>
              <CodeBlock text={oneLiner} onCopy={copy} />
            </Box>
          ) : null}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button variant="contained" onClick={onClose}>
          Done
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function CodeBlock({ text, onCopy }: { text: string; onCopy: (s: string) => void }) {
  return (
    <Box
      sx={{
        position: "relative",
        p: 2,
        pr: 6,
        bgcolor: "action.hover",
        borderRadius: 1,
        fontFamily: "monospace",
        fontSize: 12,
        whiteSpace: "pre-wrap",
        wordBreak: "break-all",
      }}
    >
      {text}
      <Tooltip title="Copy">
        <IconButton
          size="small"
          sx={{ position: "absolute", top: 8, right: 8 }}
          onClick={() => onCopy(text)}
        >
          <ContentCopyIcon fontSize="small" />
        </IconButton>
      </Tooltip>
    </Box>
  );
}
