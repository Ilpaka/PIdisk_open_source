import { useEffect, useState } from "react";

import {
  AppBar,
  Box,
  Button,
  Divider,
  IconButton,
  Stack,
  Toolbar,
  Tooltip,
  Typography,
} from "@mui/material";
import CreateNewFolderIcon from "@mui/icons-material/CreateNewFolder";
import LogoutIcon from "@mui/icons-material/Logout";
import DeleteIcon from "@mui/icons-material/Delete";
import CloudUploadIcon from "@mui/icons-material/CloudUpload";
import CloudDownloadIcon from "@mui/icons-material/CloudDownload";
import SyncIcon from "@mui/icons-material/Sync";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";
import SettingsIcon from "@mui/icons-material/Settings";
import { Drawer } from "@mui/material";
import { useNavigate } from "react-router-dom";

import { useHotkeys } from "@/hooks/useHotkeys";

import { useProfileStore } from "@/stores/profileStore";
import { useFilesStore } from "@/stores/filesStore";
import { useSnackbarStore } from "@/stores/snackbarStore";
import { useConnectionStore } from "@/stores/connectionStore";
import { useTransferStore } from "@/stores/transferStore";
import { connectionApi } from "@/api/connection";
import { filesApi } from "@/api/files";
import { transferApi } from "@/api/transfer";
import { dialogsApi } from "@/api/dialogs";
import { confirmDialog, promptDialog } from "@/stores/dialogStore";
import BreadcrumbsBar from "@/components/BreadcrumbsBar";
import FileGrid from "@/components/FileGrid";
import FolderTree from "@/components/FolderTree";
import TransferDrawer from "@/components/TransferDrawer";
import SyncPanel from "@/components/SyncPanel";
import TrashPanel from "@/components/TrashPanel";
import ResizableSidebar from "@/components/ResizableSidebar";
import FileContextMenu, {
  type FileContextMenuAnchor,
} from "@/components/FileContextMenu";
import type { FileEntry } from "@/types/domain";

