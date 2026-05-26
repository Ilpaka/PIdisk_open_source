import { Menu, MenuItem, ListItemIcon, ListItemText, Divider } from "@mui/material";
import DownloadIcon from "@mui/icons-material/Download";
import FolderZipIcon from "@mui/icons-material/FolderZip";
import DriveFileRenameOutlineIcon from "@mui/icons-material/DriveFileRenameOutline";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";
import type { FileEntry } from "@/types/domain";

export interface FileContextMenuAnchor {
  entry: FileEntry;
  mouseX: number;
  mouseY: number;
}

interface Props {
  anchor: FileContextMenuAnchor | null;
  onClose: () => void;
  onDownload: (entry: FileEntry) => void;
  onDownloadZip: (entry: FileEntry) => void;
  onRename: (entry: FileEntry) => void;
  onDelete: (entry: FileEntry) => void;
}

export default function FileContextMenu({
  anchor,
  onClose,
  onDownload,
  onDownloadZip,
  onRename,
  onDelete,
}: Props) {
  const entry = anchor?.entry;
  const open = Boolean(anchor);

  const close = (action?: () => void) => {
    onClose();
    if (action) {
      // run the action on the next tick so MUI can close the menu first
      setTimeout(action, 0);
    }
  };

  return (
    <Menu
      open={open}
      onClose={() => close()}
      anchorReference="anchorPosition"
      anchorPosition={
        anchor ? { left: anchor.mouseX, top: anchor.mouseY } : undefined
      }
    >
      {entry?.isDir ? (
        <MenuItem onClick={() => entry && close(() => onDownloadZip(entry))}>
          <ListItemIcon>
            <FolderZipIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText>Download as ZIP</ListItemText>
        </MenuItem>
      ) : (
        <MenuItem onClick={() => entry && close(() => onDownload(entry))}>
          <ListItemIcon>
            <DownloadIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText>Download</ListItemText>
        </MenuItem>
      )}
      <Divider />
      <MenuItem onClick={() => entry && close(() => onRename(entry))}>
        <ListItemIcon>
          <DriveFileRenameOutlineIcon fontSize="small" />
        </ListItemIcon>
        <ListItemText>Rename</ListItemText>
      </MenuItem>
      <MenuItem
        onClick={() => entry && close(() => onDelete(entry))}
        sx={{ color: "error.main" }}
      >
        <ListItemIcon>
          <DeleteOutlineIcon fontSize="small" color="error" />
        </ListItemIcon>
        <ListItemText>Move to trash</ListItemText>
      </MenuItem>
    </Menu>
  );
}
