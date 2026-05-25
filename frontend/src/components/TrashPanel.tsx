import { useEffect, useState } from "react";
import {
  Box,
  Button,
  IconButton,
  List,
  ListItem,
  ListItemSecondaryAction,
  ListItemText,
  Stack,
  Tooltip,
  Typography,
} from "@mui/material";
import RestoreIcon from "@mui/icons-material/Restore";
import DeleteSweepIcon from "@mui/icons-material/DeleteSweep";

import type { TrashEntry } from "@/types/domain";
import { trashApi } from "@/api/trash";
import { useSnackbarStore } from "@/stores/snackbarStore";
import { confirmDialog } from "@/stores/dialogStore";

export default function TrashPanel({ onChanged }: { onChanged?: () => void }) {
  const [entries, setEntries] = useState<TrashEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const push = useSnackbarStore((s) => s.push);

  const refresh = async () => {
    setLoading(true);
    try {
      setEntries(await trashApi.list());
    } catch (err) {
      push("error", String(err));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void refresh();
  }, []);

  const onRestore = async (id: string) => {
    try {
      await trashApi.restore(id);
      push("success", "Restored");
      await refresh();
      onChanged?.();
    } catch (err) {
      push("error", String(err));
    }
  };

  const onClear = async () => {
    const ok = await confirmDialog({
      title: "Empty trash",
      message: "Permanently delete all items currently in the trash directory? This cannot be undone.",
      confirmText: "Empty",
      destructive: true,
    });
    if (!ok) return;
    try {
      await trashApi.clear();
      push("info", "Trash cleared");
      await refresh();
      onChanged?.();
    } catch (err) {
      push("error", String(err));
    }
  };

  return (
    <Box sx={{ p: 2, minWidth: 360 }}>
      <Stack direction="row" alignItems="center" sx={{ mb: 2 }}>
        <Typography variant="subtitle1" fontWeight={600} sx={{ flexGrow: 1 }}>
          Trash
        </Typography>
        <Tooltip title="Clear all">
          <span>
            <IconButton onClick={onClear} disabled={entries.length === 0}>
              <DeleteSweepIcon />
            </IconButton>
          </span>
        </Tooltip>
      </Stack>
      {loading ? <Typography color="text.secondary">Loading...</Typography> : null}
      {!loading && entries.length === 0 ? (
        <Typography color="text.secondary">No trashed items.</Typography>
      ) : null}
      <List>
        {entries.map((e) => (
          <ListItem key={e.id}>
            <ListItemText
              primary={e.originalPath.split("/").pop()}
              secondary={e.originalPath}
            />
            <ListItemSecondaryAction>
              <Tooltip title="Restore to original location">
                <IconButton edge="end" onClick={() => onRestore(e.id)}>
                  <RestoreIcon />
                </IconButton>
              </Tooltip>
            </ListItemSecondaryAction>
          </ListItem>
        ))}
      </List>
      <Button size="small" onClick={refresh}>
        Refresh
      </Button>
    </Box>
  );
}
