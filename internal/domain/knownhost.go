package domain

import "time"

type KnownHost struct {
	Host        string    `json:"host"`
	Port        uint16    `json:"port"`
	KeyType     string    `json:"keyType"`
	Fingerprint string    `json:"fingerprint"`
	PublicKey   string    `json:"publicKey,omitempty"`
	AddedAt     time.Time `json:"addedAt"`
}
