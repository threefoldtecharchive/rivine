package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"

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
		CoinOutputs []types.CoinOutput `json:"coinoutputs`
	}
	// WalletCoinsPOSTResp Resp contains the ID of the transaction
	// that was created as a result of a POST call to /wallet/coins.
	WalletCoinsPOSTResp struct {
		TransactionID types.TransactionID `json:"transactionid"`
	}

	// WalletBlockStakesPOST is given by the user
	// to indicate to where to send how much blockstakes
	WalletBlockStakesPOST struct {
		BlockStakeOutputs []types.BlockStakeOutput `json:"blockstakeoutputs`
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
)

// walletHander handles API calls to /wallet.
func (api *API) walletHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	coinBal, blockstakeBal, err := api.wallet.ConfirmedBalance()
	if err != nil {
		WriteError(w, Error{"error after call to /wallet: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}
	coinLockBal, blockstakeLockBal, err := api.wallet.ConfirmedLockedBalance()
	if err != nil {
		WriteError(w, Error{"error after call to /wallet: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}
	coinsOut, coinsIn, err := api.wallet.UnconfirmedBalance()
	if err != nil {
		WriteError(w, Error{"error after call to /wallet: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}
	multiSigWallets, err := api.wallet.MultiSigWallets()
	if err != nil {
		WriteError(w, Error{"error after call to /wallet: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}

	WriteJSON(w, WalletGET{
		Encrypted: api.wallet.Encrypted(),
		Unlocked:  api.wallet.Unlocked(),

		ConfirmedCoinBalance:       coinBal,
		ConfirmedLockedCoinBalance: coinLockBal,
		UnconfirmedOutgoingCoins:   coinsOut,
		UnconfirmedIncomingCoins:   coinsIn,

		BlockStakeBalance:       blockstakeBal,
		LockedBlockStakeBalance: blockstakeLockBal,

		MultiSigWallets: multiSigWallets,
	})
}

// walletBlockStakeStats handles API calls to /wallet/blockstakestat.
func (api *API) walletBlockStakeStats(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	unspentBSOs, err := api.wallet.GetUnspentBlockStakeOutputs()
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
	tbclt, bsf, bc, err := api.wallet.BlockStakeStats()
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

// walletAddressHandler handles API calls to /wallet/address.
func (api *API) walletAddressHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	unlockHash, err := api.wallet.NextAddress()
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/addresses: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}
	WriteJSON(w, WalletAddressGET{
		Address: unlockHash,
	})
}

// walletAddressHandler handles API calls to /wallet/addresses.
func (api *API) walletAddressesHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	addresses, err := api.wallet.AllAddresses()
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/addresses: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}
	WriteJSON(w, WalletAddressesGET{Addresses: addresses})
}

// walletBackupHandler handles API calls to /wallet/backup.
func (api *API) walletBackupHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	destination := req.FormValue("destination")
	// Check that the destination is absolute.
	if !filepath.IsAbs(destination) {
		WriteError(w, Error{"error when calling /wallet/backup: destination must be an absolute path"}, http.StatusBadRequest)
		return
	}
	err := api.wallet.CreateBackup(destination)
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/backup: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}
	WriteSuccess(w)
}

// walletInitHandler handles API calls to /wallet/init.
func (api *API) walletInitHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	passphrase := req.FormValue("passphrase")
	if passphrase == "" {
		WriteError(w, Error{"error when calling /wallet/init: passphrase is required"},
			http.StatusUnauthorized)
		return
	}

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

	encryptionKey := crypto.TwofishKey(crypto.HashObject(passphrase))
	seed, err := api.wallet.Encrypt(encryptionKey, seed)
	if err != nil {
		WriteError(w, Error{"error when calling /wallet/init: " + err.Error()}, http.StatusBadRequest)
		return
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

// walletSeedHandler handles API calls to /wallet/seed.
func (api *API) walletSeedHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	menmonic := req.FormValue("mnemonic")
	passphrase := req.FormValue("passphrase")
	if passphrase == "" {
		WriteError(w, Error{"error when calling /wallet/seed: passphrase is required"},
			http.StatusUnauthorized)
		return
	}

	seed, err := modules.InitialSeedFromMnemonic(menmonic)
	if err != nil {
		WriteError(w, Error{"error when calling /wallet/seed: " + err.Error()}, http.StatusBadRequest)
		return
	}

	encryptionKey := crypto.TwofishKey(crypto.HashObject(passphrase))
	err = api.wallet.LoadSeed(encryptionKey, seed)
	if err == nil {
		WriteSuccess(w)
		return
	}
	if err != modules.ErrBadEncryptionKey {
		WriteError(w, Error{"error when calling /wallet/seed: " + err.Error()}, http.StatusBadRequest)
		return
	}
	WriteError(w, Error{"error when calling /wallet/seed: " + modules.ErrBadEncryptionKey.Error()}, http.StatusBadRequest)
}

// walletLockHanlder handles API calls to /wallet/lock.
func (api *API) walletLockHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	err := api.wallet.Lock()
	if err != nil {
		WriteError(w, Error{err.Error()}, http.StatusBadRequest)
		return
	}
	WriteSuccess(w)
}

// walletSeedsHandler handles API calls to /wallet/seeds.
func (api *API) walletSeedsHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	// Get the primary seed information.
	primarySeed, progress, err := api.wallet.PrimarySeed()
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
	allSeeds, err := api.wallet.AllSeeds()
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

// walletKeyHandler handles API calls to /wallet/key/:unlockhash.
func (api *API) walletKeyHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	strUH := ps.ByName("unlockhash")
	var uh types.UnlockHash
	err := uh.LoadString(strUH)
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/key/" + strUH + " : " + err.Error()},
			http.StatusBadRequest)
		return
	}

	pk, sk, err := api.wallet.GetKey(uh)
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/key/" + strUH + " : " + err.Error()},
			walletErrorToHTTPStatus(err))
		return
	}
	WriteJSON(w, WalletKeyGet{
		AlgorithmSpecifier: pk.Algorithm,
		PublicKey:          pk.Key,
		SecretKey:          sk,
	})
}

