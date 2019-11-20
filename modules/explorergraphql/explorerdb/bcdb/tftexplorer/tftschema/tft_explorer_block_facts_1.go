package tftschema

import (
	"encoding/json"
	schema "github.com/threefoldtech/zos/pkg/schema"
)

type TftExplorerBlockFacts1 struct {
	ObjectId                             int64          `json:"object_id"`
	Target                               string         `json:"target"`
	Difficulty                           schema.Numeric `json:"difficulty"`
	AggregatedTotalCoins                 schema.Numeric `json:"aggregated_total_coins"`
	AggregatedTotalCoinsLocked           schema.Numeric `json:"aggregated_total_coins_locked"`
	AggregatedTotalBlockStakes           schema.Numeric `json:"aggregated_total_block_stakes"`
	AggregatedTotalBlockStakesLocked     schema.Numeric `json:"aggregated_total_block_stakes_locked"`
	AggregatedEstimatedBlockStakesActive schema.Numeric `json:"aggregated_estimated_block_stakes_active"`
}

func NewTftExplorerBlockFacts1() (TftExplorerBlockFacts1, error) {
	const value = "{\"object_id\": 0, \"target\": \"\", \"difficulty\": 0, \"aggregated_total_coins\": 0, \"aggregated_total_coins_locked\": 0, \"aggregated_total_block_stakes\": 0, \"aggregated_total_block_stakes_locked\": 0, \"aggregated_estimated_block_stakes_active\": 0}"
	var object TftExplorerBlockFacts1
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return object, err
	}
	return object, nil
}
