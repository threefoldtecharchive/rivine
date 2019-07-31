package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/pkg/cli"
	"github.com/threefoldtech/rivine/types"
)

func createAtomicSwapCmd(client *CommandLineClient) *cobra.Command {
	atomicSwapCmd := &atomicSwapCmd{cli: client}

	// create root consensus command and all subs
	var (
		rootCmd = &cobra.Command{
			Use:   "atomicswap",
			Short: "Create and interact with atomic swap contracts.",
			Long:  "Create and audit atomic swap contracts, as well as redeem money from them.",
		}

		participateCmd = &cobra.Command{
			Use:   "participate <initiator address> <amount> <secret hash>",
			Short: "Create an atomic swap contract as participant.",
			Long: `Create an atomic swap contract as a participant,
		using the secret hash given by the initiator.
		
		Returned status codes:
		
		  0: contract created successfully as participant
		  1: generic error, automatically recovering is not possible or recommended
		  3: command cancelled by user
		  64: misusage of the command, see --help on how to use the command
		
		Example STDOUT output when using the '--encoding json' flag:
		
		  {
			"coins": "100000000000", // integer, expressed in the lowest coin unit
			"contract": { // defines the conditions of the contract
			  // address of the creator (initiator) from the initiator contract, used to refund
			  "sender": "01b49da2ff193f46ee0fc684d7a6121a8b8e324144dffc7327471a4da79f1730960edcb2ce737f",
			  // address of the recipient (participant) from the initiator contract, used to redeem
			  "receiver": "019e9b6f2d43a44046b62836ce8d75c935ff66cbba1e624b3e9755b98ac176a08dac5267b2c8ee",
			  // hashedsecret, the sha256 hash of the secret that has to be given when redeemed
			  "hashedsecret": "c22267f0f118282b15098e0b5e3a8027af64f0cce7afa19b274abb21b5555626",
			  // defines when the contract can be refunded
			  // (redeeming it will still be possible, as long as it hasn't been refunded yet)
			  "timelock": 1530169858
			},
			// unlock hash of the contract condition
			"contractid": "020ce1011596c7bf7aa63e15a6b3d1eb44087358b0670efa50d34d0509625b13d2b82887cd5bbb",
			// crypto-random generated 32-byte secret
			"secret": "6f0e3b2cdd82da7c4ddc20f03be25765fb885fb59af316cb3bbf8649c82d046d",
			// ID of the unspend output which was created to pair the atomic swap contract with a coin value
			"outputid": "4abc9c18e03b0e9f35e636d8294c8de1fd823dcaa17bce115d1452df6198cdd7",
			// ID of the transaction that contained the created coin output which contains the atomic swap contract
			"transactionid": "3faf9e80cb33b572830824b2b3341c4f0f86ca34fcad1e6604ad112c008f5fbc"
		  }
		
		Note that this output is only returned in case the command
		was successful, and thus exited with status code 0.
		`,
			Run: Wrap(atomicSwapCmd.participateCmd),
		}

		initiateCmd = &cobra.Command{
			Use:   "initiate <participant address> <amount>",
			Short: "Create an atomic swap contract as initiator.",
			Long: `Create an atomic swap contract as an initiator,
		randonly generating a secret for you, and deriving the secret hash from it.
		The used secret is returned through the STDOUT with the rest of the contract details.
		
		Returned status codes:
		
		  0: contract created successfully as initiator
		  1: generic error, automatically recovering is not possible or recommended
		  3: command cancelled by user
		  64: misusage of the command, see --help on how to use the command
		
		Example output when using the '--encoding json' flag:
		
		  {
			"coins": "100000000000", // integer, expressed in the lowest coin unit
			"contract": { // defines the conditions of the contract
			  // address of the creator (participant) from the participation contract, used to refund
			  "sender": "01b49da2ff193f46ee0fc684d7a6121a8b8e324144dffc7327471a4da79f1730960edcb2ce737f",
			  // address of the recipient (initiator) from the participation contract, used to redeem
			  "receiver": "019e9b6f2d43a44046b62836ce8d75c935ff66cbba1e624b3e9755b98ac176a08dac5267b2c8ee",
			  // hashedsecret, the sha256 hash of the secret that has to be given when redeemed
			  "hashedsecret": "c22267f0f118282b15098e0b5e3a8027af64f0cce7afa19b274abb21b5555626",
			  // defines when the contract can be refunded
			  // (redeeming it will still be possible, as long as it hasn't been refunded yet)
			  "timelock": 1530169858
			},
			// unlock hash of the contract condition
			"contractid": "020ce1011596c7bf7aa63e15a6b3d1eb44087358b0670efa50d34d0509625b13d2b82887cd5bbb",
			// ID of the unspend output which was created to pair the atomic swap contract with a coin value
			"outputid": "4abc9c18e03b0e9f35e636d8294c8de1fd823dcaa17bce115d1452df6198cdd7",
			// ID of the transaction that contained the created coin output which contains the atomic swap contract
			"transactionid": "3faf9e80cb33b572830824b2b3341c4f0f86ca34fcad1e6604ad112c008f5fbc"
		  }
		
		Note that this output is only returned in case the command
		was successful, and thus exited with status code 0.
		`,
			Run: Wrap(atomicSwapCmd.initiateCmd),
		}

		auditCmd = &cobra.Command{
			Use:   "auditcontract outputid [transactionid|jsonTransaction]",
			Short: "Audit a created atomic swap contract.",
			Long: `Audit a created atomic swap contract.
		
		Run a full audit by giving the outputid,
		in which case the outputid is looked up as an unspent atomic swap contract,
		either confirmed using the consensus or unconfirmed using the transaction pool.
		Optionally the transactionid can be given in order to speed up the process in the latter case.
		
		Run a quick audit by giving the outoutid and a raw JSON-encoded transaction.
		A quick audit only validates that the atomic swap contract exists within the
		transaction as the condition of the output identified by the given outputid.
		It does not audt whether the contract is actually created on the blockchan,
		confirmed or not.
		
		Optionally the participant's address, currency amount and secret hash are validated,
		by giving one, some or all of them as flag arguments, for both quick and full audits.
		
		Returned status codes:
		
		  0: contract found and validated successfully
		  1: generic error, automatically recovering is not possible or recommended
		  2: contract not found
		  3: command cancelled by user
		  4: temporary error: contract was found and valid but not yet confirmed
		  64: misusage of the command, see --help on how to use the command
		  128: contract invalid compared to the given criteria
		
		Example output when using the '--encoding json' flag:
		
		  {
			"coins": "100000000000", // integer, expressed in the lowest coin unit
			"contract": { // defines the conditions of the contract
			  // address of the creator (sender) from the atomic swap contract, used to refund
			  "sender": "01b49da2ff193f46ee0fc684d7a6121a8b8e324144dffc7327471a4da79f1730960edcb2ce737f",
			  // address of the recipient (receiver) from the atomic swap contract, used to redeem
			  "receiver": "019e9b6f2d43a44046b62836ce8d75c935ff66cbba1e624b3e9755b98ac176a08dac5267b2c8ee",
			  // hashedsecret, the sha256 hash of the secret that has to be given when redeemed
			  "hashedsecret": "c22267f0f118282b15098e0b5e3a8027af64f0cce7afa19b274abb21b5555626",
			  // defines when the contract can be refunded
			  // (redeeming it will still be possible, as long as it hasn't been refunded yet)
			  "timelock": 1530169858
			}
		  }
		
		This output is always returned when the contract was found,
		even if it doesn't pass the audit.
		`,
			Run: atomicSwapCmd.auditCmd,
		}

		extractSecretCmd = &cobra.Command{
			Use:   "extractsecret transactionid [outputid]",
			Short: "Extract the secret from a redeemed swap contract.",
			Long: `Extract the secret from a redeemed atomic swap contract.
		
		Look for a transaction in the consensus set, using the given transactionID.
		The transaction has to contain at least one atomic swap contract fulfillment.
		If an outputID is given, the (coin) input, from which the secret is to be extracted,
		has to have the given outputID as parent ID, otherwise the first input is used,
		which has an atomic swap contract fulfillment.
		
		If it was spend as a refund, this comment will exit with an error,
		and no secret will be extracted.
		
		Optionally, the extracted secret is validated,
		by comparing its hashed version to the secret hash given using the --secrethash flag.
		
		Returned status codes:
		
		  0: contract found and secret extracted successfully
		  1: generic error, automatically recovering is not possible or recommended
		  2: contract not found
		  3: command cancelled by user
		  64: misusage of the command, see --help on how to use the command
		  128: contract invalid compared to the given criteria
		
		Example output when using the '--encoding json' flag:
		
		  {
			// the secret that was used to redeem the funds,
			// that were locked in the atomic swap contract
			"secret": "6f0e3b2cdd82da7c4ddc20f03be25765fb885fb59af316cb3bbf8649c82d046d"
		  }
		
		Note that this output is only returned in case the command
		was successful, and thus exited with status code 0.
		`,
			Run: atomicSwapCmd.extractSecretCmd,
		}

		redeemCmd = &cobra.Command{
			Use:   "redeem outputid secret",
			Short: "Redeem the coins locked in an atomic swap contract.",
			Long: `Redeem the coins locked in an atomic swap contract intended for you.
		
		Returned status codes:
		
		  0: contract found and redeemed successfully
		  1: generic error, automatically recovering is not possible or recommended
		  2: contract not found
		  3: command cancelled by user
		  64: misusage of the command, see --help on how to use the command
		
		Example output when using the '--encoding json' flag:
		
		  {
			// id of the transaction that was created when redeeming the locked funds
			"transactionid": "9869bdb633c18fe28553c594cf5d0931d4eab0929e4aa8e6fc9f1a1250c61103"
		  }
		
		Note that this output is only returned in case the command
		was successful, and thus exited with status code 0.
		`,
			Run: Wrap(atomicSwapCmd.redeemCmd),
		}

		refundCmd = &cobra.Command{
			Use:   "refund outputid",
			Short: "Refund the coins locked in an atomic swap contract.",
			Long: `Refund the coins locked in an atomic swap contract created by you.
		
		Returned status codes:
		
		  0: contract found and refunded successfully
		  1: generic error, automatically recovering is not possible or recommended
		  2: contract not found
		  3: command cancelled by user
		  64: misusage of the command, see --help on how to use the command
		
		Example output when using the '--encoding json' flag:
		
		  {
			// id of the transaction that was created when refunding the locked funds
			"transactionid": "9869bdb633c18fe28553c594cf5d0931d4eab0929e4aa8e6fc9f1a1250c61103"
		  }
		
		Note that this output is only returned in case the command
		was successful, and thus exited with status code 0.
		`,
			Run: Wrap(atomicSwapCmd.refundCmd),
		}
	)
	rootCmd.AddCommand(
		participateCmd,
		initiateCmd,
		auditCmd,
		extractSecretCmd,
		redeemCmd,
		refundCmd,
	)

	// create flags
	rootCmd.PersistentFlags().BoolVarP(&atomicSwapCmd.rootCfg.YesToAll, "yes", "y", false,
		"answer 'yes' to all yes/no questions without asking explicitly")
	rootCmd.PersistentFlags().Var(
		cli.NewEncodingTypeFlag(0, &atomicSwapCmd.rootCfg.EncodingType, cli.EncodingTypeHuman|cli.EncodingTypeJSON),
		"encoding", cli.EncodingTypeFlagDescription(cli.EncodingTypeHuman|cli.EncodingTypeJSON))

	participateCmd.Flags().DurationVarP(
		&atomicSwapCmd.participateCfg.Duration, "duration", "d",
		time.Hour*24, "the duration of the atomic swap contract, the amount of time the initiator has to collect")
	participateCmd.Flags().Var(cli.StringLoaderFlag{StringLoader: &atomicSwapCmd.participateCfg.SourceUnlockHash}, "initiator",
		"optionally define a wallet address (unlockhash) that is to be used for refunding purposes, one will be generated for you if none is given")

	initiateCmd.Flags().DurationVarP(
		&atomicSwapCmd.initiateCfg.Duration, "duration", "d",
		time.Hour*48, "the duration of the atomic swap contract, the amount of time the participant has to collect")
	initiateCmd.Flags().Var(cli.StringLoaderFlag{StringLoader: &atomicSwapCmd.initiateCfg.SourceUnlockHash}, "initiator",
		"optionally define a wallet address (unlockhash) that is to be used for refunding purposes, one will be generated for you if none is given")

	auditCmd.Flags().Var(
		cli.StringLoaderFlag{StringLoader: &atomicSwapCmd.auditCfg.HashedSecret}, "secrethash",
		"optionally validate the secret of the found atomic swap contract condition by comparing its hashed version with this secret hash")
	auditCmd.Flags().Var(
		cli.StringLoaderFlag{StringLoader: &atomicSwapCmd.auditCfg.ReceiverAddress}, "receiver",
		"optionally validate the given receiver's address (unlockhash) to the one found in the atomic swap contract condition")
	auditCmd.Flags().StringVar(
		&atomicSwapCmd.auditCfg.CoinAmountString, "amount", "",
		"optionally validate the given coin amount to the one found in the unspent coin output")
	auditCmd.Flags().DurationVar(
		&atomicSwapCmd.auditCfg.MinDurationLeft, "min-duration", 0,
		"optionally validate the given contract has sufficient duration left, as defined by the timelock in the found atomic swap contract condition")

	extractSecretCmd.Flags().Var(
		cli.StringLoaderFlag{StringLoader: &atomicSwapCmd.extractSecretCfg.HashedSecret}, "secrethash",
		"optionally validate the secret of the found atomic swap contract condition by comparing its hashed version with this secret hash")

	// return root command
	return rootCmd
}

