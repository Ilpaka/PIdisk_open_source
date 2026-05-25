// src/components/SyncPanel.jsx

import React from 'react';
import {
  Card,
  CardContent,
  Box,
  Typography,
  Button,
  Grid,
  Chip,
  LinearProgress,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  FormControlLabel,
  Switch,
  IconButton,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import SyncIcon from '@mui/icons-material/Sync';
import StopIcon from '@mui/icons-material/Stop';
import FolderIcon from '@mui/icons-material/Folder';
import DeleteIcon from '@mui/icons-material/Delete';

function formatBytes(bytes) {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export default function SyncPanel({
  syncStatus,
  syncFolders,
  onAddFolder,
  onStartSync,
  onStopSync,
  onToggleFolder,
  onRemoveFolder,
}) {
  return (
    <Card sx={{ m: 1 }}>
      <CardContent>
        {/* Заголовок и кнопки управления */}
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Typography variant="h6">Синхронизация файлов</Typography>
          <Box>
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              onClick={onAddFolder}
              sx={{ mr: 1 }}
            >
              Добавить папку
            </Button>
            {syncStatus.is_running ? (
              <Button
                variant="contained"
                color="error"
                startIcon={<StopIcon />}
                onClick={onStopSync}
              >
                Остановить
              </Button>
            ) : (
              <Button
                variant="contained"
                color="success"
                startIcon={<SyncIcon />}
                onClick={onStartSync}
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
                      onChange={(e) => onToggleFolder(folder.name, e.target.checked)}
                    />
                  }
                  label="Активна"
                />
                <IconButton
                  edge="end"
                  onClick={() => onRemoveFolder(folder.name)}
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
  );
}
