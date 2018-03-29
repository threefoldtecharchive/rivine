package client

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rivine/rivine/api"
	"github.com/rivine/rivine/types"
	"github.com/spf13/cobra"
)

var (
	atomicSwapCmd = &cobra.Command{
		Use:   "atomicswap",
		Short: "Create and interact with atomic swap contracts.",
		Long:  "Create and audit atomic swap contract, as well as redeem money from them.",
	}

	atomicSwapCreateCmd = &cobra.Command{
		Use:   "create amount dest src",
		Short: "Create an atomic swap contract.",
		Long:  "Create an atomic swap contract, either as an initiator or a participant,",
		Run:   Wrap(atomicswapcreatecmd),
	}
)

var (
	createSwapContractConfig struct {
		duration     time.Duration
		hashedSecret string
		data         string
	}
)

func newContractUnlockHash(condition types.AtomicSwapCondition) types.UnlockHash {
	return types.NewAtomicSwapInputLock(condition).UnlockHash()
}

func atomicswapcreatecmd(amount, dest, src string) {
	var (
		condition types.AtomicSwapCondition
		secret    types.AtomicSwapSecret
	)

	cc, err := NewCurrencyConvertor(_CurrencyUnits)
	if err != nil {
		Die(err)
	}
	hastings, err := cc.ParseCoinString(amount)
	if err != nil {
		fmt.Fprintln(os.Stderr, cc.CoinArgDescription("amount"))
		Die("Could not parse amount:", err)
	}

	err = condition.Receiver.LoadString(dest)
	if err != nil {
		Die("Could not parse destination address (unlock hash):", err)
	}
	err = condition.Sender.LoadString(src)
	if err != nil {
		Die("Could not parse sender address (unlock hash):", err)
	}
	if hsl := len(createSwapContractConfig.hashedSecret); hsl == 0 {
		_, err := rand.Read(secret[:])
		if err != nil {
			Die("Could not read random secret:", err)
		}
		condition.HashedSecret = sha256.Sum256(secret[:])
		createSwapContractConfig.hashedSecret = hex.EncodeToString(condition.HashedSecret[:])
	} else if hsl != types.AtomicSwapHashedSecretLen*2 {
		Die("Invalid hashed secret length")
	} else {
		_, err := hex.Decode(condition.HashedSecret[:], []byte(createSwapContractConfig.hashedSecret))
		if err != nil {
			Die("Invalid hashed secret:", err)
		}
	}
	if createSwapContractConfig.duration == 0 {
		Die("Duration is required")
	}
	condition.TimeLock = types.OffsetTimestamp(createSwapContractConfig.duration)
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

	fmt.Println("Atomic Swap Contract Created!")
	fmt.Println("transaction is in progress...")
	fmt.Println("Contract Information:")
	fmt.Printf("\n  \tTransactionID: %s\n  \tUnlockHash: %s\n  \tAmount: %s\n",
		response.TransactionID.String(), unlockHash.String(), amount)
	fmt.Printf("  \tSender: %s\n  \tReceiver: %s\n",
		condition.Sender.String(), condition.Receiver.String())
	fmt.Printf("\n  \tTime Lock: %d (%s)\n  \tHashed Secret: %s\n",
		condition.TimeLock, condition.TimeLock.String(),
		createSwapContractConfig.hashedSecret)
	if secret != (types.AtomicSwapSecret{}) {
		fmt.Printf("  \tSecret: %s\n", hex.EncodeToString(secret[:]))
	}
}

func init() {
	atomicSwapCreateCmd.Flags().StringVarP(
		&createSwapContractConfig.hashedSecret, "hashedsecret", "p",
		"", "the hex-encoded hashed secret to use for this contract, if none is given one will be generated for you")
	atomicSwapCreateCmd.Flags().DurationVarP(
		&createSwapContractConfig.duration, "duration", "d",
		time.Hour*48, "the duration of the atomic swap contract, the amount of time the receiver has to collect")
	atomicSwapCreateCmd.Flags().StringVar(
		&createSwapContractConfig.data, "data",
		"", "optional data you can attach to the atomic swap contract's transaction")
}