type atomicSwapCmd struct {
	cli *CommandLineClient

	rootCfg struct {
		EncodingType cli.EncodingType
		YesToAll     bool
	}
	participateCfg struct {
		Duration         time.Duration
		SourceUnlockHash types.UnlockHash
	}
	initiateCfg struct {
		Duration         time.Duration
		SourceUnlockHash types.UnlockHash
	}
	auditCfg struct {
		ReceiverAddress  types.UnlockHash
		CoinAmountString string
		HashedSecret     types.AtomicSwapHashedSecret
		MinDurationLeft  time.Duration
	}
	extractSecretCfg struct {
		HashedSecret types.AtomicSwapHashedSecret
	}
}

type (
	// AtomicSwapOutputCreation represents the formatted output
	// of the atomic swap creation commands (initiate and participate).
	AtomicSwapOutputCreation struct {
		Coins         types.Currency            `json:"coins"`
		Contract      types.AtomicSwapCondition `json:"contract"`
		ContractID    types.UnlockHash          `json:"contractid"`
		Secret        *types.AtomicSwapSecret   `json:"secret,omitempty"`
		OutputID      types.CoinOutputID        `json:"outputid"`
		TransactionID types.TransactionID       `json:"transactionid"`
	}
	// AtomicSwapOutputAudit represents the formatted output
	// of the atomic swap audit command
	AtomicSwapOutputAudit struct {
		Coins    types.Currency            `json:"coins"`
		Contract types.AtomicSwapCondition `json:"contract"`
	}
	// AtomicSwapOutputExtractSecret represents the formatted output
	// of the atomic swap extract secret command
	AtomicSwapOutputExtractSecret struct {
		Secret types.AtomicSwapSecret `json:"secret"`
	}
	// AtomicSwapOutputSpendContract represents the formatted output
	// of the atomic swap spend commands (redeem and refund)
	AtomicSwapOutputSpendContract struct {
		TransactionID types.TransactionID `json:"transactionid"`
	}
)

