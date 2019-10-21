package explorerdb

import (
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

type DB interface {
	SetChainContext(ChainContext) error
	GetChainContext() (ChainContext, error)

	ApplyBlock(block Block, txs []Transaction, outputs []Output, inputs map[types.OutputID]OutputSpenditureData) error
	RevertBlock(block types.BlockID, txs []types.TransactionID, outputs []types.OutputID, inputs []types.OutputID) error

	GetBlock(types.BlockID) (Block, error)
	GetTransaction(types.TransactionID) (Transaction, error)
	GetOutput(types.OutputID) (Output, error)

	GetWallet(types.UnlockHash) (WalletData, error)
	GetMultiSignatureWallet(types.UnlockHash) (MultiSignatureWalletData, error)
	GetAtomicSwapContract(types.UnlockHash) (AtomicSwapContract, error)

	Close() error
}

type ChainContext struct {
	ConsensusChangeID modules.ConsensusChangeID
	Height            types.BlockHeight
	Timestamp         types.Timestamp
	BlockID           types.BlockID
}

// TODO: handle also chain-specific stuff, such as chains that do not have block rewards

func ApplyConsensusChange(db DB, csc modules.ConsensusChange) error {
	chainCtx, err := db.GetChainContext()
	if err != nil {
		return err
	}

	for _, revertedBlock := range csc.RevertedBlocks {
		// TODO: verify if this is correct, or if it should be done after
		chainCtx.Height--
		chainCtx.Timestamp = revertedBlock.Timestamp

		block := RivineBlockAsExplorerBlock(revertedBlock)
		chainCtx.BlockID = block.ID

		outputs := make([]types.OutputID, 0, len(revertedBlock.MinerPayouts))
		for idx := range revertedBlock.MinerPayouts {
			outputs = append(outputs, block.Payouts[idx])
		}

		var inputs []types.OutputID
		transactions := make([]types.TransactionID, 0, len(revertedBlock.Transactions))
		for idx, txn := range revertedBlock.Transactions {
			// add txn
			transactions = append(transactions, block.Transactions[idx])
			// add inputs
			for _, input := range txn.CoinInputs {
				inputs = append(inputs, types.OutputID(input.ParentID))
			}
			for _, input := range txn.BlockStakeInputs {
				inputs = append(inputs, types.OutputID(input.ParentID))
			}
			// add outputs
			for cidx := range txn.CoinOutputs {
				outputs = append(outputs, types.OutputID(txn.CoinOutputID(uint64(cidx))))
			}
			for bsidx := range txn.BlockStakeOutputs {
				outputs = append(outputs, types.OutputID(txn.BlockStakeOutputID(uint64(bsidx))))
			}
		}

		err = db.RevertBlock(block.ID, transactions, outputs, inputs)
		if err != nil {
			return err
		}
	}

	for _, appliedBlock := range csc.AppliedBlocks {
		block := RivineBlockAsExplorerBlock(appliedBlock)

		outputs := make([]Output, 0, len(appliedBlock.MinerPayouts))
		for idx, mp := range appliedBlock.MinerPayouts {
			outputs = append(outputs, RivineMinerPayoutAsOutput(
				block.ID,
				types.CoinOutputID(block.Payouts[idx]),
				mp,
				// TODO: customize this per chain network (behaviour and constants)
				idx == 0,
				chainCtx.Height,
				720,
			))
		}
		// TODO: customize this per chain network
		feePayoutID := block.Payouts[0]
		if len(block.Payouts) > 1 {
			feePayoutID = block.Payouts[1]
		}

		inputs := make(map[types.OutputID]OutputSpenditureData)
		transactions := make([]Transaction, 0, len(appliedBlock.Transactions))
		for idx, txn := range appliedBlock.Transactions {
			transaction := RivineTransactionAsTransaction(
				block.ID,
				block.Transactions[idx],
				txn,
				feePayoutID,
			)
			transactions = append(transactions, transaction)
			// add inputs
			for _, input := range txn.CoinInputs {
				inputs[types.OutputID(input.ParentID)] = OutputSpenditureData{
					Fulfillment:              input.Fulfillment,
					FulfillmentTransactionID: block.Transactions[idx],
				}
			}
			for _, input := range txn.BlockStakeInputs {
				inputs[types.OutputID(input.ParentID)] = OutputSpenditureData{
					Fulfillment:              input.Fulfillment,
					FulfillmentTransactionID: block.Transactions[idx],
				}
			}
			// add outputs
			for idx, output := range txn.CoinOutputs {
				outputs = append(outputs, RivineCoinOutputAsOutput(
					block.Transactions[idx],
					types.CoinOutputID(transaction.Outputs[idx]),
					output,
				))
			}
			outputOffset := len(txn.CoinOutputs)
			for idx, output := range txn.BlockStakeOutputs {
				outputs = append(outputs, RivineBlockStakeOutputAsOutput(
					block.Transactions[idx],
					types.BlockStakeOutputID(transaction.Outputs[outputOffset+idx]),
					output,
				))
			}
		}

		err = db.ApplyBlock(block, transactions, outputs, inputs)
		if err != nil {
			return err
		}

		// TODO: verify if this is correct, or if it should be done before
		chainCtx.Height++
		chainCtx.Timestamp = appliedBlock.Timestamp
		chainCtx.BlockID = block.ID
	}

	err = db.SetChainContext(chainCtx)
	return err
}

func RivineBlockAsExplorerBlock(block types.Block) Block {
	// aggregate payouts (as a list of identifiers)
	payouts := make([]types.OutputID, 0, len(block.MinerPayouts))
	for idx := range block.MinerPayouts {
		payouts = append(payouts, types.OutputID(block.MinerPayoutID(uint64(idx))))
	}
	// aggregate transactions (as a list of identifiers)
	transactions := make([]types.TransactionID, 0, len(block.Transactions))
	for _, txn := range block.Transactions {
		transactions = append(transactions, txn.ID())
	}
	// return the block
	return Block{
		ID:           block.ID(),
		Payouts:      payouts,
		Transactions: transactions,
	}
}

func RivineTransactionAsTransaction(parent types.BlockID, id types.TransactionID, rtxn types.Transaction, feePayoutID types.OutputID) Transaction {
	// aggregate inputs (as a list of identifiers)
	inputs := make([]types.OutputID, 0, len(rtxn.CoinInputs)+len(rtxn.BlockStakeInputs))
	for _, input := range rtxn.CoinInputs {
		inputs = append(inputs, types.OutputID(input.ParentID))
	}
	for _, input := range rtxn.BlockStakeInputs {
		inputs = append(inputs, types.OutputID(input.ParentID))
	}
	// aggregate outputs (as a list of identifiers)
	outputs := make([]types.OutputID, 0, len(rtxn.CoinOutputs)+len(rtxn.BlockStakeOutputs))
	for idx := range rtxn.CoinOutputs {
		outputs = append(inputs, types.OutputID(rtxn.CoinOutputID(uint64(idx))))
	}
	for idx := range rtxn.BlockStakeOutputs {
		outputs = append(inputs, types.OutputID(rtxn.BlockStakeOutputID(uint64(idx))))
	}
	// return transaction
	return Transaction{
		ID: id,

		ParentBlock: parent,
		Version:     rtxn.Version,

		Inputs:  inputs,
		Outputs: outputs,
		FeePayout: TransactionFeePayoutInfo{
			PayoutID: feePayoutID,
			Values:   rtxn.MinerFees,
		},

		ExtensionData: rtxn.Extension,
	}
}

func RivineMinerPayoutAsOutput(parent types.BlockID, id types.CoinOutputID, payout types.MinerPayout, reward bool, height types.BlockHeight, delay types.BlockHeight) Output {
	// define output type
	var ot OutputType
	if reward {
		ot = OutputTypeBlockCreationReward
	} else {
		ot = OutputTypeTransactionFee
	}

	// return output
	return Output{
		ID:       types.OutputID(id),
		ParentID: crypto.Hash(parent),
		Type:     ot,
		Value:    payout.Value,
		Condition: types.NewCondition(types.NewTimeLockCondition(
			uint64(height+delay),
			types.NewUnlockHashCondition(payout.UnlockHash))),
		SpenditureData: nil,
	}
}

func RivineCoinOutputAsOutput(parent types.TransactionID, id types.CoinOutputID, output types.CoinOutput) Output {
	return Output{
		ID:        types.OutputID(id),
		ParentID:  crypto.Hash(parent),
		Type:      OutputTypeCoin,
		Value:     output.Value,
		Condition: output.Condition,
	}
}

func RivineBlockStakeOutputAsOutput(parent types.TransactionID, id types.BlockStakeOutputID, output types.BlockStakeOutput) Output {
	return Output{
		ID:        types.OutputID(id),
		ParentID:  crypto.Hash(parent),
		Type:      OutputTypeBlockStake,
		Value:     output.Value,
		Condition: output.Condition,
	}
}
