package domain

import "time"

type FileEntry struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	IsDir    bool      `json:"isDir"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
	Mode     uint32    `json:"mode"`
}

type Listing struct {
	Path    string      `json:"path"`
	Entries []FileEntry `json:"entries"`
}

type DiskUsage struct {
	Path    string  `json:"path"`
	Used    uint64  `json:"used"`
	Free    uint64  `json:"free"`
	Total   uint64  `json:"total"`
	Percent float64 `json:"percent"`
	Raw     string  `json:"raw,omitempty"`
}
