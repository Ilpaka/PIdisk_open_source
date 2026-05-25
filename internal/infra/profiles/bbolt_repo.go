package profiles

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
	bucketName = "profiles"
	dbFileName = "profiles.db"
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

func (r *Repository) List(_ context.Context) ([]domain.Profile, error) {
	var out []domain.Profile
	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.ForEach(func(_, v []byte) error {
			var p domain.Profile
			if err := json.Unmarshal(v, &p); err != nil {
				return err
			}
			out = append(out, p)
			return nil
		})
	})
	return out, err
}

func (r *Repository) Get(_ context.Context, id domain.ProfileID) (domain.Profile, error) {
	var p domain.Profile
	err := r.db.View(func(tx *bolt.Tx) error {
		raw := tx.Bucket([]byte(bucketName)).Get([]byte(id))
		if raw == nil {
			return domain.ErrProfileNotFound
		}
		return json.Unmarshal(raw, &p)
	})
	return p, err
}

func (r *Repository) Save(_ context.Context, p domain.Profile) error {
	if p.ID == "" {
		return errors.New("profile id required")
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return r.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketName)).Put([]byte(p.ID), raw)
	})
}

func (r *Repository) Delete(_ context.Context, id domain.ProfileID) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketName)).Delete([]byte(id))
	})
}

var _ ports.ProfileRepository = (*Repository)(nil)
