package explorergraphql

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/modules/explorergraphql/explorerdb"
	"github.com/threefoldtech/rivine/types"
)

// TODO: check how to do better error handling

// TODO: ensure we handle errors as gracefully as possible
//      (e.g. do not return errors when a warning is sufficient,
//            and do not stop the world for one property failure)

// TODO: support transaction pool data in case data was not found yet
//  >> this should probably be done as a Wrapped explorerdb.DB,
//     which subscribes to the transaction pool, and thus knows
//     at any point what data can be used in case the wrapped DB returns ErrNotFound

type Resolver struct {
	db             explorerdb.DB
	cs             modules.ConsensusSet
	chainConstants types.ChainConstants
	blockchainInfo types.BlockchainInfo
}

func (r *Resolver) BlockHeader() BlockHeaderResolver {
	return &blockHeaderResolver{r}
}
func (r *Resolver) ChainFacts() ChainFactsResolver {
	return &chainFactsResolver{r}
}
func (r *Resolver) QueryRoot() QueryRootResolver {
	return &queryRootResolver{r}
}
func (r *Resolver) UnlockHashCondition() UnlockHashConditionResolver {
	return &unlockHashConditionResolver{r}
}
func (r *Resolver) UnlockHashPublicKeyPair() UnlockHashPublicKeyPairResolver {
	return &unlockHashPublicKeyPairResolver{r}
}

type blockHeaderResolver struct{ *Resolver }

func (r *blockHeaderResolver) Child(ctx context.Context, obj *BlockHeader) (*Block, error) {
	if obj.BlockHeight == nil {
		return nil, errors.New("internal error: block height not defined for block header")
	}
	height := (*obj.BlockHeight) + 1
	block, err := getBlockAt(ctx, r.db, &height)
	if err != nil {
		if err == explorerdb.ErrNotFound {
			return nil, nil // this is acceptable, as it might be the latest block
		}
		return nil, err
	}
	return block, nil
}

type chainFactsResolver struct{ *Resolver }

func (r *chainFactsResolver) LastBlock(ctx context.Context, obj *ChainFacts) (*Block, error) {
	chainCtx, err := r.db.GetChainContext()
	if err != nil {
		return nil, err
	}
	return NewBlock(chainCtx.BlockID, r.db), nil
}
func (r *chainFactsResolver) Aggregated(ctx context.Context, obj *ChainFacts) (*ChainAggregatedData, error) {
	if obj.Aggregated != nil && obj.Aggregated.TotalCoins.Cmp(new(big.Int)) != 0 {
		return obj.Aggregated, nil // nothing to do anymore
	}
	dbChainAggregatedFacts, err := r.db.GetChainAggregatedFacts()
	if err != nil && err != explorerdb.ErrNotFound {
		return nil, fmt.Errorf("internal DB error while fetching chain aggregated facts: %v", err)
	}
	return dbChainAggregatedFactsAsGQL(&dbChainAggregatedFacts)
}

type queryRootResolver struct{ *Resolver }

func (r *queryRootResolver) Chain(ctx context.Context) (*ChainFacts, error) {
	constants, err := rivConstantsAsGQL(r.cs, &r.chainConstants, &r.blockchainInfo)
	if err != nil {
		return nil, fmt.Errorf("internal server error: failed to resolve chain constants: %v", err)
	}
	return &ChainFacts{
		Constants:  constants,
		LastBlock:  nil, // resolved by another lazy resolver
		Aggregated: nil, // resolved by another lazy resolver
	}, nil
}

