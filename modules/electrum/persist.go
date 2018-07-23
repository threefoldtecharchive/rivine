package electrum

import (
	"os"
	"path/filepath"

	"github.com/rivine/rivine/persist"
)

const (
	logFile = "electrum.log"
)

func (e *Electrum) initLogger(persistDir string) error {
	// / Create the consensus directory.
	err := os.MkdirAll(persistDir, 0700)
	if err != nil {
		return err
	}

	// Initialize the logger.
	e.log, err = persist.NewFileLogger(e.bcInfo, filepath.Join(persistDir, logFile))
	return err
}
