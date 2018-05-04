package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/rivine/rivine/api"
	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"
)

const (
	// Amount of addresses to generate from the seed
	generateAddressAmount = 50
	// genesisOutputSeed is the mnemonic representation of the genesis seed
	// We will be using this seed to generate the address(es) to transfer the funds from
	genesisOutputSeed = "carbon boss inject cover mountain fetch fiber fit tornado cloth wing dinosaur proof joy intact fabric thumb rebel borrow poet chair network expire else"
	// unlockHashType is the string representation expected in the hashtype field when checking the unlock hashes in the explorer api
	unlockHashType = "unlockhash"
	// minerPayoutMaturityWindow is the amount of blocks that need to pass before a miner payout can be spend
	minerPayoutMaturityWindow = 144
	// gatewayPass is the password with which the gateway is protected
	gatewayPass = "test123"
)

var (
	// ErrInsufficientWalletFunds indicates that the wallet does not have enough funds for a transaction
	ErrInsufficientWalletFunds = errors.New("Wallet does not have enough funds for this transaction")

	explorerAddr = ""
)

func main() {
	// Parse the arguments
	if len(os.Args) < 4 {
		fmt.Println(`Usage: go run offline_transaction.go $gatewayIP:port $amount $target_address`)
		return
	}
	explorerAddr = os.Args[1]
	amount := types.Currency{}
	if _, err := fmt.Sscan(os.Args[2], &amount); err != nil {
		panic(err)
	}
	var addr types.UnlockHash
	if err := addr.LoadString(os.Args[3]); err != nil {
		panic(err)
	}
	// First convert the seed from the mnemonic form to the byte form
	seed, err := modules.InitialSeedFromMnemonic(genesisOutputSeed)
	if err != nil {
		panic(fmt.Sprintf("Failed to decode seed: %s", err))
	}
	// Now create the wallet
	w := NewWallet(seed)
	// Sync it with the chain
	if err = w.SyncWallet(); err != nil {
		panic(err)
	}

	// Create the transaction, inputs are taken automatically from the wallet
	// The transaction object is signed by this method as well
	txn, err := w.CreateTxn(amount, addr)
	if err != nil {
		panic(err)
	}

	// Send the transaction to the gateway
	if err = CommitTxn(txn); err != nil {
		panic(err)
	}

	fmt.Println("Transaction successfull")
}

// CreateTxn creates a new transaction of the specified ammount to a specified address. A remainder address
// to which the leftover coins will be transfered (if any) is chosen automatically. An error is returned if the coins
// available in the coininputs are insufficient to cover the amount specified for transfer (+ the miner fee).
// A miner fee is automatically added. This example does not add any arbitrary data.
func (w *MyWallet) CreateTxn(amount types.Currency, addressTo types.UnlockHash) (*types.Transaction, error) {
	// Count the funds in our wallet
	walletFunds := types.NewCurrency64(0)
	for _, uco := range w.unspentCoinOutputs {
		walletFunds = walletFunds.Add(uco.Value)
	}
	// Since this is only for demonstration purposes, lets give a fixed 10 hastings fee
	minerfee := types.NewCurrency64(10)

	// The total funds we will be spending in this transaction
	requiredFunds := amount.Add(minerfee)

	// Verify that we actually have enough funds available in the wallet to complete the transaction
	if walletFunds.Cmp(requiredFunds) == -1 {
		return nil, ErrInsufficientWalletFunds
	}

	// Create the transaction object
	txn := types.Transaction{
		Version: types.DefaultChainConstants().DefaultTransactionVersion,
	}

	// Greedily add coin inputs until we have enough to fund the output and minerfee
	inputs := []types.CoinInput{}

	// Track the amount of coins we already added via the inputs
	inputValue := types.ZeroCurrency

	for id, utxo := range w.unspentCoinOutputs {
		// If the inputValue is not smaller than the requiredFunds we added enough inputs to fund the transaction
		if inputValue.Cmp(requiredFunds) != -1 {
			break
		}
		// Append the input
		inputs = append(inputs, types.CoinInput{
			ParentID: id,
			Fulfillment: types.NewFulfillment(types.NewSingleSignatureFulfillment(
				types.Ed25519PublicKey(w.keys[utxo.Condition.UnlockHash()].PublicKey))),
		})
		// And update the value in the transaction
		inputValue = inputValue.Add(utxo.Value)
	}
	// Set the inputs
	txn.CoinInputs = inputs

	for _, inp := range inputs {
		if _, exists := w.keys[w.unspentCoinOutputs[inp.ParentID].Condition.UnlockHash()]; !exists {
			return nil, errors.New("Trying to spend unexisting output")
		}
	}
	// Add our first output
	txn.CoinOutputs = append(txn.CoinOutputs, types.CoinOutput{
		Value:     amount,
		Condition: types.NewCondition(types.NewUnlockHashCondition(addressTo)),
	})

	// So now we have enough inputs to fund everything. But we might have overshot it a little bit, so lets check that
	// and add a new output to ourself if required to consume the leftover value
	remainder := inputValue.Sub(requiredFunds)
	if !remainder.IsZero() {
		// We have leftover funds, so add a new transaction
		// Lets write to an unused address
		for addr := range w.keys {
			addrUsed := false
			for _, utxo := range w.unspentCoinOutputs {
				if utxo.Condition.UnlockHash() == addr {
					addrUsed = true
					break
				}
			}
			if addrUsed {
				continue
			}
			outputToSelf := types.CoinOutput{
				Value:     remainder,
				Condition: types.NewCondition(types.NewUnlockHashCondition(addr)),
			}
			// add our self referencing output to the transaction
			txn.CoinOutputs = append(txn.CoinOutputs, outputToSelf)
			break
		}
	}

	// Add the miner fee to the transaction
	txn.MinerFees = []types.Currency{minerfee}

	// sign transaction
	if err := w.signTxn(txn); err != nil {
		return nil, err
	}
	return &txn, nil
}

