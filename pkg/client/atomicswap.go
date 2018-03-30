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
)

func newContractUnlockHash(condition types.AtomicSwapCondition) types.UnlockHash {
	return types.NewAtomicSwapInputLock(condition).UnlockHash()
}

func atomicswapparticipatecmd(dest, amount, hashedSecret string) {
	var (
		condition types.AtomicSwapCondition
	)

	hastings, err := _CurrencyConvertor.ParseCoinString(amount)
	if err != nil {
		fmt.Fprintln(os.Stderr, _CurrencyConvertor.CoinArgDescription("amount"))
		Die("Could not parse amount:", err)
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
	unlockHash := types.NewAtomicSwapInputLock(condition).UnlockHash()
	body, err := json.Marshal(api.WalletTransactionPOST{
		UnlockHash: unlockHash,
		Amount:     hastings,
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
	coinOutputIndex := -1
	for idx, co := range response.Transaction.CoinOutputs {
		if unlockHash.Cmp(co.UnlockHash) == 0 {
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
	unlockHash := types.NewAtomicSwapInputLock(condition).UnlockHash()
	body, err := json.Marshal(api.WalletTransactionPOST{
		UnlockHash: unlockHash,
		Amount:     hastings,
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
	coinOutputIndex := -1
	for idx, co := range response.Transaction.CoinOutputs {
		if unlockHash.Cmp(co.UnlockHash) == 0 {
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
	var outputID types.OutputID
	if err = outputID.LoadString(args[0]); err == nil {
		var amount types.Currency // hastings or block stakes
		if argn == 6 {
			amount, err = _CurrencyConvertor.ParseCoinString(args[5])
			if err != nil {
				Die("failed to parse optional amount-argument:", err)
			}
		}
		// do a complete check! feels good, no?
		atomicswapauditcmdComplete(outputID, condition, amount)
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

	uh := types.NewAtomicSwapInputLock(condition).UnlockHash()
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
func atomicswapauditcmdComplete(outputID types.OutputID, condition types.AtomicSwapCondition, amount types.Currency) {
	// get outputID info from explorer
	resp := new(api.ExplorerHashGET)
	err := _DefaultClient.httpClient.GetAPI("/explorer/hashes/"+outputID.String(), resp)
	if err != nil {
		Die("Could not get blockStake/coin output from explorer:", err)
	}

	switch tln := len(resp.Transactions); tln {
	case 1:
		// when we receive 1 transaction,
		// we'll assume that the output (atomic swap contract) hasn't been spend yet
		expectedUnlockHash := types.NewAtomicSwapInputLock(condition).UnlockHash()
		switch resp.HashType {
		case api.HashTypeCoinOutputIDStr:
			auditAtomicSwapCoinOutput(resp.Transactions[0],
				types.CoinOutputID(outputID), expectedUnlockHash, amount)
		case api.HashTypeBlockStakeOutputIDStr:
			auditAtomicSwapBlockStakeOutput(resp.Transactions[0],
				types.BlockStakeOutputID(outputID), expectedUnlockHash, amount)
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
		case api.HashTypeBlockStakeOutputIDStr:
			auditAtomicSwapFindSpendBlockStakeTransactionID(types.BlockStakeOutputID(outputID), resp.Transactions)
		default:
			Die("Requested output was found, but already spend, and received unexpected hash type for given output ID:", resp.HashType)
		}

	default:
		Die("Unexpected amount (BUG?!) of returned transactions for given outputID:", tln)
	}
}
func auditAtomicSwapCoinOutput(txn api.ExplorerTransaction, expectedCoinOutputID types.CoinOutputID, expectedUH types.UnlockHash, expectedHastings types.Currency) {
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
	if coinOutput.UnlockHash.Type != types.UnlockTypeAtomicSwap {
		Die("The found coin output's unlock type is not of the atomic swap (unlock) type!")
	}
	if coinOutput.UnlockHash.Cmp(expectedUH) != 0 {
		Die("Invalid contract! The found coin output's unlock hash does not match the provided contract information!")
	}

	if expectedHastings.Equals64(0) {
		fmt.Println("No amount hastings given to audit, skipping this audit step!")
		return // no hastings to audit
	}
	if !expectedHastings.Equals(coinOutput.Value) {
		Die(fmt.Sprintf("Invalid Contract! expected %s but found %s registered in the found coin output instead!",
			_CurrencyConvertor.ToCoinStringWithUnit(expectedHastings),
			_CurrencyConvertor.ToCoinStringWithUnit(coinOutput.Value)))
	}
}
func auditAtomicSwapBlockStakeOutput(txn api.ExplorerTransaction, expectedBlockStakeID types.BlockStakeOutputID, expectedUH types.UnlockHash, expectedBlockStakes types.Currency) {
	blockStakeOutputIndex := -1
	for idx, co := range txn.BlockStakeOutputIDs {
		if bytes.Compare(co[:], expectedBlockStakeID[:]) == 0 {
			blockStakeOutputIndex = idx
			break
		}
	}
	if blockStakeOutputIndex == -1 {
		Die("Couldn't find expected block stake outputID in retrieved transaction's block stake outputs! BUG?!") // unexpected, bug?
	}
	if len(txn.RawTransaction.BlockStakeOutputs) == 0 || len(txn.RawTransaction.BlockStakeOutputs) < blockStakeOutputIndex {
		Die("Retrieved transaction for given coin output ID has returned insufficient amount of coin outputs to check! BUG?!") // unexpected, bug?
	}
	blockStakeOutput := txn.RawTransaction.BlockStakeOutputs[blockStakeOutputIndex]
	if blockStakeOutput.UnlockHash.Type != types.UnlockTypeAtomicSwap {
		Die("The found block stake output's unlock type is not of the atomic swap (unlock) type!")
	}
	if blockStakeOutput.UnlockHash.Cmp(expectedUH) != 0 {
		Die("Invalid contract! The found block stake output's unlock hash does not match the provided contract information!")
	}

	if expectedBlockStakes.Equals64(0) {
		fmt.Println("No amount of block stakes given to audit, skipping this audit step!")
		return // no block takes to audit
	}
	if !expectedBlockStakes.Equals(blockStakeOutput.Value) {
		Die(fmt.Sprintf("Invalid Contract! expected %s BS but found %s BS registered in the found block stake output instead!",
			expectedBlockStakes.String(),
			blockStakeOutput.Value.String()))
	}
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
func auditAtomicSwapFindSpendBlockStakeTransactionID(blockStakeOutputID types.BlockStakeOutputID, txns []api.ExplorerTransaction) {
	for _, txn := range txns {
		for _, bsi := range txn.RawTransaction.BlockStakeInputs {
			if bytes.Compare(blockStakeOutputID[:], bsi.ParentID[:]) == 0 {
				Die("Atomic swap contract was already spend as a block stake input! This as part of transaction:", txn.ID.String())
			}
		}
	}
	Die("Atomic swap contract was already spend as a block stake input! This as part of an unknown transaction (BUG!?)")
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

	cuh := types.NewAtomicSwapInputLock(condition).UnlockHash()

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
