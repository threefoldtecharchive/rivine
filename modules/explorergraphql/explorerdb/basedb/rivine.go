package basedb

import (
	"fmt"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/types"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
)

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
	// return the bloc
	return Block{
		ID:           block.ID(),
		ParentID:     block.ParentID,
		Height:       height,
		Timestamp:    block.Timestamp,
		Payouts:      payouts,
		Transactions: transactions,
	}
}

func RivineTransactionAsTransaction(parent types.BlockID, id types.TransactionID, rtxn types.Transaction, feePayoutID types.OutputID) Transaction {
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

	condition := types.NewCondition(types.NewTimeLockCondition(
		uint64(height+delay),
		types.NewUnlockHashCondition(payout.UnlockHash)))
	unlockReferencePoint, _ := UnlockReferencePointFromCondition(condition)

	// return output
	return Output{
		ID:                   types.OutputID(id),
		ParentID:             crypto.Hash(parent),
		Type:                 ot,
		Value:                payout.Value,
		Condition:            condition,
		UnlockReferencePoint: unlockReferencePoint,
		SpenditureData:       nil,
	}
}

func RivineCoinOutputAsOutput(parent types.TransactionID, id types.CoinOutputID, output types.CoinOutput) Output {
	unlockReferencePoint, _ := UnlockReferencePointFromCondition(output.Condition)
	return Output{
		ID:                   types.OutputID(id),
		ParentID:             crypto.Hash(parent),
		Type:                 OutputTypeCoin,
		Value:                output.Value,
		Condition:            output.Condition,
		UnlockReferencePoint: unlockReferencePoint,
	}
}

func RivineBlockStakeOutputAsOutput(parent types.TransactionID, id types.BlockStakeOutputID, output types.BlockStakeOutput) Output {
	unlockReferencePoint, _ := UnlockReferencePointFromCondition(output.Condition)
	return Output{
		ID:                   types.OutputID(id),
		ParentID:             crypto.Hash(parent),
		Type:                 OutputTypeBlockStake,
		Value:                output.Value,
		Condition:            output.Condition,
		UnlockReferencePoint: unlockReferencePoint,
	}
}
