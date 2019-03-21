package internal

import (
	"encoding/binary"

	"github.com/threefoldtech/rivine/types"
)

// EncodeBlockheight encodes the given blockheight as a sortable key
func EncodeBlockheight(height types.BlockHeight) []byte {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key[:], uint64(height))
	return key
}

// DecodeBlockheight decodes the given sortable key as a blockheight
func DecodeBlockheight(key []byte) types.BlockHeight {
	return types.BlockHeight(binary.BigEndian.Uint64(key))
}
