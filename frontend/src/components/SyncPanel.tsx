import { useEffect, useState } from "react";
import {
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  IconButton,
  List,
  ListItem,
  ListItemSecondaryAction,
  ListItemText,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import DeleteIcon from "@mui/icons-material/Delete";
import PlayArrowIcon from "@mui/icons-material/PlayArrow";
import StopIcon from "@mui/icons-material/Stop";

import { useSyncStore } from "@/stores/syncStore";
import { useSnackbarStore } from "@/stores/snackbarStore";
import { confirmDialog } from "@/stores/dialogStore";
import type { SyncFolder } from "@/types/domain";

const blank: SyncFolder = {
  name: "",
  localPath: "",
  remotePath: "",
  enabled: true,
  lastSync: "",
  direction: "both",
};

export default function SyncPanel() {
  const { folders, stats, refreshFolders, refreshStatus, add, remove, toggle, start, stop } =
    useSyncStore();
  const push = useSnackbarStore((s) => s.push);
  const [adding, setAdding] = useState(false);
  const [draft, setDraft] = useState<SyncFolder>(blank);

  useEffect(() => {
    void refreshFolders();
    void refreshStatus();
  }, [refreshFolders, refreshStatus]);

  const onAdd = async () => {
    try {
      await add(draft);
      setDraft(blank);
      setAdding(false);
      push("success", `Folder "${draft.name}" added`);
    } catch (err) {
      push("error", String(err));
    }
  };

  const onToggle = async (folder: SyncFolder) => {
    try {
      await toggle(folder.name, !folder.enabled);
    } catch (err) {
      push("error", String(err));
    }
  };

  const onRemove = async (folder: SyncFolder) => {
    const ok = await confirmDialog({
      title: "Remove sync folder",
      message: `Stop syncing "${folder.name}"? Local and remote files stay where they are.`,
      confirmText: "Remove",
      destructive: true,
    });
    if (!ok) return;
    try {
      await remove(folder.name);
      push("info", `Folder "${folder.name}" removed`);
    } catch (err) {
      push("error", String(err));
    }
  };

  return (
    <Box sx={{ p: 2 }}>
      <Stack direction="row" alignItems="center" spacing={2} sx={{ mb: 2 }}>
        <Typography variant="subtitle1" fontWeight={600} sx={{ flexGrow: 1 }}>
          Synchronization
        </Typography>
        <Chip
          size="small"
          color={stats.isRunning ? "success" : "default"}
          label={stats.isRunning ? "running" : "stopped"}
        />
        {stats.isRunning ? (
          <Button size="small" startIcon={<StopIcon />} onClick={() => stop()}>
            Stop
          </Button>
        ) : (
          <Button size="small" variant="contained" startIcon={<PlayArrowIcon />} onClick={() => start()}>
            Start
          </Button>
        )}
        <Button size="small" startIcon={<AddIcon />} onClick={() => setAdding(true)}>
          Add folder
        </Button>
      </Stack>

      <Divider />

      {folders.length === 0 ? (
        <Typography color="text.secondary" sx={{ mt: 2 }}>
          No sync folders yet.
        </Typography>
      ) : (
        <List>
          {folders.map((f) => (
            <ListItem key={f.name}>
              <Switch checked={f.enabled} onChange={() => onToggle(f)} />
              <ListItemText
                primary={f.name}
                secondary={`${f.localPath}  <->  ${f.remotePath}`}
              />
              <ListItemSecondaryAction>
                <IconButton edge="end" onClick={() => onRemove(f)}>
                  <DeleteIcon />
                </IconButton>
              </ListItemSecondaryAction>
            </ListItem>
          ))}
        </List>
      )}

      {stats.errors.length > 0 ? (
        <Box sx={{ mt: 2 }}>
          <Typography variant="overline" color="error">
            Errors
          </Typography>
          <Stack spacing={0.5}>
            {stats.errors.map((e, i) => (
              <Typography key={i} variant="caption" color="error">
                {e}
              </Typography>
            ))}
          </Stack>
        </Box>
      ) : null}

      <Dialog open={adding} onClose={() => setAdding(false)} fullWidth maxWidth="sm">
        <DialogTitle>New sync folder</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <TextField
              label="Name"
              value={draft.name}
              onChange={(e) => setDraft((d) => ({ ...d, name: e.target.value }))}
            />
            <TextField
              label="Local path"
              value={draft.localPath}
              onChange={(e) => setDraft((d) => ({ ...d, localPath: e.target.value }))}
            />
            <TextField
              label="Remote path"
              value={draft.remotePath}
              onChange={(e) => setDraft((d) => ({ ...d, remotePath: e.target.value }))}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setAdding(false)}>Cancel</Button>
          <Button variant="contained" onClick={onAdd}>
            Add
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
