package tftschema

import "encoding/json"

type TftExplorerWalletInfoMs1 struct {
	ObjectId               int64   `json:"object_id"`
	OwnerObjectIds         []int64 `json:"owner_object_ids"`
	RequiredSignatureCount int64   `json:"required_signature_count"`
}

func NewTftExplorerWalletInfoMs1() (TftExplorerWalletInfoMs1, error) {
	const value = "{\"object_id\": 0}"
	var object TftExplorerWalletInfoMs1
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return object, err
	}
	return object, nil
}
