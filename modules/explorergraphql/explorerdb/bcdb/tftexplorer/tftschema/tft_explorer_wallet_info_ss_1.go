package tftschema

import "encoding/json"

type TftExplorerWalletInfoSs1 struct {
	ObjectId                      int64   `json:"object_id"`
	MultiSignatureWalletObjectIds []int64 `json:"multi_signature_wallet_object_ids"`
}

func NewTftExplorerWalletInfoSs1() (TftExplorerWalletInfoSs1, error) {
	const value = "{\"object_id\": 0}"
	var object TftExplorerWalletInfoSs1
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return object, err
	}
	return object, nil
}
