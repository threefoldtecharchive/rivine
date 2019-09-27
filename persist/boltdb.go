package persist

import (
	"fmt"
	"sync"
	"time"

	bolt "github.com/rivine/bbolt"
)

// BoltDatabase is a persist-level wrapper for the bolt database, providing
// extra information such as a version number.
type BoltDatabase struct {
	Metadata
	*bolt.DB
}

// SaveMetadata overwrites the metadata.
func (db *BoltDatabase) SaveMetadata() error {
	return db.Update(func(tx *bolt.Tx) error {
		// Check if the database has metadata. If not, create metadata for the
		// database.
		bucket := tx.Bucket([]byte("Metadata"))
		if bucket == nil {
			return db.updateMetadata(tx)
		}
		err := bucket.Put([]byte("Header"), []byte(db.Header))
		if err != nil {
			return err
		}
		return bucket.Put([]byte("Version"), []byte(db.Version))
	})
}

// checkMetadata confirms that the metadata in the database is
// correct. If there is no metadata, correct metadata is inserted
func (db *BoltDatabase) checkMetadata(md Metadata) error {
	err := db.Update(func(tx *bolt.Tx) error {
		// Check if the database has metadata. If not, create metadata for the
		// database.
		bucket := tx.Bucket([]byte("Metadata"))
		if bucket == nil {
			return db.updateMetadata(tx)
		}

		// Verify that the metadata matches the expected metadata.
		header := bucket.Get([]byte("Header"))
		if string(header) != md.Header {
			return ErrBadHeader
		}
		version := bucket.Get([]byte("Version"))
		if string(version) != md.Version {
			return ErrBadVersion
		}
		return nil
	})
	return err
}

// updateMetadata will set the contents of the metadata bucket to the values
// in db.Metadata.
func (db *BoltDatabase) updateMetadata(tx *bolt.Tx) error {
	bucket, err := tx.CreateBucketIfNotExists([]byte("Metadata"))
	if err != nil {
		return err
	}
	err = bucket.Put([]byte("Header"), []byte(db.Header))
	if err != nil {
		return err
	}
	err = bucket.Put([]byte("Version"), []byte(db.Version))
	if err != nil {
		return err
	}
	return nil
}

// Close closes the database.
func (db *BoltDatabase) Close() error {
	return db.DB.Close()
}

// OpenDatabase opens a database and validates its metadata.
func OpenDatabase(md Metadata, filename string) (*BoltDatabase, error) {
	// Open the database using a 3 second timeout (without the timeout,
	// database will potentially hang indefinitely.
	db, err := bolt.Open(filename, 0600, &bolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		return nil, err
	}

	// Check the metadata.
	boltDB := &BoltDatabase{
		Metadata: md,
		DB:       db,
	}
	err = boltDB.checkMetadata(md)
	if err != nil {
		db.Close()
		return nil, err
	}

	return boltDB, nil
}

// LazyBoltBucket is a lazy implementation of the bolt DB,
// allowing you to only actually get the bucket from bolt db if you need it.
type LazyBoltBucket struct {
	createdBucket *bolt.Bucket
	once          sync.Once
	getter        func() (*bolt.Bucket, error)
}

// NewLazyBoltBucket creates a new lazy BoltDB bucket,
// see `LazyBoltBucket` for more information.
func NewLazyBoltBucket(getter func() (*bolt.Bucket, error)) *LazyBoltBucket {
	return &LazyBoltBucket{
		createdBucket: nil,
		getter:        getter,
	}
}

func (lb *LazyBoltBucket) AsBoltBucket() (*bolt.Bucket, error) {
	return lb.bucket()
}

func (lb *LazyBoltBucket) Bucket(name []byte) (*bolt.Bucket, error) {
	bucket, err := lb.bucket()
	if err != nil {
		return nil, err
	}
	outBucket := bucket.Bucket(name)
	if outBucket == nil {
		return nil, fmt.Errorf("no bucket found for name %s", string(name))
	}
	return outBucket, nil
}
func (lb *LazyBoltBucket) CreateBucket(key []byte) (*bolt.Bucket, error) {
	bucket, err := lb.bucket()
	if err != nil {
		return nil, err
	}
	return bucket.CreateBucket(key)
}
func (lb *LazyBoltBucket) CreateBucketIfNotExists(key []byte) (*bolt.Bucket, error) {
	bucket, err := lb.bucket()
	if err != nil {
		return nil, err
	}
	return bucket.CreateBucketIfNotExists(key)
}
func (lb *LazyBoltBucket) Cursor() (*bolt.Cursor, error) {
	bucket, err := lb.bucket()
	if err != nil {
		return nil, err
	}
	return bucket.Cursor(), nil
}
func (lb *LazyBoltBucket) Delete(key []byte) error {
	bucket, err := lb.bucket()
	if err != nil {
		return err
	}
	return bucket.Delete(key)
}
func (lb *LazyBoltBucket) DeleteBucket(key []byte) error {
	bucket, err := lb.bucket()
	if err != nil {
		return err
	}
	return bucket.DeleteBucket(key)
}
func (lb *LazyBoltBucket) ForEach(fn func(k, v []byte) error) error {
	bucket, err := lb.bucket()
	if err != nil {
		return err
	}
	return bucket.ForEach(fn)
}
func (lb *LazyBoltBucket) Get(key []byte) ([]byte, error) {
	bucket, err := lb.bucket()
	if err != nil {
		return nil, err
	}
	return bucket.Get(key), nil
}
func (lb *LazyBoltBucket) NextSequence() (uint64, error) {
	bucket, err := lb.bucket()
	if err != nil {
		return 0, err
	}
	return bucket.NextSequence()
}
func (lb *LazyBoltBucket) Put(key []byte, value []byte) error {
	bucket, err := lb.bucket()
	if err != nil {
		return err
	}
	return bucket.Put(key, value)
}
func (lb *LazyBoltBucket) Root() (uint64, error) {
	bucket, err := lb.bucket()
	if err != nil {
		return 0, err
	}
	return uint64(bucket.Root()), nil
}
func (lb *LazyBoltBucket) Sequence() (uint64, error) {
	bucket, err := lb.bucket()
	if err != nil {
		return 0, err
	}
	return bucket.Sequence(), nil
}
func (lb *LazyBoltBucket) SetSequence(v uint64) error {
	bucket, err := lb.bucket()
	if err != nil {
		return nil
	}
	return bucket.SetSequence(v)
}
func (lb *LazyBoltBucket) Stats() (bolt.BucketStats, error) {
	bucket, err := lb.bucket()
	if err != nil {
		return bolt.BucketStats{}, err
	}
	return bucket.Stats(), nil
}
func (lb *LazyBoltBucket) Tx() (*bolt.Tx, error) {
	bucket, err := lb.bucket()
	if err != nil {
		return nil, err
	}
	return bucket.Tx(), nil
}
func (lb *LazyBoltBucket) Writable() (bool, error) {
	bucket, err := lb.bucket()
	if err != nil {
		return false, err
	}
	return bucket.Writable(), nil
}

func (lb *LazyBoltBucket) bucket() (bucket *bolt.Bucket, err error) {
	lb.once.Do(func() {
		lb.createdBucket, err = lb.getter()
	})
	return lb.createdBucket, err
}
