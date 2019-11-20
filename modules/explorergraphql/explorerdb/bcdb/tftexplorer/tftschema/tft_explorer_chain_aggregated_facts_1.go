package tftschema

import (
	"encoding/json"
	schema "github.com/threefoldtech/zos/pkg/schema"
)

type TftExplorerChainAggregatedFacts1 struct {
	TotalCoins                 schema.Numeric                  `json:"total_coins"`
	TotalCoinsLocked           schema.Numeric                  `json:"total_coins_locked"`
	TotalBlockStakes           schema.Numeric                  `json:"total_block_stakes"`
	TotalBlockStakesLocked     schema.Numeric                  `json:"total_block_stakes_locked"`
	EstimatedBlockStakesActive schema.Numeric                  `json:"estimated_block_stakes_active"`
	LastBlocks                 []TftExplorerBlockFactsContext1 `json:"last_blocks"`
}

func NewTftExplorerChainAggregatedFacts1() (TftExplorerChainAggregatedFacts1, error) {
	const value = "{\"total_coins\": \"0\", \"total_coins_locked\": \"0\", \"total_block_stakes\": \"0\", \"total_block_stakes_locked\": \"0\", \"estimated_block_stakes_active\": \"0\"}"
	var object TftExplorerChainAggregatedFacts1
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return object, err
	}
	return object, nil
}

type TftExplorerBlockFactsContext1 struct {
	Target    string      `json:"target"`
	Timestamp schema.Date `json:"timestamp"`
}

func NewTftExplorerBlockFactsContext1() (TftExplorerBlockFactsContext1, error) {
	const value = "{\"target\": \"\", \"timestamp\": 0}"
	var object TftExplorerBlockFactsContext1
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return object, err
	}
	return object, nil
}
