package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/NebulousLabs/entropy-mnemonics"

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
	genesisOutputSeed = "across knife thirsty puck itches hazard enmity fainted pebbles unzip echo queen rarest aphid bugs yanks okay abbey eskimos dove orange nouns august ailments inline rebel glass tyrant acumen"
	// unlockHashType is the string representation expected in the hashtype field when checking the unlock hashes in the explorer api
	unlockHashType = "unlockhash"
	// minerPayoutMaturityWindow is the amount of blocks that need to pass before a miner payout can be spend
	minerPayoutMaturityWindow = 144
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
	seed, err := modules.StringToSeed(genesisOutputSeed, mnemonics.English)
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

	for _, a := range txn.CoinInputs {
		fmt.Println(a)
	}

	// Send the transaction to the gateway
	if err = CommitTxn(txn); err != nil {
		panic(err)
	}

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

	fmt.Println(walletFunds)

	// Verify that we actually have enough funds available in the wallet to complete the transaction
	if walletFunds.Cmp(requiredFunds) == -1 {
		return nil, ErrInsufficientWalletFunds
	}

	// Create the transaction object
	txn := &types.Transaction{}

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
		inputs = append(inputs, types.CoinInput{ParentID: id, UnlockConditions: w.keys[utxo.UnlockHash].UnlockConditions})
		// And update the value in the transaction
		inputValue = inputValue.Add(utxo.Value)
	}
	// Set the inputs
	txn.CoinInputs = inputs

	for i, inp := range inputs {
		if _, exists := w.keys[w.unspentCoinOutputs[inp.ParentID].UnlockHash]; !exists {
			panic("Tryinng to spend unexisting output")
		}
		fmt.Println("Input ", i, ": ", inp.ParentID.String())
	}
	// Add our first output
	txn.CoinOutputs = append(txn.CoinOutputs, types.CoinOutput{Value: amount, UnlockHash: addressTo})

	// So now we have enough inputs to fund everything. But we might have overshot it a little bit, so lets check that
	// and add a new output to ourself if required to consume the leftover value
	remainder := inputValue.Sub(requiredFunds)
	if !remainder.IsZero() {
		// We have leftover funds, so add a new transaction
		// Lets write to an unused address
		for addr := range w.keys {
			addrUsed := false
			for _, utxo := range w.unspentCoinOutputs {
				if utxo.UnlockHash == addr {
					addrUsed = true
					break
				}
			}
			if addrUsed {
				continue
			}
			outputToSelf := types.CoinOutput{
				Value:      remainder,
				UnlockHash: addr,
			}
			// add our self referencing output to the transaction
			txn.CoinOutputs = append(txn.CoinOutputs, outputToSelf)
			break
		}
	}

	// Add the miner fee to the transaction
	txn.MinerFees = []types.Currency{minerfee}

	// SIGN THE DAMN THING
	if err := w.signTxn(txn); err != nil {
		panic(err)
	}
	fmt.Println(txn)

	return txn, nil
}

// signTxn signs a transaction
func (w *MyWallet) signTxn(txn *types.Transaction) error {
	// We will sign the whole transaction
	coveredFields := types.CoveredFields{WholeTransaction: true}
	// Add a signature for every input
	for _, input := range txn.CoinInputs {
		// input := tb.transaction.CoinInputs[inputIndex]
		key := w.keys[input.UnlockConditions.UnlockHash()]
		_, err := addSignatures(txn, coveredFields, input.UnlockConditions, crypto.Hash(input.ParentID), key)
		if err != nil {
			return err
		}
		// txn.TransactionSignatures = append(txn.TransactionSignatures, newSigIndices...)
	}
	return nil
}