func (atomicSwapCmd *atomicSwapCmd) participateCmd(participantAddress, amount, hashedSecret string) {
	// parse hastings
	hastings := parseCoinArg(atomicSwapCmd.cli.CreateCurrencyConvertor(), amount)

	// parse receiver (=participant) and sender (=initiator)
	var (
		receiver, sender types.UnlockHash
	)
	err := receiver.LoadString(participantAddress)
	if err != nil {
		cli.DieWithExitCode(cli.ExitCodeUsage, "failed to parse participant address (unlock hash):", err)
	}
	if atomicSwapCmd.participateCfg.SourceUnlockHash.Type != 0 {
		// use the hash given by the user explicitly
		sender = atomicSwapCmd.participateCfg.SourceUnlockHash
	} else {
		// get new one from the wallet
		resp := new(api.WalletAddressGET)
		err := atomicSwapCmd.cli.GetAPI("/wallet/address", resp)
		if err != nil {
			cli.DieWithError("failed to generate new address:", err)
		}
		sender = resp.Address
	}

	// parse secret hash
	if hsl := len(hashedSecret); hsl != types.AtomicSwapHashedSecretLen*2 {
		cli.DieWithExitCode(cli.ExitCodeUsage, "invalid secret hash length")
	}
	var hash types.AtomicSwapHashedSecret
	_, err = hex.Decode(hash[:], []byte(hashedSecret))
	if err != nil {
		cli.DieWithExitCode(cli.ExitCodeUsage, "invalid secret hash:", err)
	}

	// create the contract
	atomicSwapCmd.createAtomicSwapContract(hastings, sender, receiver, hash, atomicSwapCmd.participateCfg.Duration)
}

