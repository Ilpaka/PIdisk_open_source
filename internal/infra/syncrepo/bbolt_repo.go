package syncrepo

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"time"

	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/ports"
	bolt "go.etcd.io/bbolt"
)

const (
	dbFileName = "sync.db"
	bucketName = "folders"
)

type Repository struct {
	db *bolt.DB
}

func Open(dataDir string) (*Repository, error) {
	path := filepath.Join(dataDir, dbFileName)
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, err
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists([]byte(bucketName))
		return e
	}); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
	if r.db == nil {
		return nil
	}
	return r.db.Close()
}

func key(profileID domain.ProfileID, name string) []byte {
	return []byte(string(profileID) + "\x00" + name)
}

func (r *Repository) Save(_ context.Context, profileID domain.ProfileID, folder domain.SyncFolder) error {
	if folder.Name == "" {
		return errors.New("folder name required")
	}
	raw, err := json.Marshal(folder)
	if err != nil {
		return err
	}
	return r.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketName)).Put(key(profileID, folder.Name), raw)
	})
}

func (r *Repository) Get(_ context.Context, profileID domain.ProfileID, name string) (domain.SyncFolder, error) {
	var out domain.SyncFolder
	err := r.db.View(func(tx *bolt.Tx) error {
		raw := tx.Bucket([]byte(bucketName)).Get(key(profileID, name))
		if raw == nil {
			return domain.ErrSyncFolderMissing
		}
		return json.Unmarshal(raw, &out)
	})
	return out, err
}

func (r *Repository) List(_ context.Context, profileID domain.ProfileID) ([]domain.SyncFolder, error) {
	prefix := []byte(string(profileID) + "\x00")
	var out []domain.SyncFolder
	err := r.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketName)).Cursor()
		for k, v := c.Seek(prefix); k != nil && hasPrefix(k, prefix); k, v = c.Next() {
			var f domain.SyncFolder
			if err := json.Unmarshal(v, &f); err != nil {
				return err
			}
			out = append(out, f)
		}
		return nil
	})
	return out, err
}

func (r *Repository) Delete(_ context.Context, profileID domain.ProfileID, name string) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketName)).Delete(key(profileID, name))
	})
}

func hasPrefix(b, prefix []byte) bool {
	if len(b) < len(prefix) {
		return false
	}
	for i, c := range prefix {
		if b[i] != c {
			return false
		}
	}
	return true
}

var _ ports.SyncFolderRepository = (*Repository)(nil)
