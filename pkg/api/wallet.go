package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/types"

	"github.com/julienschmidt/httprouter"
)

type (
	// WalletGET contains general information about the wallet.
	WalletGET struct {
		Encrypted bool `json:"encrypted"`
		Unlocked  bool `json:"unlocked"`

		ConfirmedCoinBalance       types.Currency `json:"confirmedcoinbalance"`
		ConfirmedLockedCoinBalance types.Currency `json:"confirmedlockedcoinbalance"`
		UnconfirmedOutgoingCoins   types.Currency `json:"unconfirmedoutgoingcoins"`
		UnconfirmedIncomingCoins   types.Currency `json:"unconfirmedincomingcoins"`

		BlockStakeBalance       types.Currency `json:"blockstakebalance"`
		LockedBlockStakeBalance types.Currency `json:"lockedblockstakebalance"`

		MultiSigWallets []modules.MultiSigWallet `json:"multisigwallets"`
	}

	// WalletBlockStakeStatsGET contains blockstake statistical info of the wallet.
	WalletBlockStakeStatsGET struct {
		TotalActiveBlockStake types.Currency             `json:"totalactiveblockstake"`
		TotalBlockStake       types.Currency             `json:"totalblockstake"`
		TotalFeeLast1000      types.Currency             `json:"totalfeelast1000"`
		TotalBCLast1000       uint64                     `json:"totalbclast1000"`
		BlockCount            uint64                     `json:"blockcount"`
		TotalBCLast1000t      uint64                     `json:"totalbclast1000t"`
		BlockStakeState       []uint64                   `json:"blockstakestate"`
		BlockStakeNumOf       []types.Currency           `json:"blockstakenumof"`
		BlockStakeUTXOAddress []types.BlockStakeOutputID `json:"blockstakeutxoaddress"`
	}

	// WalletAddressGET contains an address returned by a GET call to
	// /wallet/address.
	WalletAddressGET struct {
		Address types.UnlockHash `json:"address"`
	}

	// WalletAddressesGET contains the list of wallet addresses returned by a
	// GET call to /wallet/addresses.
	WalletAddressesGET struct {
		Addresses []types.UnlockHash `json:"addresses"`
	}

	// WalletInitPOST contains the mnemonic of the primary seed,
	// the seed which is either given by you as part of the request,
	// or generated for you if none is given. If it's the first case,
	// the returned primary seed mnemonic should be the same as the one you already know.
	// POST call to /wallet/init.
	WalletInitPOST struct {
		PrimarySeed string `json:"primaryseed"`
	}

	// WalletTransactionPOST contains the unlockhash and amount of money to send,
	// during a POST call to /wallet/transaction, funding the output,
	// using available inputs in the wallet.
	WalletTransactionPOST struct {
		Condition types.UnlockConditionProxy `json:"condition"`
		Amount    types.Currency             `json:"amount"`
		Data      string                     `json:"data,omitempty"`
	}

	// WalletTransactionPOSTResponse contains the ID of the transaction
	// that was created as a result of a POST call to /wallet/transaction.
	WalletTransactionPOSTResponse struct {
		Transaction types.Transaction `json:"transaction"`
	}

	// WalletCoinsPOST is given by the user
	// to indicate to where to send how much coins
	WalletCoinsPOST struct {
		CoinOutputs           []types.CoinOutput `json:"coinoutputs`
		Data                  []byte             `json:"data,omitempty"`
		RefundAddress         *types.UnlockHash  `json:"refundaddress,omitempty"`
		GenerateRefundAddress bool               `json:"genrefundaddress,omitempty"`
	}
	// WalletCoinsPOSTResp Resp contains the ID of the transaction
	// that was created as a result of a POST call to /wallet/coins.
	WalletCoinsPOSTResp struct {
		TransactionID types.TransactionID `json:"transactionid"`
	}

	// WalletBlockStakesPOST is given by the user
	// to indicate to where to send how much blockstakes
	WalletBlockStakesPOST struct {
		BlockStakeOutputs     []types.BlockStakeOutput `json:"blockstakeoutputs`
		Data                  []byte                   `json:"data,omitempty"`
		RefundAddress         *types.UnlockHash        `json:"refundaddress,omitempty"`
		GenerateRefundAddress bool                     `json:"genrefundaddress,omitempty"`
	}
	// WalletBlockStakesPOSTResp Resp contains the ID of the transaction
	// that was created as a result of a POST call to /wallet/blockstakes.
	WalletBlockStakesPOSTResp struct {
		TransactionID types.TransactionID `json:"transactionids"`
	}

	// WalletSeedsGET contains the seeds used by the wallet.
	WalletSeedsGET struct {
		PrimarySeed        string   `json:"primaryseed"`
		AddressesRemaining int      `json:"addressesremaining"`
		AllSeeds           []string `json:"allseeds"`
	}

	// WalletKeyGet contains the public and private key used by the wallet.
	WalletKeyGet struct {
		AlgorithmSpecifier types.Specifier `json:"specifier"`
		PublicKey          types.ByteSlice `json:"publickey"`
		SecretKey          types.ByteSlice `json:"secretkey"`
	}

	// WalletTransactionGETid contains the transaction returned by a call to
	// /wallet/transaction/$(id)
	WalletTransactionGETid struct {
		Transaction modules.ProcessedTransaction `json:"transaction"`
	}

	// WalletTransactionsGET contains the specified set of confirmed and
	// unconfirmed transactions.
	WalletTransactionsGET struct {
		ConfirmedTransactions   []modules.ProcessedTransaction `json:"confirmedtransactions"`
		UnconfirmedTransactions []modules.ProcessedTransaction `json:"unconfirmedtransactions"`
	}

	// WalletTransactionsGETaddr contains the set of wallet transactions
	// relevant to the input address provided in the call to
	// /wallet/transaction/$(addr)
	WalletTransactionsGETaddr struct {
		ConfirmedTransactions   []modules.ProcessedTransaction `json:"confirmedtransactions"`
		UnconfirmedTransactions []modules.ProcessedTransaction `json:"unconfirmedtransactions"`
	}

	// WalletListUnlockedGET contains the set of unspent, unlocked coin
	// and blockstake outputs owned by the wallet.
	WalletListUnlockedGET struct {
		UnlockedCoinOutputs       []UnspentCoinOutput       `json:"unlockedcoinoutputs"`
		UnlockedBlockstakeOutputs []UnspentBlockstakeOutput `json:"unlockedblockstakeoutputs"`
	}

	// WalletListLockedGET contains the set of unspent, locked coin and
	// blockstake outputs owned by the wallet
	WalletListLockedGET struct {
		LockedCoinOutputs       []UnspentCoinOutput       `json:"lockedcoinoutputs"`
		LockedBlockstakeOutputs []UnspentBlockstakeOutput `json:"lockedblockstakeoutputs"`
	}

	// UnspentCoinOutput is a coin output and its associated ID
	UnspentCoinOutput struct {
		ID     types.CoinOutputID `json:"id"`
		Output types.CoinOutput   `json:"coinoutput"`
	}

	// UnspentBlockstakeOutput is a blockstake output and its associated ID
	UnspentBlockstakeOutput struct {
		ID     types.BlockStakeOutputID `json:"id"`
		Output types.BlockStakeOutput   `json:"output"`
	}

	// WalletCreateTransactionPOST is a list of coin and blockstake inputs and outputs
	// The values in the coin and blockstake input and outputs pair must match exactly (also
	// accounting for miner fees)
	WalletCreateTransactionPOST struct {
		CoinInputs        []types.CoinOutputID       `json:"coininputs"`
		BlockStakeInputs  []types.BlockStakeOutputID `json:"blockstakeinputs"`
		CoinOutputs       []types.CoinOutput         `json:"coinoutputs"`
		BlockStakeOutputs []types.BlockStakeOutput   `json:"blockstakeoutputs"`
	}

	// WalletCreateTransactionRESP wraps the transaction returned by the walletcreatetransaction
	// endpoint
	WalletCreateTransactionRESP struct {
		Transaction types.Transaction `json:"transaction"`
	}

	// WalletFundCoins is the resulting object that is returned,
	// to be used by a client to fund a transaction of any type.
	WalletFundCoins struct {
		CoinInputs       []types.CoinInput `json:"coininputs"`
		RefundCoinOutput *types.CoinOutput `json:"refund"`
	}

	// WalletPublicKeyGET contains a public key returned by a GET call to
	// /wallet/publickey.
	WalletPublicKeyGET struct {
		PublicKey types.PublicKey `json:"publickey"`
	}
)

