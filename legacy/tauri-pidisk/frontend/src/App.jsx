// src/App.jsx
import React, { useState, useEffect, useRef } from "react";
import { invoke } from "@tauri-apps/api/core";
import {
  Box,
  Menu,
  MenuItem,
  AppBar,
  Toolbar,
  IconButton,
  Typography,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Button,
  Switch,
  FormControlLabel,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  Chip,
  Card,
  CardContent,
  LinearProgress,
  Grid,
} from "@mui/material";
import SettingsIcon from "@mui/icons-material/Settings";
import AddIcon from "@mui/icons-material/Add";
import ViewListIcon from '@mui/icons-material/ViewList';
import ViewModuleIcon from '@mui/icons-material/ViewModule';
import DeleteSweepIcon from "@mui/icons-material/DeleteSweep";
import SyncIcon from "@mui/icons-material/Sync";
import StopIcon from "@mui/icons-material/Stop";
import FolderIcon from "@mui/icons-material/Folder";
import DeleteIcon from "@mui/icons-material/Delete";
import CloudSyncIcon from "@mui/icons-material/CloudSync";
import Snackbar from '@mui/material/Snackbar';
import MuiAlert from '@mui/material/Alert';
import Breadcrumbs from '@mui/material/Breadcrumbs';
import Link from '@mui/material/Link';
import NavigateNextIcon from '@mui/icons-material/NavigateNext';

const { save } = window.__TAURI__.dialog;
const { open } = window.__TAURI__.dialog;

import FolderTree from "./components/FolderTree";
import FileGrid from "./components/FileGrid";
import MemoryBar from "./components/MemoryBar";

