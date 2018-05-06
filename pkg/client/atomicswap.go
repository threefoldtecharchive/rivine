package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rivine/rivine/api"
	"github.com/rivine/rivine/types"
	"github.com/spf13/cobra"
)

var (
	atomicSwapCmd = &cobra.Command{
		Use:   "atomicswap",
		Short: "Create and interact with atomic swap contracts.",
		Long:  "Create and audit atomic swap contracts, as well as redeem money from them.",
	}

	atomicSwapParticipateCmd = &cobra.Command{
		Use:   "participate dest amount hashedsecret",
		Short: "Create an atomic swap contract as participant.",
		Long: `Create an atomic swap contract as a participant,
using the hashed secret given by the initiator.`,
		Run: Wrap(atomicswapparticipatecmd),
	}

	atomicSwapInitiateCmd = &cobra.Command{
		Use:   "initiate dest amount",
		Short: "Create an atomic swap contract as initiator.",
		Long: `Create an atomic swap contract as an initiator,
such that you can share its public information with the participant.`,
		Run: Wrap(atomicswapinitiatecmd),
	}

	atomicSwapAuditCmd = &cobra.Command{
		Use:   "audit outputid|unlockhash dest src timelock hashedsecret [amount]",
		Short: "Audit a created atomic swap contract.",
		Long: `Audit a created atomic swap contract.

If the first parameter is an unlock hash,
this command will only perform a quick audit,
meaning it will simply check if the given unlock hash,
matches the expected contract information.

If the first parameter is an outputID
the unlock hash will be retrieved from an edge copy of the blockchain,
using a local/remote explorer-enabled node,
optionally validating the amount of coins attached to the contract as well.
On top of that will ensure to check that the contract hasn't been spend yet.
`,
		Run: atomicswapauditcmd,
	}

	atomicSwapExtractSecretCmd = &cobra.Command{
		Use:   "extractsecret outputid [hashedsecret]",
		Short: "Extract the secret from a claimed swap contract.",
		Long: `Extract the secret from a claimed atomic swap contract,
by looking up the outputID, ensuring it is spend (as otherwise it cannot be claimed),
and retrieving the secret that was given part of a claim.

If it was spend as a refund, than no secret can be extracted.

Optionally the expected hashedsecret can be given, as to ensure that the
hashed secret is as expected, and also matches the found/extracted secret.`,
		Run: atomicswapextractsecretcmd,
	}

	atomicSwapClaimCmd = &cobra.Command{
		Use:   "claim outputid secret [dest src timelock hashedsecret] amount",
		Short: "Claim the coins locked in an atomic swap contract.",
		Long:  "Claim the coins locked in an atomic swap contract intended for you.",
		Run:   atomicswapclaimcmd,
	}

	atomicSwapRefundCmd = &cobra.Command{
		Use:   "refund outputid [dest src timelock hashedsecret] amount",
		Short: "Refund the coins locked in an atomic swap contract.",
		Long:  "Refund the coins locked in an atomic swap contract created by you.",
		Run:   atomicswaprefundcmd,
	}
)

var (
	atomicSwapParticipatecfg struct {
		duration         time.Duration
		sourceUnlockHash unlockHashFlag
	}
	atomicSwapInitiatecfg struct {
		duration         time.Duration
		sourceUnlockHash unlockHashFlag
	}
	atomicSwapClaimcfg struct {
		audit  bool
		legacy bool
	}
	atomicSwapRefundcfg struct {
		audit  bool
		legacy bool
	}
)