// signTxn signs a transaction
func (w *MyWallet) signTxn(txn types.Transaction) error {
	// sign every coin input
	for idx, input := range txn.CoinInputs {
		key := w.keys[input.Fulfillment.UnlockHash()]
		err := input.Fulfillment.Sign(types.FulfillmentSignContext{
			InputIndex:  uint64(idx),
			Transaction: txn,
			Key:         key.SecretKey,
		})
		if err != nil {
			return err
		}
	}
	// sign every blockstake input
	for idx, input := range txn.BlockStakeInputs {
		key := w.keys[input.Fulfillment.UnlockHash()]
		err := input.Fulfillment.Sign(types.FulfillmentSignContext{
			InputIndex:  uint64(idx),
			Transaction: txn,
			Key:         key.SecretKey,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// CommitTxn sends a transaction to a gateway node
func CommitTxn(txn *types.Transaction) error {
	bodyBuff := bytes.NewBuffer(nil)
	if err := json.NewEncoder(bodyBuff).Encode(txn); err != nil {
		return err
	}
	resp, err := RivineRequest("POST", "/transactionpool/transactions", bodyBuff)
	if err != nil {
		return err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return errors.New("Failed to commit txn: " + string(body))
	}
	return nil
}

// MyWallet implements a small offline wallet. For ease of use, it holds only a single seed
type MyWallet struct {
	// seed it the main object of the wallet, everything else is derived from it
	seed modules.Seed
	// unspentCoinOutputs map output ids to the actual coin outputs. Only keep track of the unspent outputs
	unspentCoinOutputs map[types.CoinOutputID]types.CoinOutput
	// keys represent the mapping between unlockhashes (addresses) and the spendableKey object which can be used to
	// spend an output to the address
	keys map[types.UnlockHash]spendableKey
}

// NewWallet initializes a new wallet from a given seed. It assumes the seed is valid
func NewWallet(seed modules.Seed) *MyWallet {
	w := MyWallet{}
	w.seed = seed
	w.unspentCoinOutputs = make(map[types.CoinOutputID]types.CoinOutput)
	w.keys = make(map[types.UnlockHash]spendableKey)
	// Create the adress -> key map
	for i := 0; i < generateAddressAmount; i++ {
		key := generateSpendableKey(w.seed, uint64(i))
		w.keys[key.UnlockHash()] = key
	}
	return &w
}

// SyncWallet syncs the wallet with the chain
func (w *MyWallet) SyncWallet() error {
	wg := sync.WaitGroup{}
	// First get the current block height
	height, err := GetCurrentChainHeight()
	if err != nil {
		return err
	}
	fmt.Println("Chain is currently at height: ", height)
	for address := range w.keys {
		wg.Add(1)
		go func(addr types.UnlockHash) {
			defer wg.Done()
			resp, err := CheckAddress(addr)
			if err != nil {
				panic(fmt.Sprint("Error while checking address usage: ", err))
			}
			if resp == nil {
				return
			}
			if resp.HashType != unlockHashType {
				panic("Address is not recognized as an unlock hash")
			}
			// We scann the blocks here for the miner fees, and the transactions for actual transations
			for _, block := range resp.Blocks {
				// Collect the miner fees
				// But only those that have matured already
				if block.Height+minerPayoutMaturityWindow >= height {
					fmt.Println("Ignoring miner payout that hasn't matured yet")
					continue
				}
				for i, minerPayout := range block.RawBlock.MinerPayouts {
					if minerPayout.UnlockHash == addr {
						fmt.Println("Found miner output with value ", minerPayout.Value)
						fmt.Println("Block: ", block.Height)
						for _, c := range block.MinerPayoutIDs {
							fmt.Println("Adding miner payout id", c.String())
						}
						w.unspentCoinOutputs[block.MinerPayoutIDs[i]] = types.CoinOutput{
							Value: minerPayout.Value,
							Condition: types.UnlockConditionProxy{
								Condition: types.NewUnlockHashCondition(minerPayout.UnlockHash),
							},
						}
					}
				}
			}

			// Collect the transaction outputs
			for _, txn := range resp.Transactions {
				for i, utxo := range txn.RawTransaction.CoinOutputs {
					if utxo.Condition.UnlockHash() == addr {
						w.unspentCoinOutputs[txn.CoinOutputIDs[i]] = utxo
					}
				}
			}
			// Remove the ones we've spent already
			for _, txn := range resp.Transactions {
				for _, ci := range txn.RawTransaction.CoinInputs {
					if _, exists := w.unspentCoinOutputs[ci.ParentID]; exists {
						delete(w.unspentCoinOutputs, ci.ParentID)
					}
				}
			}
		}(address)
	}
	wg.Wait()
	return nil
}

// CheckAddress performs a http call to an explorer to check if an address has (an) (unspent) output(s)
func CheckAddress(addr types.UnlockHash) (*api.ExplorerHashGET, error) {
	resp, err := RivineRequest("GET", "/explorer/hashes/"+addr.String(), nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}
	body := &api.ExplorerHashGET{}
	err = json.NewDecoder(resp.Body).Decode(body)
	return body, err
}

// GetCurrentChainHeight gets the current height of the blockchain
func GetCurrentChainHeight() (types.BlockHeight, error) {
	resp, err := RivineRequest("GET", "/consensus", nil)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, nil
	}
	body := &api.ConsensusGET{}
	err = json.NewDecoder(resp.Body).Decode(body)
	return body.Height, err
}

// RivineRequest executes a request to a rivined http api
func RivineRequest(method string, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, explorerAddr+endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Rivine-Agent")
	req.SetBasicAuth("", gatewayPass)
	cl := http.Client{}
	return cl.Do(req)
}

// spendableKey is a set of secret keys plus the corresponding unlock
// conditions.  The public key can be derived from the secret key and then
// matched to the corresponding public keys in the unlock conditions.
// Copied since the type is not exported
type spendableKey struct {
	PublicKey crypto.PublicKey
	SecretKey crypto.SecretKey
}

func (sk spendableKey) UnlockHash() types.UnlockHash {
	return types.NewEd25519PubKeyUnlockHash(sk.PublicKey)
}

// generateSpendableKey creates the keys and unlock conditions a given index of a
// seed. Copied as the function is not exported
func generateSpendableKey(seed modules.Seed, index uint64) spendableKey {
	// Generate the keys and unlock conditions.
	entropy := crypto.HashAll(seed, index)
	sk, pk := crypto.GenerateKeyPairDeterministic(entropy)
	return spendableKey{
		PublicKey: pk,
		SecretKey: sk,
	}
}
