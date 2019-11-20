package tftschema

import (
	"encoding/json"
	schema "github.com/threefoldtech/zos/pkg/schema"
)

type TftExplorerWalletBalance1 struct {
	ObjectId            int64          `json:"object_id"`
	CoinsUnlocked       schema.Numeric `json:"coins_unlocked"`
	CoinsLocked         schema.Numeric `json:"coins_locked"`
	BlockStakesUnlocked schema.Numeric `json:"block_stakes_unlocked"`
	BlockStakesLocked   schema.Numeric `json:"block_stakes_locked"`
}

func NewTftExplorerWalletBalance1() (TftExplorerWalletBalance1, error) {
	const value = "{\"object_id\": 0, \"coins_unlocked\": 0, \"coins_locked\": 0, \"block_stakes_unlocked\": 0, \"block_stakes_locked\": 0}"
	var object TftExplorerWalletBalance1
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return object, err
	}
	return object, nil
}