export default function FilesPage() {
  const navigate = useNavigate();
  const active = useProfileStore((s) => s.active);
  const clearActive = useProfileStore((s) => s.clearActive);
  const connected = useConnectionStore((s) => s.connected);
  const resetConn = useConnectionStore((s) => s.reset);
  const pushSnack = useSnackbarStore((s) => s.push);
  const transferCount = useTransferStore((s) => Object.keys(s.active).length);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [syncOpen, setSyncOpen] = useState(false);
  const [trashOpen, setTrashOpen] = useState(false);
  const [ctxAnchor, setCtxAnchor] = useState<FileContextMenuAnchor | null>(null);

  const {
    cwd,
    entries,
    selection,
    loading,
    error,
    setCwd,
    refresh,
    toggleSelect,
    selectOnly,
    selectAll,
    clearSelection,
  } = useFilesStore();

  useEffect(() => {
    if (!active) {
      navigate("/login");
      return;
    }
    if (!cwd) {
      void setCwd(active.rootDir || "/");
    }
  }, [active, cwd, setCwd, navigate]);

  useEffect(() => {
    if (error) pushSnack("error", error);
  }, [error, pushSnack]);

  const onActivate = (p: string, isDir: boolean) => {
    if (isDir) void setCwd(p);
  };

  const onMkdir = async () => {
    const name = await promptDialog({
      title: "New folder",
      label: "Folder name",
      placeholder: "my-folder",
    });
    if (!name) return;
    try {
      await filesApi.mkdir(cwd, name);
      pushSnack("success", `Folder "${name}" created`);
      await refresh();
    } catch (err) {
      pushSnack("error", String(err));
    }
  };

  // Upload uses the native file picker so the user never types a path.
  const onUpload = async () => {
    let localPath: string;
    try {
      localPath = await dialogsApi.selectFile("Choose a file to upload");
    } catch (err) {
      pushSnack("error", String(err));
      return;
    }
    if (!localPath) return;
    const remoteName = localPath.split(/[/\\]/).pop() ?? "uploaded";
    try {
      await transferApi.upload(localPath, `${cwd}/${remoteName}`);
      setDrawerOpen(true);
    } catch (err) {
      pushSnack("error", String(err));
    }
  };

  // Download is "smart": a single file uses save-as, a single folder offers
  // a zip archive, multi-select falls back to a save-folder dialog.
  const startFileDownload = async (entry: FileEntry) => {
    let dest: string;
    try {
      dest = await dialogsApi.saveFile(entry.name, "Save file as");
    } catch (err) {
      pushSnack("error", String(err));
      return;
    }
    if (!dest) return;
    try {
      await transferApi.download(entry.path, dest);
      setDrawerOpen(true);
    } catch (err) {
      pushSnack("error", String(err));
    }
  };

  const startFolderZipDownload = async (entry: FileEntry) => {
    let dest: string;
    try {
      dest = await dialogsApi.saveArchive(entry.name);
    } catch (err) {
      pushSnack("error", String(err));
      return;
    }
    if (!dest) return;
    try {
      await transferApi.downloadFolderAsZip(entry.path, dest);
      setDrawerOpen(true);
    } catch (err) {
      pushSnack("error", String(err));
    }
  };

  const onDownloadFromToolbar = async () => {
    if (selection.size === 0) {
      pushSnack("warning", "Select a file or folder to download");
      return;
    }
    if (selection.size > 1) {
      pushSnack("warning", "Multi-download is not implemented yet, select one item");
      return;
    }
    const targetPath = Array.from(selection)[0];
    const entry = entries.find((e) => e.path === targetPath);
    if (!entry) return;
    if (entry.isDir) {
      await startFolderZipDownload(entry);
    } else {
      await startFileDownload(entry);
    }
  };

  const onRemoveSelected = async () => {
    if (selection.size === 0) return;
    const ok = await confirmDialog({
      title: "Remove items",
      message: `Remove ${selection.size} item(s)? They will be moved to the trash directory of this profile.`,
      confirmText: "Remove",
      destructive: true,
    });
    if (!ok) return;
    try {
      for (const target of selection) {
        const res = await filesApi.remove(target);
        pushSnack(
          "info",
          res.trashed ? `Moved to trash: ${target}` : `Removed: ${target}`,
        );
      }
      await refresh();
    } catch (err) {
      pushSnack("error", String(err));
    }
  };

  const removeEntry = async (entry: FileEntry) => {
    const ok = await confirmDialog({
      title: "Move to trash",
      message: `Move "${entry.name}" to the trash directory?`,
      confirmText: "Move",
      destructive: true,
    });
    if (!ok) return;
    try {
      await filesApi.remove(entry.path);
      pushSnack("info", `Moved to trash: ${entry.path}`);
      await refresh();
    } catch (err) {
      pushSnack("error", String(err));
    }
  };

  const renameEntry = async (entry: FileEntry) => {
    const next = await promptDialog({
      title: "Rename",
      label: "New name",
      defaultValue: entry.name,
    });
    if (!next || next === entry.name) return;
    try {
      await filesApi.rename(cwd, entry.name, next);
      pushSnack("success", `Renamed to "${next}"`);
      await refresh();
    } catch (err) {
      pushSnack("error", String(err));
    }
  };

  const onRenameSelected = async () => {
    if (selection.size !== 1) {
      pushSnack("warning", "Select a single entry to rename");
      return;
    }
    const target = Array.from(selection)[0];
    const entry = entries.find((e) => e.path === target);
    if (entry) await renameEntry(entry);
  };

  const onBack = () => {
    if (!cwd || cwd === "/" || cwd === active?.rootDir) return;
    const parent = cwd.replace(/\/[^/]+$/, "") || "/";
    void setCwd(parent);
  };

  useHotkeys({
    onRename: onRenameSelected,
    onDelete: onRemoveSelected,
    onSelectAll: selectAll,
    onEscape: clearSelection,
    onBack,
    onRefresh: refresh,
  });

  const onLogout = async () => {
    try {
      await connectionApi.lock();
    } catch {
      // ignore
    }
    resetConn();
    await clearActive();
    navigate("/login");
  };

  if (!active) return null;

  return (
    <Stack sx={{ height: "100vh" }}>
      <AppBar position="static" color="default" elevation={0}>
        <Toolbar variant="dense" sx={{ gap: 2 }}>
          <Typography variant="subtitle2" sx={{ flexGrow: 1 }}>
            {active.name} ({active.username}@{active.host}:{active.port})
            {connected ? " - connected" : " - reconnecting..."}
          </Typography>
          <Tooltip title="New folder">
            <IconButton size="small" onClick={onMkdir}>
              <CreateNewFolderIcon fontSize="small" />
            </IconButton>
          </Tooltip>
          <Tooltip title="Upload from disk">
            <IconButton size="small" onClick={onUpload}>
              <CloudUploadIcon fontSize="small" />
            </IconButton>
          </Tooltip>
          <Tooltip title="Download selected (folder downloads as ZIP)">
            <span>
              <IconButton
                size="small"
                onClick={onDownloadFromToolbar}
                disabled={selection.size === 0}
              >
                <CloudDownloadIcon fontSize="small" />
              </IconButton>
            </span>
          </Tooltip>
          <Tooltip title="Remove selected">
            <span>
              <IconButton size="small" onClick={onRemoveSelected} disabled={selection.size === 0}>
                <DeleteIcon fontSize="small" />
              </IconButton>
            </span>
          </Tooltip>
          <Button size="small" onClick={() => setDrawerOpen(true)}>
            Transfers ({transferCount})
          </Button>
          <Tooltip title="Sync folders">
            <IconButton size="small" onClick={() => setSyncOpen(true)}>
              <SyncIcon fontSize="small" />
            </IconButton>
          </Tooltip>
          <Tooltip title="Trash">
            <IconButton size="small" onClick={() => setTrashOpen(true)}>
              <DeleteOutlineIcon fontSize="small" />
            </IconButton>
          </Tooltip>
          <Tooltip title="Settings">
            <IconButton size="small" onClick={() => navigate("/settings")}>
              <SettingsIcon fontSize="small" />
            </IconButton>
          </Tooltip>
          <Divider orientation="vertical" flexItem />
          <Button startIcon={<LogoutIcon />} onClick={onLogout} size="small">
            Disconnect
          </Button>
        </Toolbar>
      </AppBar>
      <BreadcrumbsBar
        cwd={cwd}
        rootDir={active.rootDir}
        onNavigate={(p) => setCwd(p)}
        onRefresh={() => refresh()}
      />
      <Stack direction="row" sx={{ flexGrow: 1, overflow: "hidden" }}>
        <ResizableSidebar>
          <FolderTree
            rootPath={active.rootDir}
            currentPath={cwd}
            onNavigate={(p) => setCwd(p)}
            onMove={async (src, dst) => {
              try {
                await filesApi.move(src, dst);
                pushSnack("success", `Moved to ${dst}`);
                await refresh();
              } catch (err) {
                pushSnack("error", String(err));
              }
            }}
          />
        </ResizableSidebar>
        <Box sx={{ flexGrow: 1, overflow: "auto" }}>
          <FileGrid
            entries={entries}
            selection={selection}
            loading={loading}
            onActivate={(e) => onActivate(e.path, e.isDir)}
            onSelect={(e, modifier) => (modifier ? toggleSelect(e.path) : selectOnly(e.path))}
            onContextMenu={(entry, ev) => {
              selectOnly(entry.path);
              setCtxAnchor({ entry, mouseX: ev.clientX, mouseY: ev.clientY });
            }}
          />
        </Box>
      </Stack>
      <TransferDrawer open={drawerOpen} onClose={() => setDrawerOpen(false)} />
      <Drawer anchor="right" open={syncOpen} onClose={() => setSyncOpen(false)}>
        <Box sx={{ width: 420 }}>
          <SyncPanel />
        </Box>
      </Drawer>
      <Drawer anchor="right" open={trashOpen} onClose={() => setTrashOpen(false)}>
        <TrashPanel onChanged={() => refresh()} />
      </Drawer>
      <FileContextMenu
        anchor={ctxAnchor}
        onClose={() => setCtxAnchor(null)}
        onDownload={startFileDownload}
        onDownloadZip={startFolderZipDownload}
        onRename={renameEntry}
        onDelete={removeEntry}
      />
    </Stack>
  );
}
