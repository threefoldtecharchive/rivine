package tftschema

import (
	"encoding/json"
	schema "github.com/threefoldtech/zos/pkg/schema"
)

type TftExplorerContractAtomicSwap1 struct {
	ObjectId             int64                                        `json:"object_id"`
	Value                schema.Numeric                               `json:"value"`
	Condition            string                                       `json:"condition"`
	TransactionObjectIds []int64                                      `json:"transaction_object_ids"`
	CoinInputObjectId    []int64                                      `json:"coin_input_object_id"`
	SpenditureData       TftExplorerContractAtomicSwapSpenditureData1 `json:"spenditure_data"`
}

func NewTftExplorerContractAtomicSwap1() (TftExplorerContractAtomicSwap1, error) {
	const value = "{\"object_id\": 0, \"value\": 0, \"condition\": \"\"}"
	var object TftExplorerContractAtomicSwap1
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return object, err
	}
	return object, nil
}

type TftExplorerContractAtomicSwapSpenditureData1 struct {
	Fulfillment        string `json:"fulfillment"`
	CoinOutputObjectId int64  `json:"coin_output_object_id"`
}

func NewTftExplorerContractAtomicSwapSpenditureData1() (TftExplorerContractAtomicSwapSpenditureData1, error) {
	const value = "{\"fulfillment\": \"\", \"coin_output_object_id\": 0}"
	var object TftExplorerContractAtomicSwapSpenditureData1
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return object, err
	}
	return object, nil
}
