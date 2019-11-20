package tftschema

import (
	"encoding/json"
	schema "github.com/threefoldtech/zos/pkg/schema"
)

type TftExplorerOutput1 struct {
	ObjectId             int64                            `json:"object_id"`
	ParentObjectId       int64                            `json:"parent_object_id"`
	Type                 int64                            `json:"type"`
	Value                schema.Numeric                   `json:"value"`
	Condition            string                           `json:"condition"`
	UnlockReferencePoint int64                            `json:"unlock_reference_point"`
	SpenditureData       TftExplorerOutputSpenditureData1 `json:"spenditure_data"`
}

func NewTftExplorerOutput1() (TftExplorerOutput1, error) {
	const value = "{\"object_id\": 0, \"parent_object_id\": 0, \"type\": 0, \"value\": 0, \"condition\": \"\", \"unlock_reference_point\": 0}"
	var object TftExplorerOutput1
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return object, err
	}
	return object, nil
}

type TftExplorerOutputSpenditureData1 struct {
	Fulfillment                    string `json:"fulfillment"`
	FulfillmentTransactionObjectId int64  `json:"fulfillment_transaction_object_id"`
}

func NewTftExplorerOutputSpenditureData1() (TftExplorerOutputSpenditureData1, error) {
	const value = "{\"fulfillment\": \"\", \"fulfillment_transaction_object_id\": 0}"
	var object TftExplorerOutputSpenditureData1
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return object, err
	}
	return object, nil
}
