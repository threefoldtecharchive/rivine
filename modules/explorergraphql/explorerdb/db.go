package explorerdb

import (
	"fmt"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/types"

	"github.com/threefoldtech/rivine/extensions/minting"
)

// TODO: integrate context.Context in each call

type DB interface {
	SetChainContext(ChainContext) error
	GetChainContext() (ChainContext, error)

	ApplyBlock(block Block, txs []Transaction, outputs []Output, inputs map[types.OutputID]OutputSpenditureData, publicKeys map[types.UnlockHash]types.PublicKey) error
	// TODO: should we also revert public key (from UH) mapping?
	RevertBlock(blockContext BlockRevertContext, txs []types.TransactionID, outputs []types.OutputID, inputs []types.OutputID) error

	GetBlock(types.BlockID) (Block, error)
	GetBlockByReferencePoint(ReferencePoint) (Block, error)
	GetTransaction(types.TransactionID) (Transaction, error)
	GetOutput(types.OutputID) (Output, error)

	GetWallet(types.UnlockHash) (WalletData, error)
	GetMultiSignatureWallet(types.UnlockHash) (MultiSignatureWalletData, error)
	GetAtomicSwapContract(types.UnlockHash) (AtomicSwapContract, error)

	GetPublicKey(types.UnlockHash) (types.PublicKey, error)

	Close() error
}

type ChainContext struct {
	ConsensusChangeID modules.ConsensusChangeID
	Height            types.BlockHeight
	Timestamp         types.Timestamp
	BlockID           types.BlockID
}

type BlockRevertContext struct {
	ID        types.BlockID
	Height    types.BlockHeight
	Timestamp types.Timestamp
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

		block := RivineBlockAsExplorerBlock(chainCtx.Height, revertedBlock)
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

		err = db.RevertBlock(BlockRevertContext{
			ID:        block.ID,
			Height:    chainCtx.Height,
			Timestamp: chainCtx.Timestamp,
		}, transactions, outputs, inputs)
		if err != nil {
			return err
		}
	}

	for _, appliedBlock := range csc.AppliedBlocks {
		block := RivineBlockAsExplorerBlock(chainCtx.Height, appliedBlock)
		publicKeys := make(map[types.UnlockHash]types.PublicKey)

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
		var feePayoutID types.OutputID
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
				publicKeys,
			)
			transactions = append(transactions, transaction)
			// add inputs
			for _, input := range txn.CoinInputs {
				inputs[types.OutputID(input.ParentID)] = OutputSpenditureData{
					Fulfillment:              input.Fulfillment,
					FulfillmentTransactionID: block.Transactions[idx],
				}
				pairs := RivineUnlockHashPublicKeyPairsFromFulfillment(input.Fulfillment)
				for _, pair := range pairs {
					publicKeys[pair.UnlockHash] = pair.PublicKey
				}
			}
			for _, input := range txn.BlockStakeInputs {
				inputs[types.OutputID(input.ParentID)] = OutputSpenditureData{
					Fulfillment:              input.Fulfillment,
					FulfillmentTransactionID: block.Transactions[idx],
				}
				pairs := RivineUnlockHashPublicKeyPairsFromFulfillment(input.Fulfillment)
				for _, pair := range pairs {
					publicKeys[pair.UnlockHash] = pair.PublicKey
				}
			}
			// add outputs
			for idx, output := range txn.CoinOutputs {
				outputs = append(outputs, RivineCoinOutputAsOutput(
					block.Transactions[idx],
					types.CoinOutputID(transaction.CoinOutputs[idx]),
					output,
				))
			}
			for idx, output := range txn.BlockStakeOutputs {
				outputs = append(outputs, RivineBlockStakeOutputAsOutput(
					block.Transactions[idx],
					types.BlockStakeOutputID(transaction.BlockStakeOutputs[idx]),
					output,
				))
			}
		}

		err = db.ApplyBlock(block, transactions, outputs, inputs, publicKeys)
		if err != nil {
			return err
		}

		// TODO: verify if this is correct, or if it should be done before
		chainCtx.Height++
		chainCtx.Timestamp = appliedBlock.Timestamp
		chainCtx.BlockID = block.ID
	}

	chainCtx.ConsensusChangeID = csc.ID
	err = db.SetChainContext(chainCtx)
	return err
}

func RivineBlockAsExplorerBlock(height types.BlockHeight, block types.Block) Block {
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
		ParentID:     block.ParentID,
		Height:       height,
		Timestamp:    block.Timestamp,
		Payouts:      payouts,
		Transactions: transactions,
	}
}

