package consensus

import (
	"errors"
	"fmt"

	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"
)

// SetTransactionValidators sets the transaction validators used by the ConsensusSet as rules for all transactions,
// regardless of the version. Use SetTransactionVersionMappedValidators in case you want rules that only apply to a specific tx version.
// If no validators are passed, the validators returned by the `consensus.StandardTransactionValidators` function are used.
func (cs *ConsensusSet) SetTransactionValidators(validators ...modules.TransactionValidationFunction) {
	if len(validators) == 0 {
		cs.txValidators = StandardTransactionValidators()
		return
	}
	cs.txValidators = validators
}

// SetTransactionVersionMappedValidators sets the transaction validators used by the ConsensusSet as rules for
// the transactions of the defined version. If no validators are passed, the validators for the given transaction version,
// as returned by the `consensus.StandardTransactionVersionMappedValidators` function, are used.
func (cs *ConsensusSet) SetTransactionVersionMappedValidators(version types.TransactionVersion, validators ...modules.TransactionValidationFunction) {
	if len(validators) == 0 {
		validators := StandardTransactionVersionMappedValidators()
		if verValidators, ok := validators[version]; ok {
			cs.txVersionMappedValidators[version] = verValidators
		}
		return
	}
	cs.txVersionMappedValidators[version] = validators
}

// StandardTransactionVersionMappedValidators returns the standard version-mapped transaction validators that are used
// if no custom set of transaction validators (effecitively rules that apply to every transaction)
// are defined, or if `nil` rules are set, in which case these default validators will be used as well.
func StandardTransactionVersionMappedValidators() map[types.TransactionVersion][]modules.TransactionValidationFunction {
	return map[types.TransactionVersion][]modules.TransactionValidationFunction{
		types.TransactionVersionZero: []modules.TransactionValidationFunction{
			ValidateInvalidByDefault,
		},
		types.TransactionVersionOne: []modules.TransactionValidationFunction{
			ValidateCoinOutputsAreBalanced,
			ValidateBlockStakeOutputsAreBalanced,
		},
	}
}

