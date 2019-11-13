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

// TODO: differentiate between user and internal DB errors!!!

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
	height := int((*obj.BlockHeight) + 1)
	block, err := getBlockAt(ctx, r.db, &height)
	if err != nil {
		if err == explorerdb.ErrNotFound {
			return nil, nil // this is acceptable, as it might be the latest block
		}
		return nil, err
	}
	return block, nil
}

func (r *blockHeaderResolver) Payouts(ctx context.Context, obj *BlockHeader, typeArg *BlockPayoutType) ([]*BlockPayout, error) {
	// NOTE: this resolver is not lazy,
	// and is a resolver that does not save any DB calls at the moment.
	// It merely prevents from data being returned that the user did not request,
	// yet the work is fully done on our side none the less.
	// If we could prevent the unused server-side work that might be even better,
	// but this API-level filtering is a start at least...
	if typeArg == nil {
		return obj.Payouts, nil
	}
	var payouts []*BlockPayout
	for _, payout := range obj.Payouts {
		pt, err := payout.Type(ctx)
		if err != nil {
			return nil, err
		}
		if pt != nil && *pt == *typeArg {
			payouts = append(payouts, payout)
		}
	}
	return payouts, nil
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
func (r *queryRootResolver) BlockAt(ctx context.Context, position *int) (*Block, error) {
	return getBlockAt(ctx, r.db, position)
}
func getBlockAt(ctx context.Context, db explorerdb.DB, position *int) (*Block, error) {
	if position == nil || (*position) < 0 {
		// default to latest block
		chainCtx, err := db.GetChainContext()
		if err != nil {
			return nil, err
		}
		chahei := int(chainCtx.Height)
		if chahei > 0 {
			chahei-- // chainCtx.Height defines the amount of blocks (in other words, the height of the chain), not the height of latest
		}
		if position == nil {
			position = &chahei
		} else {
			*position += chahei
			if *position < 0 {
				return nil, nil // block not found
			}

		}
	}
	height := types.BlockHeight(*position)
	blockID, err := db.GetBlockIDAt(height)
	if err != nil {
		return nil, err
	}
	return NewBlock(blockID, db), nil
}

// TODO: [FATAL] fix cursor: seems to be not supper correct in paging (seems to be something wrong with our height check)
//       need to slow debug on paper what i'm doing with these heights, probably a stupid mistake
func (r *queryRootResolver) Blocks(ctx context.Context, filter *BlocksFilter) (*ResponseBlocks, error) {
	var (
		err            error
		blocksFilter   *explorerdb.BlocksFilter
		bhFilter       *explorerdb.BlockHeightFilterRange
		tsFilter       *explorerdb.TimestampFilterRange
		txLengthFilter *explorerdb.IntFilterRange
		limit          *int
		cursor         *explorerdb.Cursor
	)
	if filter != nil {
		var possibleToFetch bool
		bhFilter, possibleToFetch, err = blockHeightsFilterFromGQL(filter.Height, r.db)
		if err != nil {
			return nil, err
		}
		if !possibleToFetch {
			return new(ResponseBlocks), nil // return early, as there is nothing to fetch
		}
		tsFilter, err = timestampFilterFromGQL(filter.Timestamp)
		if err != nil {
			return nil, err
		}
		txLengthFilter, err = intFilterFromGQL(filter.TransactionLength)
		if err != nil {
			return nil, err
		}
		if bhFilter != nil || tsFilter != nil || txLengthFilter != nil {
			blocksFilter = &explorerdb.BlocksFilter{
				BlockHeight:       bhFilter,
				Timestamp:         tsFilter,
				TransactionLength: txLengthFilter,
			}
		}
		limit = filter.Limit
		cursor = filter.Cursor
	}

	blocks, nextCursor, err := r.db.GetBlocks(limit, blocksFilter, cursor)
	if err != nil {
		return nil, err
	}
	apiBlocks := make([]*Block, 0, len(blocks))
	for idx := range blocks {
		apiBlock, err := NewBlockFromDB(&blocks[idx], r.db)
		if err != nil {
			return nil, err
		}
		apiBlocks = append(apiBlocks, apiBlock)
	}
	return &ResponseBlocks{
		Blocks:     apiBlocks,
		NextCursor: nextCursor,
	}, nil
}

func blockHeightsFilterFromGQL(operators *BlockPositionOperators, db explorerdb.DB) (*explorerdb.BlockHeightFilterRange, bool, error) {
	if operators == nil {
		return nil, true, nil
	}
	if operators.Before != nil {
		if operators.After != nil || operators.Between != nil {
			return nil, false, errors.New("After and Between BlockHeigt filters cannot be defined if Before filter is defined")
		}
		position := *operators.Before
		if position < 0 {
			chainCtx, err := db.GetChainContext()
			if err != nil {
				return nil, false, err
			}
			position += int(chainCtx.Height)
			if position < 0 {
				return nil, false, errors.New("out of range before block height position given")
			}
		}
		if position == 0 {
			return nil, false, nil // no blocks to fetch
		}
		endHeight := types.BlockHeight(position - 1) // "before" filter is exclusive
		return explorerdb.NewBlockHeightFilterRange(nil, &endHeight), true, nil
	}
	if operators.After != nil {
		// we know before is nil due to previous check
		if operators.Between != nil {
			return nil, false, errors.New("Between BlockHeigt filters cannot be defined if Before or After filter is defined")
		}
		position := *operators.After
		if position < 0 {
			chainCtx, err := db.GetChainContext()
			if err != nil {
				return nil, false, err
			}
			position += int(chainCtx.Height)
			if position < 0 {
				return nil, false, errors.New("out of range before block height position given")
			}
		}
		beginHeight := types.BlockHeight(position + 1) // "after" filter is exclusive
		return explorerdb.NewBlockHeightFilterRange(&beginHeight, nil), true, nil
	}
	if operators.Between != nil && (operators.Between.Start != nil || operators.Between.End != nil) {
		// we know the other 2 are nil due to the previous checks
		begin, end := operators.Between.Start, operators.Between.End
		var (
			chainCtx     explorerdb.ChainContext
			hBegin, hEnd *types.BlockHeight
		)
		if begin != nil {
			if *begin < 0 {
				var err error
				chainCtx, err = db.GetChainContext()
				if err != nil {
					return nil, false, err
				}
				*begin += int(chainCtx.Height)
				if *begin < 0 {
					return nil, false, errors.New("out of range before block height begin position given")
				}
			}
			height := types.BlockHeight(*begin)
			hBegin = &height
		}
		if end != nil {
			if *end < 0 {
				if chainCtx.Timestamp == 0 { // if chain ctx is not fetched yet
					var err error
					chainCtx, err = db.GetChainContext()
					if err != nil {
						return nil, false, err
					}
				}
				*end += int(chainCtx.Height)
				if *end < 0 {
					return nil, false, errors.New("out of range before block height end position given")
				}
			}
			height := types.BlockHeight(*end)
			hEnd = &height
		}
		return explorerdb.NewBlockHeightFilterRange(hBegin, hEnd), true, nil
	}
	// use no filter, useless, but same as a nil filter explicitly
	return nil, false, nil
}

func timestampFilterFromGQL(operators *TimestampOperators) (*explorerdb.TimestampFilterRange, error) {
	if operators == nil {
		return nil, nil
	}
	if operators.Before != nil {
		if operators.After != nil || operators.Between != nil {
			return nil, errors.New("After and Between Timestamp filters cannot be defined if Before filter is defined")
		}
		ts := *operators.Before
		if ts > 0 {
			ts-- // "before" filter is exclusive
		}
		return explorerdb.NewTimestampFilterRange(nil, &ts), nil
	}
	if operators.After != nil {
		// we know before is nil due to previous check
		if operators.Between != nil {
			return nil, errors.New("Between Timestamp filters cannot be defined if Before or After filter is defined")
		}
		ts := *operators.After + 1 // "after" filter is exclusive
		return explorerdb.NewTimestampFilterRange(&ts, nil), nil
	}
	if operators.Between != nil && (operators.Between.Start != nil || operators.Between.End != nil) {
		// we know the other 2 are nil due to the previous checks
		return explorerdb.NewTimestampFilterRange(
			operators.Between.Start,
			operators.Between.End,
		), nil
	}
	// use no filter, useless, but same as a nil filter explicitly
	return nil, nil
}

func intFilterFromGQL(intFilter *IntFilter) (*explorerdb.IntFilterRange, error) {
	// define min-max range
	var (
		min, max *int
	)
	if intFilter.LessThan != nil {
		if intFilter.LessThanOrEqualTo != nil || intFilter.EqualTo != nil || intFilter.GreaterThanOrEqualTo != nil || intFilter.GreaterThan != nil {
			return nil, errors.New("big int filter can only use one option, multiple cannot be combined")
		}
		max = intFilter.LessThan
		*max--
	} else if intFilter.LessThanOrEqualTo != nil {
		// LessThan is already confirmed by previous check to be nil
		if intFilter.EqualTo != nil || intFilter.GreaterThanOrEqualTo != nil || intFilter.GreaterThan != nil {
			return nil, errors.New("big int filter can only use one option, multiple cannot be combined")
		}
		max = intFilter.LessThanOrEqualTo
	} else if intFilter.EqualTo != nil {
		// LessThan and LessThanOrEqualTo are already confirmed by previous check to be nil
		if intFilter.GreaterThanOrEqualTo != nil || intFilter.GreaterThan != nil {
			return nil, errors.New("big int filter can only use one option, multiple cannot be combined")
		}
		min = intFilter.EqualTo
		max = intFilter.EqualTo
	} else if intFilter.GreaterThanOrEqualTo != nil {
		// LessThan, LessThanOrEqualTo and EqualTo are already confirmed by previous check to be nil
		if intFilter.GreaterThan != nil {
			return nil, errors.New("big int filter can only use one option, multiple cannot be combined")
		}
		min = intFilter.GreaterThanOrEqualTo
	} else if intFilter.GreaterThan != nil {
		// all others are already confirmed by previous check to be nil
		min = intFilter.GreaterThan
		*min++
	}

	// return filter function based on min/max range
	return explorerdb.NewIntFilterRange(min, max), nil
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