func rivConstantsAsGQL(cs modules.ConsensusSet, chainConstants *types.ChainConstants, blockchainInfo *types.BlockchainInfo) (*ChainConstants, error) {
	coinPrecision := len(chainConstants.CurrencyUnits.OneCoin.String())
	if coinPrecision > 0 {
		coinPrecision--
	}
	var (
		blockCreatorFee       *BigInt
		minimumTransactionFee *BigInt
	)
	if !chainConstants.BlockCreatorFee.Equals64(0) {
		bi := dbCurrencyAsBigInt(chainConstants.BlockCreatorFee)
		blockCreatorFee = &bi
	}
	if !chainConstants.MinimumTransactionFee.Equals64(0) {
		bi := dbCurrencyAsBigInt(chainConstants.MinimumTransactionFee)
		minimumTransactionFee = &bi
	}
	txFeeBeneficiary, err := dbConditionAsUnlockCondition(chainConstants.TransactionFeeCondition)
	if err != nil {
		return nil, fmt.Errorf("failed to cast transaction fee breneficiary as GQL UnlockCondition: %v", err)
	}
	return &ChainConstants{
		Name:                              blockchainInfo.Name,
		NetworkName:                       blockchainInfo.NetworkName,
		CoinUnit:                          blockchainInfo.CoinUnit,
		CoinPecision:                      coinPrecision,
		ChainVersion:                      blockchainInfo.ChainVersion.String(),
		GatewayProtocolVersion:            blockchainInfo.ProtocolVersion.String(),
		DefaultTransactionVersion:         ByteVersion(chainConstants.DefaultTransactionVersion),
		ConsensusPlugins:                  cs.LoadedPlugins(),
		GenesisTimestamp:                  chainConstants.GenesisTimestamp,
		BlockSizeLimitInBytes:             int(chainConstants.BlockSizeLimit),
		AverageBlockCreationTimeInSeconds: int(chainConstants.BlockFrequency),
		GenesisTotalBlockStakes:           dbCurrencyAsBigInt(chainConstants.GenesisBlockStakeCount()),
		BlockStakeAging:                   int(chainConstants.BlockStakeAging),
		BlockCreatorFee:                   blockCreatorFee,
		MinimumTransactionFee:             minimumTransactionFee,
		TransactionFeeBeneficiary:         txFeeBeneficiary,
		PayoutMaturityDelay:               chainConstants.MaturityDelay,
	}, nil
}

