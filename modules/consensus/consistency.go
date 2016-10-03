package consensus

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/encoding"
	"github.com/rivine/rivine/types"

	"github.com/NebulousLabs/bolt"
)

// manageErr handles an error detected by the consistency checks.
func manageErr(tx *bolt.Tx, err error) {
	markInconsistency(tx)
	if build.DEBUG {
		panic(err)
	} else {
		fmt.Println(err)
	}
}

// consensusChecksum grabs a checksum of the consensus set by pushing all of
// the elements in sorted order into a merkle tree and taking the root. All
// consensus sets with the same current block should have identical consensus
// checksums.
func consensusChecksum(tx *bolt.Tx) crypto.Hash {
	// Create a checksum tree.
	tree := crypto.NewTree()

	// For all of the constant buckets, push every key and every value. Buckets
	// are sorted in byte-order, therefore this operation is deterministic.
	consensusSetBuckets := []*bolt.Bucket{
		tx.Bucket(BlockPath),
		tx.Bucket(CoinOutputs),
		tx.Bucket(BlockStakeOutputs),
	}
	for i := range consensusSetBuckets {
		err := consensusSetBuckets[i].ForEach(func(k, v []byte) error {
			tree.Push(k)
			tree.Push(v)
			return nil
		})
		if err != nil {
			manageErr(tx, err)
		}
	}

	// Iterate through all the buckets looking for buckets prefixed with
	// prefixDSCO or prefixFCEX. Buckets are presented in byte-sorted order by
	// name.
	err := tx.ForEach(func(name []byte, b *bolt.Bucket) error {
		// If the bucket is not a delayed siacoin output bucket or a file
		// contract expiration bucket, skip.
		if !bytes.HasPrefix(name, prefixDSCO) && !bytes.HasPrefix(name, prefixFCEX) {
			return nil
		}

		// The bucket is a prefixed bucket - add all elements to the tree.
		return b.ForEach(func(k, v []byte) error {
			tree.Push(k)
			tree.Push(v)
			return nil
		})
	})
	if err != nil {
		manageErr(tx, err)
	}

	return tree.Root()
}

// checkCoinCount checks that the number of coins countable within the
// consensus set equal the expected number of coins for the block height.
func checkCoinCount(tx *bolt.Tx) {

	// Add all of the siacoin outputs.
	var scoSiacoins types.Currency
	err := tx.Bucket(CoinOutputs).ForEach(func(_, scoBytes []byte) error {
		var sco types.CoinOutput
		err := encoding.Unmarshal(scoBytes, &sco)
		if err != nil {
			manageErr(tx, err)
		}
		scoSiacoins = scoSiacoins.Add(sco.Value)
		return nil
	})
	if err != nil {
		manageErr(tx, err)
	}

	expectedSiacoins := types.NewCurrency64(0) //TODO rivine
	totalSiacoins := scoSiacoins
	if totalSiacoins.Cmp(expectedSiacoins) != 0 {
		diagnostics := fmt.Sprintf("Wrong number of coins: %v\n", scoSiacoins)
		if totalSiacoins.Cmp(expectedSiacoins) < 0 {
			diagnostics += fmt.Sprintf("total: %v\nexpected: %v\n expected is bigger: %v", totalSiacoins, expectedSiacoins, expectedSiacoins.Sub(totalSiacoins))
		} else {
			diagnostics += fmt.Sprintf("total: %v\nexpected: %v\n expected is bigger: %v", totalSiacoins, expectedSiacoins, totalSiacoins.Sub(expectedSiacoins))
		}
		manageErr(tx, errors.New(diagnostics))
	}
}

// checkBlockStakeCount checks that the number of siafunds countable within the
// consensus set equal the expected number of BlockStakeOutputs for the block height.
func checkBlockStakeCount(tx *bolt.Tx) {
	var total types.Currency
	err := tx.Bucket(BlockStakeOutputs).ForEach(func(_, siafundOutputBytes []byte) error {
		var sfo types.BlockStakeOutput
		err := encoding.Unmarshal(siafundOutputBytes, &sfo)
		if err != nil {
			manageErr(tx, err)
		}
		total = total.Add(sfo.Value)
		return nil
	})
	if err != nil {
		manageErr(tx, err)
	}
	if total.Cmp(types.BlockStakeCount) != 0 {
		manageErr(tx, errors.New("wrong number if blockstakes in the consensus set"))
	}
}

// checkRevertApply reverts the most recent block, checking to see that the
// consensus set hash matches the hash obtained for the previous block. Then it
// applies the block again and checks that the consensus set hash matches the
// original consensus set hash.
func (cs *ConsensusSet) checkRevertApply(tx *bolt.Tx) {
	current := currentProcessedBlock(tx)
	// Don't perform the check if this block is the genesis block.
	if current.Block.ID() == cs.blockRoot.Block.ID() {
		return
	}

	parent, err := getBlockMap(tx, current.Block.ParentID)
	if err != nil {
		manageErr(tx, err)
	}
	if current.Height != parent.Height+1 {
		manageErr(tx, errors.New("parent structure of a block is incorrect"))
	}
	_, _, err = cs.forkBlockchain(tx, parent)
	if err != nil {
		manageErr(tx, err)
	}
	if consensusChecksum(tx) != parent.ConsensusChecksum {
		manageErr(tx, errors.New("consensus checksum mismatch after reverting"))
	}
	_, _, err = cs.forkBlockchain(tx, current)
	if err != nil {
		manageErr(tx, err)
	}
	if consensusChecksum(tx) != current.ConsensusChecksum {
		manageErr(tx, errors.New("consensus checksum mismatch after re-applying"))
	}
}

// checkConsistency runs a series of checks to make sure that the consensus set
// is consistent with some rules that should always be true.
func (cs *ConsensusSet) checkConsistency(tx *bolt.Tx) {
	if cs.checkingConsistency {
		return
	}
	cs.checkingConsistency = true
	checkCoinCount(tx)
	checkBlockStakeCount(tx)
	if build.DEBUG {
		cs.checkRevertApply(tx)
	}
	cs.checkingConsistency = false
}

// maybeCheckConsistency runs a consistency check with a small probability.
// Useful for detecting database corruption in production without needing to go
// through the extremely slow process of running a consistency check every
// block.
func (cs *ConsensusSet) maybeCheckConsistency(tx *bolt.Tx) {
	n, err := crypto.RandIntn(1000)
	if err != nil {
		manageErr(tx, err)
	}
	if n == 0 {
		cs.checkConsistency(tx)
	}
}

// TODO: Check that every file contract has an expiration too, and that the
// number of file contracts + the number of expirations is equal.