// RegisterWalletHTTPHandlers registers the default Rivine handlers for all default Rivine Wallet HTTP endpoints.
func RegisterWalletHTTPHandlers(router Router, wallet modules.Wallet, requiredPassword string) {
	if wallet == nil {
		build.Critical("no wallet module given")
	}
	if router == nil {
		build.Critical("no httprouter Router given")
	}

	router.GET("/wallet", RequirePasswordHandler(NewWalletRootHandler(wallet), requiredPassword))
	router.GET("/wallet/blockstakestats", RequirePasswordHandler(NewWalletBlockStakeStatsHandler(wallet), requiredPassword))
	router.GET("/wallet/address", RequirePasswordHandler(NewWalletAddressHandler(wallet), requiredPassword))
	router.GET("/wallet/addresses", RequirePasswordHandler(NewWalletAddressesHandler(wallet), requiredPassword))
	router.GET("/wallet/backup", RequirePasswordHandler(NewWalletBackupHandler(wallet), requiredPassword))
	router.POST("/wallet/init", RequirePasswordHandler(NewWalletInitHandler(wallet), requiredPassword))
	router.POST("/wallet/lock", RequirePasswordHandler(NewWalletLockHandler(wallet), requiredPassword))
	router.POST("/wallet/seed", RequirePasswordHandler(NewWalletSeedHandler(wallet), requiredPassword))
	router.GET("/wallet/seeds", RequirePasswordHandler(NewWalletSeedsHandler(wallet), requiredPassword))
	router.GET("/wallet/key/:unlockhash", RequirePasswordHandler(NewWalletKeyHandler(wallet), requiredPassword))
	router.POST("/wallet/transaction", RequirePasswordHandler(NewWalletTransactionCreateHandler(wallet), requiredPassword))
	router.POST("/wallet/coins", RequirePasswordHandler(NewWalletCoinsHandler(wallet), requiredPassword))
	router.POST("/wallet/blockstakes", RequirePasswordHandler(NewWalletBlockStakesHandler(wallet), requiredPassword))
	router.GET("/wallet/transaction/:id", NewWalletTransactionHandler(wallet))
	router.GET("/wallet/transactions", NewWalletTransactionsHandler(wallet))
	router.GET("/wallet/transactions/:addr", NewWalletTransactionsAddrHandler(wallet))
	router.POST("/wallet/unlock", RequirePasswordHandler(NewWalletUnlockHandler(wallet), requiredPassword))
	router.GET("/wallet/unlocked", RequirePasswordHandler(NewWalletListUnlockedHandler(wallet), requiredPassword))
	router.GET("/wallet/locked", RequirePasswordHandler(NewWalletListLockedHandler(wallet), requiredPassword))
	router.POST("/wallet/create/transaction", RequirePasswordHandler(NewWalletCreateTransactionHandler(wallet), requiredPassword))
	router.POST("/wallet/sign", RequirePasswordHandler(NewWalletSignHandler(wallet), requiredPassword))
	router.GET("/wallet/publickey", RequirePasswordHandler(NewWalletGetPublicKeyHandler(wallet), requiredPassword))
	router.GET("/wallet/fund/coins", RequirePasswordHandler(NewWalletFundCoinsHandler(wallet), requiredPassword))
}

