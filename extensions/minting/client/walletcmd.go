package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	minting "github.com/threefoldtech/rivine/extensions/minting"
	"github.com/threefoldtech/rivine/types"

	"github.com/threefoldtech/rivine/pkg/cli"
	client "github.com/threefoldtech/rivine/pkg/client"
)

//WalletCmdsOpts defines chain-specific options for the wallet commands
type WalletCmdsOpts struct {
	CoinDestructionTxVersion types.TransactionVersion
	RequireMinerFees         bool
}

//CreateWalletCmds adds the wallet cli subcommands for the minting plugin
func CreateWalletCmds(ccli *client.CommandLineClient, mintingDefinitionTxVersion, coinCreationTxVersion types.TransactionVersion, opts *WalletCmdsOpts) error {
	bc, err := client.NewLazyBaseClientFromCommandLineClient(ccli)
	if err != nil {
		return err
	}
	walletCmd := &walletCmd{
		cli:                        ccli,
		walletClient:               client.NewWalletClient(bc),
		txPoolClient:               client.NewTransactionPoolClient(bc),
		mintingDefinitionTxVersion: mintingDefinitionTxVersion,
		coinCreationTxVersion:      coinCreationTxVersion,
	}
	if opts != nil {
		walletCmd.requireMinerFees = opts.RequireMinerFees
		walletCmd.coinDestructionTxVersion = opts.CoinDestructionTxVersion
	}

	// create root explore command and all subs
	var (
		createMinterDefinitionTxCmd = &cobra.Command{
			Use:   "minterdefinitiontransaction <dest>|<rawCondition>",
			Short: "Create a new minter definition transaction",
			Long: `Create a new minter definition transaction using the given mint condition.
The mint condition is used to overwrite the current globally defined mint condition,
and can be given as a raw output condition (or address, which resolves to a singlesignature condition).

The returned (raw) MinterDefinitionTransaction still has to be signed, prior to sending.
	`,
			Run: walletCmd.createMinterDefinitionTxCmd,
		}
		createCoinCreationTxCmd = &cobra.Command{
			Use:   "coincreationtransaction <dest>|<rawCondition> <amount> [<dest>|<rawCondition> <amount>]...",
			Short: "Create a new coin creation transaction",
			Long: `Create a new coin creation transaction using the given outputs.
The outputs can be given as a pair of value and a raw output condition (or
address, which resolves to a singlesignature condition).

Amounts have to be given expressed in the OneCoin unit, and without the unit of currency.
Decimals are possible and have to be defined using the decimal point.

The Minimum Miner Fee will be added on top of the total given amount automatically.

The returned (raw) CoinCreationTransaction still has to be signed, prior to sending.
	`,
			Run: walletCmd.createCoinCreationTxCmd,
		}
	)

	// add commands as wallet sub commands
	ccli.WalletCmd.RootCmdCreate.AddCommand(
		createMinterDefinitionTxCmd,
		createCoinCreationTxCmd,
	)

	// client.ExploreCmd.AddCommand(getMintConditionCmd)
	cli.ArbitraryDataFlagVar(createMinterDefinitionTxCmd.Flags(), &walletCmd.minterDefinitionTxCfg.Description,
		"description", "optionally add a description to describe the reasons of transfer of minting power, added as arbitrary data")
	cli.ArbitraryDataFlagVar(createCoinCreationTxCmd.Flags(), &walletCmd.coinCreationTxCfg.Description,
		"description", "optionally add a description to describe the origins of the coin creation, added as arbitrary data")

	if opts != nil && opts.CoinDestructionTxVersion > 0 {
		var (
			burnCoinsCmd = &cobra.Command{
				Use:   "coins <amount>",
				Short: "burn the given amount of coins",
				Args:  cobra.ExactArgs(1),
				Run:   walletCmd.burnCoinsCmd,
			}
		)

		// set the flags
		cli.ArbitraryDataFlagVar(burnCoinsCmd.Flags(), &walletCmd.coinDestructionTxCfg.Description,
			"description", "optionally add a description to describe the reasons of transfer of minting power, added as arbitrary data")
		burnCoinsCmd.Flags().StringVar(
			&walletCmd.coinDestructionTxCfg.RefundAddress,
			"refund-address", "", "define a custom refund address")
		burnCoinsCmd.Flags().BoolVar(
			&walletCmd.coinDestructionTxCfg.RefundAddressNew,
			"refund-address-new", false, "generate a new refund address if a refund needs to happen")

		// get the root command or create it
		var burnRootCmd *cobra.Command
		for _, cmd := range ccli.WalletCmd.Commands() {
			if cmd.Name() == "burn" {
				burnRootCmd = cmd
				break
			}
		}
		if burnRootCmd == nil {
			burnRootCmd = &cobra.Command{
				Use:   "burn",
				Short: "burn resources, using an available command",
			}
			ccli.WalletCmd.AddCommand(burnRootCmd)
		}

		// attach our burn command to the burn root cmd
		burnRootCmd.AddCommand(burnCoinsCmd)
	}
	return nil
}

