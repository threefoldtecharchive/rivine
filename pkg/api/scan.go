package api

import (
	"math/big"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/types"
)

// ScanAmount scans a types.Currency from a string.
func ScanAmount(amount string) (types.Currency, bool) {
	// use SetString manually to ensure that amount does not contain
	// multiple values, which would confuse fmt.Scan
	i, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return types.Currency{}, ok
	}
	return types.NewCurrency(i), true
}

// ScanAddress scans a types.UnlockHash from a string.
func ScanAddress(addrStr string) (addr types.UnlockHash, err error) {
	err = addr.LoadString(addrStr)
	if err != nil {
		return types.UnlockHash{}, err
	}
	return addr, nil
}

// ScanHash scans a crypto.Hash from a string.
func ScanHash(s string) (h crypto.Hash, err error) {
	err = h.LoadString(s)
	if err != nil {
		return crypto.Hash{}, err
	}
	return h, nil
}
