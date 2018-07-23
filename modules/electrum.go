package modules

const (
	// ElectrumDir is the name of the subdirectory to create the persistent files
	ElectrumDir = "electrum"
)

type (
	// Electrum implements the electrum protocol,
	// see https://electrumx.readthedocs.io/en/latest/protocol.html
	Electrum interface {
		// Close closes the electrum server. It will also close all currently
		// open connections
		Close() error
		// Start all the listeners which accept incomming connections to serve
		// the electrum protocol on
		Start()
	}
)