// NewWalletRootHandler creates a handler to handle API calls to /wallet.
func NewWalletRootHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		coinBal, blockstakeBal, err := wallet.ConfirmedBalance()
		if err != nil {
			WriteError(w, Error{"error after call to /wallet: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		coinLockBal, blockstakeLockBal, err := wallet.ConfirmedLockedBalance()
		if err != nil {
			WriteError(w, Error{"error after call to /wallet: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		coinsOut, coinsIn, err := wallet.UnconfirmedBalance()
		if err != nil {
			WriteError(w, Error{"error after call to /wallet: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		multiSigWallets, err := wallet.MultiSigWallets()
		if err != nil {
			WriteError(w, Error{"error after call to /wallet: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}

		WriteJSON(w, WalletGET{
			Encrypted: wallet.Encrypted(),
			Unlocked:  wallet.Unlocked(),

			ConfirmedCoinBalance:       coinBal,
			ConfirmedLockedCoinBalance: coinLockBal,
			UnconfirmedOutgoingCoins:   coinsOut,
			UnconfirmedIncomingCoins:   coinsIn,

			BlockStakeBalance:       blockstakeBal,
			LockedBlockStakeBalance: blockstakeLockBal,

			MultiSigWallets: multiSigWallets,
		})
	}
}

// NewWalletBlockStakeStatsHandler creates a new handler to handle API calls to /wallet/blockstakestat.
func NewWalletBlockStakeStatsHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		unspentBSOs, err := wallet.GetUnspentBlockStakeOutputs()
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/blockstakestat: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		count := len(unspentBSOs)
		bss := make([]uint64, count)
		bsn := make([]types.Currency, count)
		bsutxoa := make([]types.BlockStakeOutputID, count)
		tabs := types.NewCurrency64(1000000) //TODO rivine change this to estimated num of BS
		tbs := types.NewCurrency64(0)

		num := 0
		tbclt, bsf, bc, err := wallet.BlockStakeStats()
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/blockstakestat: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}

		for _, ubso := range unspentBSOs {
			bss[num] = 1
			bsn[num] = ubso.Value
			tbs = tbs.Add(bsn[num])
			bsutxoa[num] = ubso.BlockStakeOutputID
			num++
		}

		tbcltt := uint64(0)
		if !tbs.IsZero() {
			tbcltt, _ = tabs.Mul64(bc).Div(tbs).Uint64()
		}

		WriteJSON(w, WalletBlockStakeStatsGET{
			TotalActiveBlockStake: tabs,
			TotalBlockStake:       tbs,
			TotalFeeLast1000:      bsf,
			TotalBCLast1000:       tbclt,
			BlockCount:            bc,
			TotalBCLast1000t:      tbcltt,
			BlockStakeState:       bss,
			BlockStakeNumOf:       bsn,
			BlockStakeUTXOAddress: bsutxoa,
		})
	}
}

// NewWalletAddressHandler creates a handler to handle API calls to /wallet/address.
func NewWalletAddressHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		unlockHash, err := wallet.NextAddress()
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/addresses: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		WriteJSON(w, WalletAddressGET{
			Address: unlockHash,
		})
	}
}

// NewWalletAddressesHandler creates a handler to handle API calls to /wallet/addresses.
func NewWalletAddressesHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		addresses, err := wallet.AllAddresses()
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/addresses: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		WriteJSON(w, WalletAddressesGET{Addresses: addresses})
	}
}

// NewWalletBackupHandler creates a handler to handle API calls to /wallet/backup.
func NewWalletBackupHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		destination := req.FormValue("destination")
		// Check that the destination is absolute.
		if !filepath.IsAbs(destination) {
			WriteError(w, Error{"error when calling /wallet/backup: destination must be an absolute path"}, http.StatusBadRequest)
			return
		}
		err := wallet.CreateBackup(destination)
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/backup: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		WriteSuccess(w)
	}
}

// NewWalletInitHandler creates a handler to handle API calls to /wallet/init.
func NewWalletInitHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		passphrase := req.FormValue("passphrase")

		seedStr := req.FormValue("seed")
		var seed modules.Seed
		if seedStr != "" {
			err := seed.LoadString(seedStr)
			if err != nil {
				WriteError(w, Error{
					"error when calling /wallet/init: invalid seed given: " + err.Error()},
					http.StatusBadRequest)
				return
			}
		}

		var err error
		if passphrase == "" {
			seed, err = wallet.Init(seed)
			if err != nil {
				WriteError(w, Error{"error when calling /wallet/init: " + err.Error()}, http.StatusBadRequest)
				return
			}
		} else {
			ph, err := crypto.HashObject(passphrase)
			if err != nil {
				WriteError(w, Error{"error when calling /wallet/init: " + err.Error()}, http.StatusBadRequest)
				return
			}
			encryptionKey := crypto.TwofishKey(ph)
			seed, err = wallet.Encrypt(encryptionKey, seed)
			if err != nil {
				WriteError(w, Error{"error when calling /wallet/init: " + err.Error()}, http.StatusBadRequest)
				return
			}
		}

		mnemonic, err := modules.NewMnemonic(seed)
		if err != nil {
			WriteError(w, Error{"error when calling /wallet/init: " + err.Error()}, http.StatusBadRequest)
			return
		}
		WriteJSON(w, WalletInitPOST{
			PrimarySeed: mnemonic,
		})
	}
}