func atomicswapparticipatecmd(dest, amount, hashedSecret string) {
	var (
		condition types.AtomicSwapCondition
	)

	hastings, err := _CurrencyConvertor.ParseCoinString(amount)
	if err != nil {
		fmt.Fprintln(os.Stderr, _CurrencyConvertor.CoinArgDescription("amount"))
		Die("Could not parse amount:", err)
	}
	if hastings.Cmp(_MinimumTransactionFee) != 1 {
		Die("Cannot create atomic swap contract! Contracts which lock a value less than or equal to miner fees are currently not supported!")
	}

	err = condition.Receiver.LoadString(dest)
	if err != nil {
		Die("Could not parse destination address (unlock hash):", err)
	}

	if atomicSwapParticipatecfg.sourceUnlockHash.UnlockHash.Type != 0 {
		// use the hash given by the user explicitly
		condition.Sender = atomicSwapParticipatecfg.sourceUnlockHash.UnlockHash
	} else {
		// get new one from the wallet
		resp := new(api.WalletAddressGET)
		err := _DefaultClient.httpClient.GetAPI("/wallet/address", resp)
		if err != nil {
			Die("Could not generate new address:", err)
		}
		condition.Sender = resp.Address
	}

	if hsl := len(hashedSecret); hsl == types.AtomicSwapHashedSecretLen*2 {
		_, err := hex.Decode(condition.HashedSecret[:], []byte(hashedSecret))
		if err != nil {
			Die("Invalid hashed secret:", err)
		}
	} else {
		Die("Invalid hashed secret length")
	}
	if atomicSwapParticipatecfg.duration == 0 {
		Die("Duration is required")
	}
	condition.TimeLock = types.OffsetTimestamp(atomicSwapParticipatecfg.duration)

	// print contract for review
	printContractInfo(hastings, condition, types.AtomicSwapSecret{})
	fmt.Println("")

	// ensure user wants to continue with creating the contract as it is (aka publishing it)
	if !askYesNoQuestion("Publish atomic swap (participation) transaction?") {
		Die("Atomic swap participation cancelled!")
	}

	// publish contract
	body, err := json.Marshal(api.WalletTransactionPOST{
		Condition: types.NewCondition(&condition),
		Amount:    hastings,
	})
	if err != nil {
		Die("Couldn't create/marshal JSOn body:", err)
	}
	var response api.WalletTransactionPOSTResponse
	err = _DefaultClient.httpClient.PostResp("/wallet/transaction", string(body), &response)
	if err != nil {
		Die("Couldn't create transaction:", err)
	}

	// find coinOutput and return its ID if possible
	coinOutputIndex, unlockHash := -1, condition.UnlockHash()
	for idx, co := range response.Transaction.CoinOutputs {
		if unlockHash.Cmp(co.Condition.UnlockHash()) == 0 {
			coinOutputIndex = idx
			break
		}
	}
	if coinOutputIndex == -1 {
		panic("didn't find atomic swap contract registered in any returned coin output")
	}
	fmt.Println("published contract transaction")
	fmt.Println("OutputID:", response.Transaction.CoinOutputID(uint64(coinOutputIndex)))
}

func atomicswapinitiatecmd(dest, amount string) {
	var (
		condition types.AtomicSwapCondition
	)

	hastings, err := _CurrencyConvertor.ParseCoinString(amount)
	if err != nil {
		fmt.Fprintln(os.Stderr, _CurrencyConvertor.CoinArgDescription("amount"))
		Die("Could not parse amount:", err)
	}
	if hastings.Cmp(_MinimumTransactionFee) != 1 {
		Die("Cannot create atomic swap contract! Contracts which lock a value less than or equal to miner fees are currently not supported!")
	}

	err = condition.Receiver.LoadString(dest)
	if err != nil {
		Die("Could not parse destination address (unlock hash):", err)
	}

	if atomicSwapInitiatecfg.sourceUnlockHash.UnlockHash.Type != 0 {
		// use the hash given by the user explicitly
		condition.Sender = atomicSwapInitiatecfg.sourceUnlockHash.UnlockHash
	} else {
		// get new one from the wallet
		resp := new(api.WalletAddressGET)
		err := _DefaultClient.httpClient.GetAPI("/wallet/address", resp)
		if err != nil {
			Die("Could not generate new address:", err)
		}
		condition.Sender = resp.Address
	}

	secret, err := types.NewAtomicSwapSecret()
	if err != nil {
		Die("failed to crypto-generate secret:", err)
	}
	condition.HashedSecret = types.NewAtomicSwapHashedSecret(secret)

	if atomicSwapInitiatecfg.duration == 0 {
		Die("Duration is required")
	}
	condition.TimeLock = types.OffsetTimestamp(atomicSwapInitiatecfg.duration)

	// print contract for review
	printContractInfo(hastings, condition, secret)
	fmt.Println("")

	// ensure user wants to continue with creating the contract as it is (aka publishing it)
	if !askYesNoQuestion("Publish atomic swap (initiating) transaction?") {
		Die("Atomic swap initiating cancelled!")
	}

	// publish contract
	body, err := json.Marshal(api.WalletTransactionPOST{
		Condition: types.NewCondition(&condition),
		Amount:    hastings,
	})
	if err != nil {
		Die("Couldn't create/marshal JSOn body:", err)
	}
	var response api.WalletTransactionPOSTResponse
	err = _DefaultClient.httpClient.PostResp("/wallet/transaction", string(body), &response)
	if err != nil {
		Die("Couldn't create transaction:", err)
	}

	// find coinOutput and return its ID if possible
	coinOutputIndex, unlockHash := -1, condition.UnlockHash()
	for idx, co := range response.Transaction.CoinOutputs {
		if unlockHash.Cmp(co.Condition.UnlockHash()) == 0 {
			coinOutputIndex = idx
			break
		}
	}
	if coinOutputIndex == -1 {
		panic("didn't find atomic swap contract registered in any returned coin output")
	}
	fmt.Println("published contract transaction")
	fmt.Println("OutputID:", response.Transaction.CoinOutputID(uint64(coinOutputIndex)))
}