func (r *queryRootResolver) Object(ctx context.Context, id *ObjectID) (Object, error) {
	if id == nil {
		// default to latest block if no ID is given, the only thing that makes sense
		return r.BlockAt(ctx, nil)
	}
	objID := explorerdb.ObjectID(*id)
	dbObjectInfo, err := r.db.GetObjectInfo(objID)
	if err != nil {
		return nil, fmt.Errorf("internal DB error while fetching object %x: %v", objID, err)
	}
	switch dbObjectInfo.Type {
	case explorerdb.ObjectTypeBlock:
		h, err := objID.AsHash()
		if err != nil {
			return nil, fmt.Errorf("internal server error: failed to convert confirmed object ID as block hash: %v", err)
		}
		return NewBlock(types.BlockID(h), r.db), nil
	case explorerdb.ObjectTypeTransaction:
		h, err := objID.AsHash()
		if err != nil {
			return nil, fmt.Errorf("internal server error: failed to convert confirmed object ID as transaction hash: %v", err)
		}
		return NewTransactionWithVersion(
			types.TransactionID(h), types.TransactionVersion(dbObjectInfo.Version),
			nil, r.db)
	case explorerdb.ObjectTypeOutput:
		h, err := objID.AsHash()
		if err != nil {
			return nil, fmt.Errorf("internal server error: failed to convert confirmed object ID as output hash: %v", err)
		}
		return NewOutput(types.OutputID(h), nil, nil, r.db), nil
	case explorerdb.ObjectTypeFreeForAllWallet:
		uh, err := objID.AsUnlockHash()
		if err != nil {
			return nil, fmt.Errorf("internal server error: failed to convert confirmed object ID as free-for-all wallet unlock hash: %v", err)
		}
		return NewFreeForAllWallet(uh, r.db), nil
	case explorerdb.ObjectTypeSingleSignatureWallet:
		uh, err := objID.AsUnlockHash()
		if err != nil {
			return nil, fmt.Errorf("internal server error: failed to convert confirmed object ID as single signature wallet unlock hash: %v", err)
		}
		return NewSingleSignatureWallet(uh, r.db), nil
	case explorerdb.ObjectTypeMultiSignatureWallet:
		uh, err := objID.AsUnlockHash()
		if err != nil {
			return nil, fmt.Errorf("internal server error: failed to convert confirmed object ID as multi signature wallet unlock hash: %v", err)
		}
		return NewSingleSignatureWallet(uh, r.db), nil
	case explorerdb.ObjectTypeAtomicSwapContract:
		uh, err := objID.AsUnlockHash()
		if err != nil {
			return nil, fmt.Errorf("internal server error: failed to convert confirmed object ID as multi signature wallet unlock hash: %v", err)
		}
		return NewAtomicSwapContract(uh, r.db), nil
	default:
		return nil, fmt.Errorf("internal server error: unsupported object type %d (object version %d)", dbObjectInfo.Type, dbObjectInfo.Version)
	}
}
func (r *queryRootResolver) Transaction(ctx context.Context, id crypto.Hash) (Transaction, error) {
	transactionID := types.TransactionID(id)
	return NewTransaction(transactionID, nil, r.db)
}
func (r *queryRootResolver) Output(ctx context.Context, id crypto.Hash) (*Output, error) {
	outputID := types.OutputID(id)
	return NewOutput(outputID, nil, nil, r.db), nil
}
func (r *queryRootResolver) Block(ctx context.Context, id *crypto.Hash) (*Block, error) {
	if id == nil {
		return getBlockAt(ctx, r.db, nil)
	}
	blockID := types.BlockID(*id)
	return NewBlock(blockID, r.db), nil
}
func (r *queryRootResolver) BlockAt(ctx context.Context, height *types.BlockHeight) (*Block, error) {
	return getBlockAt(ctx, r.db, height)
}
func getBlockAt(ctx context.Context, db explorerdb.DB, height *types.BlockHeight) (*Block, error) {
	if height == nil {
		// default to latest block
		chainCtx, err := db.GetChainContext()
		if err != nil {
			return nil, err
		}
		chahei := chainCtx.Height
		if chahei > 0 {
			chahei-- // chainCtx.Height defines the amount of blocks (in other words, the height of the chain), not the height of latest
		}
		height = &chahei
	}
	blockID, err := db.GetBlockIDAt(*height)
	if err != nil {
		return nil, err
	}
	return NewBlock(blockID, db), nil
}

func (r *queryRootResolver) Wallet(ctx context.Context, unlockhash types.UnlockHash) (Wallet, error) {
	switch unlockhash.Type {
	case types.UnlockTypePubKey:
		return NewSingleSignatureWallet(unlockhash, r.db), nil
	case types.UnlockTypeMultiSig:
		return NewMultiSignatureWallet(unlockhash, r.db), nil
	case types.UnlockTypeNil:
		return NewFreeForAllWallet(unlockhash, r.db), nil
	default:
		return nil, fmt.Errorf("unsupported wallet type %d", unlockhash.Type)
	}
}
func (r *queryRootResolver) Contract(ctx context.Context, unlockhash types.UnlockHash) (Contract, error) {
	return NewAtomicSwapContract(unlockhash, r.db), nil
}

type unlockHashConditionResolver struct{ *Resolver }

func (r *unlockHashConditionResolver) PublicKey(ctx context.Context, obj *UnlockHashCondition) (*types.PublicKey, error) {
	pk, err := r.db.GetPublicKey(obj.UnlockHash)
	if err != nil {
		if err == explorerdb.ErrNotFound {
			return nil, nil // no error
		}
		return nil, err
	}
	return &pk, nil
}

type unlockHashPublicKeyPairResolver struct{ *Resolver }

func (r *unlockHashPublicKeyPairResolver) PublicKey(ctx context.Context, obj *UnlockHashPublicKeyPair) (*types.PublicKey, error) {
	pk, err := r.db.GetPublicKey(obj.UnlockHash)
	if err != nil {
		if err == explorerdb.ErrNotFound {
			return nil, nil // no error
		}
		return nil, err
	}
	return &pk, nil
}