func (atomicSwapCmd *atomicSwapCmd) initiateCmd(participantAddress, amount string) {
	// parse hastings
	hastings := parseCoinArg(atomicSwapCmd.cli.CreateCurrencyConvertor(), amount)

	// parse receiver (=participant) and sender (=initiator)
	var (
		receiver, sender types.UnlockHash
	)
	err := receiver.LoadString(participantAddress)
	if err != nil {
		cli.DieWithExitCode(cli.ExitCodeUsage, "failed to parse participant address (unlock hash):", err)
	}
	if atomicSwapCmd.initiateCfg.SourceUnlockHash.Type != 0 {
		// use the hash given by the user explicitly
		sender = atomicSwapCmd.initiateCfg.SourceUnlockHash
	} else {
		// get new one from the wallet
		resp := new(api.WalletAddressGET)
		err := atomicSwapCmd.cli.GetAPI("/wallet/address", resp)
		if err != nil {
			cli.DieWithError("failed to generate new address:", err)
		}
		sender = resp.Address
	}

	// create the contract
	atomicSwapCmd.createAtomicSwapContract(hastings, sender, receiver,
		types.AtomicSwapHashedSecret{}, atomicSwapCmd.initiateCfg.Duration)
}

func (atomicSwapCmd *atomicSwapCmd) createAtomicSwapContract(hastings types.Currency, sender, receiver types.UnlockHash, hash types.AtomicSwapHashedSecret, duration time.Duration) {
	if hastings.Cmp(atomicSwapCmd.cli.Config.MinimumTransactionFee) != 1 {
		cli.DieWithExitCode(cli.ExitCodeUsage, "an atomic swap contract has to have a coin value higher than the minimum transaction fee of 1")
	}

	var (
		err    error
		secret types.AtomicSwapSecret
	)

	if hash == (types.AtomicSwapHashedSecret{}) {
		secret, err = types.NewAtomicSwapSecret()
		if err != nil {
			cli.Die("failed to crypto-generate secret:", err)
		}
		hash = types.NewAtomicSwapHashedSecret(secret)
	}

	if duration == 0 {
		cli.DieWithExitCode(cli.ExitCodeUsage, "duration is required and has to be greater than 0")
	}

	condition := types.AtomicSwapCondition{
		Sender:       sender,
		Receiver:     receiver,
		HashedSecret: hash,
		TimeLock:     types.OffsetTimestamp(duration),
	}
	if !atomicSwapCmd.rootCfg.YesToAll {
		// print contract for review
		atomicSwapCmd.printContractInfo(os.Stderr, hastings, condition, secret)
		// ensure user wants to continue with creating the contract as it is (aka publishing it)
		if !askYesNoQuestion("Publish atomic swap transaction?") {
			cli.DieWithExitCode(cli.ExitCodeCancelled, "cancelled atomic swap contract")
		}
	}
	// publish contract
	body, err := json.Marshal(api.WalletTransactionPOST{
		Condition: types.NewCondition(&condition),
		Amount:    hastings,
	})
	if err != nil {
		cli.Die("failed to create/marshal JSON body:", err)
	}
	var response api.WalletTransactionPOSTResponse
	err = atomicSwapCmd.cli.PostResp("/wallet/transaction", string(body), &response)
	if err != nil {
		cli.DieWithError("failed to create transaction:", err)
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
		cli.Die("didn't find atomic swap contract registered in any returned coin output")
	}

	if atomicSwapCmd.rootCfg.EncodingType == cli.EncodingTypeJSON {
		// if encoding type is JSON, simply print all information as JSON
		output := AtomicSwapOutputCreation{
			Coins:         hastings,
			Contract:      condition,
			ContractID:    condition.UnlockHash(),
			OutputID:      response.Transaction.CoinOutputID(uint64(coinOutputIndex)),
			TransactionID: response.Transaction.ID(),
		}
		if secret != (types.AtomicSwapSecret{}) {
			output.Secret = &secret
		}
		json.NewEncoder(os.Stdout).Encode(output)
		return
	}

	// otherwise print it for a human, in a more verbose and friendly way
	fmt.Println("")
	fmt.Println("published contract transaction")
	fmt.Println("")
	fmt.Println("OutputID:", response.Transaction.CoinOutputID(uint64(coinOutputIndex)))
	fmt.Println("TransactionID:", response.Transaction.ID())
	fmt.Println("")
	fmt.Println("Contract Info:")
	fmt.Println("")
	atomicSwapCmd.printContractInfo(os.Stdout, hastings, condition, secret)
}