// walletTransactionCreateHandler handles API calls to POST /wallet/transaction.
func (api *API) walletTransactionCreateHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var body WalletTransactionPOST
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		WriteError(w, Error{"error decoding the supplied transaction output: " + err.Error()}, http.StatusBadRequest)
		return
	}

	tx, err := api.wallet.SendCoins(body.Amount, body.Condition, []byte(body.Data))
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/transaction: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}
	WriteJSON(w, WalletTransactionPOSTResponse{
		Transaction: tx,
	})
}

// walletSiacoinsHandler handles API calls to /wallet/coins.
func (api *API) walletCoinsHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var body WalletCoinsPOST
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		WriteError(w, Error{"error decoding the supplied coin outputs: " + err.Error()}, http.StatusBadRequest)
		return
	}
	tx, err := api.wallet.SendOutputs(body.CoinOutputs, nil, nil)
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/coins: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}
	WriteJSON(w, WalletCoinsPOSTResp{
		TransactionID: tx.ID(),
	})
}

// walletSiafundsHandler handles API calls to /wallet/blockstake.
func (api *API) walletBlockStakesHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var body WalletBlockStakesPOST
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		WriteError(w, Error{"error decoding the supplied blockstake outputs: " + err.Error()}, http.StatusBadRequest)
		return
	}
	tx, err := api.wallet.SendOutputs(nil, body.BlockStakeOutputs, nil)
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/blockstakes: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}
	WriteJSON(w, WalletBlockStakesPOSTResp{
		TransactionID: tx.ID(),
	})
}

// walletDataHandler handles the API calls to /wallet/data
func (api *API) walletDataHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	dest, err := scanAddress(req.FormValue("destination"))
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/coins: " + err.Error()}, http.StatusBadRequest)
		return
	}
	dataString := req.FormValue("data")
	data, err := base64.StdEncoding.DecodeString(dataString)
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/coins: Failed to decode arbitrary data"}, http.StatusBadRequest)
		return
	}
	// Since zero outputs are not allowed, just send one of the smallest unit, the minimal amount.
	// The transaction fee should be much higher anyway
	tx, err := api.wallet.SendCoins(types.NewCurrency64(1),
		types.NewCondition(types.NewUnlockHashCondition(dest)), data)
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/coins: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}
	WriteJSON(w, WalletCoinsPOSTResp{
		TransactionID: tx.ID(),
	})
}

