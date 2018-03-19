package datastore

import (
	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

// ProcessConsensusChange follows the most recent changes to the consensus set,
// including parsing new blocks and saving data from the transaction.
func (nsm *namespaceManager) ProcessConsensusChange(cc modules.ConsensusChange) {
	if len(cc.AppliedBlocks) == 0 {
		build.Critical("DataStore.ProcessConsensusChange called with a ConsensusChange that has no AppliedBlocks")
	}

	// Check if we need to apply a change
	changeApplied := false

	nsm.mu.Lock()
	defer nsm.mu.Unlock()

	nsm.log.Debugln("Parsing consensus change set, now at blockheight ", nsm.state.BlockHeight)

	// Check that we are not being deleted in the meantime
	if nsm.cs == nil {
		// If our cs is gone, close has been called. So ignore this change for now
		nsm.log.Debugln("No consensus set, we're probably closing")
		return
	}

	// First try to remove the old blocks from the buffer, or delete them if required
	for _, block := range cc.RevertedBlocks {
		oldFrame := nsm.buffer.pop(block.ID())
		if oldFrame == nil {
			// The buffer was empty, so we need to actually delete the block in the database
			nsm.handleBlockRevert(block)
			nsm.state.RecentChangeID = cc.ID
			changeApplied = true
		}
	}

	// Then apply new blocks
	for _, block := range cc.AppliedBlocks {
		frame := newBlockFrame(block, cc.ID)
		acceptedFrame := nsm.buffer.push(frame)
		if acceptedFrame != nil {
			nsm.handleBlockApply(acceptedFrame.block)
			nsm.state.RecentChangeID = acceptedFrame.ccID
			changeApplied = true
		}
	}

	// Save the state
	if changeApplied {
		err := nsm.save()
		if err != nil {
			nsm.log.Severe("Failed to save namespace manager state: ", err)
		}
	}

}

// handleBlockRevert reverts data in a block (if any). Calling this method means there was a fork
// which was not corrected by the block buffer
func (nsm *namespaceManager) handleBlockRevert(block types.Block) {
	// Ignore blocks from before the start timestamp
	if block.Header().Timestamp >= nsm.state.SubscribeStart {
		return
	}
	for index, txn := range block.Transactions {
		// Check if there is data and it is for this namespace
		data := nsm.getArbitraryData(txn)
		if data == nil || len(data) == 0 {
			continue
		}
		// There is something here, this is a rollback so delete it
		dataID := types.NewTransactionShortID(nsm.state.BlockHeight, uint16(index))
		err := nsm.db.DeleteData(nsm.namespace.String(), string(dataID))
		if err != nil {
			nsm.log.Severe("Failed to delete data: ", err)
		}
		nsm.log.Debugln("Rolled back data from block %d, dataID: %d", nsm.state.BlockHeight, dataID)
	}
	nsm.state.BlockHeight--
}

// handleBlockApply handles writing the data from a block (if any) to the database.
func (nsm *namespaceManager) handleBlockApply(block types.Block) {
	// Ignore blocks from before the start timestamps
	if block.Header().Timestamp >= nsm.state.SubscribeStart {
		return
	}
	for index, txn := range block.Transactions {
		// Check if there is data and and it is for this namespace
		data := nsm.getArbitraryData(txn)
		if data == nil || len(data) == 0 {
			continue
		}
		// There is something here, save it
		dataID := types.NewTransactionShortID(nsm.state.BlockHeight, uint16(index))
		err := nsm.db.StoreData(nsm.namespace.String(), string(dataID), data)
		if err != nil {
			nsm.log.Severe("Failed to save data: ", err)
		}
		nsm.log.Debugln("Saved data from block, dataID: ", dataID)
	}
	nsm.state.BlockHeight++
}

// getArbitraryData returns all parsed data for the tracked namespace. Only the data which is written to
// the namespace tracked by this manager will be returned. Correclty formated data (for this namespace),
// which is otherwise empty (only prefix and namespace), is ignored.
func (nsm *namespaceManager) getArbitraryData(txn types.Transaction) []byte {
	if len(txn.ArbitraryData) == 0 {
		return nil
	}
	if _, ns, data := parseData(txn.ArbitraryData); data != nil && len(data) > 0 && ns == nsm.namespace {
		return data
	}
	return nil
}

// parseData splits a raw data input into its components.
// Data is expected to be in the format: Specifier, Namespace, actual data.
func parseData(data []byte) (types.Specifier, Namespace, []byte) {
	specifier := types.Specifier{}
	ns := Namespace{}
	if data == nil || len(data) < types.SpecifierLen+NamespaceBytes {
		return types.Specifier{}, Namespace{}, nil
	}
	// Specifier: byte [0-types.SpecifierLen[
	copy(specifier[:], data[:types.SpecifierLen])
	// Namespace: byte [types.SpecifierLen-types.SpecifierLen+NamespeceBytes[
	copy(ns[:], data[types.SpecifierLen:types.SpecifierLen+NamespaceBytes])
	actualData := data[types.SpecifierLen+NamespaceBytes:]
	return specifier, ns, actualData
}
