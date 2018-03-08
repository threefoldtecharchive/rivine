package datastore

// Database is the common interface which must be implemented to be compatible with the datastore
type Database interface {
	// Ping tests the database connection
	Ping() error
	// SaveManager saves a managers state
	SaveManager(*NamespaceManager) error
	// GetManagers loads all active managers
	GetManagers() (map[Namespace]*NamespaceManager, error)
	// DeleteManager removes a manager in case it unsubscribes
	DeleteManager(*NamespaceManager) error
	// StoreData stores the specified data, linked to the namespace and the generated ID
	// Multiple section of data must be storeable for the same namespace. The ID is unique
	// for the namespace, but other namespaces might have the same ID's for different data
	StoreData(Namespace, DataID, []byte) error
	// DeleteData removes the data specified by the namespace and ID.
	DeleteData(Namespace, DataID) error
	// Subscribe continuously manages a channel for messages
	Subscribe(SubEventCallback)
	// Close gracefully closes the database connection
	Close() error
}
