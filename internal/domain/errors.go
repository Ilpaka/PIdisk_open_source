package domain

import "errors"

var (
	ErrProfileNotFound    = errors.New("profile not found")
	ErrProfileLocked      = errors.New("profile is locked")
	ErrProfileExists      = errors.New("profile with the same name exists")
	ErrNoActiveProfile    = errors.New("no active profile")
	ErrInvalidProfileName = errors.New("invalid profile name")
	ErrInvalidHost        = errors.New("invalid host")
	ErrInvalidPort        = errors.New("invalid port")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidKeyPath     = errors.New("invalid private key path")
	ErrInvalidRootDir     = errors.New("invalid root directory")
	ErrInvalidTrashDir    = errors.New("invalid trash directory")

	ErrNotConnected     = errors.New("ssh client is not connected")
	ErrConnectionClosed = errors.New("ssh connection closed")
	ErrHostKeyMismatch  = errors.New("host key mismatch, possible MITM")
	ErrHostKeyRejected  = errors.New("host key rejected by user")
	ErrHostKeyPending   = errors.New("host key awaiting user confirmation")

	ErrInvalidPath    = errors.New("invalid path")
	ErrPathTraversal  = errors.New("path contains traversal segments")
	ErrTransferCancel = errors.New("transfer cancelled")
	ErrTransferNotFound = errors.New("transfer not found")

	ErrTrashEntryNotFound = errors.New("trash entry not found")

	ErrSyncFolderExists   = errors.New("sync folder with the same name exists")
	ErrSyncFolderMissing  = errors.New("sync folder not found")
	ErrSyncAlreadyRunning = errors.New("sync engine already running")
	ErrSyncNotRunning     = errors.New("sync engine is not running")
)
