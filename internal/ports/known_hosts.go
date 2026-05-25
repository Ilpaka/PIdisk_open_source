package ports

import "github.com/pidisk/pidisk/internal/domain"

type KnownHostsStore interface {
	Find(host string, port uint16) (domain.KnownHost, bool, error)
	List() ([]domain.KnownHost, error)
	Add(entry domain.KnownHost) error
	Remove(host string, port uint16) error
	Path() string
}
