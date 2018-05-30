package wallet

import "github.com/rivine/rivine/types"

// UnlockedUnspendOutputs returns all unlocked coinoutput and blockstakeoutputs
func (w *Wallet) UnlockedUnspendOutputs() (map[types.CoinOutputID]types.CoinOutput, map[types.BlockStakeOutputID]types.BlockStakeOutput) {
	w.mu.Lock()
	defer w.mu.Unlock()

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
	return ucom, ubsom
}

// LockedUnspendOutputs returnas all locked coinoutput and blockstakeoutputs
func (w *Wallet) LockedUnspendOutputs() (map[types.CoinOutputID]types.CoinOutput, map[types.BlockStakeOutputID]types.BlockStakeOutput) {
	w.mu.Lock()
	defer w.mu.Unlock()

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
	return ucom, ubsom
}
