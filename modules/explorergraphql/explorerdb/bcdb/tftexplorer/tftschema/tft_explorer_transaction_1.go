package tftschema

import (
	"encoding/json"
	schema "github.com/threefoldtech/zos/pkg/schema"
)

type TftExplorerTransaction1 struct {
	ObjectId                  int64                                `json:"object_id"`
	ParentBlockObjectId       int64                                `json:"parent_block_object_id"`
	Version                   int64                                `json:"version"`
	CoinInputObjectIds        []int64                              `json:"coin_input_object_ids"`
	CoinOutputObjectIds       []int64                              `json:"coin_output_object_ids"`
	BlockStakeInputObjectIds  []int64                              `json:"block_stake_input_object_ids"`
	BlockStakeOutputObjectIds []int64                              `json:"block_stake_output_object_ids"`
	FeePayout                 TftExplorerTransactionPayoutFeeInfo1 `json:"fee_payout"`
	ArbitraryData             string                               `json:"arbitrary_data"`
	EncodedExtensionData      string                               `json:"encoded_extension_data"`
}

func NewTftExplorerTransaction1() (TftExplorerTransaction1, error) {
	const value = "{\"object_id\": 0, \"parent_block_object_id\": 0, \"version\": 0, \"arbitrary_data\": \"\", \"encoded_extension_data\": \"\"}"
	var object TftExplorerTransaction1
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return object, err
	}
	return object, nil
}

type TftExplorerTransactionPayoutFeeInfo1 struct {
	PayoutOutputObjectId int64            `json:"payout_output_object_id"`
	PayoutValues         []schema.Numeric `json:"payout_values"`
}

func NewTftExplorerTransactionPayoutFeeInfo1() (TftExplorerTransactionPayoutFeeInfo1, error) {
	const value = "{\"payout_output_object_id\": 0}"
	var object TftExplorerTransactionPayoutFeeInfo1
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return object, err
	}
	return object, nil
}
