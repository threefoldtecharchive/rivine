package client

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/threefoldtech/rivine/extensions/authcointx"
	"github.com/threefoldtech/rivine/types"

	"github.com/threefoldtech/rivine/pkg/cli"
	client "github.com/threefoldtech/rivine/pkg/client"
)

// CreateWalletCmds creates the auth coin wallet root command as well as its transaction creation sub commands.
func CreateWalletCmds(client *client.CommandLineClient, conditionUpdateTransactionVersion, addressUpdateTransactionVersion types.TransactionVersion) {
	walletCmd := &walletCmd{
		cli: client,

		conditionUpdateTransactionVersion: conditionUpdateTransactionVersion,
		addressUpdateTransactionVersion:   addressUpdateTransactionVersion,
	}

	coinAuthRootCmd := &cobra.Command{
		Use:   "authcoin",
		Short: "root command for all auth coin transaction commands",
	}
	// coin auth sub cmds
	var (
		createAuthAddressUpdateTxCmd = &cobra.Command{
			Use:   "authaddresses",
			Short: "authorize and/or deauthorize the given addresses",
			Args:  cobra.MaximumNArgs(0),
			Run:   walletCmd.authAddressUpdateTxCreateCmd,
		}
		createAuthConditionUpdateTxCmd = &cobra.Command{
			Use:   "updatecondition address [address...] [signaturecount]",
			Short: "Update the auth condition by defining a single or multisig condition",
			Args:  cobra.MinimumNArgs(1),
			Run:   walletCmd.authConditionUpdateTxCreateCmd,
		}
	)

	// add commands as coin auth sub commands
	coinAuthRootCmd.AddCommand(
		createAuthAddressUpdateTxCmd,
		createAuthConditionUpdateTxCmd,
	)

	// add commands as wallet sub commands
	client.WalletCmd.AddCommand(
		coinAuthRootCmd,
	)

	// client.ExploreCmd.AddCommand(getMintConditionCmd)
	cli.ArbitraryDataFlagVar(createAuthAddressUpdateTxCmd.Flags(), &walletCmd.authAddressUpdateTxCfg.Description,
		"description", "optionally add a description to describe the reasons of auth address update, added as arbitrary data")
	createAuthAddressUpdateTxCmd.Flags().StringSliceVarP(
		&walletCmd.authAddressUpdateTxCfg.AuthAddresses,
		"auth", "e", nil, "add addresses to authorize, allowing (Enabling) the addresses to receive and send coins",
	)
	createAuthAddressUpdateTxCmd.Flags().StringSliceVarP(
		&walletCmd.authAddressUpdateTxCfg.DeauthAddresses,
		"deauth", "d", nil, "add addresses to deauthorize, no longer allowing the addresses to receive and send coins",
	)
	cli.ArbitraryDataFlagVar(createAuthConditionUpdateTxCmd.Flags(), &walletCmd.authConditionUpdateTxCfg.Description,
		"description", "optionally add a description to describe the reasons of transfer of coin auth powers, added as arbitrary data")
}

type walletCmd struct {
	cli                                                                *client.CommandLineClient
	conditionUpdateTransactionVersion, addressUpdateTransactionVersion types.TransactionVersion
	authAddressUpdateTxCfg                                             struct {
		AuthAddresses   []string
		DeauthAddresses []string
		Description     []byte
	}
	authConditionUpdateTxCfg struct {
		Description []byte
	}
}

func (walletCmd *walletCmd) authConditionUpdateTxCreateCmd(cmd *cobra.Command, args []string) {
	var (
		err       error
		condition types.UnlockConditionProxy
	)
	if len(args) == 1 {
		// create a single sig condition
		var uh types.UnlockHash
		err = uh.LoadString(args[0])
		if err != nil {
			cmd.UsageFunc()(cmd)
			cli.DieWithError("invalid address cannot be turned into an UnlockHashCondition", err)
			return
		}
		condition = types.NewCondition(types.NewUnlockHashCondition(uh))
	} else {
		// create a multi sig condition
		var sigsRequired uint64
		if len(args) > 2 {
			finalPos := len(args) - 1
			finalArg := args[finalPos]
			if u, err := strconv.ParseUint(finalArg, 10, 64); err == nil {
				sigsRequired = u
				args = args[:finalPos]
			}
		}
		addresses := make([]types.UnlockHash, len(args))
		for index, arg := range args {
			err = addresses[index].LoadString(arg)
			if err != nil {
				cli.DieWithError(fmt.Sprintf("invalid address %q cannot be used as part of a MultiSigCondition", arg), err)
			}
		}
		if sigsRequired == 0 {
			sigsRequired = uint64(len(addresses))
		}
		condition = types.NewCondition(types.NewMultiSignatureCondition(addresses, sigsRequired))
	}

	// create an auth condition update tx with a random nonce and the minimum required miner fee
	tx := authcointx.AuthConditionUpdateTransaction{
		Nonce:         types.RandomTransactionNonce(),
		AuthCondition: condition,
	}

	if n := len(walletCmd.authConditionUpdateTxCfg.Description); n > 0 {
		tx.ArbitraryData = make([]byte, n)
		copy(tx.ArbitraryData[:], walletCmd.authConditionUpdateTxCfg.Description[:])
	}

	// encode the transaction as a JSON-encoded string and print it to the STDOUT
	json.NewEncoder(os.Stdout).Encode(tx.Transaction(walletCmd.conditionUpdateTransactionVersion))
}

func (walletCmd *walletCmd) authAddressUpdateTxCreateCmd(cmd *cobra.Command, _ []string) {
	if len(walletCmd.authAddressUpdateTxCfg.AuthAddresses) == 0 && len(walletCmd.authAddressUpdateTxCfg.DeauthAddresses) == 0 {
		cmd.UsageFunc()(cmd)
		cli.Die("at least one address needs to be authorized or deauthorized, but no addresses are given")
	}

	var err error
	tx := authcointx.AuthAddressUpdateTransaction{
		Nonce: types.RandomTransactionNonce(),
	}
	// add authorized addresses
	tx.AuthAddresses = make([]types.UnlockHash, len(walletCmd.authAddressUpdateTxCfg.AuthAddresses))
	for index, address := range walletCmd.authAddressUpdateTxCfg.AuthAddresses {
		err = tx.AuthAddresses[index].LoadString(address)
		if err != nil {
			cli.DieWithError(fmt.Sprintf("invalid address %q cannot be authorized", address), err)
		}
	}
	// add deauthorized addresses
	tx.DeauthAddresses = make([]types.UnlockHash, len(walletCmd.authAddressUpdateTxCfg.DeauthAddresses))
	for index, address := range walletCmd.authAddressUpdateTxCfg.DeauthAddresses {
		err = tx.DeauthAddresses[index].LoadString(address)
		if err != nil {
			cli.DieWithError(fmt.Sprintf("invalid address %q cannot be deauthorized", address), err)
		}
	}

	// print raw transaction, ready to be signed
	err = json.NewEncoder(os.Stdout).Encode(tx.Transaction(walletCmd.addressUpdateTransactionVersion))
	if err != nil {
		panic(err)
	}
}
