package explorergraphql

import (
	"fmt"
	"math/big"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules/explorergraphql/explorerdb"
	"github.com/threefoldtech/rivine/types"
)

func dbChainAggregatedFactsAsGQL(facts *explorerdb.ChainAggregatedFacts) (*ChainAggregatedData, error) {
	return &ChainAggregatedData{
		TotalCoins:                 dbCurrencyAsBigIntRef(facts.TotalCoins),
		TotalLockedCoins:           dbCurrencyAsBigIntRef(facts.TotalLockedCoins),
		TotalBlockStakes:           dbCurrencyAsBigIntRef(facts.TotalBlockStakes),
		TotalLockedBlockStakes:     dbCurrencyAsBigIntRef(facts.TotalLockedBlockStakes),
		EstimatedActiveBlockStakes: dbCurrencyAsBigIntRef(facts.EstimatedActiveBlockStakes),
	}, nil
}

func dbBalanceAsGQL(dbBalance *explorerdb.Balance) *Balance {
	if dbBalance.Unlocked.IsZero() && dbBalance.Locked.IsZero() {
		return nil
	}
	return &Balance{
		Unlocked: dbCurrencyAsBigInt(dbBalance.Unlocked),
		Locked:   dbCurrencyAsBigInt(dbBalance.Locked),
	}
}

func dbOutputTypeAsGQL(outputType explorerdb.OutputType) *OutputType {
	switch outputType {
	case explorerdb.OutputTypeCoin:
		ot := OutputTypeCoin
		return &ot
	case explorerdb.OutputTypeBlockStake:
		ot := OutputTypeBlockStake
		return &ot
	case explorerdb.OutputTypeBlockCreationReward:
		ot := OutputTypeBlockCreationReward
		return &ot
	case explorerdb.OutputTypeTransactionFee:
		ot := OutputTypeTransactionFee
		return &ot
	default:
		return nil
	}
}

func dbFulfillmentAsUnlockFulfillment(fulfillment types.UnlockFulfillmentProxy, parentCondition UnlockCondition) (UnlockFulfillment, error) {
	switch ft := fulfillment.FulfillmentType(); ft {
	case types.FulfillmentTypeSingleSignature:
		sft := fulfillment.Fulfillment.(*types.SingleSignatureFulfillment)
		return &SingleSignatureFulfillment{
			Version:         ByteVersion(ft),
			ParentCondition: parentCondition,
			PublicKey:       sft.PublicKey,
			Signature:       Signature(sft.Signature[:]),
		}, nil
	case types.FulfillmentTypeMultiSignature:
		msft := fulfillment.Fulfillment.(*types.MultiSignatureFulfillment)
		pairs := make([]*PublicKeySignaturePair, 0, len(msft.Pairs))
		for _, pair := range msft.Pairs {
			pairs = append(pairs, &PublicKeySignaturePair{
				PublicKey: pair.PublicKey,
				Signature: Signature(pair.Signature[:]),
			})
		}
		return &MultiSignatureFulfillment{
			Version:         ByteVersion(ft),
			ParentCondition: parentCondition,
			Pairs:           pairs,
		}, nil
	case types.FulfillmentTypeAtomicSwap:
		asft := fulfillment.Fulfillment.(*types.AtomicSwapFulfillment)
		return dbAtomicSwapFulfillmentAsGQL(asft, parentCondition), nil
	default:
		return nil, fmt.Errorf("unsupported fulfillment type %d: %v", ft, fulfillment)
	}
}

func dbAtomicSwapFulfillmentAsGQL(fulfillment *types.AtomicSwapFulfillment, parentCondition UnlockCondition) *AtomicSwapFulfillment {
	asfGQL := &AtomicSwapFulfillment{
		Version:         ByteVersion(fulfillment.FulfillmentType()),
		ParentCondition: parentCondition,
		PublicKey:       fulfillment.PublicKey,
		Signature:       Signature(fulfillment.Signature[:]),
	}
	if fulfillment.Secret != (types.AtomicSwapSecret{}) {
		asfGQL.Secret = dbByteSliceAsBinaryData(fulfillment.Secret[:])
	}
	return asfGQL
}

