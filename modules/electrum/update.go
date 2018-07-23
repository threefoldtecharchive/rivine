package electrum

import (
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

type (
	// Update is a collection of info processed from consensus set/
	// explorer/transactionpool updates
	Update struct {
		addressStates map[types.UnlockHash]string
	}
)

// ProcessConsensusChange parses updates from the cs to a general format,
// which is then passed on to all connections.
func (e *Electrum) ProcessConsensusChange(cc modules.ConsensusChange) {
	// TODO: Ideally this would be a subscription to the explorer module,
	// but that requires implementing subsriptions in that module first. For
	// now rely on the fact that the explorer module subsribes to the cs first,
	// so relevant info should already be processed there
	e.log.Debug("Processing consensus change", cc.ID)

	change := &Update{addressStates: make(map[types.UnlockHash]string)}

	for _, cod := range cc.CoinOutputDiffs {
		address := cod.CoinOutput.Condition.UnlockHash()
		// avoid computing status for an address multiple times
		if _, exists := change.addressStates[address]; !exists {
			change.addressStates[address] = e.AddressStatus(address)
		}
	}

	for _, bsod := range cc.BlockStakeOutputDiffs {
		address := bsod.BlockStakeOutput.Condition.UnlockHash()
		// avoid computing status for an address multiple times
		if _, exists := change.addressStates[address]; !exists {
			change.addressStates[address] = e.AddressStatus(address)
		}
	}

	for _, ch := range e.connections {
		ch <- change
	}
}