// NewWalletSeedHandler creates a handler to handle API calls to /wallet/seed.
func NewWalletSeedHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		menmonic := req.FormValue("mnemonic")
		passphrase := req.FormValue("passphrase")

		seed, err := modules.InitialSeedFromMnemonic(menmonic)
		if err != nil {
			WriteError(w, Error{"error when calling /wallet/seed: " + err.Error()}, http.StatusBadRequest)
			return
		}

		if passphrase == "" {
			err = wallet.LoadPlainSeed(seed)
			if err != nil {
				WriteError(w, Error{"error when calling /wallet/seed: " +
					modules.ErrBadEncryptionKey.Error()}, http.StatusBadRequest)
				return
			}
		} else {
			ph, err := crypto.HashObject(passphrase)
			if err != nil {
				WriteError(w, Error{"error when calling /wallet/seed: " + err.Error()}, http.StatusBadRequest)
				return
			}
			encryptionKey := crypto.TwofishKey(ph)
			err = wallet.LoadSeed(encryptionKey, seed)
			if err == modules.ErrBadEncryptionKey {
				WriteError(w, Error{"error when calling /wallet/seed: " +
					modules.ErrBadEncryptionKey.Error()}, http.StatusBadRequest)
				return
			}
			if err != nil {
				WriteError(w, Error{"error when calling /wallet/seed: " +
					err.Error()}, http.StatusBadRequest)
				return
			}
		}

		WriteSuccess(w)
	}
}

