package tftschema

import (
	"encoding/json"
	schema "github.com/threefoldtech/zos/pkg/schema"
)

type TftExplorerChainContext1 struct {
	ConsensusChangeId string      `json:"consensus_change_id"`
	Height            int64       `json:"height"`
	Timestamp         schema.Date `json:"timestamp"`
	BlockId           string      `json:"block_id"`
}

func NewTftExplorerChainContext1() (TftExplorerChainContext1, error) {
	const value = "{\"consensus_change_id\": \"\", \"height\": 0, \"timestamp\": 0, \"block_id\": \"\"}"
	var object TftExplorerChainContext1
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return object, err
	}
	return object, nil
}
