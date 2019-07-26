package types

// block.go defines the Block type for Sia, and provides some helper functions
// for working with blocks.

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/threefoldtech/rivine/build"
	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/pkg/encoding/rivbin"
	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
)

const (
	// BlockHeaderSize is the size, in bytes, of a block header.
	// 32 (ParentID) + 8 (Timestamp) + 24 (8 BlockHeight + 8 TransactionIndex + 8 OutputIndex) +  32 (MerkleRoot)
	BlockHeaderSize = 96
)

type (
	// A Block is a summary of changes to the state that have occurred since the
	// previous block. Blocks reference the ID of the previous block (their
	// "parent"), creating the linked-list commonly known as the blockchain. Their
	// primary function is to bundle together transactions on the network. Blocks
	// are created by "blockcreators," who collect transactions from other nodes, and
	// then use there BlockStake and some other parameters to come below a given
	// target.
	Block struct {
		ParentID     BlockID                 `json:"parentid"`
		Timestamp    Timestamp               `json:"timestamp"`
		POBSOutput   BlockStakeOutputIndexes `json:"pobsindexes"`
		MinerPayouts []MinerPayout           `json:"minerpayouts"`
		Transactions []Transaction           `json:"transactions"`
	}

	// MinerPayout defines a miner payout, as (to be) paid out,
	// as orchestrated by this block's transactions.
	//
	// It is a structure created as to keep the Block Structure unchanged,
	// to how it used to be.
	MinerPayout struct {
		Value      Currency   `json:"value"`
		UnlockHash UnlockHash `json:"unlockhash"`
	}

	// A BlockHeader, when encoded, is an 96-byte constant size field
	// containing enough information to do headers-first block downloading.
	// Hashing the header results in the block ID.
	BlockHeader struct {
		ParentID   BlockID                 `json:"parentid"`
		POBSOutput BlockStakeOutputIndexes `json:"pobsindexes"`
		Timestamp  Timestamp               `json:"timestamp"`
		MerkleRoot crypto.Hash             `json:"merkleroot"`
	}

	BlockHeight uint64
	BlockID     crypto.Hash
)

// ID returns the ID of a Block, which is calculated by hashing the header.
func (h BlockHeader) ID() BlockID {
	hash, err := crypto.HashObject(h)
	if err != nil {
		build.Severe(fmt.Errorf("failed to crypto-hash block header into a block ID: %v", err))
	}
	return BlockID(hash)
}

// Header returns the header of a block.
func (b Block) Header() BlockHeader {
	return BlockHeader{
		ParentID:   b.ParentID,
		Timestamp:  b.Timestamp,
		POBSOutput: b.POBSOutput,
		MerkleRoot: b.MerkleRoot(),
	}
}

// CalculateTotalMinerFees determines the sum of the minerfees
func (b Block) CalculateTotalMinerFees() Currency {
	sum := NewCurrency64(0)
	for _, txn := range b.Transactions {
		for _, fee := range txn.MinerFees {
			sum = sum.Add(fee)
		}
	}
	return sum
}

// ID returns the ID of a Block, which is calculated by hashing the
// concatenation of the block's parent's ID, nonce, and the result of the
// b.MerkleRoot(). It is equivalent to calling block.Header().ID()
func (b Block) ID() BlockID {
	return b.Header().ID()
}

// MerkleRoot calculates the Merkle root of a Block. The leaves of the Merkle
// tree are composed of the miner outputs (one leaf per payout), and the
// transactions (one leaf per transaction).
func (b Block) MerkleRoot() crypto.Hash {
	tree := crypto.NewTree()
	var err error
	for _, payout := range b.MinerPayouts {
		err = tree.PushObject(payout)
		if err != nil {
			build.Severe(err)
		}
	}
	for _, txn := range b.Transactions {
		err = tree.PushObject(txn)
		if err != nil {
			build.Severe(err)
		}
	}
	return tree.Root()
}

// MinerPayoutID returns the ID of the miner payout at the given index, which
// is calculated by hashing the concatenation of the BlockID and the payout
// index.
func (b Block) MinerPayoutID(i uint64) CoinOutputID {
	id := b.ID()
	hash, err := crypto.HashAll(id, i)
	if err != nil {
		build.Severe("failed to crypto hash block ID and the miner payout index as a crypto hash", err)
	}
	return CoinOutputID(hash)
}

// MarshalSia implements the siabin.SiaMarshaler interface.
func (b Block) MarshalSia(w io.Writer) error {
	w.Write(b.ParentID[:])
	w.Write(siabin.EncUint64(uint64(b.Timestamp)))
	return siabin.NewEncoder(w).EncodeAll(b.POBSOutput, b.MinerPayouts, b.Transactions)
}

// UnmarshalSia implements the siabin.SiaUnmarshaler interface.
func (b *Block) UnmarshalSia(r io.Reader) error {
	io.ReadFull(r, b.ParentID[:])
	tsBytes := make([]byte, 8)
	io.ReadFull(r, tsBytes)
	b.Timestamp = Timestamp(siabin.DecUint64(tsBytes))
	return siabin.NewDecoder(r).DecodeAll(&b.POBSOutput, &b.MinerPayouts, &b.Transactions)
}

// MarshalRivine implements the rivbin.RivineMarshaler interface.
func (b Block) MarshalRivine(w io.Writer) error {
	w.Write(b.ParentID[:])
	rivbin.MarshalUint64(w, uint64(b.Timestamp))
	return rivbin.NewEncoder(w).EncodeAll(b.POBSOutput, b.MinerPayouts, b.Transactions)
}

// UnmarshalRivine implements the rivbin.RivineUnmarshaler interface.
func (b *Block) UnmarshalRivine(r io.Reader) error {
	io.ReadFull(r, b.ParentID[:])
	i, err := rivbin.UnmarshalUint64(r)
	if err != nil {
		return err
	}
	b.Timestamp = Timestamp(i)
	return rivbin.NewDecoder(r).DecodeAll(&b.POBSOutput, &b.MinerPayouts, &b.Transactions)
}

// UnmarshalBlockHeadersParentIDAndTS
// The MerkleRoot is not unmarshalled from the header because
func (b *Block) UnmarshalBlockHeadersParentIDAndTS(raw []byte) (BlockID, Timestamp) {
	var ParentID BlockID
	copy(ParentID[:], raw[:32])
	return ParentID, Timestamp(siabin.DecUint64(raw[32:40]))
}

// MarshalJSON marshales a block id as a hex string.
func (bid BlockID) MarshalJSON() ([]byte, error) {
	return json.Marshal(bid.String())
}

// String prints the block id in hex.
func (bid BlockID) String() string {
	return fmt.Sprintf("%x", bid[:])
}

// UnmarshalJSON decodes the json hex string of the block id.
func (bid *BlockID) UnmarshalJSON(b []byte) error {
	return (*crypto.Hash)(bid).UnmarshalJSON(b)
}