// CommitTxn sends a transaction to a gateway node
func CommitTxn(txn *types.Transaction) error {
	endpoint := "/transactionpool/transactions"
	bodyBuff := bytes.NewBuffer(nil)
	if err := json.NewEncoder(bodyBuff).Encode(txn); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", explorerAddr+endpoint, bodyBuff)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Rivine-Agent")
	cl := http.Client{}
	resp, err := cl.Do(req)
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
		w.keys[key.UnlockConditions.UnlockHash()] = key
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
			// TODO: FIGURE OUT WHY THIS DOES NOT WORK
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
						w.unspentCoinOutputs[block.MinerPayoutIDs[i]] = minerPayout
					}
				}
			}

			// Collect the transaction outputs
			for _, txn := range resp.Transactions {
				for i, utxo := range txn.RawTransaction.CoinOutputs {
					if utxo.UnlockHash == addr {
						fmt.Println("Found txn output with value ", utxo.Value)
						fmt.Println("Block ", txn.Height)
						for _, c := range txn.CoinOutputIDs {
							fmt.Println("Adding ", c.String())
						}
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
	endpoint := "/explorer/hashes/"
	req, err := http.NewRequest("GET", explorerAddr+endpoint+addr.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Rivine-Agent")
	cl := http.Client{}
	resp, err := cl.Do(req)
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
	endpoint := "/consensus"
	req, err := http.NewRequest("GET", explorerAddr+endpoint, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "Rivine-Agent")
	cl := http.Client{}
	resp, err := cl.Do(req)
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

// spendableKey is a set of secret keys plus the corresponding unlock
// conditions.  The public key can be derived from the secret key and then
// matched to the corresponding public keys in the unlock conditions.
// Copied since the type is not exported
type spendableKey struct {
	UnlockConditions types.UnlockConditions
	SecretKeys       []crypto.SecretKey
}

// generateSpendableKey creates the keys and unlock conditions a given index of a
// seed. Copied as the function is not exported
func generateSpendableKey(seed modules.Seed, index uint64) spendableKey {
	// Generate the keys and unlock conditions.
	entropy := crypto.HashAll(seed, index)
	sk, pk := crypto.GenerateKeyPairDeterministic(entropy)
	return spendableKey{
		UnlockConditions: generateUnlockConditions(pk),
		SecretKeys:       []crypto.SecretKey{sk},
	}
}

// generateUnlockConditions provides the unlock conditions that would be
// automatically generated from the input public key. Copied as the function is not exported
func generateUnlockConditions(pk crypto.PublicKey) types.UnlockConditions {
	return types.UnlockConditions{
		PublicKeys: []types.SiaPublicKey{{
			Algorithm: types.SignatureEd25519,
			Key:       pk[:],
		}},
		SignaturesRequired: 1,
	}
}

// addSignatures will sign a transaction using a spendable key, with support
// for multisig spendable keys. Because of the restricted input, the function
// is compatible with both coin inputs and blockstake inputs.
// Copied as the function is not exported
func addSignatures(txn *types.Transaction, cf types.CoveredFields, uc types.UnlockConditions, parentID crypto.Hash, spendKey spendableKey) (newSigIndices []int, err error) {
	// Try to find the matching secret key for each public key - some public
	// keys may not have a match. Some secret keys may be used multiple times,
	// which is why public keys are used as the outer loop.
	totalSignatures := uint64(0)
	for i, siaPubKey := range uc.PublicKeys {
		// Search for the matching secret key to the public key.
		for j := range spendKey.SecretKeys {
			pubKey := spendKey.SecretKeys[j].PublicKey()
			if bytes.Compare(siaPubKey.Key, pubKey[:]) != 0 {
				continue
			}

			// Found the right secret key, add a signature.
			sig := types.TransactionSignature{
				ParentID:       parentID,
				CoveredFields:  cf,
				PublicKeyIndex: uint64(i),
			}
			newSigIndices = append(newSigIndices, len(txn.TransactionSignatures))
			txn.TransactionSignatures = append(txn.TransactionSignatures, sig)
			sigIndex := len(txn.TransactionSignatures) - 1
			sigHash := txn.SigHash(sigIndex)
			encodedSig, err := crypto.SignHash(sigHash, spendKey.SecretKeys[j])
			if err != nil {
				return nil, err
			}
			txn.TransactionSignatures[sigIndex].Signature = encodedSig[:]

			// Count that the signature has been added, and break out of the
			// secret key loop.
			totalSignatures++
			break
		}

		// If there are enough signatures to satisfy the unlock conditions,
		// break out of the outer loop.
		if totalSignatures == uc.SignaturesRequired {
			break
		}
	}
	return newSigIndices, nil
}