func (atomicSwapCmd *atomicSwapCmd) auditCmd(cmd *cobra.Command, args []string) {
	argn := len(args)
	if argn < 1 || argn > 2 {
		cmd.UsageFunc()(cmd)
		os.Exit(cli.ExitCodeUsage)
	}

	var (
		outputID              types.CoinOutputID
		transactionID         types.TransactionID
		transactionIDGiven    bool
		unspentCoinOutputResp api.ConsensusGetUnspentCoinOutput
		txnPoolGetResp        api.TransactionPoolGET
	)

	err := outputID.LoadString(args[0])
	if err != nil {
		cli.DieWithExitCode(cli.ExitCodeUsage, "failed to parse required positional (coin) outputID argument:", err)
	}
	if argn == 2 {
		err = transactionID.LoadString(args[1])
		if err != nil {
			// try to parse it as a transaction instead
			var txn types.Transaction
			err = txn.UnmarshalJSON([]byte(args[1]))
			if err != nil {
				cli.DieWithExitCode(cli.ExitCodeUsage,
					errors.New("second position argument is optional and has to be either a transactionID "+
						"or a raw json-encoded transaction: "+err.Error()))
			}
			for idx, co := range txn.CoinOutputs {
				coid := txn.CoinOutputID(uint64(idx))
				if coid == outputID {
					atomicSwapCmd.auditAtomicSwapContract(co, auditSourceUser)
					return
				}
			}
			goto failure
		}
		transactionIDGiven = true
	}

	// get unspent output from consensus
	err = atomicSwapCmd.cli.GetAPI("/consensus/unspent/coinoutputs/"+outputID.String(), &unspentCoinOutputResp)
	if err == nil {
		atomicSwapCmd.auditAtomicSwapContract(unspentCoinOutputResp.Output, auditSourceConsensus)
		return
	}
	if err != api.ErrStatusNotFound {
		cli.Die("unexpected error occurred while getting (unspent) coin output from consensus:", err)
	}
	// output couldn't be found as an unspent coin output
	// therefore the last positive hope is if it wasn't yet part of the transaction pool
	err = atomicSwapCmd.cli.GetAPI("/transactionpool/transactions", &txnPoolGetResp)
	if err != nil {
		cli.DieWithExitCode(cli.ExitCodeNotFound,
			"contract no found as part of an unspent coin output, and getting unconfirmed transactions from the transactionpool failed:", err)
	}
	for _, txn := range txnPoolGetResp.Transactions {
		var lastTransaction bool
		if transactionIDGiven {
			if txn.ID() != transactionID {
				continue
			}
			lastTransaction = true
		}
		for idx, co := range txn.CoinOutputs {
			coid := txn.CoinOutputID(uint64(idx))
			if coid == outputID {
				atomicSwapCmd.auditAtomicSwapContract(co, auditSourceTransactionPool)
				return
			}
		}
		if lastTransaction {
			// if a transactionID and this was the txn we were looking for
			// we know that the outputID will not be found for the given transactionID
			goto failure
		}
	}
	// given that we could have just hit the unlucky window,
	// where the block might have been just created in between our 2 calls,
	// let's try to get the coin output one last time from the consensus
	// contract couldn't be found as either
	err = atomicSwapCmd.cli.GetAPI("/consensus/unspent/coinoutputs/"+outputID.String(), &unspentCoinOutputResp)
	if err == nil {
		atomicSwapCmd.auditAtomicSwapContract(unspentCoinOutputResp.Output, auditSourceConsensus)
		return
	}
	if err != api.ErrStatusNotFound {
		cli.Die("unexpected error occurred while getting (unspent) coin output from consensus:", err)
	}
failure:
	fmt.Fprintf(os.Stderr, `Failed to find atomic swap contract using outputid %s.
It wasn't found as part of a confirmed unspent coin output in the consensus set,
neither was it found as an unconfirmed coin output in the transaction pool.

This might mean one of two things:

+ Most likely it means that the given outputID is invalid;
+ Another possibility is that the atomic swap contract was already refunded or redeemed,
  this can be confirmed by looking the outputID up in a local, remote or public explorer;
`, outputID)
	cli.DieWithExitCode(cli.ExitCodeNotFound, "no unspent coin output could be found for ID "+outputID.String())
}

type auditSource uint8

const (
	auditSourceConsensus auditSource = iota
	auditSourceTransactionPool
	auditSourceUser
)

func (atomicSwapCmd *atomicSwapCmd) auditAtomicSwapContract(co types.CoinOutput, source auditSource) {
	currencyConverter := atomicSwapCmd.cli.CreateCurrencyConvertor()

	condition, ok := co.Condition.Condition.(*types.AtomicSwapCondition)
	if !ok {
		cli.Die(fmt.Sprintf(
			"received unexpected condition of type %T, while type *types.AtomicSwapCondition was expected in order to be able to audit",
			co.Condition.Condition))
	}
	durationLeft := time.Unix(int64(condition.TimeLock), 0).Sub(computeTimeNow())

	if atomicSwapCmd.rootCfg.EncodingType == cli.EncodingTypeJSON {
		json.NewEncoder(os.Stdout).Encode(AtomicSwapOutputAudit{
			Coins:    co.Value,
			Contract: *condition,
		})
	} else {
		fmt.Printf(`Atomic Swap Contract (condition) found:

Contract value: %s

Receiver's address: %s
Sender's (contract creator) address: %s
Secret Hash: %s
TimeLock: %[5]d (%[5]s)
TimeLock reached in: %s

`, currencyConverter.ToCoinStringWithUnit(co.Value), condition.Receiver,
			condition.Sender, condition.HashedSecret, condition.TimeLock, durationLeft)
	}

	var invalidContract bool
	if atomicSwapCmd.auditCfg.CoinAmountString != "" && atomicSwapCmd.auditCfg.CoinAmountString != "0" {
		amount, err := currencyConverter.ParseCoinString(atomicSwapCmd.auditCfg.CoinAmountString)
		if err != nil {
			cli.DieWithError("failed to parse amount string: ", err)
		}
		// optionally validate coin amount
		if !amount.Equals(co.Value) {
			invalidContract = true
			fmt.Fprintln(os.Stderr, "unspent out's value "+
				currencyConverter.ToCoinStringWithUnit(co.Value)+
				" does not match the expected value "+
				currencyConverter.ToCoinStringWithUnit(amount))
		}
	}
	if atomicSwapCmd.auditCfg.HashedSecret != (types.AtomicSwapHashedSecret{}) {
		// optionally validate hashed secret
		if atomicSwapCmd.auditCfg.HashedSecret != condition.HashedSecret {
			invalidContract = true
			fmt.Fprintln(os.Stderr, "found contract's secret hash "+
				condition.HashedSecret.String()+
				" does not match the expected secret hash "+
				atomicSwapCmd.auditCfg.HashedSecret.String())
		}
	}
	if atomicSwapCmd.auditCfg.ReceiverAddress != (types.UnlockHash{}) {
		// optionally validate receiver's address (unlockhash)
		if atomicSwapCmd.auditCfg.ReceiverAddress.Cmp(condition.Receiver) != 0 {
			invalidContract = true
			fmt.Fprintln(os.Stderr, "found contract's receiver's address "+
				condition.Receiver.String()+
				" does not match the expected receiver's address "+
				atomicSwapCmd.auditCfg.ReceiverAddress.String())
		}
	}
	if atomicSwapCmd.auditCfg.MinDurationLeft != 0 {
		// optionally validate locktime
		if durationLeft < atomicSwapCmd.auditCfg.MinDurationLeft {
			invalidContract = true
			fmt.Fprintln(os.Stderr, "found contract's duration left "+
				durationLeft.String()+
				" is not sufficient, when compared the expected duration left of "+
				atomicSwapCmd.auditCfg.MinDurationLeft.String())
		}
	}
	if invalidContract {
		cli.DieWithExitCode(AuditContractExitCodeInvalidContract,
			"found Atomic Swap Contract does not meet the given expectations")
	}
	fmt.Fprintln(os.Stderr, "found Atomic Swap Contract is valid")
	switch source {
	case auditSourceTransactionPool:
		fmt.Fprintln(os.Stderr, "NOTE: this contract is still in the transaction pool and thus unconfirmed")
		cli.DieWithExitCode(cli.ExitCodeTemporaryError, "contract is not yet confirmed")
	case auditSourceUser:
		fmt.Fprintln(os.Stderr, `NOTE: this contract was given as part of a raw JSON-encoded transaction,
and it is therefore possible that this contract is not yet created,
confirmed or not`)
	}
}

