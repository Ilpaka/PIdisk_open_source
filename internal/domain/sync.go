package domain

import "time"

type SyncDirection string

const (
	SyncBoth    SyncDirection = "both"
	SyncToRemote SyncDirection = "to_remote"
	SyncToLocal  SyncDirection = "to_local"
)

type SyncFolder struct {
	Name       string    `json:"name"`
	LocalPath  string    `json:"localPath"`
	RemotePath string    `json:"remotePath"`
	Enabled    bool      `json:"enabled"`
	LastSync   time.Time `json:"lastSync"`
	Direction  SyncDirection `json:"direction"`
}

type SyncStats struct {
	IsRunning     bool           `json:"isRunning"`
	LastSyncTime  time.Time      `json:"lastSyncTime"`
	SyncedFolders []string       `json:"syncedFolders"`
	Errors        []string       `json:"errors"`
	FilesSynced   uint32         `json:"filesSynced"`
	BytesSynced   uint64         `json:"bytesSynced"`
	Conflicts     []SyncConflict `json:"conflicts"`
}

type SyncConflict struct {
	Folder      string    `json:"folder"`
	Path        string    `json:"path"`
	LocalMTime  time.Time `json:"localMtime"`
	RemoteMTime time.Time `json:"remoteMtime"`
	Resolution  string    `json:"resolution"`
}

type SyncEvent struct {
	Folder      string    `json:"folder"`
	FilesSynced uint32    `json:"filesSynced"`
	BytesSynced uint64    `json:"bytesSynced"`
	DurationMs  int64     `json:"durationMs"`
	Error       string    `json:"error,omitempty"`
	StartedAt   time.Time `json:"startedAt"`
}
