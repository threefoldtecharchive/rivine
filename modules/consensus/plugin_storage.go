package consensus

import (
	"errors"

	"github.com/threefoldtech/rivine/persist"

	bolt "github.com/rivine/bbolt"
)

// A PluginViewStorage struct is a definition for the storage tool for a plugin bucket
type PluginViewStorage struct {
	db   *persist.BoltDatabase
	name string
}

// NewPluginStorage creates a new plugin storage for a given plugin bucket.
// PluginViewStorage abstracts the way the plugin bucket manages its data.
func NewPluginStorage(db *persist.BoltDatabase, name string) *PluginViewStorage {
	return &PluginViewStorage{
		db:   db,
		name: name,
	}
}

// View takes in a callback to a bucket and takes care of managing the database knowledge
func (ps *PluginViewStorage) View(callback func(bucket *bolt.Bucket) error) error {
	return ps.db.View(func(tx *bolt.Tx) error {
		rootbucket := tx.Bucket([]byte(BucketPlugins))
		if rootbucket == nil {
			return errors.New("Plugins bucket does not exist")
		}
		childbucket := rootbucket.Bucket([]byte(ps.name))
		if childbucket == nil {
			return errors.New("Bucket is not present in plugins bucket")
		}
		return callback(childbucket)
	})
}
