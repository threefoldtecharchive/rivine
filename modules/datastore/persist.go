package datastore

import (
	"os"
	"path/filepath"

	"github.com/rivine/rivine/persist"
)

const (
	logFile = "datastore.log"
)

func (ds *DataStore) initLogger(persistDir string) error {
	// / Create the consensus directory.
	err := os.MkdirAll(persistDir, 0700)
	if err != nil {
		return err
	}

	// Initialize the logger.
	ds.log, err = persist.NewFileLogger(ds.bcInfo, filepath.Join(persistDir, logFile))
	return err
}