func atomicswapauditcmd(cmd *cobra.Command, args []string) {
	argn := len(args)
	if argn < 5 || argn > 6 {
		cmd.UsageFunc()(cmd)
		os.Exit(exitCodeUsage)
	}

	var condition types.AtomicSwapCondition
	err := condition.Receiver.LoadString(args[1])
	if err != nil {
		Die("failed to parse dst-argument:", err)
	}
	err = condition.Sender.LoadString(args[2])
	if err != nil {
		Die("failed to parse src-argument:", err)
	}
	err = condition.TimeLock.LoadString(args[3])
	if err != nil {
		Die("failed to parse timelock-argument:", err)
	}
	err = condition.HashedSecret.LoadString(args[4])
	if err != nil {
		Die("failed to parse hashedsecret-argument:", err)
	}

	// try to interpret the first param as an unlock hash
	var expectedUH types.UnlockHash
	if err = expectedUH.LoadString(args[0]); err == nil {
		// do a quick check!
		atomicswapauditcmdFast(expectedUH, condition)
		return
	}

	// try to interpret the first param as an outputID
	var outputID types.CoinOutputID
	if err = outputID.LoadString(args[0]); err == nil {
		var amount types.Currency // hastings or block stakes
		if argn == 6 {
			amount, err = _CurrencyConvertor.ParseCoinString(args[5])
			if err != nil {
				Die("failed to parse optional amount-argument:", err)
			}
		}
		// do a complete check! feels good, no?
		atomicswapauditcmdComplete(outputID, &condition, amount)
		return
	}

	cmd.UsageFunc()(cmd)
	Die("first parameter has to be valid and be either an unlock hash or an outputID")
}

// atomicswapauditcmdFast only ensures that the given unlock hash matches the given AS condition
func atomicswapauditcmdFast(expectedUH types.UnlockHash, condition types.AtomicSwapCondition) {
	if expectedUH.Type != types.UnlockTypeAtomicSwap {
		Die("The given unlock has is not of the atomic swap (unlock) type!")
	}

	uh := condition.UnlockHash()
	if expectedUH.Cmp(uh) != 0 {
		Die("Invalid contract! Given contract information does NOT match the given unlock hash!")
	}

	fmt.Println("Given unlock hash matches the given contract information :)")
	fmt.Println("")
	printContractInfo(types.Currency{}, condition, types.AtomicSwapSecret{})
	fmt.Println("")
	fmt.Println("This was a quick check only, whether it has been spend already or not is unclear.")
	fmt.Println("You can do a complete/thorough check when auditing using the output ID instead.")
}