const (
	// AuditContractExitCodeInvalidContract is returned as exit code,
	// in case the contract was found, but failed to validate against the given validation flags.
	AuditContractExitCodeInvalidContract = 128
)

// extractsecret transactionid [outputid]
func (atomicSwapCmd *atomicSwapCmd) extractSecretCmd(cmd *cobra.Command, args []string) {
	argn := len(args)
	if argn < 1 || argn > 2 {
		cmd.UsageFunc()(cmd)
		os.Exit(cli.ExitCodeUsage)
	}

	var (
		txnID         types.TransactionID
		outputID      types.CoinOutputID
		outputIDGiven bool
		secret        types.AtomicSwapSecret
	)
	err := txnID.LoadString(args[0])
	if err != nil {
		cli.Die("failed to parse first argment as a transaction (long) ID:", err)
	}
	if argn == 2 {
		err = outputID.LoadString(args[1])
		if err != nil {
			cli.Die("failed to parse optional second argment as a coin outputID:", err)
		}
		outputIDGiven = true
	}

	var (
		txnPoolGetResp api.TransactionPoolGET
		txnResp        api.ConsensusGetTransaction
	)

	// first try to get the transaction from transaction pool,
	// this is OK for extracting the secret, as the secret will already be validated
	// against the condition's secret hash, prior to being able to add it to the transaction pool.
	// ALl we care here is extracting the secret, as soon as possible.
	err = atomicSwapCmd.cli.GetAPI("/transactionpool/transactions", &txnPoolGetResp)
	if err != nil {
		fmt.Fprintln(os.Stderr, "getting unconfirmed transactions from the transactionpool failed: "+err.Error())
	}
	for _, txn := range txnPoolGetResp.Transactions {
		for _, ci := range txn.CoinInputs {
			if outputIDGiven && ci.ParentID != outputID {
				continue
			}
			if ft := ci.Fulfillment.FulfillmentType(); ft != types.FulfillmentTypeAtomicSwap {
				if outputIDGiven && ci.ParentID == outputID {
					cli.Die(fmt.Sprintf(
						"received unexpected fulfillment type of type %d (%T)", ft, ci.Fulfillment.Fulfillment))
				}
				continue
			}
			getter, ok := ci.Fulfillment.Fulfillment.(atomicSwapSecretGetter)
			if !ok {
				cli.Die(fmt.Sprintf(
					"received unexpected fulfillment type of type %T", ci.Fulfillment.Fulfillment))
			}
			secret = getter.AtomicSwapSecret()
			goto secretCheck
		}
	}

	// get transaction from consensus, assuming that the transactionID is valid,
	// it should mean that the transaction is already part of a created block
	err = atomicSwapCmd.cli.GetAPI("/consensus/transactions/"+txnID.String(), &txnResp)
	if err != nil {
		if err == api.ErrStatusNotFound {
			cli.DieWithExitCode(cli.ExitCodeUsage,
				"failed to find transaction:", err, "; Long ID:", txnID)
		}
		cli.Die("failed to get transaction:", err, "; Long ID:", txnID)
	}

	// get the secret from any of the inputs within this transaction, if possible,
	// or from an input which doesn't just define the right fulfillment but also has the right parentID
	for _, ci := range txnResp.CoinInputs {
		if outputIDGiven && ci.ParentID != outputID {
			continue
		}
		if ft := ci.Fulfillment.FulfillmentType(); ft != types.FulfillmentTypeAtomicSwap {
			if outputIDGiven && ci.ParentID == outputID {
				cli.Die(fmt.Sprintf(
					"received unexpected fulfillment type of type %d (%T)", ft, ci.Fulfillment.Fulfillment))
			}
			continue
		}
		getter, ok := ci.Fulfillment.Fulfillment.(atomicSwapSecretGetter)
		if !ok {
			cli.Die(fmt.Sprintf(
				"received unexpected fulfillment type of type %T", ci.Fulfillment.Fulfillment))
		}
		secret = getter.AtomicSwapSecret()
		break
	}

secretCheck:
	if secret == (types.AtomicSwapSecret{}) {
		cli.DieWithExitCode(cli.ExitCodeNotFound,
			"failed to find a matching atomic swap contract fulfillment in transaction with LongID: ", txnID)
	}
	if atomicSwapCmd.extractSecretCfg.HashedSecret != (types.AtomicSwapHashedSecret{}) {
		hs := types.NewAtomicSwapHashedSecret(secret)
		if hs != atomicSwapCmd.extractSecretCfg.HashedSecret {
			cli.DieWithExitCode(AuditContractExitCodeInvalidContract,
				fmt.Sprintf("found secret %s does not match expected and given secret hash %s",
					secret, atomicSwapCmd.extractSecretCfg.HashedSecret))
		}
	}

	if atomicSwapCmd.rootCfg.EncodingType == cli.EncodingTypeJSON {
		// if encoding type is JSON, simply print all information as JSON
		json.NewEncoder(os.Stdout).Encode(AtomicSwapOutputExtractSecret{
			Secret: secret,
		})
		return
	}

	// otherwise print it for a human, in a more verbose and friendly way
	fmt.Println("atomic swap contract was redeemed")
	fmt.Println("extracted secret:", secret.String())
}