function formatBytes(bytes) {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function renderBreadcrumbs(currentPath, onNavigate) {
  const pathSegments = currentPath.replace(/^\/+/, '').split('/');
  let crumbs = [];
  let accPath = '';
  pathSegments.forEach((seg, idx) => {
    accPath += '/' + seg;
    crumbs.push({
      name: seg,
      path: accPath
    });
  });

  return (
    <Breadcrumbs aria-label="breadcrumb" separator={<NavigateNextIcon fontSize="small" />} sx={{ mb: 1 }}>
      {crumbs.map((crumb, idx) => {
        const isLast = idx === crumbs.length - 1;
        return isLast ? (
          <Typography color="text.primary" key={crumb.path} noWrap>
            {crumb.name}
          </Typography>
        ) : (
          <Link
            key={crumb.path}
            color="inherit"
            underline="hover"
            sx={{ cursor: 'pointer' }}
            onClick={e => {
              e.preventDefault();
              onNavigate(crumb.path);
            }}
            noWrap
            href="#"
          >
            {crumb.name}
          </Link>
        );
      })}
    </Breadcrumbs>
  );
}

export default function App() {
  // ===== states =====
  const [files, setFiles] = useState([]);
  const [currentPath, setCurrentPath] = useState(".");
  const [error, setError] = useState(null);
  const [menu, setMenu] = useState(null);
  const [createdFolder, setCreatedFolder] = useState(null);
  const [renameTarget, setRenameTarget] = useState(null);
  const [renameValue, setRenameValue] = useState("");
  const [deletedItem, setDeletedItem] = useState(null);
  const fileInputRef = useRef(null);
  
  // ===== settings dialog state =====
  const [openSettings, setOpenSettings] = useState(false);
  const [host, setHost] = useState("");
  const [port, setPort] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [rootDir, setRootDir] = useState("");
  const [trashDir, setTrashDir] = useState("");
  const [trashCleared, setTrashCleared] = useState(false);
  const [viewMode, setViewMode] = useState("grid");
  const [snackbar, setSnackbar] = useState({open: false, message: '', severity: 'success'});
  const [selectedFiles, setSelectedFiles] = useState([]);

  // ===== sync states =====
  const [syncFolders, setSyncFolders] = useState([]);
  const [syncStatus, setSyncStatus] = useState({
    is_running: false,
    last_sync_time: 0,
    synced_folders: [],
    errors: [],
    files_synced: 0,
    bytes_synced: 0
  });
  const [openSyncDialog, setOpenSyncDialog] = useState(false);
  const [newSyncFolder, setNewSyncFolder] = useState({
    name: '',
    local_path: '',
    remote_path: ''
  });
  
  // ИСПРАВЛЕНО: Правильная инициализация состояния
  const [showSyncPanel, setShowSyncPanel] = useState(false);

  const Alert = React.forwardRef(function Alert(props, ref) {
    return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
  });

  // ===== initial load =====
  useEffect(() => {
    loadDirectory("/root/PIdisk");
    loadSettings();
    loadSyncFolders();
    
    const interval = setInterval(loadSyncStatus, 5000);
    return () => clearInterval(interval);
  }, []);

  // ===== load functions =====
  async function loadSettings() {
    try {
      const cfg = await invoke("get_settings");
      setHost(cfg.host);
      setPort(cfg.port);
      setUsername(cfg.username);
      setPassword(cfg.password);
      setTrashDir(cfg.trash_dir);
    } catch (e) {
      setError(String(e));
    }
  }

  async function loadSyncFolders() {
    try {
      const folders = await invoke("get_sync_folders");
      setSyncFolders(folders);
    } catch (e) {
      setError(String(e));
    }
  }

  async function loadSyncStatus() {
    try {
      const status = await invoke("get_sync_status");
      setSyncStatus(status);
    } catch (e) {
      console.error("Ошибка загрузки статуса синхронизации:", e);
    }
  }

  async function loadDirectory(dir) {
    try {
      const [newPath, list] = await invoke("read_dir", { dir });
      setCurrentPath(newPath);
      setFiles(list || []);
      setSelectedFiles([]);
      setError(null);
    } catch (e) {
      setError(String(e));
      setFiles([]);
    }
  }

  // ===== sync functions =====
  async function handleAddSyncFolder() {
    if (!newSyncFolder.name || !newSyncFolder.local_path || !newSyncFolder.remote_path) {
      showSnackbar('Заполните все поля!', 'error');
      return;
    }

    try {
      await invoke("add_sync_folder", {
        name: newSyncFolder.name,
        localPath: newSyncFolder.local_path,
        remotePath: newSyncFolder.remote_path
      });
      
      setNewSyncFolder({ name: '', local_path: '', remote_path: '' });
      setOpenSyncDialog(false);
      loadSyncFolders();
      showSnackbar('Папка добавлена для синхронизации!', 'success');
    } catch (e) {
      showSnackbar('Ошибка добавления папки!', 'error');
    }
  }

  async function handleToggleSyncFolder(name, enabled) {
    try {
      await invoke("toggle_sync_folder", { name, enabled });
      loadSyncFolders();
      showSnackbar(`Синхронизация ${enabled ? 'включена' : 'выключена'}!`, 'success');
    } catch (e) {
      showSnackbar('Ошибка изменения настроек!', 'error');
    }
  }

  async function handleRemoveSyncFolder(name) {
    try {
      await invoke("remove_sync_folder", { name });
      loadSyncFolders();
      showSnackbar('Папка удалена из синхронизации!', 'success');
    } catch (e) {
      showSnackbar('Ошибка удаления папки!', 'error');
    }
  }

  async function handleStartSync() {
    try {
      await invoke("start_sync");
      loadSyncStatus();
      showSnackbar('Синхронизация запущена!', 'success');
    } catch (e) {
      showSnackbar(`Ошибка запуска синхронизации: ${e}`, 'error');
    }
  }

  async function handleStopSync() {
    try {
      await invoke("stop_sync");
      loadSyncStatus();
      showSnackbar('Синхронизация остановлена!', 'success');
    } catch (e) {
      showSnackbar('Ошибка остановки синхронизации!', 'error');
    }
  }

  async function selectLocalFolder() {
    try {
      const folderPath = await open({
        directory: true,
        multiple: false,
        title: "Выберите локальную папку для синхронизации"
      });
      if (folderPath) {
        setNewSyncFolder(prev => ({ ...prev, local_path: folderPath }));
      }
    } catch (e) {
      showSnackbar('Ошибка выбора папки!', 'error');
    }
  }

  // ===== context menus =====
  function onFileContext(e, name) {
    e.preventDefault();
    e.stopPropagation();
    const isMultiSelected = selectedFiles.length > 1 && selectedFiles.includes(name);
    setMenu({
      mouseX: e.clientX + 2,
      mouseY: e.clientY + 4,
      name,
      isFolder: !/\.[^/.]+$/.test(name),
      multi: isMultiSelected,
    });
  }

  function onMainContext(e) {
    e.preventDefault();
    setMenu({ mouseX: e.clientX + 2, mouseY: e.clientY + 4, name: null, isFolder: false });
  }

  function showSnackbar(message, severity = 'success') {
    setSnackbar({ open: true, message, severity });
  }

  const handleClose = () => setMenu(null);

  // ===== menu actions =====
  const handleOpen = (name) => {
    handleClose();
    const next = currentPath === "." ? name : `${currentPath}/${name}`;
    loadDirectory(next);
  };

  const handleDownloadFolder = async (name) => {
    handleClose();
    try {
      const folderPath = await open({
        directory: true,
        multiple: false,
        title: "Выберите папку для сохранения"
      });
      if (!folderPath) return;

      const savePath = `${folderPath}/${name}`;
      await invoke('download_folder', {
        serverFolderName: name,
        savePath: savePath,
      });

      showSnackbar('Папка скачана!', 'success');
    } catch (err) {
      showSnackbar('Ошибка при скачивании папки!', 'error');
    }
  };

  const handleDelete = async (name) => {
    handleClose();
    try {
      await invoke("rm", { target: name });
      loadDirectory(currentPath);
      setDeletedItem({ parentPath: currentPath, name });
      showSnackbar('Файл(-ы) удалён(-ы)!', 'success');
    } catch (e) {
      showSnackbar('Ошибка при удаление!', 'error');
    }
  };

  const handleDeleteSelected = async () => {
    handleClose();
    if (!selectedFiles.length) return;
    try {
      for (const name of selectedFiles) {
        await invoke("rm", { target: name });
      }
      await loadDirectory(currentPath);
      setDeletedItem({ parentPath: currentPath, names: selectedFiles });
      showSnackbar('Файлы удалены!', 'success');
    } catch (e) {
      showSnackbar('Ошибка при удалении!', 'error');
    }
  };

  const handleRenameMenu = (name) => {
    handleClose();
    setTimeout(() => {
      setRenameTarget(name);
      setRenameValue(name);
    }, 0);
  };

  const handleRenameConfirm = async () => {
    if (!renameTarget) return;
    const oldName = renameTarget;
    const newName = renameValue.trim();

    setRenameTarget(null);
    setRenameValue("");

    if (newName && newName !== oldName) {
      if (files.includes(newName)) {
        showSnackbar("Файл или папка с таким именем уже есть!", "error");
        return;
      }
      try {
        await invoke("rename", { old: oldName, new: newName });
        showSnackbar("Файл переименован!", "success");
        setCreatedFolder({ parentPath: currentPath, oldName, newName });
      } catch (e) {
        showSnackbar("Ошибка при переименовании!", "error");
        setError(String(e));
      }
    }
    await loadDirectory(currentPath);
  };

  const handleNewFolder = async () => {
    handleClose();
    const base = "новая папка";
    let idx = 1, def = base;
    const exists = new Set(files);
    while (exists.has(def)) {
      idx++;
      def = `${base} ${idx}`;
    }
    await invoke("mkdir", { name: def });
    await loadDirectory(currentPath);
    setCreatedFolder({ parentPath: currentPath, newName: def });
    setTimeout(() => {
      setRenameTarget(def);
      setRenameValue(def);
    }, 0);
  };

  // ===== settings dialog handlers =====
  const handleOpenSettings = () => setOpenSettings(true);
  const handleCloseSettings = () => setOpenSettings(false);

  const handleSettingsSave = async () => {
    try {
      await invoke("update_settings", {
        host,
        port,
        username,
        password,
      });
      setOpenSettings(false);
      showSnackbar('Успешно!', 'success');
      await loadDirectory(currentPath);
    } catch (e) {
      setError(
        String(e).includes("Неверные данные")
          ? "Неверные данные для подключения"
          : String(e)
      );
    }
  };

  const handleUploadClick = async () => {
    const inp = document.createElement("input");
    inp.type = "file";
    inp.multiple = true;
    inp.onchange = async (e) => {
      const files = e.target.files;
      if (!files || files.length === 0) return;
      try {
        for (let i = 0; i < files.length; i++) {
          const file = files[i];
          const buf = await file.arrayBuffer();
          const bytes = Array.from(new Uint8Array(buf));
          await invoke("upload_file", {
            filename: file.name,
            data: bytes,
          });
        }
        showSnackbar('Загрузка завершена!', 'success');
        await loadDirectory(currentPath);
      } catch (err) {
        showSnackbar('Ошибка при загрузке!', 'error');
      }
    };
    inp.click();
  };

  const handleDownload = async (fileName) => {
    try {
      const savePath = await save({ defaultPath: fileName });
      if (!savePath) return;
      handleClose();
      await invoke('download_and_save', {
        serverFileName: fileName,
        savePath,
      });
      showSnackbar('Файл(-ы) скачен(-ы)!', 'success');
    } catch (err) {
      showSnackbar('Ошибка при скачивание!', 'error');
    }
  };

  const handleDownloadSelected = async () => {
    if (!selectedFiles.length) return;
    if (selectedFiles.length === 1) {
      await handleDownload(selectedFiles[0]);
      return;
    }
    try {
      const folderPath = await open({
        directory: true,
        multiple: false,
        title: "Выберите папку для сохранения файлов"
      });
      if (!folderPath) return;
      handleClose();
      for (const fileName of selectedFiles) {
        const savePath = `${folderPath}/${fileName}`;
        await invoke('download_and_save', {
          serverFileName: fileName,
          savePath,
        });
      }
      showSnackbar('Файлы скачаны!', 'success');
    } catch (err) {
      showSnackbar('Ошибка при скачивании!', 'error');
    }
  };

  async function handleClearAll() {
    try {
      await invoke("clear_all");
      showSnackbar('Корзина очищена!', 'success');
      await loadDirectory(currentPath);
      setTrashCleared(true);
    } catch (e) {
      showSnackbar('Ошибка при очистке корзины!', 'error');
    }
  }

  // ИСПРАВЛЕНО: Функция переключения панели синхронизации
  const toggleSyncPanel = () => {
    console.log("Переключение панели синхронизации. Текущее состояние:", showSyncPanel);
    setShowSyncPanel(prev => {
      const newState = !prev;
      console.log("Новое состояние:", newState);
      return newState;
    });
  };

  return (
    <Box display="flex" flexDirection="column" height="100vh" width="100vw">
      {/* TOP BAR */}
      <AppBar position="static" color="default" elevation={1}>
        <Toolbar variant="dense" sx={{ justifyContent: "space-between" }}>
          <Typography variant="h6">PIdisk</Typography>
          <Box>
            {/* ИСПРАВЛЕНО: Кнопка синхронизации */}
            <IconButton
              size="small"
              color={showSyncPanel ? "primary" : "inherit"}
              onClick={toggleSyncPanel}
              title="Синхронизация"
            >
              <CloudSyncIcon />
            </IconButton>
            
            <IconButton
              size="small"
              color={viewMode === "grid" ? "primary" : "inherit"}
              onClick={() => setViewMode("grid")}
              title="Сетка"
            >
              <ViewModuleIcon />
            </IconButton>
            <IconButton
              color={viewMode === "list" ? "primary" : "default"}
              onClick={() => setViewMode("list")}
              title="Показать списком"
            >
              <ViewListIcon />
            </IconButton>
            <IconButton size="small" color="inherit" onClick={handleOpenSettings}>
              <SettingsIcon />
            </IconButton>

            {currentPath === trashDir && (
              <IconButton size="small" color="inherit" onClick={handleClearAll}>
                <DeleteSweepIcon />
              </IconButton>
            )}

            <IconButton size="small" onClick={handleUploadClick}>
              <AddIcon />
            </IconButton>
          </Box>
        </Toolbar>
      </AppBar>

      {/* ИСПРАВЛЕНО: SYNC PANEL */}
      {showSyncPanel && (
        <Box sx={{ width: '100%' }}>
          <Card sx={{ m: 1 }}>
            <CardContent>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                <Typography variant="h6">Синхронизация файлов</Typography>
                <Box>
                  <Button
                    variant="contained"
                    startIcon={<AddIcon />}
                    onClick={() => setOpenSyncDialog(true)}
                    sx={{ mr: 1 }}
                  >
                    Добавить папку
                  </Button>
                  {syncStatus.is_running ? (
                    <Button
                      variant="contained"
                      color="error"
                      startIcon={<StopIcon />}
                      onClick={handleStopSync}
                    >
                      Остановить
                    </Button>
                  ) : (
                    <Button
                      variant="contained"
                      color="success"
                      startIcon={<SyncIcon />}
                      onClick={handleStartSync}
                    >
                      Запустить
                    </Button>
                  )}
                </Box>
              </Box>

              {/* Статус синхронизации */}
              <Grid container spacing={2} sx={{ mb: 2 }}>
                <Grid item xs={12} md={6}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Chip 
                      label={syncStatus.is_running ? "Активна" : "Остановлена"} 
                      color={syncStatus.is_running ? "success" : "default"}
                      icon={syncStatus.is_running ? <SyncIcon /> : <StopIcon />}
                    />
                    {syncStatus.last_sync_time > 0 && (
                      <Chip 
                        label={`Последняя: ${new Date(syncStatus.last_sync_time * 1000).toLocaleTimeString()}`}
                        variant="outlined"
                        size="small"
                      />
                    )}
                  </Box>
                </Grid>
                <Grid item xs={12} md={6}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Chip 
                      label={`Файлов: ${syncStatus.files_synced}`}
                      variant="outlined"
                      size="small"
                    />
                    <Chip 
                      label={`Размер: ${formatBytes(syncStatus.bytes_synced)}`}
                      variant="outlined"
                      size="small"
                    />
                  </Box>
                </Grid>
              </Grid>

              {/* Прогресс-бар при активной синхронизации */}
              {syncStatus.is_running && (
                <LinearProgress sx={{ mb: 2 }} />
              )}

              {/* Ошибки синхронизации */}
              {syncStatus.errors.length > 0 && (
                <Box sx={{ mb: 2 }}>
                  <Typography variant="subtitle2" color="error" gutterBottom>
                    Ошибки синхронизации:
                  </Typography>
                  {syncStatus.errors.slice(-3).map((error, idx) => (
                    <Chip key={idx} label={error} color="error" size="small" sx={{ mr: 1, mb: 1 }} />
                  ))}
                </Box>
              )}

              {/* Список папок для синхронизации */}
              <Typography variant="subtitle1" gutterBottom>
                Папки для синхронизации ({syncFolders.length})
              </Typography>
              <List>
                {syncFolders.map((folder) => (
                  <ListItem key={folder.name} divider>
                    <FolderIcon sx={{ mr: 2, color: folder.enabled ? 'primary.main' : 'grey.500' }} />
                    <ListItemText
                      primary={folder.name}
                      secondary={
                        <Box>
                          <Typography variant="body2" color="text.secondary">
                            Локально: {folder.local_path}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            Сервер: {folder.remote_path}
                          </Typography>
                          {folder.last_sync > 0 && (
                            <Typography variant="caption" color="text.secondary">
                              Последняя синхронизация: {new Date(folder.last_sync * 1000).toLocaleString()}
                            </Typography>
                          )}
                        </Box>
                      }
                    />
                    <ListItemSecondaryAction>
                      <FormControlLabel
                        control={
                          <Switch
                            checked={folder.enabled}
                            onChange={(e) => handleToggleSyncFolder(folder.name, e.target.checked)}
                          />
                        }
                        label="Активна"
                      />
                      <IconButton
                        edge="end"
                        onClick={() => handleRemoveSyncFolder(folder.name)}
                      >
                        <DeleteIcon />
                      </IconButton>
                    </ListItemSecondaryAction>
                  </ListItem>
                ))}
                {syncFolders.length === 0 && (
                  <ListItem>
                    <ListItemText 
                      primary="Нет папок для синхронизации"
                      secondary="Добавьте папки для начала синхронизации"
                    />
                  </ListItem>
                )}
              </List>
            </CardContent>
          </Card>
        </Box>
      )}

      {/* SYNC FOLDER DIALOG */}
      <Dialog open={openSyncDialog} onClose={() => setOpenSyncDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Добавить папку для синхронизации</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Название папки"
            fullWidth
            variant="outlined"
            value={newSyncFolder.name}
            onChange={(e) => setNewSyncFolder(prev => ({ ...prev, name: e.target.value }))}
            sx={{ mb: 2 }}
            helperText="Уникальное название для идентификации папки"
          />
          
          <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
            <TextField
              margin="dense"
              label="Локальная папка"
              fullWidth
              variant="outlined"
              value={newSyncFolder.local_path}
              onChange={(e) => setNewSyncFolder(prev => ({ ...prev, local_path: e.target.value }))}
              sx={{ mr: 1 }}
              helperText="Папка на вашем компьютере"
            />
            <Button variant="outlined" onClick={selectLocalFolder}>
              Выбрать
            </Button>
          </Box>

          <TextField
            margin="dense"
            label="Удаленная папка (на сервере)"
            fullWidth
            variant="outlined"
            value={newSyncFolder.remote_path}
            onChange={(e) => setNewSyncFolder(prev => ({ ...prev, remote_path: e.target.value }))}
            placeholder="/root/PIdisk/sync_folder"
            helperText="Путь к папке на SSH сервере"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenSyncDialog(false)}>Отмена</Button>
          <Button onClick={handleAddSyncFolder} variant="contained">Добавить</Button>
        </DialogActions>
      </Dialog>

      {/* SETTINGS DIALOG */}
      <Dialog open={openSettings} onClose={handleCloseSettings} maxWidth="xs" fullWidth>
        <DialogTitle>Настройки</DialogTitle>
        <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 2, pt: 1 }}>
          <TextField label="IP-адрес" value={host} onChange={e => setHost(e.target.value)} fullWidth />
          <TextField label="Порт" value={port} onChange={e => setPort(+e.target.value)} fullWidth type="number" />
          <TextField label="Пользователь" value={username} onChange={e => setUsername(e.target.value)} fullWidth />
          <TextField label="Пароль" value={password} onChange={e => setPassword(e.target.value)} fullWidth type="password" />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseSettings}>Отмена</Button>
          <Button variant="contained" onClick={handleSettingsSave}>Сохранить</Button>
        </DialogActions>
      </Dialog>

      {/* MAIN LAYOUT */}
      <Box display="flex" flexGrow={1}>
        {/* LEFT PANEL */}
        <Box
          width="25%"
          borderRight={1}
          borderColor="divider"
          display="flex"
          flexDirection="column"
        >
          <Box
            flex="1 1 auto"
            sx={{ height: 0, overflowY: "auto" }}
          >
            <FolderTree
              currentPath={currentPath}
              onNavigate={loadDirectory}
              onDropFile={async (src, dest) => {
                await invoke("mv", { src, dest });
                await loadDirectory(currentPath);
                setCreatedFolder({ parentPath: currentPath, newName: src });
              }}
              createdFolder={createdFolder}
              onFolderCreated={() => setCreatedFolder(null)}
              deletedItem={deletedItem}
              onItemDeleted={() => setDeletedItem(null)}
              trashDir={trashDir}
              trashCleared={trashCleared}
              onTrashCleared={() => setTrashCleared(false)}
            />
          </Box>
          <Box p={1} borderTop={1} borderColor="divider">
            <MemoryBar />
          </Box>
        </Box>

        {/* RIGHT PANEL */}
        <Box flex={1} p={2} overflow="auto" onContextMenu={onMainContext} sx={{ height: "calc(100vh - 48px)", overflow: "auto" }}>
          {error && <Box color="error.main" mb={1}>{error}</Box>}
          {renderBreadcrumbs(currentPath, loadDirectory)}
          <FileGrid
            items={files}
            onDoubleClick={handleOpen}
            onContextMenu={onFileContext}
            renameTarget={renameTarget}
            renameValue={renameValue}
            onRenameChange={setRenameValue}
            onRenameConfirm={handleRenameConfirm}
            viewMode={viewMode}
            selectedFiles={selectedFiles}
            setSelectedFiles={setSelectedFiles}
          />
          <Menu
            open={!!menu}
            onClose={handleClose}
            anchorReference="anchorPosition"
            anchorPosition={menu ? { top: menu.mouseY, left: menu.mouseX } : undefined}
          >
            {menu?.name ? (
              selectedFiles.length > 1 && selectedFiles.includes(menu.name) ? (
                <>
                  <MenuItem onClick={handleDownloadSelected}>
                    Скачать выделенное ({selectedFiles.length})
                  </MenuItem>
                  <MenuItem onClick={handleDeleteSelected}>
                    Удалить выделенное ({selectedFiles.length})
                  </MenuItem>
                </>
              ) : (
                <>
                  {menu.isFolder && (
                    <MenuItem onClick={() => handleOpen(menu.name)}>Открыть</MenuItem>
                  )}
                  <MenuItem onClick={() => handleDownloadFolder(menu.name)}>Скачать папку</MenuItem>
                  <MenuItem onClick={() => handleDownload(menu.name)}>Скачать</MenuItem>
                  <MenuItem onClick={() => { setRenameTarget(menu.name); setRenameValue(menu.name); handleClose(); }}>Переименовать</MenuItem>
                  <MenuItem onClick={() => handleDelete(menu.name)}>Удалить</MenuItem>
                </>
              )
            ) : (
              <MenuItem onClick={handleNewFolder}>Новая папка</MenuItem>
            )}
          </Menu>
        </Box>
      </Box>
      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert
          onClose={() => setSnackbar({ ...snackbar, open: false })}
          severity={snackbar.severity}
          sx={{ width: '100%' }}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}