// atomicswapauditcmdComplete ensures that an output exists for the given outputID,
// that the unlock hash of that output matches the given AS condition
// and, last but not least, it ensures that this found output hasn't been spend yet
// (but seriously, why would it?)
func atomicswapauditcmdComplete(outputID types.CoinOutputID, conditionRef *types.AtomicSwapCondition, amount types.Currency) (parentTransaction types.Transaction) {
	// get output info from explorer
	resp := new(api.ExplorerHashGET)
	err := _DefaultClient.httpClient.GetAPI("/explorer/hashes/"+outputID.String(), resp)
	if err != nil {
		Die("Could not get blockStake/coin output from explorer:", err)
	}

	switch tln := len(resp.Transactions); tln {
	case 1:
		txn := resp.Transactions[0]
		parentTransaction = txn.RawTransaction

		// when we receive 1 transaction,
		// we'll assume that the output (atomic swap contract) hasn't been spend yet
		var expectedUnlockHash types.UnlockHash
		if conditionRef != nil {
			expectedUnlockHash = conditionRef.UnlockHash()
		} else {
			expectedUnlockHash = types.NilUnlockHash
		}
		var condition types.AtomicSwapCondition
		switch resp.HashType {
		case api.HashTypeCoinOutputIDStr:
			condition = auditAtomicSwapCoinOutput(txn, outputID, expectedUnlockHash, amount)
		default:
			Die("Received unexpected hash type for given output ID:", resp.HashType)
		}

		fmt.Println("An unspend atomic swap contract could be found for the given outputID,")
		fmt.Println("and the given contract information matches the found contract's information, all good! :)")
		fmt.Println("")
		printContractInfo(types.Currency{}, condition, types.AtomicSwapSecret{})

	case 2:
		// when we receive 2 transactions,
		// we'll assume that output (atomic swap contract) has been spend already
		// try to find the transactionID, so we can tell which one it was spend
		switch resp.HashType {
		case api.HashTypeCoinOutputIDStr:
			auditAtomicSwapFindSpendCoinTransactionID(types.CoinOutputID(outputID), resp.Transactions)
		default:
			Die("Requested output was found, but already spend, and received unexpected hash type for given output ID:", resp.HashType)
		}

	default:
		Die("Unexpected amount (BUG?!) of returned transactions for given outputID:", tln)
	}

	return
}
func auditAtomicSwapCoinOutput(txn api.ExplorerTransaction, expectedCoinOutputID types.CoinOutputID, expectedUH types.UnlockHash, expectedHastings types.Currency) types.AtomicSwapCondition {
	coinOutputIndex := -1
	for idx, co := range txn.CoinOutputIDs {
		if bytes.Compare(co[:], expectedCoinOutputID[:]) == 0 {
			coinOutputIndex = idx
			break
		}
	}
	if coinOutputIndex == -1 {
		Die("Couldn't find expected coin outputID in retrieved transaction's coin outputs! BUG?!") // unexpected, bug?
	}
	if len(txn.RawTransaction.CoinOutputs) == 0 || len(txn.RawTransaction.CoinOutputs) < coinOutputIndex {
		Die("Retrieved transaction for given coin output ID has returned insufficient amount of coin outputs to check! BUG?!") // unexpected, bug?
	}
	coinOutput := txn.RawTransaction.CoinOutputs[coinOutputIndex]

	condition, ok := coinOutput.Condition.Condition.(*types.AtomicSwapCondition)
	if !ok {
		Die(fmt.Sprintf("The found coin output's condition type (%T) is not of the atomic swap type!", coinOutput.Condition.Condition))
	}

	if condition.UnlockHash().Cmp(expectedUH) != 0 {
		Die("Invalid contract! The found coin output's unlock hash does not match the provided contract information!")
	}

	if expectedHastings.Equals64(0) {
		fmt.Println("No amount hastings given to audit, skipping this audit step!")
		return *condition // no hastings to audit
	}
	if !expectedHastings.Equals(coinOutput.Value) {
		Die(fmt.Sprintf("Invalid Contract! expected %s but found %s registered in the found coin output instead!",
			_CurrencyConvertor.ToCoinStringWithUnit(expectedHastings),
			_CurrencyConvertor.ToCoinStringWithUnit(coinOutput.Value)))
	}

	return *condition
}
func auditAtomicSwapFindSpendCoinTransactionID(coinOutputID types.CoinOutputID, txns []api.ExplorerTransaction) {
	for _, txn := range txns {
		for _, ci := range txn.RawTransaction.CoinInputs {
			if bytes.Compare(coinOutputID[:], ci.ParentID[:]) == 0 {
				Die("Atomic swap contract was already spend as a coin input! This as part of transaction:", txn.ID.String())
			}
		}
	}
	Die("Atomic swap contract was already spend as a coin input! This as part of an unknown transaction (BUG!?)")
}

// extractsecret outputid [hashedsecret]
func atomicswapextractsecretcmd(cmd *cobra.Command, args []string) {
	argn := len(args)
	if argn < 1 || argn > 2 {
		cmd.UsageFunc()(cmd)
		os.Exit(exitCodeUsage)
	}

	var (
		outputID             types.OutputID
		expectedHashedSecret types.AtomicSwapHashedSecret
	)
	err := outputID.LoadString(args[0])
	if err != nil {
		Die("Couldn't parse first argment as an output ID:", err)
	}
	if argn == 2 {
		err := expectedHashedSecret.LoadString(args[1])
		if err != nil {
			Die("Couldn't parse second argment as a hashed secret:", err)
		}
	}

	// get output info from explorer
	resp := new(api.ExplorerHashGET)
	err = _DefaultClient.httpClient.GetAPI("/explorer/hashes/"+outputID.String(), resp)
	if err != nil {
		Die("Could not get blockStake/coin output from explorer:", err)
	}

	switch tln := len(resp.Transactions); tln {
	case 1:
		Die("Cannot extract secret! Atomic swap contract has not yet been claimed!")

	case 2:
		// when we receive 2 transactions,
		// we'll assume that output (atomic swap contract) has been spend already
		// try to get the secret, which is possible if it was clamed
		var secret types.AtomicSwapSecret
		switch resp.HashType {
		case api.HashTypeCoinOutputIDStr:
			secret = auditAtomicSwapFindSpendCoinSecret(types.CoinOutputID(outputID), resp.Transactions)
		case api.HashTypeBlockStakeOutputIDStr:
			secret = auditAtomicSwapFindSpendBlockStakeSecret(types.BlockStakeOutputID(outputID), resp.Transactions)
		default:
			Die("Cannot extract secret! Requested output was found and already spend, but received unexpected hash type for given output ID:", resp.HashType)
		}
		if secret == (types.AtomicSwapSecret{}) {
			Die("Atomic swap contract was spend as a refund, not a claim! No secret can be extracted!")
		}

		// no need to verify the secret ourselves,
		// as the secret is already guaranteed to be correct,
		// given it is part of the blockchain and validated as part of the creation and syncing.

		fmt.Println("Atomic swap contract was claimed by recipient! Success! :)")
		fmt.Println("Extracted secret:", secret.String())

	default:
		Die("Cannot extract secret! Unexpected amount (BUG?!) of returned transactions for given outputID:", tln)
	}
}

