package wallet

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"math/big"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/modules/consensus"
	"github.com/threefoldtech/rivine/modules/gateway"
	"github.com/threefoldtech/rivine/modules/transactionpool"
	"github.com/threefoldtech/rivine/types"
)

// A Wallet tester contains a ConsensusTester and has a bunch of helpful
// functions for facilitating wallet integration testing.
type walletTester struct {
	cs      modules.ConsensusSet
	gateway modules.Gateway
	tpool   modules.TransactionPool
	// miner   modules.TestMiner
	wallet *Wallet

	walletMasterKey crypto.TwofishKey

	persistDir string
}

// createWalletTester takes a testing.T and creates a WalletTester.
func createWalletTester(name string) (*walletTester, error) {
	bcInfo := types.DefaultBlockchainInfo()
	chainCts := types.TestnetChainConstants()
	// Create the modules
	testdir := build.TempDir(modules.WalletDir, name)
	g, err := gateway.New("localhost:0", false, 1, filepath.Join(testdir, modules.GatewayDir), bcInfo, chainCts, nil, false)
	if err != nil {
		return nil, err
	}

	cs, err := consensus.New(g, false, filepath.Join(testdir, modules.ConsensusDir), bcInfo, chainCts, false, "")
	if err != nil {
		return nil, err
	}
	tp, err := transactionpool.New(cs, g, filepath.Join(testdir, modules.TransactionPoolDir), bcInfo, chainCts, false)
	if err != nil {
		return nil, err
	}
	w, err := New(cs, tp, filepath.Join(testdir, modules.WalletDir), bcInfo, chainCts, false)
	if err != nil {
		return nil, err
	}
	var masterKey crypto.TwofishKey
	_, err = rand.Read(masterKey[:])
	if err != nil {
		return nil, err
	}
	_, err = w.Encrypt(masterKey, modules.Seed{})
	if err != nil {
		return nil, err
	}
	err = w.Unlock(masterKey)
	if err != nil {
		return nil, err
	}

	// m, err := miner.New(cs, tp, w, filepath.Join(testdir, modules.WalletDir))
	// if err != nil {
	// 	return nil, err
	// }

	// Assemble all components into a wallet tester.
	wt := &walletTester{
		cs:      cs,
		gateway: g,
		tpool:   tp,
		// miner:   m,
		wallet: w,

		walletMasterKey: masterKey,

		persistDir: testdir,
	}
	//
	// // Mine blocks until there is money in the wallet.
	// for i := types.BlockHeight(0); i <= types.MaturityDelay; i++ {
	// 	b, _ := wt.miner.FindBlock()
	// 	err := wt.cs.AcceptBlock(b)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	return wt, nil
}

// createWalletTester takes a testing.T and creates a WalletTester.
// todo rename this and use if for all `createWalletTester` scenarios
func createWalletTesterWithStubCS(name string, cs *consensusSetStub) (*walletTester, error) {
	if cs == nil {
		return nil, errors.New("no stub consensus set given")
	}
	bcInfo := types.DefaultBlockchainInfo()
	chainCts := types.TestnetChainConstants()
	// Create the modules
	testdir := build.TempDir(modules.WalletDir, name)
	g, err := gateway.New("localhost:0", false, 1, filepath.Join(testdir, modules.GatewayDir), bcInfo, chainCts, nil, false)
	if err != nil {
		return nil, err
	}

	tp, err := transactionpool.New(cs, g, filepath.Join(testdir, modules.TransactionPoolDir), bcInfo, chainCts, false)
	if err != nil {
		return nil, err
	}
	w, err := New(cs, tp, filepath.Join(testdir, modules.WalletDir), bcInfo, chainCts, false)
	if err != nil {
		return nil, err
	}
	var masterKey crypto.TwofishKey
	_, err = rand.Read(masterKey[:])
	if err != nil {
		return nil, err
	}
	_, err = w.Encrypt(masterKey, modules.Seed{})
	if err != nil {
		return nil, err
	}
	err = w.Unlock(masterKey)
	if err != nil {
		return nil, err
	}

	// m, err := miner.New(cs, tp, w, filepath.Join(testdir, modules.WalletDir))
	// if err != nil {
	// 	return nil, err
	// }

	// Assemble all components into a wallet tester.
	wt := &walletTester{
		cs:      cs,
		gateway: g,
		tpool:   tp,
		// miner:   m,
		wallet: w,

		walletMasterKey: masterKey,

		persistDir: testdir,
	}
	//
	// // Mine blocks until there is money in the wallet.
	// for i := types.BlockHeight(0); i <= types.MaturityDelay; i++ {
	// 	b, _ := wt.miner.FindBlock()
	// 	err := wt.cs.AcceptBlock(b)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	return wt, nil
}

