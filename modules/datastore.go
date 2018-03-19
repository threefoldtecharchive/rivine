package modules

const (
	// DataStoreDir is the name of the subdirectory to create the persistent files
	DataStoreDir = "datastore"
	// DataStoreDatabaseSubDir is the name of the subdirectory (within the DataStoreDir)
	// to create the persistent redis files
	DataStoreDatabaseSubDir = "database"
)

type (
	// DataStore pulls arbitrary data, stored in transactions, from the blockchain, and saves
	// it in a database
	DataStore interface {
		// Close closes the datstore. It will also close the connection to the database
		Close() error
	}
)
