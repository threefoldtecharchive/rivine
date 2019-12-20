package wallet

import (
	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

// UnlockedUnspendOutputs returns all unlocked coinoutput and blockstakeoutputs
func (w *Wallet) UnlockedUnspendOutputs() (map[types.CoinOutputID]types.CoinOutput, map[types.BlockStakeOutputID]types.BlockStakeOutput, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		return nil, nil, modules.ErrLockedWallet
	}

	ucom := make(map[types.CoinOutputID]types.CoinOutput)
	ubsom := make(map[types.BlockStakeOutputID]types.BlockStakeOutput)

	// prepare fulfillable context
	ctx := w.getFulfillableContextForLatestBlock()

	// get all coin and block stake stum
	for id, co := range w.coinOutputs {
		if co.Condition.Fulfillable(ctx) {
			ucom[id] = co
		}
	}
	// same for multisig
	for id, co := range w.multiSigCoinOutputs {
		if co.Condition.Fulfillable(ctx) {
			ucom[id] = co
		}
	}
	// block stakes
	for id, bso := range w.blockstakeOutputs {
		if bso.Condition.Fulfillable(ctx) {
			ubsom[id] = bso
		}
	}
	// block stake multisigs
	for id, bso := range w.multiSigBlockStakeOutputs {
		if bso.Condition.Fulfillable(ctx) {
			ubsom[id] = bso
		}
	}
	return ucom, ubsom, nil
}

// LockedUnspendOutputs returns all locked coinoutput and blockstakeoutputs
func (w *Wallet) LockedUnspendOutputs() (map[types.CoinOutputID]types.CoinOutput, map[types.BlockStakeOutputID]types.BlockStakeOutput, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		return nil, nil, modules.ErrLockedWallet
	}

	ucom := make(map[types.CoinOutputID]types.CoinOutput)
	ubsom := make(map[types.BlockStakeOutputID]types.BlockStakeOutput)

	// prepare fulfillable context
	ctx := w.getFulfillableContextForLatestBlock()

	// get all coin and block stake stum
	for id, co := range w.coinOutputs {
		if !co.Condition.Fulfillable(ctx) {
			ucom[id] = co
		}
	}
	// same for multisig
	for id, co := range w.multiSigCoinOutputs {
		if !co.Condition.Fulfillable(ctx) {
			ucom[id] = co
		}
	}
	// block stakes
	for id, bso := range w.blockstakeOutputs {
		if !bso.Condition.Fulfillable(ctx) {
			ubsom[id] = bso
		}
	}
	// block stake multisigs
	for id, bso := range w.multiSigBlockStakeOutputs {
		if !bso.Condition.Fulfillable(ctx) {
			ubsom[id] = bso
		}
	}
	return ucom, ubsom, nil
}

// sortedOutputs is a struct containing a slice of coin outputs and their
// corresponding ids. sortedOutputs can be sorted using the sort package.
type sortedOutputs struct {
	ids     []types.CoinOutputID
	outputs []types.CoinOutput
}

// Len returns the number of elements in the sortedOutputs struct.
func (so sortedOutputs) Len() int {
	if len(so.ids) != len(so.outputs) {
		build.Severe("sortedOutputs object is corrupt")
	}
	return len(so.ids)
}

// Less returns whether element 'i' is less than element 'j'. The currency
// value of each output is used for comparison.
func (so sortedOutputs) Less(i, j int) bool {
	return so.outputs[i].Value.Cmp(so.outputs[j].Value) < 0
}

// Swap swaps two elements in the sortedOutputs set.
func (so sortedOutputs) Swap(i, j int) {
	so.ids[i], so.ids[j] = so.ids[j], so.ids[i]
	so.outputs[i], so.outputs[j] = so.outputs[j], so.outputs[i]
}
