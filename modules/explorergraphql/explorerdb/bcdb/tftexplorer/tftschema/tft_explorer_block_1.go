package tftschema

import (
	"encoding/json"
	schema "github.com/threefoldtech/zos/pkg/schema"
)

type TftExplorerBlock1 struct {
	ObjectId             int64       `json:"object_id"`
	ParentObjectId       int64       `json:"parent_object_id"`
	Height               int64       `json:"height"`
	Timestamp            schema.Date `json:"timestamp"`
	PayoutObjectIds      []int64     `json:"payout_object_ids"`
	TransactionObjectIds []int64     `json:"transaction_object_ids"`
}

func NewTftExplorerBlock1() (TftExplorerBlock1, error) {
	const value = "{\"object_id\": 0, \"parent_object_id\": 0, \"height\": 0, \"timestamp\": 0}"
	var object TftExplorerBlock1
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return object, err
	}
	return object, nil
}