// NewWalletLockHandler creates a handler to handle API calls to /wallet/lock.
func NewWalletLockHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		err := wallet.Lock()
		if err != nil {
			WriteError(w, Error{err.Error()}, http.StatusBadRequest)
			return
		}
		WriteSuccess(w)
	}
}

// NewWalletSeedsHandler creates a handler to handle API calls to /wallet/seeds.
func NewWalletSeedsHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		// Get the primary seed information.
		primarySeed, progress, err := wallet.PrimarySeed()
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/seeds: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		primarySeedStr, err := modules.NewMnemonic(primarySeed)
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/seeds: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}

		// Get the list of seeds known to the wallet.
		allSeeds, err := wallet.AllSeeds()
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/seeds: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		var allSeedsStrs []string
		for _, seed := range allSeeds {
			str, err := modules.NewMnemonic(seed)
			if err != nil {
				WriteError(w, Error{"error after call to /wallet/seeds: " + err.Error()}, http.StatusBadRequest)
				return
			}
			allSeedsStrs = append(allSeedsStrs, str)
		}
		WriteJSON(w, WalletSeedsGET{
			PrimarySeed:        primarySeedStr,
			AddressesRemaining: int(modules.PublicKeysPerSeed - progress),
			AllSeeds:           allSeedsStrs,
		})
	}
}