// createBlankWalletTester creates a wallet tester that has not mined any
// blocks or encrypted the wallet.
func createBlankWalletTester(name string) (*walletTester, error) {
	bcInfo := types.DefaultBlockchainInfo()
	chainCts := types.TestnetChainConstants()
	// Create the modules
	testdir := build.TempDir(modules.WalletDir, name)
	g, err := gateway.New("localhost:0", false, 1, filepath.Join(testdir, modules.GatewayDir), bcInfo, chainCts, nil, false)
	if err != nil {
		return nil, err
	}
	cs, err := consensus.New(g, false, filepath.Join(testdir, modules.ConsensusDir), bcInfo, chainCts, false, "")
	if err != nil {
		return nil, err
	}
	tp, err := transactionpool.New(cs, g, filepath.Join(testdir, modules.TransactionPoolDir), bcInfo, chainCts, false)
	if err != nil {
		return nil, err
	}
	w, err := New(cs, tp, filepath.Join(testdir, modules.WalletDir), bcInfo, chainCts, false)
	if err != nil {
		return nil, err
	}
	// m, err := miner.New(cs, tp, w, filepath.Join(testdir, modules.WalletDir))
	// if err != nil {
	// 	return nil, err
	// }

	// Assemble all components into a wallet tester.
	wt := &walletTester{
		gateway: g,
		cs:      cs,
		tpool:   tp,
		// miner:   m,
		wallet: w,

		persistDir: testdir,
	}
	return wt, nil
}

// closeWt closes all of the modules in the wallet tester.
func (wt *walletTester) closeWt() {
	errs := []error{
		wt.gateway.Close(),
		wt.cs.Close(),
		wt.tpool.Close(),
		//		wt.miner.Close(),
		wt.wallet.Close(),
	}
	if err := build.JoinErrors(errs, "; "); err != nil {
		build.Critical(err)
	}
}

