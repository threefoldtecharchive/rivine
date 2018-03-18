package datastore

// Database is the common interface which must be implemented to be compatible with the datastore
type Database interface {
	// Ping tests the database connection
	Ping() error
	// StoreData stores the specified data, linked to a key and a field in that key
	// Multiple section of data must be storeable for the same key. The field is unique
	// for the key, but other keys might have the same fields for different data
	StoreData(string, string, []byte) error
	// DeleteData removes the data specified by the key and field.
	DeleteData(string, string) error
	// LoadFieldsForKey loads a mapping of all fields and the associated data for a given key
	LoadFieldsForKey(string) (map[string][]byte, error)
	// Subscribe continuously manages a channel for messages
	// Subscribe(SubEventCallback)
	Subscribe(chan<- *SubEvent)
	// Unsubscribe ends the subscription to the message channel and closes the
	// event channel provided by the subscribe function
	Unsubscribe() error
	// Close gracefully closes the database connection
	Close() error
}