type atomicSwapSecretGetter interface {
	AtomicSwapSecret() types.AtomicSwapSecret
}

func auditAtomicSwapFindSpendCoinSecret(coinOutputID types.CoinOutputID, txns []api.ExplorerTransaction) types.AtomicSwapSecret {
	for _, txn := range txns {
		for _, ci := range txn.RawTransaction.CoinInputs {
			if bytes.Compare(coinOutputID[:], ci.ParentID[:]) == 0 {
				asf, ok := ci.Fulfillment.Fulfillment.(atomicSwapSecretGetter)
				if !ok {
					Die("Given output was spend as a coin input, but fulfillment was not of type atomic swap!")
				}
				return asf.AtomicSwapSecret()
			}
		}
	}
	Die("Output was spend as a coin input, but couldn't find the transaction! (BUG?!")
	return types.AtomicSwapSecret{}
}
func auditAtomicSwapFindSpendBlockStakeSecret(blockStakeOutputID types.BlockStakeOutputID, txns []api.ExplorerTransaction) types.AtomicSwapSecret {
	for _, txn := range txns {
		for _, bsi := range txn.RawTransaction.BlockStakeInputs {
			if bytes.Compare(blockStakeOutputID[:], bsi.ParentID[:]) == 0 {
				asf, ok := bsi.Fulfillment.Fulfillment.(atomicSwapSecretGetter)
				if !ok {
					Die("Given output was spend as a block stake input, but fulfillment was not of type atomic swap!")
				}
				return asf.AtomicSwapSecret()
			}
		}
	}
	Die("Output was spend as a block stake input, but couldn't find the transaction! (BUG?!")
	return types.AtomicSwapSecret{}
}