// StandardTransactionValidators returns the standard transaction validators that are used
// if no custom set of transaction validators (effecitively rules that apply to every transaction)
// are defined, or if `nil` rules are set, in which case these default validators will be used as well.
func StandardTransactionValidators() []modules.TransactionValidationFunction {
	return []modules.TransactionValidationFunction{
		ValidateTransactionFitsInABlock,
		ValidateTransactionArbitraryData,
		ValidateCoinInputsAreValid,
		ValidateCoinOutputsAreValid,
		ValidateBlockStakeInputsAreValid,
		ValidateBlockStakeOutputsAreValid,
		ValidateMinerFeesAreValid,
		ValidateDoubleCoinSpends,
		ValidateDoubleBlockStakeSpends,
		ValidateCoinInputsAreFulfilled,
		ValidateBlockStakeInputsAreFulfilled,
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////
/// Transaction Validators (Validation at any height/time)
//////////////////////////////////////////////////////////////////////////////////////////////////////////////

// ValidateTransactionFitsInABlock is a validator function that checks
// if a transaction fits in a block
func ValidateTransactionFitsInABlock(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext) error {
	return types.TransactionFitsInABlock(tx.Transaction, ctx.BlockSizeLimit)
}

// ValidateTransactionArbitraryData is a validator function that checks
// if a transaction's arbitrary data is valid
func ValidateTransactionArbitraryData(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext) error {
	return types.ArbitraryDataFits(tx.ArbitraryData, ctx.ArbitraryDataSizeLimit)
}

// ValidateCoinOutputsAreValid is a validator function that checks if all coin outputs are standard,
// meaning their condition is considered standard (== known) and their (coin) value is individually greater than zero.
func ValidateCoinOutputsAreValid(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext) error {
	var err error
	for _, co := range tx.CoinOutputs {
		if co.Value.IsZero() {
			return types.ErrZeroOutput
		}
		err = co.Condition.IsStandardCondition(ctx.ValidationContext)
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateCoinInputsAreValid is a validator function that checks if all coin inputs are standard,
// meaning their fulfillment is considered standard (== known) and their parent ID is defined.
func ValidateCoinInputsAreValid(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext) error {
	var err error
	for _, ci := range tx.CoinInputs {
		if ci.ParentID == (types.CoinOutputID{}) {
			return errors.New("no parent ID defined for coin input")
		}
		err = ci.Fulfillment.IsStandardFulfillment(ctx.ValidationContext)
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateBlockStakeOutputsAreValid is a validator function that checks if all block stake output is standard,
// meaning their condition is considered standard (== known) and their (block stake) value is individually greater than zero.
func ValidateBlockStakeOutputsAreValid(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext) error {
	var err error
	for _, bso := range tx.BlockStakeOutputs {
		if bso.Value.IsZero() {
			return types.ErrZeroOutput
		}
		err = bso.Condition.IsStandardCondition(ctx.ValidationContext)
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateBlockStakeInputsAreValid is a validator function that checks if all block stake inputs are standard,
// meaning their fulfillment is considered standard (== known) and their parent ID is defined.
func ValidateBlockStakeInputsAreValid(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext) error {
	var err error
	for _, bsi := range tx.BlockStakeInputs {
		if bsi.ParentID == (types.BlockStakeOutputID{}) {
			return errors.New("no parent ID defined for block stake input")
		}
		err = bsi.Fulfillment.IsStandardFulfillment(ctx.ValidationContext)
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateMinerFeeIsPresent is a validator function that checks
// that at least one miner fee is present
func ValidateMinerFeeIsPresent(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext) error {
	if ctx.IsBlockCreatingTx {
		return nil // validation does not apply to to block creation tx
	}
	if len(tx.MinerFees) == 0 {
		return fmt.Errorf("tx %s does not contain any miner fees while at least one was expected", tx.ID().String())
	}
	return nil
}

// ValidateMinerFeesAreValid is a validator function that checks if all miner fees are valid,
// meaning their (coin) value is individually greater than zero.
func ValidateMinerFeesAreValid(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext) error {
	for _, fee := range tx.MinerFees {
		if fee.Cmp(ctx.MinimumMinerFee) == -1 {
			return types.ErrTooSmallMinerFee
		}
	}
	return nil
}

// ValidateDoubleCoinSpends validates that no coin output is spent twice.
func ValidateDoubleCoinSpends(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext) error {
	spendCoins := make(map[types.CoinOutputID]struct{}, len(tx.CoinInputs))
	for _, ci := range tx.CoinInputs {
		if _, found := spendCoins[ci.ParentID]; found {
			return types.ErrDoubleSpend
		}
		spendCoins[ci.ParentID] = struct{}{}
	}
	return nil
}

// ValidateDoubleBlockStakeSpends validates that no block stake output is spent twice.
func ValidateDoubleBlockStakeSpends(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext) error {
	spendBlockStakes := make(map[types.BlockStakeOutputID]struct{}, len(tx.BlockStakeInputs))
	for _, bsi := range tx.BlockStakeInputs {
		if _, found := spendBlockStakes[bsi.ParentID]; found {
			return types.ErrDoubleSpend
		}
		spendBlockStakes[bsi.ParentID] = struct{}{}
	}
	return nil
}

// ValidateInvalidByDefault returns always an error and can be used to not allow transactions to be validated.
func ValidateInvalidByDefault(_ modules.ConsensusTransaction, _ types.TransactionValidationContext) error {
	return errors.New("transaction is invalid as it has been disabled for validation using the ValidateInvalidByDefault function")
}

// ValidateCoinInputsAreFulfilled validates that all coin outputs are validated
func ValidateCoinInputsAreFulfilled(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext) error {
	var (
		ok bool
		co types.CoinOutput
	)
	for index, ci := range tx.CoinInputs {
		co, ok = tx.SpentCoinOutputs[ci.ParentID]
		if !ok {
			return fmt.Errorf(
				"unable to find parent ID %s as an unspent coin output in the current consensus transaction at block height %d",
				ci.ParentID.String(), ctx.BlockHeight)
		}
		// check if the referenced output's condition has been fulfilled
		err := co.Condition.Fulfill(ci.Fulfillment, types.FulfillContext{
			ExtraObjects: []interface{}{uint64(index)},
			BlockHeight:  ctx.BlockHeight,
			BlockTime:    ctx.BlockTime,
			Transaction:  tx.Transaction,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateCoinOutputsAreBalanced is a validator function that checks if the sum of
// all types of coin outputs equals the sum of coin inputs.
func ValidateCoinOutputsAreBalanced(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext) error {
	// collect the coin input sum
	var coinInputSum types.Currency
	for _, ci := range tx.CoinInputs {
		co, ok := tx.SpentCoinOutputs[ci.ParentID]
		if !ok {
			return fmt.Errorf(
				"unable to find parent ID %s as an unspent coin output in the current consensus transaction at block height %d",
				ci.ParentID.String(), ctx.BlockHeight)
		}
		coinInputSum = coinInputSum.Add(co.Value)
	}

	// collect the coin output sum
	coinOutputSum := tx.CoinOutputSum()

	// ensure the tx is balanced within the context of coin outputs
	r := coinInputSum.Cmp(coinOutputSum)
	if r < 0 {
		return fmt.Errorf(
			"unbalanced coin outputs: the sum of coin inputs (%s) for tx %s is less than its sum of coin outputs (%s)",
			coinInputSum.String(), tx.ID().String(), coinOutputSum.String())
	}
	if r > 0 {
		return fmt.Errorf(
			"unbalanced coin outputs: the sum of coin inputs (%s) for tx %s is greater than its sum of coin outputs (%s)",
			coinInputSum.String(), tx.ID().String(), coinOutputSum.String())
	}
	return nil
}

// ValidateBlockStakeInputsAreFulfilled validates that all block stake inputs are fulfilled
func ValidateBlockStakeInputsAreFulfilled(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext) error {
	var (
		ok  bool
		err error
		bso types.BlockStakeOutput
	)
	for index, bsi := range tx.BlockStakeInputs {
		bso, ok = tx.SpentBlockStakeOutputs[bsi.ParentID]
		if !ok {
			return fmt.Errorf(
				"unable to find parent ID %s as an unspent blockstake output in the current consensus transaction at block height %d",
				bsi.ParentID.String(), ctx.BlockHeight)
		}
		// check if the referenced output's condition has been fulfilled
		err = bso.Condition.Fulfill(bsi.Fulfillment, types.FulfillContext{
			ExtraObjects: []interface{}{uint64(index)},
			BlockHeight:  ctx.BlockHeight,
			BlockTime:    ctx.BlockTime,
			Transaction:  tx.Transaction,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateBlockStakeOutputsAreBalanced is a validator function that checks if the sum of
// all block stakes outputs equals the sum of all block stake inputs.
func ValidateBlockStakeOutputsAreBalanced(tx modules.ConsensusTransaction, ctx types.TransactionValidationContext) error {
	// collect the block stake input sum
	var blockStakeInputSum types.Currency
	for _, bsi := range tx.BlockStakeInputs {
		bso, ok := tx.SpentBlockStakeOutputs[bsi.ParentID]
		if !ok {
			return fmt.Errorf(
				"unable to find parent ID %s as an unspent blockstake output in the current consensus transaction at block height %d",
				bsi.ParentID.String(), ctx.BlockHeight)
		}
		blockStakeInputSum = blockStakeInputSum.Add(bso.Value)
	}

	// collect the block stkae output sum
	var blockStakeOutputSum types.Currency
	for _, bso := range tx.BlockStakeOutputs {
		blockStakeOutputSum = blockStakeOutputSum.Add(bso.Value)
	}

	// ensure the tx is balanced within the context of block stakes outputs
	r := blockStakeInputSum.Cmp(blockStakeOutputSum)
	if r < 0 {
		return fmt.Errorf(
			"unbalanced block stake outputs: the sum of block stake inputs (%s) for tx %s is less than its sum of block stake outputs (%s)",
			blockStakeInputSum.String(), tx.ID().String(), blockStakeOutputSum.String())
	}
	if r > 0 {
		return fmt.Errorf(
			"unbalanced block stake outputs: the sum of block stake inputs (%s) for tx %s is greater than its sum of block stake outputs (%s)",
			blockStakeInputSum.String(), tx.ID().String(), blockStakeOutputSum.String())
	}
	return nil
}
