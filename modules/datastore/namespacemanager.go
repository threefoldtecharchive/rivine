package datastore

import (
	"bytes"
	"sync"

	"github.com/rivine/rivine/persist"

	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

const (
	// Namespace to sto
	subscriberSet = "s"
)

type (
	// namespaceManager keeps track of data for a namespace
	namespaceManager struct {
		// namespace is the namepsace we are tracking
		namespace Namespace
		// state of the manager
		state namespaceManagerState
		// buffer for the incomming blocks so replication is delayed
		// and we can prevent deletion of data
		buffer *blockBuffer
		// cs we are subscribed to
		cs modules.ConsensusSet
		// db used to replicate the data
		db Database

		log *persist.Logger

		mu sync.Mutex
	}

	// namespaceManagerState is the internal state of a namespace manager
	namespaceManagerState struct {
		// BlockHeight is the current block height of the tracker
		BlockHeight types.BlockHeight
		// RecentChangeID is the last consensuschangeid we received
		RecentChangeID modules.ConsensusChangeID
		// SubscribeStart is an optional timestamp, all data before this is ignored
		SubscribeStart types.Timestamp
	}
)

// newNamespaceManager initializes a new namespace manager for a given namespace,
// with a timestamp to start tracking from
func (ds *DataStore) newNamespaceManager(namespace Namespace, ts types.Timestamp) *namespaceManager {
	nsm := &namespaceManager{
		state: namespaceManagerState{
			SubscribeStart: ts,
		},
		namespace: namespace,
		buffer:    ds.newBlockBuffer(),
		db:        ds.db,
		log:       ds.log,
	}
	go ds.cs.ConsensusSetSubscribe(nsm, nsm.state.RecentChangeID)
	return nsm
}

// newNamespaceManagerFromSerializedState creates a new namespacemaneger with a pre-created state
func (ds *DataStore) newNamespaceManagerFromSerializedState(namespace Namespace, stateBytes []byte) (*namespaceManager, error) {
	nsm := &namespaceManager{
		namespace: namespace,
		buffer:    ds.newBlockBuffer(),
		db:        ds.db,
		log:       ds.log,
	}
	err := nsm.deserialize(stateBytes)
	if err != nil {
		return nil, err
	}
	go ds.cs.ConsensusSetSubscribe(nsm, nsm.state.RecentChangeID)
	return nsm, err
}

// save updates the state of the namespace manager in the database
func (nsm *namespaceManager) save() error {
	serializedData, err := nsm.serialize()
	if err != nil {
		return err
	}
	return nsm.db.StoreData(subscriberSet, nsm.namespace.String(), serializedData)
}

// delete removes the namespace manager from the database
func (nsm *namespaceManager) delete() error {
	return nsm.db.DeleteData(subscriberSet, nsm.namespace.String())
}

// close shuts down this namespace manager
func (nsm *namespaceManager) close() {
	nsm.mu.Lock()
	defer nsm.mu.Unlock()

	if nsm.cs == nil {
		return
	}
	nsm.cs.Unsubscribe(nsm)
	nsm.cs = nil
	return
}

// serialize converts the state of a namespace manager to byte form
func (nsm *namespaceManager) serialize() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	st := nsm.state
	err := encoding.NewEncoder(buffer).Encode(st)
	return buffer.Bytes(), err
}

// deserialize loads state from a previously serialized state
func (nsm *namespaceManager) deserialize(state []byte) error {
	st := &nsm.state
	buf := bytes.NewBuffer(state)
	return encoding.NewDecoder(buf).Decode(st)
}