// claim outputid secret [dest src timelock hashedsecret amount]
func atomicswapclaimcmd(cmd *cobra.Command, args []string) {
	var (
		err          error
		outputID     types.CoinOutputID
		secret       types.AtomicSwapSecret
		conditionRef *types.AtomicSwapCondition
		hastings     types.Currency
	)

	// step 1: parse all arguments and prepare transaction primitives
	switch len(args) {
	case 3:
		hastings, err = _CurrencyConvertor.ParseCoinString(args[2])
		if err != nil {
			Die("failed to parse amount-argument as coins:", err)
		}
		atomicSwapClaimcfg.audit = true // enforce auditing, such that we can get condition via there

	case 7:
		condition := getAtomicSwapRedeemConditionFromOptPosArgs(args[2:6])
		conditionRef = &condition

		hastings, err = _CurrencyConvertor.ParseCoinString(args[6])
		if err != nil {
			Die("failed to parse amount-argument as coins:", err)
		}

	default:
		cmd.UsageFunc()(cmd)
		Die("Invalid amount of positional arguments given!")
	}

	// parse common pos args
	err = outputID.LoadString(args[0])
	if err != nil {
		Die("failed to parse outputid-argument:", err)
	}
	err = secret.LoadString(args[1])
	if err != nil {
		Die("failed to parse secret-argument:", err)
	}

	// optional step: audit contract
	if atomicSwapClaimcfg.audit {
		// overwrite our flag
		parentTransaction := atomicswapauditcmdComplete(outputID, conditionRef, hastings)
		atomicSwapClaimcfg.legacy = parentTransaction.Version == types.TransactionVersionZero

		// NOTE: for now we hardcode this for coins, as we only support contracts with coins in this cmd
		for idx, co := range parentTransaction.CoinOutputs {
			if parentTransaction.CoinOutputID(uint64(idx)) == outputID {
				var ok bool
				conditionRef, ok = co.Condition.Condition.(*types.AtomicSwapCondition)
				if !ok {
					Die("unexpected condition for coin output parent ID")
				}
			}
		}

		if conditionRef == nil {
			Die("atomic swap condition could not be find in parent coin output with ID " + outputID.String())
		}
	}

	// step 2: get correct spendable key from wallet
	pk, sk := getSpendableKey(conditionRef.Receiver)
	// quickly validate if returned sk matches the known unlock hash (sanity check)
	uh := types.NewPubKeyUnlockHash(pk)
	if uh.Cmp(conditionRef.Receiver) != 0 {
		Die("Unexpected wallet public key returned:", sk)
	}

	if hastings.Cmp(_MinimumTransactionFee) != 1 {
		Die("Cannot claim atomic swap contract! Contracts which lock a value less than or equal to miner fees are currently not supported!")
	}

	// step 3: confirm contract details with user, before continuing
	// print contract for review
	if !atomicSwapClaimcfg.audit {
		// only print again, if not printed already
		printContractInfo(hastings, *conditionRef, secret)
	}
	fmt.Println("")
	// ensure user wants to continue with claiming the contract!
	if !askYesNoQuestion("Publish atomic swap claim transaction?") {
		Die("Atomic swap claim transaction cancelled!")
	}

	// step 4: create a transaction
	txn := types.Transaction{
		Version: _DefaultTransactionVersion,
		CoinInputs: []types.CoinInput{
			{
				ParentID: outputID,
				Fulfillment: types.NewFulfillment(func() types.MarshalableUnlockFulfillment {
					if atomicSwapClaimcfg.legacy {
						return &types.LegacyAtomicSwapFulfillment{
							Sender:       conditionRef.Sender,
							Receiver:     conditionRef.Receiver,
							HashedSecret: conditionRef.HashedSecret,
							TimeLock:     conditionRef.TimeLock,
							PublicKey:    pk,
							Secret:       secret,
						}
					}
					return &types.LegacyAtomicSwapFulfillment{
						PublicKey: pk,
						Secret:    secret,
					}
				}()),
			},
		},
		CoinOutputs: []types.CoinOutput{
			{
				Condition: types.NewCondition(types.NewUnlockHashCondition(uh)),
				Value:     hastings.Sub(_MinimumTransactionFee),
			},
		},
		MinerFees: []types.Currency{_MinimumTransactionFee},
	}

	// step 5: sign transaction's only input
	err = txn.CoinInputs[0].Fulfillment.Sign(types.FulfillmentSignContext{
		InputIndex:  0,
		Transaction: txn,
		Key:         sk,
	})
	if err != nil {
		Die("Cannot claim atomic swap's locked coins! Couldn't sign transaction:", err)
	}
	if uh.Cmp(conditionRef.Receiver) != 0 {
		Die("Cannot claim atomic swap's locked coins! Wrong wallet key-pair received:", uh)
	}

	// step 6: submit transaction to transaction pool and celebrate if possible
	txnid, err := commitTxn(txn)
	if err != nil {
		Die("Failed to claim atomic swaps locked tokens, as transaction couldn't commit:", err)
	}

	fmt.Println("")
	fmt.Println("Published atomic swap claim transaction!")
	fmt.Println("Transaction ID:", txnid)
	fmt.Println(`>   NOTE that this does NOT mean for 100% you'll have the money!
> Due to potential forks, double spending, and any other possible issues your
> claim might be declined by the network. Please check the network
> (e.g. using a public explorer node or your own full node) to ensure
> your payment went through. If not, try to audit the contract (again).`)
}

