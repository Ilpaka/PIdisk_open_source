package domain

import "time"

type TransferID string

type TransferKind string

const (
	KindUpload      TransferKind = "upload"
	KindDownload    TransferKind = "download"
	KindDownloadDir TransferKind = "download_dir"
)

type TransferState string

const (
	StateRunning   TransferState = "running"
	StateDone      TransferState = "done"
	StateError     TransferState = "error"
	StateCancelled TransferState = "cancelled"
)

type TransferProgress struct {
	ID          TransferID    `json:"id"`
	Kind        TransferKind  `json:"kind"`
	Name        string        `json:"name"`
	LocalPath   string        `json:"localPath,omitempty"`
	RemotePath  string        `json:"remotePath,omitempty"`
	BytesDone   int64         `json:"bytesDone"`
	BytesTotal  int64         `json:"bytesTotal"`
	Speed       int64         `json:"speed"`
	State       TransferState `json:"state"`
	StartedAt   time.Time     `json:"startedAt"`
	CompletedAt time.Time     `json:"completedAt,omitempty"`
	Error       string        `json:"error,omitempty"`
}
