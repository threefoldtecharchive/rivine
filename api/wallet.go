package api

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/modules"
	"github.com/rivine/rivine/types"

	"github.com/NebulousLabs/entropy-mnemonics"
	"github.com/julienschmidt/httprouter"
)

type (
	// WalletGET contains general information about the wallet.
	WalletGET struct {
		Encrypted bool `json:"encrypted"`
		Unlocked  bool `json:"unlocked"`

		ConfirmedCoinBalance     types.Currency `json:"confirmedcoinbalance"`
		UnconfirmedOutgoingCoins types.Currency `json:"unconfirmedoutgoingcoins"`
		UnconfirmedIncomingCoins types.Currency `json:"unconfirmedincomingcoins"`

		BlockStakeBalance types.Currency `json:"blockstakebalance"`
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

	// WalletInitPOST contains the primary seed that gets generated during a
	// POST call to /wallet/init.
	WalletInitPOST struct {
		PrimarySeed string `json:"primaryseed"`
	}

	// WalletCoinsPOST contains the transaction sent in the POST call to
	// /wallet/siafunds.
	WalletCoinsPOST struct {
		TransactionIDs []types.TransactionID `json:"transactionids"`
	}

	// WalletBlockStakesPOST contains the transaction sent in the POST call to
	// /wallet/blockstakes.
	WalletBlockStakesPOST struct {
		TransactionIDs []types.TransactionID `json:"transactionids"`
	}

	// WalletSeedsGET contains the seeds used by the wallet.
	WalletSeedsGET struct {
		PrimarySeed        string   `json:"primaryseed"`
		AddressesRemaining int      `json:"addressesremaining"`
		AllSeeds           []string `json:"allseeds"`
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
)

// encryptionKeys enumerates the possible encryption keys that can be derived
// from an input string.
func encryptionKeys(seedStr string) (validKeys []crypto.TwofishKey) {
	dicts := []mnemonics.DictionaryID{"english", "german", "japanese"}
	for _, dict := range dicts {
		seed, err := modules.StringToSeed(seedStr, dict)
		if err != nil {
			continue
		}
		validKeys = append(validKeys, crypto.TwofishKey(crypto.HashObject(seed)))
	}
	validKeys = append(validKeys, crypto.TwofishKey(crypto.HashObject(seedStr)))
	return validKeys
}

// walletHander handles API calls to /wallet.
func (api *API) walletHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	coinBal, blockstakeBal := api.wallet.ConfirmedBalance()
	coinsOut, coinsIn := api.wallet.UnconfirmedBalance()
	WriteJSON(w, WalletGET{
		Encrypted: api.wallet.Encrypted(),
		Unlocked:  api.wallet.Unlocked(),

		ConfirmedCoinBalance:     coinBal,
		UnconfirmedOutgoingCoins: coinsOut,
		UnconfirmedIncomingCoins: coinsIn,

		BlockStakeBalance: blockstakeBal,
	})
}

// walletHander handles API calls to /wallet.
func (api *API) walletBlockStakeStats(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	if !api.wallet.Unlocked() {
		WriteError(w, Error{"error after call to /wallet/blockstakestat: wallet must be unlocked before it can be used"}, http.StatusBadRequest)
		return
	}
	count := len(api.wallet.GetUnspentBlockStakeOutputs())
	bss := make([]uint64, count)
	bsn := make([]types.Currency, count)
	bsutxoa := make([]types.BlockStakeOutputID, count)
	tabs := types.NewCurrency64(1000000) //TODO rivine change this to estimated num of BS
	tbs := types.NewCurrency64(0)

	num := 0
	tbclt, bsf, bc := api.wallet.BlockStakeStats()

	for _, ubso := range api.wallet.GetUnspentBlockStakeOutputs() {
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
	unlockConditions, err := api.wallet.NextAddress()
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/addresses: " + err.Error()}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, WalletAddressGET{
		Address: unlockConditions.UnlockHash(),
	})
}

// walletAddressHandler handles API calls to /wallet/addresses.
func (api *API) walletAddressesHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	WriteJSON(w, WalletAddressesGET{
		Addresses: api.wallet.AllAddresses(),
	})
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
		WriteError(w, Error{"error after call to /wallet/backup: " + err.Error()}, http.StatusBadRequest)
		return
	}
	WriteSuccess(w)
}

// walletInitHandler handles API calls to /wallet/init.
func (api *API) walletInitHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var encryptionKey crypto.TwofishKey
	if req.FormValue("encryptionpassword") != "" {
		encryptionKey = crypto.TwofishKey(crypto.HashObject(req.FormValue("encryptionpassword")))
	}
	seed, err := api.wallet.Encrypt(encryptionKey)
	if err != nil {
		WriteError(w, Error{"error when calling /wallet/init: " + err.Error()}, http.StatusBadRequest)
		return
	}

	dictID := mnemonics.DictionaryID(req.FormValue("dictionary"))
	if dictID == "" {
		dictID = "english"
	}
	seedStr, err := modules.SeedToString(seed, dictID)
	if err != nil {
		WriteError(w, Error{"error when calling /wallet/init: " + err.Error()}, http.StatusBadRequest)
		return
	}
	WriteJSON(w, WalletInitPOST{
		PrimarySeed: seedStr,
	})
}

