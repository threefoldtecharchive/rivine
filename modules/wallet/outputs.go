package wallet

import (
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

// UnlockedUnspendOutputs returns all unlocked coinoutput and blockstakeoutputs
func (w *Wallet) UnlockedUnspendOutputs() (map[types.CoinOutputID]types.CoinOutput, map[types.BlockStakeOutputID]types.BlockStakeOutput, error) {
	if err := w.tg.Add(); err != nil {
		return nil, nil, modules.ErrWalletShutdown
	}
	defer w.tg.Done()

	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		return nil, nil, modules.ErrLockedWallet
	}

	// ensure durability of reported balance
	err := w.syncDB()
	if err != nil {
		return nil, nil, err
	}

	ucom := make(map[types.CoinOutputID]types.CoinOutput)
	ubsom := make(map[types.BlockStakeOutputID]types.BlockStakeOutput)

	// prepare fulfillable context
	ctx := w.getFulfillableContextForLatestBlock()

	// get all (multisig) coin/blockStake unspend outputs
	err = dbForEachCoinOutput(w.dbTx, func(id types.CoinOutputID, co types.CoinOutput) {
		if co.Condition.Fulfillable(ctx) {
			ucom[id] = co
		}
	})
	if err != nil {
		return nil, nil, err
	}
	err = dbForEachBlockStakeOutput(w.dbTx, func(id types.BlockStakeOutputID, bso types.BlockStakeOutput) {
		if bso.Condition.Fulfillable(ctx) {
			ubsom[id] = bso
		}
	})
	if err != nil {
		return nil, nil, err
	}
	err = dbForEachMultisigCoinOutput(w.dbTx, func(id types.CoinOutputID, co types.CoinOutput) {
		if co.Condition.Fulfillable(ctx) {
			ucom[id] = co
		}
	})
	if err != nil {
		return nil, nil, err
	}
	err = dbForEachMultisigBlockStakeOutput(w.dbTx, func(id types.BlockStakeOutputID, bso types.BlockStakeOutput) {
		if bso.Condition.Fulfillable(ctx) {
			ubsom[id] = bso
		}
	})
	if err != nil {
		return nil, nil, err
	}
	return ucom, ubsom, nil
}

// LockedUnspendOutputs returnas all locked coinoutput and blockstakeoutputs
func (w *Wallet) LockedUnspendOutputs() (map[types.CoinOutputID]types.CoinOutput, map[types.BlockStakeOutputID]types.BlockStakeOutput, error) {
	if err := w.tg.Add(); err != nil {
		return nil, nil, modules.ErrWalletShutdown
	}
	defer w.tg.Done()

	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.unlocked {
		return nil, nil, modules.ErrLockedWallet
	}

	// ensure durability of reported balance
	err := w.syncDB()
	if err != nil {
		return nil, nil, err
	}

	ucom := make(map[types.CoinOutputID]types.CoinOutput)
	ubsom := make(map[types.BlockStakeOutputID]types.BlockStakeOutput)

	// prepare fulfillable context
	ctx := w.getFulfillableContextForLatestBlock()

	// get all (multisig) coin/blockStake unspend outputs
	err = dbForEachCoinOutput(w.dbTx, func(id types.CoinOutputID, co types.CoinOutput) {
		if !co.Condition.Fulfillable(ctx) {
			ucom[id] = co
		}
	})
	if err != nil {
		return nil, nil, err
	}
	err = dbForEachBlockStakeOutput(w.dbTx, func(id types.BlockStakeOutputID, bso types.BlockStakeOutput) {
		if !bso.Condition.Fulfillable(ctx) {
			ubsom[id] = bso
		}
	})
	if err != nil {
		return nil, nil, err
	}
	err = dbForEachMultisigCoinOutput(w.dbTx, func(id types.CoinOutputID, co types.CoinOutput) {
		if !co.Condition.Fulfillable(ctx) {
			ucom[id] = co
		}
	})
	if err != nil {
		return nil, nil, err
	}
	err = dbForEachMultisigBlockStakeOutput(w.dbTx, func(id types.BlockStakeOutputID, bso types.BlockStakeOutput) {
		if !bso.Condition.Fulfillable(ctx) {
			ubsom[id] = bso
		}
	})
	if err != nil {
		return nil, nil, err
	}
	return ucom, ubsom, nil
}
