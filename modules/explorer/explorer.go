// Package explorer provides a glimpse into what the network currently
// looks like.
package explorer

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"runtime"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/persist"
	"github.com/threefoldtech/rivine/types"

	certmagic "github.com/mholt/certmagic"
)

const (
	// ActiveBSEstimationBlocks is the number of blocks that are used to
	// estimate the active block stake used to generate blocks.
	ActiveBSEstimationBlocks = 200
)

var (
	errNilCS = errors.New("explorer cannot use a nil consensus set")
)

type (
	// blockFacts contains a set of facts about the consensus set related to a
	// certain block. The explorer needs some additional information in the
	// history so that it can calculate certain values, which is one of the
	// reasons that the explorer uses a separate struct instead of
	// modules.BlockFacts.
	blockFacts struct {
		modules.BlockFacts

		Timestamp types.Timestamp
	}

	// An Explorer contains a more comprehensive view of the blockchain,
	// including various statistics and metrics.
	Explorer struct {
		cs             modules.ConsensusSet
		db             *persist.BoltDatabase
		persistDir     string
		bcInfo         types.BlockchainInfo
		chainCts       types.ChainConstants
		rootTarget     types.Target
		genesisBlock   types.Block
		genesisBlockID types.BlockID
	}
)

// New creates the internal data structures, and subscribes to
// consensus for changes to the blockchain
func New(cs modules.ConsensusSet, persistDir string, bcInfo types.BlockchainInfo, chainCts types.ChainConstants) (*Explorer, error) {
	// Check that input modules are non-nil
	if cs == nil {
		return nil, errNilCS
	}

	// Initialize the explorer.
	genesisBlock := chainCts.GenesisBlock()
	e := &Explorer{
		cs:             cs,
		persistDir:     persistDir,
		bcInfo:         bcInfo,
		chainCts:       chainCts,
		rootTarget:     chainCts.RootTarget(),
		genesisBlock:   genesisBlock,
		genesisBlockID: genesisBlock.ID(),
	}

	// Initialize the persistent structures, including the database.
	err := e.initPersist()
	if err != nil {
		return nil, err
	}

	// retrieve the current ConsensusChangeID
	var recentChange modules.ConsensusChangeID
	err = e.db.View(dbGetInternal(internalRecentChange, &recentChange))
	if err != nil {
		return nil, err
	}

	err = cs.ConsensusSetSubscribe(e, recentChange, nil)
	if err != nil {
		// TODO: restart from 0
		return nil, errors.New("explorer subscription failed: " + err.Error())
	}

	return e, nil
}

// Close closes the explorer.
func (e *Explorer) Close() error {
	e.cs.Unsubscribe(e)
	return e.db.Close()
}

// ServeFrontend serves all static files in /frontend/explorer/public
// Which will host the explorer locally on :port 2015
func ServeFrontend(isLocalhost bool, domainNames []string, email string) error {
	// read and agree to your CA's legal documents
	certmagic.Agreed = true

	// provide an email address
	certmagic.Email = email

	// use the staging endpoint while we're developing
	certmagic.CA = certmagic.LetsEncryptStagingCA

	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return errors.New("Calling runtime failed")
	}

	// Path to frontend static files
	filepath := path.Join(path.Dir(filename), "./../../frontend/explorer/public")

	http.Handle("/", http.FileServer(http.Dir(filepath)))

	if !isLocalhost {
		for _, domainName := range domainNames {
			if !certmagic.HostQualifies(domainName) {
				return fmt.Errorf("Domain name %s is not valid for automatic https", domainName)
			}
		}
		return certmagic.HTTPS(domainNames, nil)
	}
	return http.ListenAndServe(":2015", nil)
}
