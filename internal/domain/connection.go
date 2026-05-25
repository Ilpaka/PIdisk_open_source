package domain

import "time"

type ConnectionState struct {
	ProfileID ProfileID `json:"profileId"`
	Connected bool      `json:"connected"`
	LastError string    `json:"lastError,omitempty"`
	LastPing  time.Time `json:"lastPing"`
}