// refund outputid [dest src timelock hashedsecret amount]
func atomicswaprefundcmd(cmd *cobra.Command, args []string) {
	var (
		err          error
		outputID     types.CoinOutputID
		secret       types.AtomicSwapSecret
		conditionRef *types.AtomicSwapCondition
		hastings     types.Currency
	)

	// step 1: parse all arguments and prepare transaction primitivesswitch len(args) {
	switch len(args) {
	case 2:
		hastings, err = _CurrencyConvertor.ParseCoinString(args[1])
		if err != nil {
			Die("failed to parse amount-argument as coins:", err)
		}
		atomicSwapRefundcfg.audit = true // enforce auditing, such that we can get condition via there

	case 6:
		condition := getAtomicSwapRedeemConditionFromOptPosArgs(args[1:5])
		conditionRef = &condition

		hastings, err = _CurrencyConvertor.ParseCoinString(args[5])
		if err != nil {
			Die("failed to parse amount-argument as coins:", err)
		}

	default:
		cmd.UsageFunc()(cmd)
		Die("Invalid amount of positional arguments given!")
	}

	// parse common pos args
	err = outputID.LoadString(args[0])
	if err != nil {
		Die("failed to parse outputid-argument:", err)
	}

	// optional step: audit contract
	if atomicSwapRefundcfg.audit {
		// overwrite our flag
		parentTransaction := atomicswapauditcmdComplete(outputID, conditionRef, hastings)
		atomicSwapRefundcfg.legacy = parentTransaction.Version == types.TransactionVersionZero

		// NOTE: for now we hardcode this for coins, as we only support contracts with coins in this cmd
		for idx, co := range parentTransaction.CoinOutputs {
			if parentTransaction.CoinOutputID(uint64(idx)) == outputID {
				var ok bool
				conditionRef, ok = co.Condition.Condition.(*types.AtomicSwapCondition)
				if !ok {
					Die("unexpected condition for coin output parent ID")
				}
			}
		}

		if conditionRef == nil {
			Die("atomic swap condition could not be find in parent coin output with ID " + outputID.String())
		}
	}

	// step 2: get correct spendable key from wallet
	pk, sk := getSpendableKey(conditionRef.Sender)
	// quickly validate if returned sk matches the known unlock hash (sanity check)
	uh := types.NewPubKeyUnlockHash(pk)
	if uh.Cmp(conditionRef.Sender) != 0 {
		Die("Unexpected wallet public key returned:", sk)
	}
	if hastings.Cmp(_MinimumTransactionFee) == -1 {
		Die("Cannot refund atomic swap contract! Contracts which lock a value less than or equal to miner fees are currently not supported!")
	}

	// step 3: confirm contract details with user, before continuing
	// print contract for review
	if !atomicSwapRefundcfg.audit {
		// only print again, if not printed already
		printContractInfo(hastings, *conditionRef, secret)
	}
	fmt.Println("")
	// ensure user wants to continue with refunding the contract!
	if !askYesNoQuestion("Publish atomic swap refund transaction?") {
		Die("Atomic swap refund transaction cancelled!")
	}

	// step 4: create a transaction
	txn := types.Transaction{
		Version: _DefaultTransactionVersion,
		CoinInputs: []types.CoinInput{
			{
				ParentID: outputID,
				Fulfillment: types.NewFulfillment(func() types.MarshalableUnlockFulfillment {
					if atomicSwapRefundcfg.legacy {
						return &types.LegacyAtomicSwapFulfillment{
							Sender:       conditionRef.Sender,
							Receiver:     conditionRef.Receiver,
							HashedSecret: conditionRef.HashedSecret,
							TimeLock:     conditionRef.TimeLock,
							PublicKey:    pk,
							// secret not needed for refund
						}
					}
					return &types.LegacyAtomicSwapFulfillment{
						PublicKey: pk,
						// secret not needed for refund
					}
				}()),
			},
		},
		CoinOutputs: []types.CoinOutput{
			{
				Condition: types.NewCondition(types.NewUnlockHashCondition(uh)),
				Value:     hastings.Sub(_MinimumTransactionFee),
			},
		},
		MinerFees: []types.Currency{_MinimumTransactionFee},
	}

	// step 5: sign transaction's only input
	err = txn.CoinInputs[0].Fulfillment.Sign(types.FulfillmentSignContext{
		InputIndex:  0,
		Transaction: txn,
		Key:         sk,
	})
	if err != nil {
		Die("Cannot refund atomic swap's locked coins! Couldn't sign transaction:", err)
	}
	if uh.Cmp(conditionRef.Sender) != 0 {
		Die("Cannot refund atomic swap's locked coins! Wrong wallet key-pair received:", uh)
	}

	// step 6: submit transaction to transaction pool and celebrate if possible
	txnid, err := commitTxn(txn)
	if err != nil {
		Die("Failed to refund atomic swaps locked tokens, as transaction couldn't commit:", err)
	}

	fmt.Println("")
	fmt.Println("Published atomic swap refund transaction!")
	fmt.Println("Transaction ID:", txnid)
	fmt.Println(`>   NOTE that this does NOT mean for 100% you'll have the money!
> Due to potential forks, double spending, and any other possible issues your
> refund might be declined by the network. Please check the network
> (e.g. using a public explorer node or your own full node) to ensure
> your payment went through. If not, try to audit the contract (again).`)
}

// getAtomicSwapRedeemConditionFromOptPosArgs parses the following 4 arguments in order:
// dest src timelock hashedseret
func getAtomicSwapRedeemConditionFromOptPosArgs(args []string) (condition types.AtomicSwapCondition) {
	err := condition.Receiver.LoadString(args[0])
	if err != nil {
		Die("failed to parse dest-argument:", err)
	}
	err = condition.Sender.LoadString(args[1])
	if err != nil {
		Die("failed to parse src-argument:", err)
	}
	err = condition.TimeLock.LoadString(args[2])
	if err != nil {
		Die("failed to parse timelock-argument:", err)
	}
	err = condition.HashedSecret.LoadString(args[3])
	if err != nil {
		Die("failed to parse hashedsecret-argument:", err)
	}

	return
}

