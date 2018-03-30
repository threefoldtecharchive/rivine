package client

import (
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

func printContractInfo(hastings types.Currency, condition types.AtomicSwapCondition, secret types.AtomicSwapSecret) {
	var secretStr string
	if secret != (types.AtomicSwapSecret{}) {
		secretStr = fmt.Sprintf(`
Secret: %s`, secret)
	}

	cuh := types.NewAtomicSwapInputLock(condition).UnlockHash()

	fmt.Printf(`Contract address: %s
Contract value: %s
Recipient address: %s
Refund address: %s

Hashed Secret: %s%s

Locktime: %[7]d (%[7]s)
Locktime reached in: %s
`, cuh, _CurrencyConvertor.ToCoinStringWithUnit(hastings), condition.Receiver, condition.Sender, condition.HashedSecret,
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