// walletSeedHandler handles API calls to /wallet/seed.
func (api *API) walletSeedHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	// Get the seed using the ditionary + phrase
	dictID := mnemonics.DictionaryID(req.FormValue("dictionary"))
	if dictID == "" {
		dictID = "english"
	}
	seed, err := modules.StringToSeed(req.FormValue("seed"), dictID)
	if err != nil {
		WriteError(w, Error{"error when calling /wallet/seed: " + err.Error()}, http.StatusBadRequest)
		return
	}

	potentialKeys := encryptionKeys(req.FormValue("encryptionpassword"))
	for _, key := range potentialKeys {
		err := api.wallet.LoadSeed(key, seed)
		if err == nil {
			WriteSuccess(w)
			return
		}
		if err != nil && err != modules.ErrBadEncryptionKey {
			WriteError(w, Error{"error when calling /wallet/seed: " + err.Error()}, http.StatusBadRequest)
			return
		}
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
	dictionary := mnemonics.DictionaryID(req.FormValue("dictionary"))
	if dictionary == "" {
		dictionary = mnemonics.English
	}

	// Get the primary seed information.
	primarySeed, progress, err := api.wallet.PrimarySeed()
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/seeds: " + err.Error()}, http.StatusBadRequest)
		return
	}
	primarySeedStr, err := modules.SeedToString(primarySeed, dictionary)
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/seeds: " + err.Error()}, http.StatusBadRequest)
		return
	}

	// Get the list of seeds known to the wallet.
	allSeeds, err := api.wallet.AllSeeds()
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/seeds: " + err.Error()}, http.StatusBadRequest)
		return
	}
	var allSeedsStrs []string
	for _, seed := range allSeeds {
		str, err := modules.SeedToString(seed, dictionary)
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

// walletSiacoinsHandler handles API calls to /wallet/coins.
func (api *API) walletCoinsHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	amount, ok := scanAmount(req.FormValue("amount"))
	if !ok {
		WriteError(w, Error{"could not read 'amount' from POST call to /wallet/coins"}, http.StatusBadRequest)
		return
	}
	dest, err := scanAddress(req.FormValue("destination"))
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/coins: " + err.Error()}, http.StatusBadRequest)
		return
	}

	txns, err := api.wallet.SendCoins(amount, dest)
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/coins: " + err.Error()}, http.StatusInternalServerError)
		return
	}
	var txids []types.TransactionID
	for _, txn := range txns {
		txids = append(txids, txn.ID())
	}
	WriteJSON(w, WalletCoinsPOST{
		TransactionIDs: txids,
	})
}

// walletSiafundsHandler handles API calls to /wallet/blockstake.
func (api *API) walletBlockStakesHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	amount, ok := scanAmount(req.FormValue("amount"))
	if !ok {
		WriteError(w, Error{"could not read 'amount' from POST call to /wallet/blockstakes"}, http.StatusBadRequest)
		return
	}
	dest, err := scanAddress(req.FormValue("destination"))
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/blockstakes: " + err.Error()}, http.StatusBadRequest)
		return
	}

	txns, err := api.wallet.SendBlockStakes(amount, dest)
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/blockstakes: " + err.Error()}, http.StatusInternalServerError)
		return
	}
	var txids []types.TransactionID
	for _, txn := range txns {
		txids = append(txids, txn.ID())
	}
	WriteJSON(w, WalletBlockStakesPOST{
		TransactionIDs: txids,
	})
}

// walletTransactionHandler handles API calls to /wallet/transaction/:id.
func (api *API) walletTransactionHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	// Parse the id from the url.
	var id types.TransactionID
	jsonID := "\"" + ps.ByName("id") + "\""
	err := id.UnmarshalJSON([]byte(jsonID))
	if err != nil {
		WriteError(w, Error{"error after call to /wallet/history: " + err.Error()}, http.StatusBadRequest)
		return
	}

	txn, ok := api.wallet.Transaction(id)
	if !ok {
		WriteError(w, Error{"error when calling /wallet/transaction/$(id): transaction not found"}, http.StatusBadRequest)
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
		WriteError(w, Error{"error after call to /wallet/transactions: " + err.Error()}, http.StatusBadRequest)
		return
	}
	unconfirmedTxns := api.wallet.UnconfirmedTransactions()

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

	confirmedATs := api.wallet.AddressTransactions(addr)
	unconfirmedATs := api.wallet.AddressUnconfirmedTransactions(addr)
	WriteJSON(w, WalletTransactionsGETaddr{
		ConfirmedTransactions:   confirmedATs,
		UnconfirmedTransactions: unconfirmedATs,
	})
}

// walletUnlockHandler handles API calls to /wallet/unlock.
func (api *API) walletUnlockHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	potentialKeys := encryptionKeys(req.FormValue("encryptionpassword"))
	for _, key := range potentialKeys {
		err := api.wallet.Unlock(key)
		if err == nil {
			WriteSuccess(w)
			return
		}
		if err != nil && err != modules.ErrBadEncryptionKey {
			WriteError(w, Error{"error when calling /wallet/unlock: " + err.Error()}, http.StatusBadRequest)
			return
		}
	}
	WriteError(w, Error{"error when calling /wallet/unlock: " + modules.ErrBadEncryptionKey.Error()}, http.StatusBadRequest)
}
