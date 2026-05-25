import {
  Box,
  Paper,
  Skeleton,
  Stack,
  Typography,
} from "@mui/material";
import type { FileEntry } from "@/types/domain";
import FileIcon from "@/components/FileIcon";

interface Props {
  entries: FileEntry[];
  selection: Set<string>;
  loading: boolean;
  onActivate: (entry: FileEntry) => void;
  onSelect: (entry: FileEntry, modifier: boolean) => void;
  onContextMenu?: (entry: FileEntry, ev: React.MouseEvent) => void;
}

const cardSx = (selected: boolean) => ({
  width: 130,
  height: 110,
  display: "flex",
  flexDirection: "column" as const,
  alignItems: "center",
  justifyContent: "center",
  cursor: "pointer",
  border: selected ? "2px solid" : undefined,
  borderColor: selected ? "primary.main" : undefined,
  bgcolor: selected ? "action.selected" : undefined,
});

export default function FileGrid({
  entries,
  selection,
  loading,
  onActivate,
  onSelect,
  onContextMenu,
}: Props) {
  if (loading) {
    return (
      <Box sx={{ display: "flex", flexWrap: "wrap", gap: 2, p: 2 }}>
        {Array.from({ length: 8 }).map((_, i) => (
          <Skeleton key={i} variant="rounded" width={130} height={110} />
        ))}
      </Box>
    );
  }

  if (entries.length === 0) {
    return (
      <Stack alignItems="center" justifyContent="center" sx={{ p: 6, color: "text.secondary" }}>
        <Typography>Empty directory.</Typography>
      </Stack>
    );
  }

  return (
    <Box sx={{ display: "flex", flexWrap: "wrap", gap: 2, p: 2 }}>
      {entries.map((e) => (
        <Paper
          key={e.path}
          variant="outlined"
          sx={cardSx(selection.has(e.path))}
          onClick={(ev) => onSelect(e, ev.shiftKey || ev.metaKey || ev.ctrlKey)}
          onDoubleClick={() => onActivate(e)}
          onContextMenu={(ev) => {
            ev.preventDefault();
            onContextMenu?.(e, ev);
          }}
          draggable
          onDragStart={(ev) => {
            ev.dataTransfer.setData("application/x-pidisk-file", e.path);
            ev.dataTransfer.setData("text/plain", e.name);
            ev.dataTransfer.effectAllowed = "move";
          }}
        >
          <FileIcon isDir={e.isDir} name={e.name} sx={{ fontSize: 42 }} />
          <Typography
            variant="caption"
            sx={{
              mt: 0.5,
              maxWidth: 120,
              textAlign: "center",
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {e.name}
          </Typography>
        </Paper>
      ))}
    </Box>
  );
}
