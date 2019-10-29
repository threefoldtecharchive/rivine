package explorergraphql

import (
	"errors"
	"net/http"
	"os"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/modules/explorergraphql/explorerdb"
	"github.com/threefoldtech/rivine/types"

	"github.com/99designs/gqlgen/handler"
	"github.com/julienschmidt/httprouter"
)

// TODO: support extensions somehow

// An Explorer contains a more comprehensive view of the blockchain,
// including various statistics and metrics.
type Explorer struct {
	db             explorerdb.DB
	cs             modules.ConsensusSet
	chainConstants types.ChainConstants
	blockChainInfo types.BlockchainInfo
}

var (
	errNilCS = errors.New("explorer cannot use a nil consensus set")
)

// New creates the internal data structures, and subscribes to
// consensus for changes to the blockchain
func New(cs modules.ConsensusSet, persistDir string, bcInfo types.BlockchainInfo, chainConstants types.ChainConstants, verboseLogging bool) (*Explorer, error) {
	// Check that input modules are non-nil
	if cs == nil {
		return nil, errNilCS
	}

	// Make the persist directory
	err := os.MkdirAll(persistDir, 0700)
	if err != nil {
		return nil, err
	}

	db, err := explorerdb.NewStormDB(persistDir, bcInfo, chainConstants, verboseLogging)
	if err != nil {
		return nil, err
	}

	chainCtx, err := db.GetChainContext()
	if err != nil {
		return nil, err
	}
	e := &Explorer{
		db:             db,
		cs:             cs,
		chainConstants: chainConstants,
		blockChainInfo: bcInfo,
	}

	err = cs.ConsensusSetSubscribe(e, chainCtx.ConsensusChangeID, nil)
	if err != nil {
		// TODO: restart from 0
		return nil, errors.New("explorer subscription failed: " + err.Error())
	}

	return e, nil
}

func (e *Explorer) SetHTTPHandlers(router *httprouter.Router, endpoint string) {
	rootHandler := handler.Playground("GraphQL playground", endpoint+"/query")
	router.Handle("GET", endpoint, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		rootHandler(w, r)
	})
	queryHandler := handler.GraphQL(NewExecutableSchema(Config{Resolvers: &Resolver{
		db:             e.db,
		cs:             e.cs,
		chainConstants: e.chainConstants,
		blockchainInfo: e.blockChainInfo,
	}}))
	router.Handle("POST", endpoint+"/query", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		queryHandler(w, r)
	})
}

// ProcessConsensusChange follows the most recent changes to the consensus set,
// including parsing new blocks and updating the utxo sets.
func (e *Explorer) ProcessConsensusChange(cc modules.ConsensusChange) {
	if len(cc.AppliedBlocks) == 0 {
		build.Critical("Explorer.ProcessConsensusChange called with a ConsensusChange that has no AppliedBlocks")
	}
	err := explorerdb.ApplyConsensusChange(e.db, e.cs, cc, &e.chainConstants)
	if err != nil {
		build.Critical("Explorer.ProcessConsensusChange failed", err)
	}
}

// Close closes the explorer.
func (e *Explorer) Close() error {
	e.cs.Unsubscribe(e)
	return e.db.Close()
}