// NewWalletKeyHandler creates a handler to handle API calls to /wallet/key/:unlockhash.
func NewWalletKeyHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		strUH := ps.ByName("unlockhash")
		uh, err := ScanAddress(strUH)
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/key/" + strUH + " : " + err.Error()},
				http.StatusBadRequest)
			return
		}

		pk, sk, err := wallet.GetKey(uh)
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/key/" + strUH + " : " + err.Error()},
				walletErrorToHTTPStatus(err))
			return
		}
		WriteJSON(w, WalletKeyGet{
			AlgorithmSpecifier: pk.Algorithm.Specifier(),
			PublicKey:          pk.Key,
			SecretKey:          sk,
		})
	}
}

// NewWalletTransactionCreateHandler creates a handler to handle API calls to POST /wallet/transaction.
func NewWalletTransactionCreateHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		var body WalletTransactionPOST
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			WriteError(w, Error{"error decoding the supplied transaction output: " + err.Error()}, http.StatusBadRequest)
			return
		}

		tx, err := wallet.SendCoins(body.Amount, body.Condition, []byte(body.Data))
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/transaction: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		WriteJSON(w, WalletTransactionPOSTResponse{
			Transaction: tx,
		})
	}
}

// NewWalletCoinsHandler creates a handler to handle API calls to /wallet/coins.
func NewWalletCoinsHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		var body WalletCoinsPOST
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			WriteError(w, Error{"error decoding the supplied coin outputs: " + err.Error()}, http.StatusBadRequest)
			return
		}
		tx, err := wallet.SendOutputs(body.CoinOutputs, nil, body.Data, body.RefundAddress, !body.GenerateRefundAddress)
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/coins: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		WriteJSON(w, WalletCoinsPOSTResp{
			TransactionID: tx.ID(),
		})
	}
}

// NewWalletBlockStakesHandler creates a handler to handle API calls to /wallet/blockstake.
func NewWalletBlockStakesHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		var body WalletBlockStakesPOST
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			WriteError(w, Error{"error decoding the supplied blockstake outputs: " + err.Error()}, http.StatusBadRequest)
			return
		}
		tx, err := wallet.SendOutputs(nil, body.BlockStakeOutputs, body.Data, body.RefundAddress, !body.GenerateRefundAddress)
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/blockstakes: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		WriteJSON(w, WalletBlockStakesPOSTResp{
			TransactionID: tx.ID(),
		})
	}
}

// NewWalletTransactionHandler creates a handler to handle API calls to /wallet/transaction/:id.
func NewWalletTransactionHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		// Parse the id from the url.
		var id types.TransactionID
		err := id.LoadString(ps.ByName("id"))
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/transaction/$(id): " + err.Error()}, http.StatusBadRequest)
			return
		}

		txn, ok, err := wallet.Transaction(id)
		if err != nil {
			WriteError(w, Error{"error when calling /wallet/transaction/$(id): " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		if !ok {
			WriteError(w, Error{"error when calling /wallet/transaction/$(id): transaction not found"}, http.StatusNotFound)
			return
		}
		WriteJSON(w, WalletTransactionGETid{
			Transaction: txn,
		})
	}
}

// NewWalletTransactionsHandler creates a handler to handle API calls to /wallet/transactions.
func NewWalletTransactionsHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		startheightStr, endheightStr := req.FormValue("startheight"), req.FormValue("endheight")
		if startheightStr == "" || endheightStr == "" {
			WriteError(w, Error{"startheight and endheight must be provided to a /wallet/transactions call."}, http.StatusBadRequest)
			return
		}
		// Get the start and end blocks.
		start, err := strconv.Atoi(startheightStr)
		if err != nil {
			WriteError(w, Error{"parsing integer value for parameter `startheight` failed: " + err.Error()}, http.StatusBadRequest)
			return
		}
		end, err := strconv.Atoi(endheightStr)
		if err != nil {
			WriteError(w, Error{"parsing integer value for parameter `endheight` failed: " + err.Error()}, http.StatusBadRequest)
			return
		}
		confirmedTxns, err := wallet.Transactions(types.BlockHeight(start), types.BlockHeight(end))
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/transactions: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		unconfirmedTxns, err := wallet.UnconfirmedTransactions()
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/transactions: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}

		WriteJSON(w, WalletTransactionsGET{
			ConfirmedTransactions:   confirmedTxns,
			UnconfirmedTransactions: unconfirmedTxns,
		})
	}
}

