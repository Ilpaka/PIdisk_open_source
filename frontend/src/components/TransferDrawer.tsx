import { useMemo } from "react";
import {
  Box,
  Button,
  Chip,
  Drawer,
  IconButton,
  LinearProgress,
  Stack,
  Tooltip,
  Typography,
} from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";
import CancelIcon from "@mui/icons-material/Cancel";
import { useTransferStore } from "@/stores/transferStore";
import { transferApi } from "@/api/transfer";

interface Props {
  open: boolean;
  onClose: () => void;
}

const formatBytes = (n: number): string => {
  if (n === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const idx = Math.min(Math.floor(Math.log(n) / Math.log(1024)), units.length - 1);
  return `${(n / 1024 ** idx).toFixed(1)} ${units[idx]}`;
};

const formatSpeed = (n: number): string => `${formatBytes(n)}/s`;

export default function TransferDrawer({ open, onClose }: Props) {
  const active = useTransferStore((s) => s.active);
  const finished = useTransferStore((s) => s.finished);
  const clearFinished = useTransferStore((s) => s.clearFinished);

  const activeList = useMemo(() => Object.values(active), [active]);

  return (
    <Drawer
      anchor="bottom"
      open={open}
      onClose={onClose}
      PaperProps={{ sx: { maxHeight: "55vh" } }}
    >
      <Stack sx={{ p: 2 }} spacing={2}>
        <Stack direction="row" alignItems="center">
          <Typography variant="subtitle1" fontWeight={600} sx={{ flexGrow: 1 }}>
            Transfers
          </Typography>
          <Button size="small" disabled={finished.length === 0} onClick={clearFinished}>
            Clear finished
          </Button>
          <IconButton onClick={onClose} size="small" sx={{ ml: 1 }}>
            <CloseIcon fontSize="small" />
          </IconButton>
        </Stack>

        {activeList.length === 0 && finished.length === 0 ? (
          <Typography color="text.secondary">No transfers yet.</Typography>
        ) : null}

        {activeList.map((t) => {
          const pct = t.bytesTotal > 0 ? (t.bytesDone / t.bytesTotal) * 100 : 0;
          return (
            <Box key={t.id}>
              <Stack direction="row" spacing={1} alignItems="center">
                <Chip size="small" label={t.kind} />
                <Typography variant="body2" sx={{ flexGrow: 1 }} noWrap>
                  {t.name}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  {formatBytes(t.bytesDone)} / {formatBytes(t.bytesTotal)} ({formatSpeed(t.speed)})
                </Typography>
                <Tooltip title="Cancel">
                  <IconButton size="small" onClick={() => transferApi.cancel(t.id)}>
                    <CancelIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              </Stack>
              <LinearProgress variant="determinate" value={pct} sx={{ mt: 0.5 }} />
            </Box>
          );
        })}

        {finished.length > 0 ? (
          <Stack spacing={1}>
            <Typography variant="overline" color="text.secondary">
              Recent
            </Typography>
            {finished.map((t) => (
              <Stack key={t.id} direction="row" spacing={1} alignItems="center">
                <Chip
                  size="small"
                  color={t.state === "done" ? "success" : t.state === "cancelled" ? "default" : "error"}
                  label={t.state}
                />
                <Typography variant="body2" noWrap sx={{ flexGrow: 1 }}>
                  {t.name}
                </Typography>
                {t.error ? (
                  <Typography variant="caption" color="error">
                    {t.error}
                  </Typography>
                ) : (
                  <Typography variant="caption" color="text.secondary">
                    {formatBytes(t.bytesDone)}
                  </Typography>
                )}
              </Stack>
            ))}
          </Stack>
        ) : null}
      </Stack>
    </Drawer>
  );
}