type atomicSwapSecretGetter interface {
	AtomicSwapSecret() types.AtomicSwapSecret
}

// redeem outputid secret
func (atomicSwapCmd *atomicSwapCmd) redeemCmd(outputIDStr, secretStr string) {
	var (
		err      error
		outputID types.CoinOutputID
		secret   types.AtomicSwapSecret
	)

	// parse pos args
	err = outputID.LoadString(outputIDStr)
	if err != nil {
		cli.DieWithExitCode(cli.ExitCodeUsage, "failed to parse outputid-argument:", err)
	}
	err = secret.LoadString(secretStr)
	if err != nil {
		cli.DieWithExitCode(cli.ExitCodeUsage, "failed to parse secret-argument:", err)
	}
	if secret == (types.AtomicSwapSecret{}) {
		cli.DieWithExitCode(cli.ExitCodeUsage, "secret cannot be all-nil when redeeming an atomic swap contract")
	}

	atomicSwapCmd.spendAtomicSwapContract(outputID, secret)
}

// refund outputid
func (atomicSwapCmd *atomicSwapCmd) refundCmd(outputIDStr string) {
	var (
		err      error
		outputID types.CoinOutputID
	)

	// parse pos arg
	err = outputID.LoadString(outputIDStr)
	if err != nil {
		cli.DieWithExitCode(cli.ExitCodeUsage, "failed to parse outputid-argument:", err)
	}

	atomicSwapCmd.spendAtomicSwapContract(outputID, types.AtomicSwapSecret{})
}

func (atomicSwapCmd *atomicSwapCmd) spendAtomicSwapContract(outputID types.CoinOutputID, secret types.AtomicSwapSecret) {
	var (
		isSender bool
		keyWord  string // define keyword for communication purposes
	)
	if secret == (types.AtomicSwapSecret{}) {
		keyWord = "refund"
		isSender = true
	} else {
		keyWord = "redeem"
	}

	// get unspent output from consensus
	var unspentCoinOutputResp api.ConsensusGetUnspentCoinOutput
	err := atomicSwapCmd.cli.GetAPI("/consensus/unspent/coinoutputs/"+outputID.String(), &unspentCoinOutputResp)
	if err != nil {
		if err == api.ErrStatusNotFound {
			cli.DieWithExitCode(cli.ExitCodeNotFound,
				"failed to get unspent coinoutput from consensus: no output with ID "+outputID.String()+" exists")
		} else {
			cli.Die("failed to get unspent coinoutput from consensus:", err)
		}
	}

	// step 2: get correct spendable key from wallet
	if ct := unspentCoinOutputResp.Output.Condition.ConditionType(); ct != types.ConditionTypeAtomicSwap {
		cli.Die("only atomic swap conditions are supported, while referenced output is of type: ", ct)
	}
	condition, ok := unspentCoinOutputResp.Output.Condition.Condition.(*types.AtomicSwapCondition)
	if !ok {
		cli.Die(fmt.Sprintf(
			"received unexpected condition of type %T, while type *types.AtomicSwapCondition was expected",
			unspentCoinOutputResp.Output.Condition.Condition))
	}
	var ourUH types.UnlockHash
	if isSender {
		ourUH = condition.Sender
	} else {
		ourUH = condition.Receiver
	}
	pk, sk := atomicSwapCmd.getSpendableKey(ourUH)
	// quickly validate if returned sk matches the known unlock hash (sanity check)
	uh, err := types.NewPubKeyUnlockHash(pk)
	if err != nil {
		cli.Die("error while creating new public-key-based unlockhash", err)
	}
	if uh.Cmp(ourUH) != 0 {
		cli.Die("unexpected wallet public key returned:", sk)
	}

	if unspentCoinOutputResp.Output.Value.Cmp(atomicSwapCmd.cli.Config.MinimumTransactionFee) != 1 {
		cli.Die("failed to " + keyWord + " atomic swap contract, as it locks a value less than or equal to the minimum transaction fee of 1")
	}

	// step 3: confirm contract details with user, before continuing
	// print contract for review
	if !atomicSwapCmd.rootCfg.YesToAll {
		atomicSwapCmd.printContractInfo(os.Stderr, unspentCoinOutputResp.Output.Value, *condition, secret)
		// ensure user wants to continue with redeeming the contract!
		if !askYesNoQuestion("Publish atomic swap " + keyWord + " transaction?") {
			cli.DieWithExitCode(cli.ExitCodeCancelled, "atomic swap "+keyWord+" transaction cancelled")
		}
	}
	// step 4: create a transaction
	txn := types.Transaction{
		Version: atomicSwapCmd.cli.Config.DefaultTransactionVersion,
		CoinInputs: []types.CoinInput{
			{
				ParentID: outputID,
				Fulfillment: types.NewFulfillment(&types.AtomicSwapFulfillment{
					PublicKey: pk,
					Secret:    secret,
				}),
			},
		},
		CoinOutputs: []types.CoinOutput{
			{
				Condition: types.NewCondition(types.NewUnlockHashCondition(uh)),
				Value:     unspentCoinOutputResp.Output.Value.Sub(atomicSwapCmd.cli.Config.MinimumTransactionFee),
			},
		},
		MinerFees: []types.Currency{atomicSwapCmd.cli.Config.MinimumTransactionFee},
	}

	// step 5: sign transaction's only input
	err = txn.CoinInputs[0].Fulfillment.Sign(types.FulfillmentSignContext{
		ExtraObjects: []interface{}{uint64(0)},
		Transaction:  txn,
		Key:          sk,
	})
	if err != nil {
		cli.Die("failed to "+keyWord+" atomic swap's locked coins, couldn't sign transaction:", err)
	}

	// step 6: submit transaction to transaction pool and celebrate if possible
	txnid, err := atomicSwapCmd.commitTxn(txn)
	if err != nil {
		cli.Die("failed to "+keyWord+" atomic swaps locked tokens, as transaction couldn't commit:", err)
	}

	if atomicSwapCmd.rootCfg.EncodingType == cli.EncodingTypeJSON {
		// if encoding type is JSON, simply print all information as JSON
		json.NewEncoder(os.Stdout).Encode(AtomicSwapOutputSpendContract{
			TransactionID: txnid,
		})
		return
	}

	// otherwise print it for a human, in a more verbose and friendly way
	fmt.Println("")
	fmt.Println("published atomic swap " + keyWord + " transaction")
	fmt.Println("transaction ID:", txnid)
	fmt.Println(`>   Note that this does not mean for 100% you'll have the money.
> Due to potential forks, double spending, and any other possible issues your
> ` + keyWord + ` might be declined by the network. Please check the network
> (e.g. using a public explorer node or your own full node) to ensure
> your payment went through. If not, try to audit the contract (again).`)
}