type walletCmd struct {
	cli          *client.CommandLineClient
	walletClient *client.WalletClient
	txPoolClient *client.TransactionPoolClient

	mintingDefinitionTxVersion, coinCreationTxVersion types.TransactionVersion
	minterDefinitionTxCfg                             struct {
		Description []byte
	}
	coinCreationTxCfg struct {
		Description []byte
	}

	coinDestructionTxVersion types.TransactionVersion
	coinDestructionTxCfg     struct {
		Description      []byte
		RefundAddress    string
		RefundAddressNew bool
	}

	requireMinerFees bool
}

func (walletCmd *walletCmd) createMinterDefinitionTxCmd(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.UsageFunc()(cmd)
		cli.Die("Invalid amount of arguments. One argume has to be given: <dest>|<rawCondition>")
	}

	// create a minter definition tx with a random nonce and the minimum required miner fee
	tx := minting.MinterDefinitionTransaction{
		Nonce: types.RandomTransactionNonce(),
	}
	if walletCmd.requireMinerFees {
		tx.MinerFees = []types.Currency{walletCmd.cli.Config.MinimumTransactionFee}
	}

	if n := len(walletCmd.minterDefinitionTxCfg.Description); n > 0 {
		tx.ArbitraryData = make([]byte, n)
		copy(tx.ArbitraryData[:], walletCmd.minterDefinitionTxCfg.Description[:])
	}

	// parse the given mint condition
	var err error
	tx.MintCondition, err = parseConditionString(args[0])
	if err != nil {
		cmd.UsageFunc()(cmd)
		cli.Die(err)
	}

	// encode the transaction as a JSON-encoded string and print it to the STDOUT
	json.NewEncoder(os.Stdout).Encode(tx.Transaction(walletCmd.mintingDefinitionTxVersion))
}

func (walletCmd *walletCmd) createCoinCreationTxCmd(cmd *cobra.Command, args []string) {
	currencyConvertor := walletCmd.cli.CreateCurrencyConvertor()

	// Check that the remaining args are condition + value pairs
	if len(args)%2 != 0 {
		cmd.UsageFunc()(cmd)
		cli.Die("Invalid arguments. Arguments must be of the form <dest>|<rawCondition> <amount> [<dest>|<rawCondition> <amount>]...")
	}

	// parse the remainder as output coditions and values
	pairs, err := parsePairedOutputs(args, currencyConvertor.ParseCoinString)
	if err != nil {
		cmd.UsageFunc()(cmd)
		cli.Die(err)
	}

	tx := minting.CoinCreationTransaction{
		Nonce: types.RandomTransactionNonce(),
	}
	if walletCmd.requireMinerFees {
		tx.MinerFees = []types.Currency{walletCmd.cli.Config.MinimumTransactionFee}
	}

	if n := len(walletCmd.coinCreationTxCfg.Description); n > 0 {
		tx.ArbitraryData = make([]byte, n)
		copy(tx.ArbitraryData[:], walletCmd.coinCreationTxCfg.Description[:])
	}

	for _, pair := range pairs {
		tx.CoinOutputs = append(tx.CoinOutputs, types.CoinOutput{
			Value:     pair.Value,
			Condition: pair.Condition,
		})
	}
	err = json.NewEncoder(os.Stdout).Encode(tx.Transaction(walletCmd.coinCreationTxVersion))
	if err != nil {
		panic(err)
	}
}

