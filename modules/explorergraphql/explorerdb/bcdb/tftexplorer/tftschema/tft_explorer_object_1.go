package tftschema

import "encoding/json"

type TftExplorerObject1 struct {
	ObjectType    int64  `json:"object_type"`
	ObjectVersion int64  `json:"object_version"`
	ObjectHash    string `json:"object_hash"`
}

func NewTftExplorerObject1() (TftExplorerObject1, error) {
	const value = "{\"object_type\": 0, \"object_version\": 0, \"object_hash\": \"\"}"
	var object TftExplorerObject1
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return object, err
	}
	return object, nil
}