func RivineTransactionAsTransaction(parent types.BlockID, id types.TransactionID, rtxn types.Transaction, feePayoutID types.OutputID, publicKeys map[types.UnlockHash]types.PublicKey) Transaction {
	// aggregate inputs (as a list of identifiers)
	coinInputs := make([]types.OutputID, 0, len(rtxn.CoinInputs))
	for _, input := range rtxn.CoinInputs {
		coinInputs = append(coinInputs, types.OutputID(input.ParentID))
	}
	blockStakeInputs := make([]types.OutputID, 0, len(rtxn.BlockStakeInputs))
	for _, input := range rtxn.BlockStakeInputs {
		blockStakeInputs = append(blockStakeInputs, types.OutputID(input.ParentID))
	}
	// aggregate outputs (as a list of identifiers)
	coinOutputs := make([]types.OutputID, 0, len(rtxn.CoinOutputs))
	for idx := range rtxn.CoinOutputs {
		coinOutputs = append(coinOutputs, types.OutputID(rtxn.CoinOutputID(uint64(idx))))
	}
	blockStakeOutputs := make([]types.OutputID, 0, len(rtxn.BlockStakeOutputs))
	for idx := range rtxn.BlockStakeOutputs {
		blockStakeOutputs = append(blockStakeOutputs, types.OutputID(rtxn.BlockStakeOutputID(uint64(idx))))
	}
	// encode extension data
	var encodedExtensionData []byte
	if rtxn.Extension != nil {
		var err error
		encodedExtensionData, err = rivbin.Marshal(rtxn.Extension)
		if err != nil {
			build.Severe("failed to encode rivine txn extension data", err)
		}
	}
	// TODO: handle this in a transactionControllerSupportedManner
	if rtxn.Version == 128 {
		data := rtxn.Extension.(*minting.MinterDefinitionTransactionExtension)
		pairs := RivineUnlockHashPublicKeyPairsFromFulfillment(data.MintFulfillment)
		for _, pair := range pairs {
			publicKeys[pair.UnlockHash] = pair.PublicKey
		}
	} else if rtxn.Version == 129 {
		data := rtxn.Extension.(*minting.CoinCreationTransactionExtension)
		pairs := RivineUnlockHashPublicKeyPairsFromFulfillment(data.MintFulfillment)
		for _, pair := range pairs {
			publicKeys[pair.UnlockHash] = pair.PublicKey
		}
	}
	// return transaction
	return Transaction{
		ID: id,

		ParentBlock: parent,
		Version:     rtxn.Version,

		CoinInputs:  coinInputs,
		CoinOutputs: coinOutputs,

		BlockStakeInputs:  blockStakeInputs,
		BlockStakeOutputs: blockStakeOutputs,

		FeePayout: TransactionFeePayoutInfo{
			PayoutID: feePayoutID,
			Values:   rtxn.MinerFees,
		},

		ArbitraryData:        rtxn.ArbitraryData,
		EncodedExtensionData: encodedExtensionData,
	}
}

type UnlockHashPublicKeyPair struct {
	UnlockHash types.UnlockHash
	PublicKey  types.PublicKey
}

func RivineUnlockHashPublicKeyPairsFromFulfillment(fulfillment types.UnlockFulfillmentProxy) []UnlockHashPublicKeyPair {
	switch ft := fulfillment.FulfillmentType(); ft {
	case types.FulfillmentTypeSingleSignature:
		ssft := fulfillment.Fulfillment.(*types.SingleSignatureFulfillment)
		return []UnlockHashPublicKeyPair{
			{
				UnlockHash: RivineUnlockHashFromPublicKey(ssft.PublicKey),
				PublicKey:  ssft.PublicKey,
			},
		}
	case types.FulfillmentTypeMultiSignature:
		msft := fulfillment.Fulfillment.(*types.MultiSignatureFulfillment)
		var pairs []UnlockHashPublicKeyPair
		for _, pair := range msft.Pairs {
			pairs = append(pairs, UnlockHashPublicKeyPair{
				UnlockHash: RivineUnlockHashFromPublicKey(pair.PublicKey),
				PublicKey:  pair.PublicKey,
			})
		}
		return pairs
	default:
		build.Critical(fmt.Sprintf("unsupported fulfillment type %d: %v", ft, fulfillment))
	}
	// should never reach here
	return nil
}

func RivineUnlockHashFromPublicKey(pk types.PublicKey) types.UnlockHash {
	uh, err := types.NewPubKeyUnlockHash(pk)
	if err != nil {
		build.Severe("failed to convert unlock hash to public key", pk, err)
	}
	return uh
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
