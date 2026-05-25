package domain

import "time"

type TrashEntry struct {
	ID           string    `json:"id"`
	OriginalPath string    `json:"originalPath"`
	TrashedPath  string    `json:"trashedPath"`
	DeletedAt    time.Time `json:"deletedAt"`
	ProfileID    ProfileID `json:"profileId"`
	IsDir        bool      `json:"isDir"`
	Size         int64     `json:"size"`
}