func (walletCmd *walletCmd) burnCoinsCmd(cmd *cobra.Command, args []string) {
	currencyConvertor := walletCmd.cli.CreateCurrencyConvertor()
	amount, err := currencyConvertor.ParseCoinString(args[0])
	if err != nil {
		cmd.UsageFunc()(cmd)
		cli.Die(err)
	}
	// add the minimum miner fee
	amount = amount.Add(walletCmd.cli.Config.MinimumTransactionFee)

	// define the optional user-defined refund address
	var refundAddress *types.UnlockHash
	if walletCmd.coinDestructionTxCfg.RefundAddress != "" {
		refundAddress = new(types.UnlockHash)
		err = refundAddress.LoadString(walletCmd.coinDestructionTxCfg.RefundAddress)
		if err != nil {
			cmd.UsageFunc()(cmd)
			cli.Die(err)
		}
	}
	// fund the burn Tx
	coinInputs, refundCoinOutput, err := walletCmd.walletClient.FundCoins(amount, refundAddress, walletCmd.coinDestructionTxCfg.RefundAddressNew)
	if err != nil {
		cli.DieWithError("failed to fund burn transaction", err)
	}

	// assemble the transaction
	cdTx := minting.CoinDestructionTransaction{
		CoinInputs: coinInputs,
		MinerFees:  []types.Currency{walletCmd.cli.Config.MinimumTransactionFee},
	}
	if refundCoinOutput != nil {
		cdTx.CoinOutputs = append(cdTx.CoinOutputs, *refundCoinOutput)
	}
	if n := len(walletCmd.coinCreationTxCfg.Description); n > 0 {
		cdTx.ArbitraryData = make([]byte, n)
		copy(cdTx.ArbitraryData[:], walletCmd.coinCreationTxCfg.Description[:])
	}

	// sign the transaction
	tx := cdTx.Transaction(walletCmd.coinDestructionTxVersion)
	err = walletCmd.walletClient.GreedySignTx(&tx)
	if err != nil {
		cli.DieWithError("failed to sign burn transaction", err)
	}

	// send the transaction
	txID, err := walletCmd.txPoolClient.AddTransactiom(tx)
	if err != nil {
		cli.DieWithError("failed to send burn transaction", err)
	}

	// print transaction ID
	fmt.Println(txID.String())
}

// try to parse the string first as an unlock hash,
// if that fails parse it as a
func parseConditionString(str string) (condition types.UnlockConditionProxy, err error) {
	// try to parse it as an unlock hash
	var uh types.UnlockHash
	err = uh.LoadString(str)
	if err == nil {
		// parsing as an unlock hash was succesfull, store the pair and continue to the next pair
		condition = types.NewCondition(types.NewUnlockHashCondition(uh))
		return
	}

	// try to parse it as a JSON-encoded unlock condition
	err = condition.UnmarshalJSON([]byte(str))
	if err != nil {
		return types.UnlockConditionProxy{}, fmt.Errorf(
			"condition has to be UnlockHash or JSON-encoded UnlockCondition, output %q is neither", str)
	}
	return
}

type (
	// parseCurrencyString takes the string representation of a currency value
	parseCurrencyString func(string) (types.Currency, error)

	outputPair struct {
		Condition types.UnlockConditionProxy
		Value     types.Currency
	}
)

func parsePairedOutputs(args []string, parseCurrency parseCurrencyString) (pairs []outputPair, err error) {
	argn := len(args)
	if argn < 2 {
		err = errors.New("not enough arguments, at least 2 required")
		return
	}
	if argn%2 != 0 {
		err = errors.New("arguments have to be given in pairs of '<dest>|<rawCondition>'+'<value>'")
		return
	}

	for i := 0; i < argn; i += 2 {
		// parse value first, as it's the one without any possibility of ambiguity
		var pair outputPair
		pair.Value, err = parseCurrency(args[i+1])
		if err != nil {
			err = fmt.Errorf("failed to parse amount/value for output #%d: %v", i/2, err)
			return
		}

		// parse condition second
		pair.Condition, err = parseConditionString(args[i])
		if err != nil {
			err = fmt.Errorf("failed to parse condition for output #%d: %v", i/2, err)
			return
		}

		// append succesfully parsed pair
		pairs = append(pairs, pair)
	}
	return
}