// get public- and private key from wallet module
func getSpendableKey(unlockHash types.UnlockHash) (types.SiaPublicKey, types.ByteSlice) {
	resp := new(api.WalletKeyGet)
	err := _DefaultClient.httpClient.GetAPI("/wallet/key/"+unlockHash.String(), resp)
	if err != nil {
		Die("Could not get a matching wallet public/secret key pair for the given unlock hash:", err)
	}
	return types.SiaPublicKey{
		Algorithm: resp.AlgorithmSpecifier,
		Key:       resp.PublicKey,
	}, resp.SecretKey
}

// commitTxn sends a transaction to the used node's transaction pool
func commitTxn(txn types.Transaction) (types.TransactionID, error) {
	bodyBuff := bytes.NewBuffer(nil)
	err := json.NewEncoder(bodyBuff).Encode(&txn)
	if err != nil {
		return types.TransactionID{}, err
	}

	resp := new(api.TransactionPoolPOST)
	err = _DefaultClient.httpClient.PostResp("/transactionpool/transactions", bodyBuff.String(), resp)
	return resp.TransactionID, err
}

func printContractInfo(hastings types.Currency, condition types.AtomicSwapCondition, secret types.AtomicSwapSecret) {
	var amountStr string
	if !hastings.Equals(types.Currency{}) {
		amountStr = fmt.Sprintf(`
Contract value: %s`, _CurrencyConvertor.ToCoinStringWithUnit(hastings))
	}

	var secretStr string
	if secret != (types.AtomicSwapSecret{}) {
		secretStr = fmt.Sprintf(`
Secret: %s`, secret)
	}

	cuh := condition.UnlockHash()

	fmt.Printf(`Contract address: %s%s
Recipient address: %s
Refund address: %s

Hashed Secret: %s%s

Locktime: %[7]d (%[7]s)
Locktime reached in: %s
`, cuh, amountStr, condition.Receiver, condition.Sender, condition.HashedSecret,
		secretStr, condition.TimeLock,
		time.Unix(int64(condition.TimeLock), 0).Sub(time.Now()))
}

func askYesNoQuestion(str string) bool {
	fmt.Printf("%s [Y/N] ", str)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		Die("failed to scan response:", err)
	}
	response = strings.ToLower(response)
	if containsString(okayResponses, response) {
		return true
	}
	if containsString(nokayResponses, response) {
		return false
	}

	fmt.Println("Please answer using 'yes' or 'no'")
	return askYesNoQuestion(str)
}

// posString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

// containsString returns true iff slice contains element
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}

var (
	okayResponses  = []string{"y", "ye", "yes"}
	nokayResponses = []string{"n", "no", "noo", "nope"}
)

func init() {
	atomicSwapParticipateCmd.Flags().DurationVarP(
		&atomicSwapParticipatecfg.duration, "duration", "d",
		time.Hour*24, "the duration of the atomic swap contract, the amount of time the receiver has to collect")
	atomicSwapParticipateCmd.Flags().Var(&atomicSwapParticipatecfg.sourceUnlockHash, "src",
		"optionally define a source address that is to be used for refunding purposes, one will be generated for you if none is given")

	atomicSwapInitiateCmd.Flags().DurationVarP(
		&atomicSwapInitiatecfg.duration, "duration", "d",
		time.Hour*48, "the duration of the atomic swap contract, the amount of time the receiver has to collect")
	atomicSwapInitiateCmd.Flags().Var(&atomicSwapInitiatecfg.sourceUnlockHash, "src",
		"optionally define a source address that is to be used for refunding purposes, one will be generated for you if none is given")

	atomicSwapClaimCmd.Flags().BoolVar(&atomicSwapClaimcfg.audit, "audit", false,
		"optionally audit the given contract information against the known contract info on the used explorer node")
	atomicSwapClaimCmd.Flags().BoolVar(&atomicSwapClaimcfg.legacy, "legacy", false,
		"defines if the to be created fulfillment has to be a legacy one, fulfilling a legacy atomic swap condition")

	atomicSwapRefundCmd.Flags().BoolVar(&atomicSwapRefundcfg.audit, "audit", false,
		"optionally audit the given contract information against the known contract info on the used explorer node")
	atomicSwapRefundCmd.Flags().BoolVar(&atomicSwapRefundcfg.legacy, "legacy", false,
		"defines if the to be created fulfillment has to be a legacy one, fulfilling a legacy atomic swap condition")
}

type unlockHashFlag struct {
	types.UnlockHash
}

func (uhf *unlockHashFlag) Set(str string) error {
	return uhf.LoadString(str)
}

func (uhf unlockHashFlag) Type() string {
	return "UnlockHash"
}