// TestNilInputs tries starting the wallet using nil inputs.
func TestNilInputs(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	bcInfo := types.DefaultBlockchainInfo()
	chainCts := types.TestnetChainConstants()
	testdir := build.TempDir(modules.WalletDir, t.Name())
	g, err := gateway.New("localhost:0", false, 1, filepath.Join(testdir, modules.GatewayDir), bcInfo, chainCts, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	cs, err := consensus.New(g, false, filepath.Join(testdir, modules.ConsensusDir), bcInfo, chainCts, false, "")
	if err != nil {
		t.Fatal(err)
	}
	tp, err := transactionpool.New(cs, g, filepath.Join(testdir, modules.TransactionPoolDir), bcInfo, chainCts, false)
	if err != nil {
		t.Fatal(err)
	}

	wdir := filepath.Join(testdir, modules.WalletDir)
	_, err = New(cs, nil, wdir, bcInfo, chainCts, false)
	if err != errNilTpool {
		t.Error(err)
	}
	_, err = New(nil, tp, wdir, bcInfo, chainCts, false)
	if err != errNilConsensusSet {
		t.Error(err)
	}
	_, err = New(nil, nil, wdir, bcInfo, chainCts, false)
	if err != errNilConsensusSet {
		t.Error(err)
	}
}

// TestAllAddresses checks that AllAddresses returns all of the wallet's
// addresses in sorted order.
func TestAllAddresses(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	wt, err := createWalletTester(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer wt.closeWt()

	wt.wallet.keys = map[types.UnlockHash]spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(1, crypto.Hash{0})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(0, crypto.Hash{1})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(1, crypto.Hash{5})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(0, crypto.Hash{5})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(0, crypto.Hash{0})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(1, crypto.Hash{1})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(1, crypto.Hash{2})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(0, crypto.Hash{3})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(0, crypto.Hash{2})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(0, crypto.Hash{4})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(1, crypto.Hash{3})] = spendableKey{}
	wt.wallet.keys[types.NewUnlockHash(1, crypto.Hash{4})] = spendableKey{}
	addrs, err := wt.wallet.AllAddresses()
	if err != nil {
		t.Fatal(err)
	}
	for i := range addrs[:5] {
		if addrs[i].Hash[0] != byte(i) {
			t.Error("address sorting failed:", i, addrs[i].Hash[0])
		}
		if addrs[i+6].Hash[0] != byte(i) {
			t.Error("address sorting failed:", i+6, addrs[i+6].Hash[0])
		}
	}
}

// TestCloseWallet tries to close the wallet.
func TestCloseWallet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	bcInfo := types.DefaultBlockchainInfo()
	chainCts := types.TestnetChainConstants()
	testdir := build.TempDir(modules.WalletDir, t.Name())
	g, err := gateway.New("localhost:0", false, 1, filepath.Join(testdir, modules.GatewayDir), bcInfo, chainCts, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	cs, err := consensus.New(g, false, filepath.Join(testdir, modules.ConsensusDir), bcInfo, chainCts, false, "")
	if err != nil {
		t.Fatal(err)
	}
	tp, err := transactionpool.New(cs, g, filepath.Join(testdir, modules.TransactionPoolDir), bcInfo, chainCts, false)
	if err != nil {
		t.Fatal(err)
	}
	wdir := filepath.Join(testdir, modules.WalletDir)
	w, err := New(cs, tp, wdir, bcInfo, chainCts, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

func newConsensusSetStub() *consensusSetStub {
	chainCts := types.TestnetChainConstants()
	return &consensusSetStub{
		blocks: []types.Block{
			chainCts.GenesisBlock(),
		},
		subscribers: make(map[modules.ConsensusSetSubscriber]struct{}),
	}
}

type consensusSetStub struct {
	blocks      []types.Block
	subscribers map[modules.ConsensusSetSubscriber]struct{}
}

func (css *consensusSetStub) Start() {
	// For testing we don't have any syncing logic
}

func (css *consensusSetStub) addTransactionAsBlock(unlockHash types.UnlockHash, value types.Currency) error {
	l := len(css.blocks)
	if l == 0 {
		return errors.New("invalid block list in consensus set")
	}
	return css.AcceptBlock(types.Block{
		ParentID:  css.blocks[l-1].ID(),
		Timestamp: types.CurrentTimestamp(),
		Transactions: []types.Transaction{
			{
				Version: types.TestnetChainConstants().DefaultTransactionVersion,
				CoinOutputs: []types.CoinOutput{
					{
						Value:     value,
						Condition: types.NewCondition(types.NewUnlockHashCondition(unlockHash)),
					},
				},
			},
		},
	})
}

func (css *consensusSetStub) AcceptBlock(block types.Block) error {
	id := block.ID()
	for _, b := range css.blocks {
		if b.ID() == id {
			return errors.New("block seen before")
		}
	}
	css.blocks = append(css.blocks, block)

	for subscriber := range css.subscribers {
		processAppliedBlock(block, subscriber)
	}

	return nil
}

func processAppliedBlock(block types.Block, subscriber modules.ConsensusSetSubscriber) {
	bh, err := crypto.HashObject(block)
	if err != nil {
		panic(err)
	}
	cc := modules.ConsensusChange{
		ID:            modules.ConsensusChangeID(bh),
		AppliedBlocks: []types.Block{block},
	}
	for _, tx := range block.Transactions {
		for _, co := range tx.CoinOutputs {
			coh, err := crypto.HashObject(co)
			if err != nil {
				panic(err)
			}
			cc.CoinOutputDiffs = append(cc.CoinOutputDiffs, modules.CoinOutputDiff{
				Direction:  modules.DiffApply,
				ID:         types.CoinOutputID(coh),
				CoinOutput: co,
			})
		}
	}
	subscriber.ProcessConsensusChange(cc)
}

func (css *consensusSetStub) BlockAtHeight(height types.BlockHeight) (types.Block, bool) {
	if height >= types.BlockHeight(len(css.blocks)) {
		return types.Block{}, false
	}
	return css.blocks[height], true
}

func (css *consensusSetStub) BlockHeightOfBlock(block types.Block) (types.BlockHeight, bool) {
	id := block.ID()
	for height, b := range css.blocks {
		if b.ID() == id {
			return types.BlockHeight(height), true
		}
	}
	return 0, false
}

func (css *consensusSetStub) TransactionAtShortID(shortID types.TransactionShortID) (types.Transaction, bool) {
	height := shortID.BlockHeight()
	block, found := css.BlockAtHeight(height)
	if !found {
		return types.Transaction{}, false
	}

	txSeqID := int(shortID.TransactionSequenceIndex())
	if len(block.Transactions) <= txSeqID {
		return types.Transaction{}, false
	}

	return block.Transactions[txSeqID], true
}

func (css *consensusSetStub) TransactionAtID(id types.TransactionID) (types.Transaction, types.TransactionShortID, bool) {
	for i, b := range css.blocks {
		for j, t := range b.Transactions {
			if t.ID() == id {
				return t, types.NewTransactionShortID(types.BlockHeight(i), uint16(j)), true
			}
		}
	}
	return types.Transaction{}, 0, false
}

func (css *consensusSetStub) FindParentBlock(b types.Block, depth types.BlockHeight) (block types.Block, exists bool) {
	var blockIndex int
	for i, block := range css.blocks {
		if block.Header().ID() == b.Header().ID() {
			blockIndex = i
			break
		}
	}
	if int(depth) > blockIndex {
		return types.Block{}, false
	}
	return css.blocks[blockIndex-int(depth)], true
}

func (css *consensusSetStub) ChildTarget(id types.BlockID) (types.Target, bool) {
	// TODO: return a more sensible value if required
	return types.Target{}, false
}

func (css *consensusSetStub) Close() error {
	return nil
}

func (css *consensusSetStub) CurrentBlock() types.Block {
	l := len(css.blocks)
	if l == 0 {
		return types.Block{}
	}
	return css.blocks[l-1]
}

func (css *consensusSetStub) Flush() error {
	return nil
}

func (css *consensusSetStub) Height() types.BlockHeight {
	l := len(css.blocks)
	if l == 0 {
		return 0
	}
	return types.BlockHeight(l - 1)
}

func (css *consensusSetStub) Synced() bool {
	return true
}

func (css *consensusSetStub) InCurrentPath(id types.BlockID) bool {
	for _, b := range css.blocks {
		if b.ID() == id {
			return true
		}
	}
	return false
}

func (css *consensusSetStub) MinimumValidChildTimestamp(id types.BlockID) (types.Timestamp, bool) {
	if len(css.blocks) == 0 {
		return 0, false
	}
	return css.blocks[0].Timestamp, true
}

func (css *consensusSetStub) CalculateStakeModifier(height types.BlockHeight, block types.Block, delay types.BlockHeight) *big.Int {
	//TODO: check if a new Stakemodifier needs to be calculated. The stakemodifier
	// only change when a new block is created, and this calculation is also needed
	// to validate an incomming new block

	// make a signed version of the current height because sub genesis block is
	// possible here.
	signedHeight := int64(height)
	signedHeight -= int64(types.TestnetChainConstants().StakeModifierDelay)

	mask := big.NewInt(1)
	var BlockIDHash *big.Int
	stakemodifier := big.NewInt(0)
	var buffer bytes.Buffer

	// Rollback the required amount of blocks, minus 1. This way we end up at the direct child of the
	// block we use to calculate the stakemodifer, rather than the actual first block. Simplifies
	// the main loop a bit
	block, _ = css.FindParentBlock(block, delay-1)

	// We have the direct child of the first block used in the stake modifier calculation. As such
	// we can follow the parentID in the block to retrieve all the blocks required, using 1 bit
	// of each blocks ID to calculate the stake modifier
	for i := 0; i < 256; i++ {
		if signedHeight >= 0 {
			var exist bool
			block, exist = css.FindParentBlock(block, 1)
			if build.DEBUG && !exist {
				panic("block to be used for stakemodifier does not yet exist")
			}
			hashof := block.ID()
			BlockIDHash = big.NewInt(0).SetBytes(hashof[:])
		} else {
			// if the counter goes sub genesis block , calculate a predefined hash
			// from the sub genesis distance.
			buffer.WriteString("genesis" + strconv.FormatInt(signedHeight, 10))
			hashof := sha256.Sum256(buffer.Bytes())
			BlockIDHash = big.NewInt(0).SetBytes(hashof[:])
		}

		stakemodifier.Or(stakemodifier, big.NewInt(0).And(BlockIDHash, mask))
		mask.Mul(mask, big.NewInt(2)) //shift 1 bit to left (more close to msb)
		signedHeight--
	}
	return stakemodifier
}

func (css *consensusSetStub) TryTransactionSet(txs []types.Transaction) (change modules.ConsensusChange, err error) {
	l := len(css.blocks)
	if l == 0 {
		return modules.ConsensusChange{}, errors.New("invalid block list in consensus set")
	}
	block := types.Block{
		ParentID:     css.blocks[l-1].ID(),
		Timestamp:    types.CurrentTimestamp(),
		Transactions: txs,
	}
	bh, err := crypto.HashObject(block)
	if err != nil {
		return change, err
	}
	cc := modules.ConsensusChange{
		ID: modules.ConsensusChangeID(bh),
	}
	for _, tx := range block.Transactions {
		for _, co := range tx.CoinOutputs {
			coh, err := crypto.HashObject(co)
			if err != nil {
				return change, err
			}
			cc.CoinOutputDiffs = append(cc.CoinOutputDiffs, modules.CoinOutputDiff{
				Direction:  modules.DiffApply,
				ID:         types.CoinOutputID(coh),
				CoinOutput: co,
			})
		}
	}
	return cc, nil
}

func (css *consensusSetStub) ConsensusSetSubscribe(subscriber modules.ConsensusSetSubscriber, changeID modules.ConsensusChangeID, cancel <-chan struct{}) error {
	if _, ok := css.subscribers[subscriber]; ok {
		return errors.New("subscriber already registered to stub consensus set")
	}
	css.subscribers[subscriber] = struct{}{}

	var i int
	if changeID != modules.ConsensusChangeID(crypto.Hash{}) {
		for ; i < len(css.blocks); i++ {
			select {
			case <-cancel:
				return errors.New("abort (stub) ConsensusSetSubscribe")
			default:
			}
			bh, err := crypto.HashObject(css.blocks[i])
			if err != nil {
				return err
			}
			if modules.ConsensusChangeID(bh) == changeID {
				break
			}
		}
	}
	for _, block := range css.blocks[i:] {
		select {
		case <-cancel:
			return errors.New("abort (stub) ConsensusSetSubscribe")
		default:
			processAppliedBlock(block, subscriber)
		}
	}
	return nil
}

func (css *consensusSetStub) Unsubscribe(subscriber modules.ConsensusSetSubscriber) {
	delete(css.subscribers, subscriber)
}

func (css *consensusSetStub) GetCoinOutput(id types.CoinOutputID) (co types.CoinOutput, err error) {
	for _, block := range css.blocks {
		for _, txn := range block.Transactions {
			for i, co := range txn.CoinOutputs {
				if txn.CoinOutputID(uint64(i)) == id {
					return co, nil
				}
			}
		}
	}
	return types.CoinOutput{}, errors.New("Coin output not found in database")
}

func (css *consensusSetStub) GetBlockStakeOutput(id types.BlockStakeOutputID) (bso types.BlockStakeOutput, err error) {
	for _, block := range css.blocks {
		for _, txn := range block.Transactions {
			for i, bso := range txn.BlockStakeOutputs {
				if txn.BlockStakeOutputID(uint64(i)) == id {
					return bso, nil
				}
			}
		}
	}
	return types.BlockStakeOutput{}, errors.New("BlockStake output not found in database")
}

func (css *consensusSetStub) RegisterPlugin(ctx context.Context, name string, plugin modules.ConsensusSetPlugin) (err error) {
	return nil
}

func (css *consensusSetStub) UnregisterPlugin(name string, plugin modules.ConsensusSetPlugin) {
	// Do nothing
}

func (css *consensusSetStub) SetTransactionValidators(validators ...modules.TransactionValidationFunction) {
	// Do nothing
}

func (css *consensusSetStub) SetTransactionVersionMappedValidators(version types.TransactionVersion, validators ...modules.TransactionValidationFunction) {
	// Do nothing
}
