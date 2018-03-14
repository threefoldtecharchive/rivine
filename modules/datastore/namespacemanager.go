package datastore

import (
	"bytes"
	"sync"

	"github.com/rivine/rivine/persist"

	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

type (
	// NamespaceManager keeps track of data for a namespace
	NamespaceManager struct {
		// Namespace is the namepsace we are tracking
		Namespace Namespace
		// State of the manager
		State NamespaceManagerState
		// Buffer for the incomming blocks so replication is delayed
		// and we can prevent deletion of data
		Buffer *BlockBuffer
		// Cs we are subscribed to
		Cs modules.ConsensusSet
		// DB used to replicate the data
		DB Database

		log *persist.Logger

		mu sync.Mutex
	}

	// NamespaceManagerState is the internal state of a namespace manager
	NamespaceManagerState struct {
		// BlockHeight is the current block height of the tracker
		BlockHeight types.BlockHeight
		// RecentChangeID is the last consensuschangeid we received
		RecentChangeID modules.ConsensusChangeID
		// DataID is the last data id we stored
		DataID DataID
		// SubscribeStart is an optional timestamp, all data before this is ignored
		SubscribeStart types.Timestamp
	}
)

// NewNamespaceManager initializes a new namespace manager for a given namespace,
// with a timestamp to start tracking from
func NewNamespaceManager(namespace Namespace, db Database, ts types.Timestamp, log *persist.Logger) *NamespaceManager {
	nsm := &NamespaceManager{
		State: NamespaceManagerState{
			SubscribeStart: ts,
		},
		Namespace: namespace,
		Buffer:    NewBlockBuffer(),
		DB:        db,
		log:       log,
	}
	return nsm
}

// Save updates the state of the namespace manager in the database
func (nsm *NamespaceManager) Save() error {
	return nsm.DB.SaveManager(nsm)
}

// Delete removes the namespace manager from the database
func (nsm *NamespaceManager) Delete() error {
	return nsm.DB.DeleteManager(nsm)
}

// SubscribeCs subscribes this namespace manager to the given consensus set
func (nsm *NamespaceManager) SubscribeCs(cs modules.ConsensusSet) {
	nsm.mu.Lock()
	nsm.Cs = cs
	nsm.mu.Unlock()

	// We can't lock here since ConsensusSetSubscribe will push all changes until
	// we are caught up with the current thread, and the update function already locks.
	// Then again, we should only ever call subscribe during the creation of a namespace
	// manager in the datastore, (either loading or a new subscription), so this should be fine.

	err := cs.ConsensusSetSubscribe(nsm, nsm.State.RecentChangeID)
	if err != nil {
		nsm.log.Severe("Namespace manager failed to subscribe to consensus set: ", err)
	}
}

// UnSubscribeCs removes this namespace manager from the previously subscribed consensus set
func (nsm *NamespaceManager) UnSubscribeCs() {
	nsm.mu.Lock()
	defer nsm.mu.Unlock()

	if nsm.Cs == nil {
		return
	}
	nsm.Cs.Unsubscribe(nsm)
	nsm.Cs = nil
}

// Serialize converts the state of a namespace manager to byte form
func (nsm *NamespaceManager) Serialize() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	st := nsm.State
	err := encoding.NewEncoder(buffer).Encode(st)
	return buffer.Bytes(), err
}

// Deserialize loads state from a previously serialized state
func (nsm *NamespaceManager) Deserialize(state []byte) error {
	st := &nsm.State
	buf := bytes.NewBuffer(state)
	return encoding.NewDecoder(buf).Decode(st)
}