// get public- and private key from wallet module
func (atomicSwapCmd *atomicSwapCmd) getSpendableKey(unlockHash types.UnlockHash) (types.PublicKey, types.ByteSlice) {
	resp := new(api.WalletKeyGet)
	err := atomicSwapCmd.cli.GetAPI("/wallet/key/"+unlockHash.String(), resp)
	if err != nil {
		cli.DieWithError("failed to get a matching wallet public/secret key pair for the given unlock hash:", err)
	}
	if isNilByteSlice(resp.PublicKey) {
		cli.Die("failed to get a wallet public key pair for the given unlock hash")
	}
	if isNilByteSlice(resp.SecretKey) {
		cli.Die("received matching public key, but no secret key was returned")
	}
	pk := types.PublicKey{
		Key: resp.PublicKey,
	}
	err = pk.Algorithm.LoadSpecifier(resp.AlgorithmSpecifier)
	if err != nil {
		cli.DieWithError("invalid public key algorithm specifier loaded:", err)
	}
	return pk, resp.SecretKey
}

func isNilByteSlice(bs types.ByteSlice) bool {
	for _, b := range bs {
		if b != 0 {
			return false
		}
	}
	return true
}

// commitTxn sends a transaction to the used node's transaction pool
func (atomicSwapCmd *atomicSwapCmd) commitTxn(txn types.Transaction) (types.TransactionID, error) {
	bodyBuff := bytes.NewBuffer(nil)
	err := json.NewEncoder(bodyBuff).Encode(&txn)
	if err != nil {
		return types.TransactionID{}, err
	}

	resp := new(api.TransactionPoolPOST)
	err = atomicSwapCmd.cli.PostResp("/transactionpool/transactions", bodyBuff.String(), resp)
	return resp.TransactionID, err
}

func (atomicSwapCmd *atomicSwapCmd) printContractInfo(w io.Writer, hastings types.Currency, condition types.AtomicSwapCondition, secret types.AtomicSwapSecret) {
	currencyConvertor := atomicSwapCmd.cli.CreateCurrencyConvertor()

	var amountStr string
	if !hastings.Equals(types.Currency{}) {
		amountStr = fmt.Sprintf(`
Contract value: %s`, currencyConvertor.ToCoinStringWithUnit(hastings))
	}

	var secretStr string
	if secret != (types.AtomicSwapSecret{}) {
		secretStr = fmt.Sprintf(`
Secret: %s`, secret)
	}

	cuh := condition.UnlockHash()

	fmt.Fprintf(w, `Contract address: %s%s
Receiver's address: %s
Sender's (contract creator) address: %s

SecretHash: %s%s

TimeLock: %[7]d (%[7]s)
TimeLock reached in: %s
`, cuh, amountStr, condition.Receiver, condition.Sender, condition.HashedSecret,
		secretStr, condition.TimeLock,
		time.Unix(int64(condition.TimeLock), 0).Sub(time.Now()))
}

func askYesNoQuestion(str string) bool {
	fmt.Fprintf(os.Stderr, "%s [Y/N] ", str)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		cli.Die("failed to scan response:", err)
	}
	response = strings.ToLower(response)
	if containsString(okayResponses, response) {
		return true
	}
	if containsString(nokayResponses, response) {
		return false
	}

	fmt.Fprintln(os.Stderr, "please answer using 'yes' or 'no'")
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

var computeTimeNow = func() time.Time {
	return time.Now()
}