// NewWalletTransactionsAddrHandler creates a handler to handle API calls to /wallet/transactions/:addr.
func NewWalletTransactionsAddrHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		// Parse the address being input.
		jsonAddr := "\"" + ps.ByName("addr") + "\""
		var addr types.UnlockHash
		err := addr.UnmarshalJSON([]byte(jsonAddr))
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/transactions: " + err.Error()}, http.StatusBadRequest)
			return
		}

		confirmedATs, err := wallet.AddressTransactions(addr)
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/transactions: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		unconfirmedATs, err := wallet.AddressUnconfirmedTransactions(addr)
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/transactions: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		WriteJSON(w, WalletTransactionsGETaddr{
			ConfirmedTransactions:   confirmedATs,
			UnconfirmedTransactions: unconfirmedATs,
		})
	}
}

// NewWalletUnlockHandler creates a handler to handle API calls to /wallet/unlock.
func NewWalletUnlockHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		passphrase := req.FormValue("passphrase")
		if passphrase == "" {
			WriteError(w, Error{"error when calling /wallet/unlock: passphrase is required"},
				http.StatusUnauthorized)
			return
		}
		ph, err := crypto.HashObject(passphrase)
		if err != nil {
			WriteError(w, Error{"error when calling /wallet/unlock:" + err.Error()},
				http.StatusUnauthorized)
			return
		}
		encryptionKey := crypto.TwofishKey(ph)
		err = wallet.Unlock(encryptionKey)
		if err == nil {
			WriteSuccess(w)
			return
		}
		if err != modules.ErrBadEncryptionKey {
			WriteError(w, Error{"error when calling /wallet/unlock: " + err.Error()}, http.StatusBadRequest)
			return
		}

		WriteError(w, Error{"error when calling /wallet/unlock: " + modules.ErrBadEncryptionKey.Error()}, http.StatusBadRequest)
	}
}

// NewWalletListUnlockedHandler creates a handler to handle API calls to /wallet/unlocked
func NewWalletListUnlockedHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		ucos, ubsos, err := wallet.UnlockedUnspendOutputs()
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/unlocked: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		ucor := []UnspentCoinOutput{}
		ubsor := []UnspentBlockstakeOutput{}

		for id, co := range ucos {
			ucor = append(ucor, UnspentCoinOutput{ID: id, Output: co})
		}

		for id, bso := range ubsos {
			ubsor = append(ubsor, UnspentBlockstakeOutput{ID: id, Output: bso})
		}

		WriteJSON(w, WalletListUnlockedGET{
			UnlockedCoinOutputs:       ucor,
			UnlockedBlockstakeOutputs: ubsor,
		})
	}
}

// NewWalletListLockedHandler creates a handler to handle API calls to /wallet/locked
func NewWalletListLockedHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		ucos, ubsos, err := wallet.LockedUnspendOutputs()
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/locked: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		ucor := []UnspentCoinOutput{}
		ubsor := []UnspentBlockstakeOutput{}

		for id, co := range ucos {
			ucor = append(ucor, UnspentCoinOutput{ID: id, Output: co})
		}

		for id, bso := range ubsos {
			ubsor = append(ubsor, UnspentBlockstakeOutput{ID: id, Output: bso})
		}

		WriteJSON(w, WalletListLockedGET{
			LockedCoinOutputs:       ucor,
			LockedBlockstakeOutputs: ubsor,
		})
	}
}

// NewWalletCreateTransactionHandler creates a handler to handle API calls to POST /wallet/create/transaction
func NewWalletCreateTransactionHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		var body WalletCreateTransactionPOST
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			WriteError(w, Error{"error decoding the supplied inputs and outputs: " + err.Error()}, http.StatusBadRequest)
			return
		}
		tx, err := wallet.CreateRawTransaction(body.CoinInputs, body.BlockStakeInputs, body.CoinOutputs, body.BlockStakeOutputs, nil)
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/create/transaction: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		WriteJSON(w, WalletCreateTransactionRESP{
			Transaction: tx,
		})
	}
}