// walletTransactionHandler handles API calls to /wallet/transaction/:id.
func (api *API) walletTransactionHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	// Parse the id from the url.
	var id types.TransactionID
	err := id.LoadString(ps.ByName("id"))
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/transaction/$(id): " + err.Error()}, http.StatusBadRequest)
		return
	}

	txn, ok, err := api.wallet.Transaction(id)
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

// walletTransactionsHandler handles API calls to /wallet/transactions.
func (api *API) walletTransactionsHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
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
	confirmedTxns, err := api.wallet.Transactions(types.BlockHeight(start), types.BlockHeight(end))
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/transactions: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}
	unconfirmedTxns, err := api.wallet.UnconfirmedTransactions()
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/transactions: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}

	WriteJSON(w, WalletTransactionsGET{
		ConfirmedTransactions:   confirmedTxns,
		UnconfirmedTransactions: unconfirmedTxns,
	})
}

// walletTransactionsAddrHandler handles API calls to
// /wallet/transactions/:addr.
func (api *API) walletTransactionsAddrHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	// Parse the address being input.
	jsonAddr := "\"" + ps.ByName("addr") + "\""
	var addr types.UnlockHash
	err := addr.UnmarshalJSON([]byte(jsonAddr))
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/transactions: " + err.Error()}, http.StatusBadRequest)
		return
	}

	confirmedATs, err := api.wallet.AddressTransactions(addr)
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/transactions: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}
	unconfirmedATs, err := api.wallet.AddressUnconfirmedTransactions(addr)
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/transactions: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}
	WriteJSON(w, WalletTransactionsGETaddr{
		ConfirmedTransactions:   confirmedATs,
		UnconfirmedTransactions: unconfirmedATs,
	})
}

// walletUnlockHandler handles API calls to /wallet/unlock.
func (api *API) walletUnlockHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	passphrase := req.FormValue("passphrase")
	if passphrase == "" {
		WriteError(w, Error{"error when calling /wallet/unlock: passphrase is required"},
			http.StatusUnauthorized)
		return
	}
	encryptionKey := crypto.TwofishKey(crypto.HashObject(passphrase))
	err := api.wallet.Unlock(encryptionKey)
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

// walletListUnlcokedHandler handles API calls to /wallet/unlocked
func (api *API) walletListUnlockedHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	ucos, ubsos, err := api.wallet.UnlockedUnspendOutputs()
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

// walletListUnlcokedHandler handles API calls to /wallet/locked
func (api *API) walletListLockedHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	ucos, ubsos, err := api.wallet.LockedUnspendOutputs()
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

	WriteJSON(w, WalletListUnlockedGET{
		UnlockedCoinOutputs:       ucor,
		UnlockedBlockstakeOutputs: ubsor,
	})
}

// walletCreateTransactionHandler handles API calls to POST /wallet/create/transaction
func (api *API) walletCreateTransactionHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var body WalletCreateTransactionPOST
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		WriteError(w, Error{"error decoding the supplied inputs and outputs: " + err.Error()}, http.StatusBadRequest)
		return
	}
	tx, err := api.wallet.CreateRawTransaction(body.CoinInputs, body.BlockStakeInputs, body.CoinOutputs, body.BlockStakeOutputs, nil)
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/create/transaction: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}
	WriteJSON(w, WalletCreateTransactionRESP{
		Transaction: tx,
	})
}

// walletSignHandler handles API calls to POST /wallet/sign
func (api *API) walletSignHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var body types.Transaction
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		WriteError(w, Error{"error decoding the supplied transaction: " + err.Error()}, http.StatusBadRequest)
		return
	}
	txn, err := api.wallet.GreedySign(body)
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/sign: " + err.Error()}, walletErrorToHTTPStatus(err))
		return
	}
	WriteJSON(w, txn)
}

func walletErrorToHTTPStatus(err error) int {
	if err == modules.ErrLockedWallet {
		return http.StatusForbidden
	}
	return http.StatusInternalServerError
}
