package data

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	dataRedis "github.com/keratin/authn-server/data/redis"
	"github.com/keratin/authn-server/data/sqlite3"
	"github.com/keratin/authn-server/ops"
)

type BlobStore interface {
	// Read fetches a blob from the store.
	Read(name string) ([]byte, error)

	// WLock acquires a global mutex that will either timeout or be
	// released by a successful Write
	WLock(name string) (bool, error)

	// Write puts a blob into the store.
	Write(name string, blob []byte) error
}

func NewBlobStore(interval time.Duration, redis *redis.Client, db *sqlx.DB, reporter ops.ErrorReporter) (BlobStore, error) {
	// the lifetime of a key should be slightly more than two intervals
	ttl := interval*2 + 10*time.Second

	// the write lock should be greater than the peak time necessary to generate and
	// encrypt a key, plus send it back over the wire to redis.
	lockTime := time.Duration(500) * time.Millisecond

	if redis != nil {
		return &dataRedis.BlobStore{
			TTL:      ttl,
			LockTime: lockTime,
			Client:   redis,
		}, nil
	}

	switch db.DriverName() {
	case "sqlite3":
		store := &sqlite3.BlobStore{
			TTL:      ttl,
			LockTime: lockTime,
			DB:       db,
		}
		store.Clean(reporter)
		return store, nil
	default:
		return nil, fmt.Errorf("unsupported driver: %v", db.DriverName())
	}
}
