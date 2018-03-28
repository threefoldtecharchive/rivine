package datastore

import (
	"errors"
	"fmt"
	"sync"

	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/persist"
	"github.com/rivine/rivine/types"
)

// local error variables
var (
	errNilCS = errors.New("datastore cannot use a nil consensus set")
	errNilDB = errors.New("datastore cannot use a nil database")
)

type (
	// DataStore pulls arbitrary data, stored in transactions, from the blockchain, and saves
	// it in a connected database
	DataStore struct {
		cs modules.ConsensusSet
		db Database

		log *persist.Logger

		// Keep a reference to all running namespace managers. Key is the Namespace
		managers map[Namespace]*namespaceManager
		mu       sync.Mutex // Add some protection to the map

		bcInfo   types.BlockchainInfo
		chainCts types.ChainConstants
	}
)

// New creates a new DataStore from a consensus set and a Database.
// If the connection opts specify an unknown driver, initialization fails
// and an error is returned. Currently only `redis` is supported
func New(cs modules.ConsensusSet, db Database, persistDir string, bcInfo types.BlockchainInfo, chainCts types.ChainConstants) (*DataStore, error) {
	// Check that we have a valid consensus set
	if cs == nil {
		return nil, errNilCS
	}
	// Check that we have a valid database
	if db == nil {
		return nil, errNilDB
	}
	// Check that the database is currently reachable
	if err := db.Ping(); err != nil {
		return nil, err
	}

	ds := &DataStore{
		cs:       cs,
		db:       db,
		bcInfo:   bcInfo,
		chainCts: chainCts,
	}

	if err := ds.initLogger(persistDir); err != nil {
		return nil, errors.New("Failed to initialize datastorelogger: " + err.Error())
	}

	ds.log.SetPrefix("[DataStore]:")
	ds.log.Println("Datastore initialized")

	// Load the already existing managers
	ds.managers = make(map[Namespace]*namespaceManager)
	mgrs, err := ds.db.LoadFieldsForKey(subscriberSet)
	if err != nil {
		return nil, errors.New("Failed to load existing namespace managers: " + err.Error())
	}
	for nsString, sb := range mgrs {
		ns := Namespace{}
		err := ns.LoadString(nsString)
		if err != nil {
			ds.log.Severe("Failed to load manager - namespace: ", err)
			continue
		}
		mgr, err := ds.newNamespaceManagerFromSerializedState(ns, sb)
		if err != nil {
			ds.log.Severe("Failed to load manager - state: ", err)
			continue
		}
		ds.managers[ns] = mgr
	}

	ds.log.Printf("Loaded %d namespace managers from db", len(mgrs))

	// set up the event subscription
	subChan := make(chan *SubEvent)
	go ds.messageCollector(subChan)
	// Subscribe to redis and start/stop managers
	ds.db.Subscribe(subChan)

	return ds, nil
}

// Close closes the datastore, all namespace managers, and finally its connection to the database
func (ds *DataStore) Close() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// First unsubscribe from the replication channel
	if err := ds.db.Unsubscribe(); err != nil {
		// Log a possible failure, but don't stop as the database connection can still be
		// closed later on
		ds.log.Severe("Failed to unsubscribe from subscription channel: ", err)
	}

	ds.log.Println("Datastore shutting down...")

	wg := sync.WaitGroup{}
	for _, nsm := range ds.managers {
		wg.Add(1)
		go func(n *namespaceManager) {
			ds.log.Debugln("Shutting down namespace manager...")
			n.close()
			wg.Done()
			ds.log.Debugln("Namespace manager shut down")
		}(nsm)
	}
	wg.Wait()
	ds.log.Debugln("All namespace managers shut down!")
	err := ds.db.Close()
	if err != nil {
		ds.log.Severe("Failed to close db connection: ", err)
		fmt.Println("Failed to close db connection: ", err)
	}
	ds.log.Debugln("DB connection closed")
	err = ds.log.Close()
	if err != nil {
		// State of the logger is unknown, a println will suffice.
		fmt.Println("Error shutting down datastore logger:", err)
	}
	return err
}

func (ds *DataStore) messageCollector(ch <-chan *SubEvent) {
	for ev := range ch {
		if ev == nil {
			// Closed channel
			return
		}
		switch ev.Action {
		case SubStart:
			// Check if we track this namespace already
			// For now, remove the old manager and fire up the new one
			// This can be improved by also writing the first registered ID in the tracked set
			// and ignoring the others
			if eNsm, exists := ds.managers[ev.Namespace]; exists {
				ds.log.Debugln("Removing duplicate namespace manager")
				eNsm.close()
			}
			nsm := ds.newNamespaceManager(ev.Namespace, ev.Start)
			if err := nsm.save(); err != nil {
				ds.log.Severe("Failed to save namespace manager during initializtion: ", err)
				return
			}
			ds.managers[ev.Namespace] = nsm
		case SubEnd:
			nsm, exists := ds.managers[ev.Namespace]
			if !exists {
				ds.log.Debugln("Failed to unsubscribe from namespace, not subscribed to namespace", ev.Namespace)
				return
			}
			nsm.close()
			err := nsm.delete()
			if err != nil {
				ds.log.Severe("Failed to delete namespace manager: ", err)
				return
			}
			ds.log.Debugln("Deleted namespace manager for namespace", ev.Namespace)
		}
	}
}
