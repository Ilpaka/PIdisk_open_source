package trashrepo

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
	dbFileName = "trash.db"
	bucketName = "trash"
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

func (r *Repository) Add(_ context.Context, e domain.TrashEntry) error {
	raw, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return r.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketName)).Put([]byte(e.ID), raw)
	})
}

func (r *Repository) Get(_ context.Context, id string) (domain.TrashEntry, error) {
	var entry domain.TrashEntry
	err := r.db.View(func(tx *bolt.Tx) error {
		raw := tx.Bucket([]byte(bucketName)).Get([]byte(id))
		if raw == nil {
			return domain.ErrTrashEntryNotFound
		}
		return json.Unmarshal(raw, &entry)
	})
	return entry, err
}

func (r *Repository) List(_ context.Context, profileID domain.ProfileID) ([]domain.TrashEntry, error) {
	var out []domain.TrashEntry
	err := r.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketName)).ForEach(func(_, v []byte) error {
			var e domain.TrashEntry
			if err := json.Unmarshal(v, &e); err != nil {
				return err
			}
			if profileID == "" || e.ProfileID == profileID {
				out = append(out, e)
			}
			return nil
		})
	})
	return out, err
}

func (r *Repository) Delete(_ context.Context, id string) error {
	if id == "" {
		return errors.New("empty trash id")
	}
	return r.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketName)).Delete([]byte(id))
	})
}

func (r *Repository) Clear(_ context.Context, profileID domain.ProfileID) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		var stale [][]byte
		err := bucket.ForEach(func(k, v []byte) error {
			var e domain.TrashEntry
			if err := json.Unmarshal(v, &e); err != nil {
				stale = append(stale, append([]byte(nil), k...))
				return nil
			}
			if profileID == "" || e.ProfileID == profileID {
				stale = append(stale, append([]byte(nil), k...))
			}
			return nil
		})
		if err != nil {
			return err
		}
		for _, k := range stale {
			if err := bucket.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})
}

var _ ports.TrashRepository = (*Repository)(nil)