// NewWalletSignHandler creates a handler to handle API calls to POST /wallet/sign
func NewWalletSignHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		var body types.Transaction
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			WriteError(w, Error{"error decoding the supplied transaction: " + err.Error()}, http.StatusBadRequest)
			return
		}
		txn, err := wallet.GreedySign(body)
		if err != nil {
			WriteError(w, Error{"error after call to /wallet/sign: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		WriteJSON(w, txn)
	}
}

// NewWalletFundCoinsHandler creates a handler to handle the API calls to /wallet/fund/coins?amount=.
// While it might be handy for other use cases, it is needed for 3bot registration
func NewWalletFundCoinsHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		q := req.URL.Query()
		// parse the amount
		amountStr := q.Get("amount")
		if amountStr == "" || amountStr == "0" {
			WriteError(w, Error{Message: "an amount has to be specified, greater than 0"}, http.StatusBadRequest)
			return
		}
		var amount types.Currency
		err := amount.LoadString(amountStr)
		if err != nil {
			WriteError(w, Error{Message: "invalid amount given: " + err.Error()}, http.StatusBadRequest)
			return
		}

		// parse optional refund address and reuseRefundAddress from query params
		var (
			refundAddress    *types.UnlockHash
			newRefundAddress bool
		)
		refundStr := q.Get("refund")
		if refundStr != "" {
			// try as a bool
			var b bool
			n, err := fmt.Sscanf(refundStr, "%t", &b)
			if err == nil && n == 1 {
				newRefundAddress = b
			} else {
				// try as an address
				var uh types.UnlockHash
				err = uh.LoadString(refundStr)
				if err != nil {
					WriteError(w, Error{Message: fmt.Sprintf("refund query param has to be a boolean or unlockhash, %s is invalid", refundStr)}, http.StatusBadRequest)
					return
				}
				refundAddress = &uh
			}
		}

		// start a transaction and fund the requested amount
		txbuilder := wallet.StartTransaction()
		err = txbuilder.FundCoins(amount, refundAddress, !newRefundAddress)
		if err != nil {
			WriteError(w, Error{Message: "failed to fund the requested coins: " + err.Error()}, http.StatusInternalServerError)
			return
		}

		// build the dummy Txn, as to view the Txn
		txn, _ := txbuilder.View()
		// defer drop the Txn
		defer txbuilder.Drop()

		// compose the result object and validate it
		result := WalletFundCoins{CoinInputs: txn.CoinInputs}
		if len(result.CoinInputs) == 0 {
			WriteError(w, Error{Message: "no coin inputs could be generated"}, http.StatusInternalServerError)
			return
		}
		switch len(txn.CoinOutputs) {
		case 0:
			// ignore, valid, but nothing to do
		case 1:
			// add as refund
			result.RefundCoinOutput = &txn.CoinOutputs[0]
		case 2:
			WriteError(w, Error{Message: "more than 2 coin outputs were generated, while maximum 1 was expected"}, http.StatusInternalServerError)
			return
		}
		// all good, return the resulting object
		WriteJSON(w, result)
	}
}

// NewWalletGetPublicKeyHandler creates a handler to handle API calls to /wallet/publickey.
// While it might be handy for other use cases, it is needed for 3bot.
func NewWalletGetPublicKeyHandler(wallet modules.Wallet) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		unlockHash, err := wallet.NextAddress()
		if err != nil {
			WriteError(w, Error{Message: "error after call to /wallet/publickey: " + err.Error()}, walletErrorToHTTPStatus(err))
			return
		}
		pk, _, err := wallet.GetKey(unlockHash)
		if err != nil {
			WriteError(w, Error{Message: "failed to fetch newly created public key: " + err.Error()}, http.StatusInternalServerError)
			return
		}
		WriteJSON(w, WalletPublicKeyGET{PublicKey: pk})
	}
}

func walletErrorToHTTPStatus(err error) int {
	if err == modules.ErrLockedWallet {
		return http.StatusForbidden
	}
	if cErr, ok := err.(types.ClientError); ok {
		return cErr.Kind.AsHTTPStatusCode()
	}
	return http.StatusInternalServerError
}