func dbConditionAsUnlockCondition(condition types.UnlockConditionProxy) (UnlockCondition, error) {
	switch ct := condition.ConditionType(); ct {
	case types.ConditionTypeNil:
		return dbNilConditionAsGQL(condition.Condition.(*types.NilCondition)), nil
	case types.ConditionTypeUnlockHash:
		return dbUnlockHashConditionAsGQL(condition.Condition), nil
	case types.ConditionTypeAtomicSwap:
		asc := condition.Condition.(*types.AtomicSwapCondition)
		return dbAtomicSwapConditionAsGQL(asc), nil
	case types.ConditionTypeTimeLock:
		tlc := condition.Condition.(*types.TimeLockCondition)
		lt := LockTypeTimestamp
		if tlc.LockTime < types.LockTimeMinTimestampValue {
			lt = LockTypeBlockHeight
		}
		uh := condition.UnlockHash()
		ltc := &LockTimeCondition{
			Version:    ByteVersion(ct),
			UnlockHash: &uh,
			LockValue:  LockTime(tlc.LockTime),
			LockType:   lt,
		}
		switch ict := tlc.Condition.ConditionType(); ict {
		case types.ConditionTypeNil:
			ltc.Condition = dbNilConditionAsGQL(tlc.Condition.(*types.NilCondition))
		case types.ConditionTypeUnlockHash:
			ltc.Condition = dbUnlockHashConditionAsGQL(tlc.Condition)
		case types.ConditionTypeMultiSignature:
			ltc.Condition = dbMultiSignatureConditionAsGQL(tlc.Condition.(types.MultiSignatureConditionOwnerInfoGetter))
		default:
			return nil, fmt.Errorf("unsupported inner LockTime condition type %d: %v", ict, tlc.Condition)
		}
		return ltc, nil
	case types.ConditionTypeMultiSignature:
		return dbMultiSignatureConditionAsGQL(condition.Condition.(types.MultiSignatureConditionOwnerInfoGetter)), nil
	default:
		return nil, fmt.Errorf("unsupported condition type %d: %v", ct, condition)
	}
}

func dbNilConditionAsGQL(condition *types.NilCondition) *NilCondition {
	return &NilCondition{
		Version:    ByteVersion(condition.ConditionType()),
		UnlockHash: types.NilUnlockHash,
	}
}

func dbAtomicSwapConditionAsGQL(condition *types.AtomicSwapCondition) *AtomicSwapCondition {
	return &AtomicSwapCondition{
		Version:      ByteVersion(condition.ConditionType()),
		UnlockHash:   condition.UnlockHash(),
		Sender:       dbUnlockHashAsUnlockHashPublicKeyPair(condition.Sender),
		Receiver:     dbUnlockHashAsUnlockHashPublicKeyPair(condition.Receiver),
		HashedSecret: *dbByteSliceAsBinaryData(condition.HashedSecret[:]),
		TimeLock:     LockTime(condition.TimeLock),
	}
}

func dbUnlockHashConditionAsGQL(condition types.MarshalableUnlockCondition) *UnlockHashCondition {
	return &UnlockHashCondition{
		Version:    ByteVersion(condition.ConditionType()),
		UnlockHash: condition.UnlockHash(),
	}
}

func dbMultiSignatureConditionAsGQL(condition types.MultiSignatureConditionOwnerInfoGetter) *MultiSignatureCondition {
	uhSlice := condition.UnlockHashSlice()
	owners := make([]*UnlockHashPublicKeyPair, 0, len(uhSlice))
	for i := range uhSlice {
		owners = append(owners, dbUnlockHashAsUnlockHashPublicKeyPair(uhSlice[i]))
	}
	return &MultiSignatureCondition{
		Version:                ByteVersion(condition.ConditionType()),
		UnlockHash:             condition.UnlockHash(),
		Owners:                 owners,
		RequiredSignatureCount: int(condition.GetMinimumSignatureCount()),
	}
}

func dbUnlockHashAsUnlockHashPublicKeyPair(uh types.UnlockHash) *UnlockHashPublicKeyPair {
	return &UnlockHashPublicKeyPair{
		UnlockHash: uh,
		PublicKey:  nil, // field is populated by another lazy resolver
	}
}

func dbTxFeePayoutInfoAsGQL(fpInfo *explorerdb.TransactionFeePayoutInfo, db explorerdb.DB) []*TransactionFeePayout {
	payouts := make([]*TransactionFeePayout, 0, len(fpInfo.Values))
	blockPayout := NewBlockPayout(fpInfo.PayoutID, nil, db)
	for _, v := range fpInfo.Values {
		payouts = append(payouts, &TransactionFeePayout{
			BlockPayout: blockPayout,
			Value:       dbCurrencyAsBigInt(v),
		})
	}
	return payouts
}

func dbOutputTypeAsBlockPayoutType(outputType explorerdb.OutputType) *BlockPayoutType {
	switch outputType {
	case explorerdb.OutputTypeBlockCreationReward:
		bpt := BlockPayoutTypeBlockReward
		return &bpt
	case explorerdb.OutputTypeTransactionFee:
		bpt := BlockPayoutTypeTransactionFee
		return &bpt
	default:
		return nil // == unknown
	}
}

func dbOutputIDAsHash(outputID types.OutputID) *crypto.Hash {
	h := crypto.Hash(outputID)
	return &h
}

func dbBlockIDAsHash(blockID types.BlockID) *crypto.Hash {
	h := crypto.Hash(blockID)
	return &h
}

func dbTargetAsHash(target types.Target) *crypto.Hash {
	h := crypto.Hash(target)
	return &h
}

func dbBigIntAsGQLRef(bi *big.Int) *BigInt {
	return &BigInt{
		Int: bi,
	}
}

func dbCurrencyAsBigIntRef(c types.Currency) *BigInt {
	return &BigInt{
		Int: c.Big(),
	}
}

func dbCurrencyAsBigInt(c types.Currency) BigInt {
	return BigInt{
		Int: c.Big(),
	}
}

func dbByteSliceAsBinaryData(b []byte) *BinaryData {
	bd := BinaryData(b)
	return &bd
}
