package types

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/encoding"
)

// TestInputSigHash tests the input signature hash algorithm.
// It is the algorithm used to compute a signature for an input within a v1 transaction.
//
// Pseudo code of signature algorithm, as documented:
//
//    blake2b_256_hash(BinaryEncoding(
//      - transactionVersion: byte,
//      - inputIndex: int64 (8 bytes, little endian),
//      extraObjects:
//        if atomicSwap:
//          SiaPublicKey:
//            - Algorithm: 16 bytes fixed-size array
//            - Key: 8 bytes length + n bytes
//        if atomicSwap as claimer (owner of receiver pub key):
//          - Secret: 32 bytes fixed-size array
//      - length(coinInputs): int64 (8 bytes, little endian)
//      for each coinInput:
//        - parentID: 32 bytes fixed-size array
//      - length(coinOutputs): int64 (8 bytes, little endian)
//      for each coinOutput:
//        - value: Currency (8 bytes length + n bytes, little endian encoded)
//        - binaryEncoding(condition)
//      - length(blockStakeInputs): int64 (8 bytes, little endian)
//      for each blockStakeInput:
//        - parentID: 32 bytes fixed-size array
//      - length(blockStakeOutputs): int64 (8 bytes, little endian)
//      for each blockStakeOutput:
//        - value: Currency (8 bytes length + n bytes, little endian encoded)
//        - binaryEncoding(condition)
//      - length(minerFees): int64 (8 bytes, little endian)
//      for each minerFee:
//        - fee: Currency (8 bytes length + n bytes, little endian encoded)
//      - arbitraryData: 8 bytes length + n bytes
//    )) : 32 bytes fixed-size crypto hash
//
// BinaryEncoding encodes all bytes as described above,
// concatenating all output bytes together, in order as given.
// The total output byte slice (result of encoding and concatenation)
// is used as the input to make the crypto hashsum using the blake2b 256bit algorithm,
// resulting in a 32 bytes fixed-size crypto hash.
//
// The binary encoding of an output condition depends upon the unlock condition type:
//
//   ConditionTypeNil: 0x000000000000000000 (type 0 as 1 byte + length 0 as 8 bytes)
//   ConditionTypeUnlockHash:
//     - 0x01: 1, 1 byte (type)
//     - 0x2100000000000000: 33, 8 bytes (length of unlockHash)
//     - unlockHash: 33 bytes fixed-size array
//   ConditionTypeAtomicSwap:
//     - 0x02: 2, 1 byte (type)
//     - 0x6a00000000000000: 106, 8 bytes (length of following condition properties)
//     - sender (unlockHash): 33 bytes fixed-size array
//     - receiver (unlockHash): 33 bytes fixed-size array
//     - hashedSecret: 32 bytes fixed-size array
//     - timeLock: uint64 (8 bytes, little endian)
//   ConditionTypeTimeLock:
//     - 0x03: 3, 1 byte (type)
//     - length(binaryEncoding(condition.type, condition.data)): int64 (8 bytes, little endian)
//     - lockTime: uint64 (8 bytes, little endian)
//     - binaryEncoding(condition.type, condition.data)
//   ConditionTypeMultiSignature:
//     - 0x04: 4, 1 byte (type)
//     - binaryEncoding(unlockHashSlice)
//
// A TimeLock wraps around another condition.
// For now the only valid condition types that can be used as the internal
// condition of a TimeLock are a `ConditionTypeUnlockHash` and `ConditionTypeMultiSignature`.
// A future version however might also allow other internal condition,
// if so, this document will be updated as well to clarify that.
//
// See /doc/Encoding.md for more information
// about the Binary Encoding Algorithm and how each (primitive) type is encoded.
//
// ENSURE TO KEEP THIS TEST UP TO DATE WITH BOTH THE IMPLEMENTATION
// AS WELL AS THE EXTERNAL (/doc/**/*) DOCUMENTATION !!
func TestInputSigHash(t *testing.T) {
	// utility funcs
	hbs := func(str string) []byte { // hexStr -> byte slice
		bs, err := hex.DecodeString(str)
		if err != nil {
			t.Error(err)
		}
		return bs
	}
	hs := func(str string) (hash crypto.Hash) { // hbs -> crypto.Hash
		copy(hash[:], hbs(str))
		return
	}
	sliceToHex := func(b []byte) string {
		return hex.EncodeToString(b[:])
	}

	// deterministic private/public key pair
	var entropy [crypto.EntropySize]byte
	b, err := hex.DecodeString("5722127ae85b120d15d5998d7591a603f4cbe2ac47a69b4d16bd7c90e21aeedd")
	if err != nil {
		t.Fatal(err)
	}
	copy(entropy[:], b[:])
	sk, rpk := crypto.GenerateKeyPairDeterministic(entropy)
	pk := Ed25519PublicKey(rpk)

	// deterministic atomic swap secret
	var secret AtomicSwapSecret
	b, err = hex.DecodeString("8a1ed1737c0a84598bbf429ba351e03b5f4e00d7ab5d8635da832957f14bb272")
	if err != nil {
		t.Fatal(err)
	}
	copy(secret[:], b[:])
	hashedSecret := NewAtomicSwapHashedSecret(secret)

	// signature computation algo, isolated from all other logic
	sig := func(input string) crypto.Signature { // hexStr -> signature
		b := hbs(input)
		hash := crypto.HashBytes(b)
		return crypto.SignHash(hash, sk)
	}

	// transaction containing all possible fields,
	// as well as all valid unlock hashes and input locks.
	txn := Transaction{
		Version: 1,
		CoinInputs: []CoinInput{
			{
				ParentID: CoinOutputID(hs("a3331dc5ad8fef504e148e1608db0f05156dc72283dc2b1b083afa8bfdcdb6d9")),
				Fulfillment: UnlockFulfillmentProxy{
					Fulfillment: NewSingleSignatureFulfillment(pk),
				},
			},
			{
				ParentID: CoinOutputID(hs("fd166ead847ae80a5754700980130ee7bd23a94bd2a8fa14807955142719eb05")),
				Fulfillment: UnlockFulfillmentProxy{
					Fulfillment: NewAtomicSwapClaimFulfillment(pk, secret),
				},
			},
		},
		CoinOutputs: []CoinOutput{
			{
				Value:     NewCurrency64(42000000000),
				Condition: UnlockConditionProxy{}, // nil condition
			},
			{
				Value: NewCurrency64(684465979878000000),
				Condition: UnlockConditionProxy{
					Condition: NewUnlockHashCondition(unlockHashFromHex("0118bc060e820e8ae35bafae2a3b85e600caa9cb2e50cfd46a2d98253d7d7780c940ea682e7b40")),
				},
			},
			{
				Value: NewCurrency64(50000000000000),
				Condition: UnlockConditionProxy{
					Condition: &AtomicSwapCondition{
						Sender:       unlockHashFromHex("01fea3ae2854f6e497c92a1cdd603a0bc92ada717200e74f64731e86a923479883519804b18d9d"),
						Receiver:     unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893"),
						HashedSecret: hashedSecret,
						TimeLock:     2524608000,
					},
				},
			},
			{
				Value: NewCurrency64(42000000000),
				Condition: UnlockConditionProxy{
					Condition: &TimeLockCondition{
						LockTime:  2524608000,
						Condition: NewUnlockHashCondition(unlockHashFromHex("0118bc060e820e8ae35bafae2a3b85e600caa9cb2e50cfd46a2d98253d7d7780c940ea682e7b40")),
					},
				},
			},
		},
		BlockStakeInputs: []BlockStakeInput{
			{
				ParentID: BlockStakeOutputID(hs("dca04a18b64ba012218d28d265a852625416567932f5eda3bcb0a0bc66da8b09")),
				Fulfillment: UnlockFulfillmentProxy{
					Fulfillment: NewSingleSignatureFulfillment(pk),
				},
			},
		},
		BlockStakeOutputs: []BlockStakeOutput{
			{
				Value: NewCurrency64(390),
				Condition: UnlockConditionProxy{
					Condition: NewUnlockHashCondition(unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893")),
				},
			},
		},
		MinerFees: []Currency{
			NewCurrency64(100000000),
		},
		ArbitraryData: []byte("All Creatures Great and Small Will Get Stuck at Brexit’s Border, 08/05/2018 06:00 CEST"),
	}

	var (
		signature         ByteSlice
		expectedSignature crypto.Signature
	)

	// BlockStake Fulfillment, input #0: SingleSignature
	err = txn.BlockStakeInputs[0].Fulfillment.Sign(FulfillmentSignContext{
		InputIndex:  0,
		Transaction: txn,
		Key:         sk,
	})
	if err != nil {
		t.Error(err)
	}
	// decode expected signature
	expectedSignature = sig(
		`01` + // transaction version
			`0000000000000000` + // input index
			`0200000000000000` + // length(coinInputs), 2
			`a3331dc5ad8fef504e148e1608db0f05156dc72283dc2b1b083afa8bfdcdb6d9` + // CI#1 - parentid only
			`fd166ead847ae80a5754700980130ee7bd23a94bd2a8fa14807955142719eb05` + // CI#2 - parentid only
			`0400000000000000` + // length(coinOutputs), 4
			`050000000000000009c7652400` + // CO#1
			`00` + `0000000000000000` +
			`0800000000000000097fb62ea777c580` + // CO#2
			`01` + `2100000000000000` + `0118bc060e820e8ae35bafae2a3b85e600caa9cb2e50cfd46a2d98253d7d7780c9` +
			`06000000000000002d79883d2000` + // CO#3
			`02` + `6a00000000000000` + `01fea3ae2854f6e497c92a1cdd603a0bc92ada717200e74f64731e86a923479883` + `01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a` + sliceToHex(hashedSecret[:]) + `00767a9600000000` +
			`050000000000000009c7652400` + // CO#4
			`03` + `2a00000000000000` + `00767a9600000000` + `01` + `0118bc060e820e8ae35bafae2a3b85e600caa9cb2e50cfd46a2d98253d7d7780c9` +
			`0100000000000000` + // length(blockStakeInputs), 1
			`dca04a18b64ba012218d28d265a852625416567932f5eda3bcb0a0bc66da8b09` + // BSI#1
			`0100000000000000` + // length(blockStakeOutputs), 1
			`02000000000000000186` + // BSO#1
			`01` + `2100000000000000` +
			`01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a` +
			`0100000000000000` + // length(minerFees), 1
			`040000000000000005f5e100` + // MF#1
			`5800000000000000` + // length(arbitraryData)
			`416c6c2043726561747572657320477265617420616e6420536d616c6c2057696c6c2047657420537475636b20617420427265786974e280997320426f726465722c2030382f30352f323031382030363a30302043455354`, // arbitraryData
	)
	signature = txn.BlockStakeInputs[0].Fulfillment.Fulfillment.(*SingleSignatureFulfillment).Signature
	if bytes.Compare(expectedSignature[:], signature[:]) != 0 {
		t.Error("blockStake input #0: ",
			hex.EncodeToString(expectedSignature[:]), "!=",
			hex.EncodeToString(signature[:]))
	}

	// Coin Fulfillment, input #0: SingleSignature
	err = txn.CoinInputs[0].Fulfillment.Sign(FulfillmentSignContext{
		InputIndex:  0,
		Transaction: txn,
		Key:         sk,
	})

	if err != nil {
		t.Error(err)
	}
	signature = txn.CoinInputs[0].Fulfillment.Fulfillment.(*SingleSignatureFulfillment).Signature
	// signature is same as previous signature, so we can compare directly
	if bytes.Compare(expectedSignature[:], signature[:]) != 0 {
		t.Error("coin input #0: ",
			hex.EncodeToString(expectedSignature[:]), "!=",
			hex.EncodeToString(signature[:]))
	}

	// Coin Fulfillment, input #1: AtomicSwapFulfillment
	err = txn.CoinInputs[1].Fulfillment.Sign(FulfillmentSignContext{
		InputIndex:  1,
		Transaction: txn,
		Key:         sk,
	})
	if err != nil {
		t.Error(err)
	}
	signature = txn.CoinInputs[1].Fulfillment.Fulfillment.(*AtomicSwapFulfillment).Signature
	// decode expected signature
	expectedSignature = sig(
		`01` + // transaction version
			`0100000000000000` + // input index
			sliceToHex(encoding.Marshal(pk)) + // AS: PublicKey
			sliceToHex(secret[:]) + // AS: Claim: secret
			`0200000000000000` + // length(coinInputs), 2
			`a3331dc5ad8fef504e148e1608db0f05156dc72283dc2b1b083afa8bfdcdb6d9` + // CI#1 - parentid only
			`fd166ead847ae80a5754700980130ee7bd23a94bd2a8fa14807955142719eb05` + // CI#2 - parentid only
			`0400000000000000` + // length(coinOutputs), 4
			`050000000000000009c7652400` + // CO#1
			`00` + `0000000000000000` +
			`0800000000000000097fb62ea777c580` + // CO#2
			`01` + `2100000000000000` + `0118bc060e820e8ae35bafae2a3b85e600caa9cb2e50cfd46a2d98253d7d7780c9` +
			`06000000000000002d79883d2000` + // CO#3
			`02` + `6a00000000000000` + `01fea3ae2854f6e497c92a1cdd603a0bc92ada717200e74f64731e86a923479883` + `01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a` + sliceToHex(hashedSecret[:]) + `00767a9600000000` +
			`050000000000000009c7652400` + // CO#4
			`03` + `2a00000000000000` + `00767a9600000000` + `01` + `0118bc060e820e8ae35bafae2a3b85e600caa9cb2e50cfd46a2d98253d7d7780c9` +
			`0100000000000000` + // length(blockStakeInputs), 1
			`dca04a18b64ba012218d28d265a852625416567932f5eda3bcb0a0bc66da8b09` + // BSI#1
			`0100000000000000` + // length(blockStakeOutputs), 1
			`02000000000000000186` + // BSO#1
			`01` + `2100000000000000` +
			`01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a` +
			`0100000000000000` + // length(minerFees), 1
			`040000000000000005f5e100` + // MF#1
			`5800000000000000` + // length(arbitraryData)
			`416c6c2043726561747572657320477265617420616e6420536d616c6c2057696c6c2047657420537475636b20617420427265786974e280997320426f726465722c2030382f30352f323031382030363a30302043455354`, // arbitraryData
	)
	if bytes.Compare(expectedSignature[:], signature[:]) != 0 {
		t.Error("coin input #1: ",
			hex.EncodeToString(expectedSignature[:]), "!=",
			hex.EncodeToString(signature[:]))
	}
}

// TestLegacyInputSigHash tests the legacy input signature hash algorithm.
// It is the algorithm used to compute a signature for an input within a v0 transaction.
//
// Pseudo code of signature algorithm, as documented:
//
//    blake2b_256_hash(BinaryEncoding(
//      - inputIndex: int64 (8 bytes, little endian),
//      extraObjects:
//        if atomicSwap:
//          SiaPublicKey:
//            - Algorithm: 16 bytes fixed-size array
//            - Key: 8 bytes length + n bytes
//        if atomicSwap as claimer (owner of receiver pub key):
//          - Secret: 32 bytes fixed-size array
//      for each coinInput:
//        - parentID: 32 bytes fixed-size array
//        - unlockHash: 33 bytes fixed-size array
//      - length(coinOutputs): int64 (8 bytes, little endian)
//      for each coinOutput:
//        - value: Currency (8 bytes length + n bytes, little endian encoded)
//        - unlockHash: 33 bytes fixed-size array
//      for each blockStakeInput:
//        - parentID: 32 bytes fixed-size array
//        - unlockHash: 33 bytes fixed-size array
//      - length(blockStakeOutputs): int64 (8 bytes, little endian)
//      for each blockStakeOutput:
//        - value: Currency (8 bytes length + n bytes, little endian encoded)
//        - unlockHash: 33 bytes fixed-size array
//      - length(minerFees): int64 (8 bytes, little endian)
//      for each minerFee:
//        - fee: Currency (8 bytes length + n bytes, little endian encoded)
//      - arbitraryData: 8 bytes length + n bytes
//    )) : 32 bytes fixed-size crypto hash
//
// BinaryEncoding encodes all bytes as described above,
// concatenating all output bytes together, in order as given.
// The total output byte slice (result of encoding and concatenation)
// is used as the input to make the crypto hashsum using the blake2b 256bit algorithm,
// resulting in a 32 bytes fixed-size crypto hash.
//
// See /doc/Encoding.md for more information
// about the Binary Encoding Algorithm and how each (primitive) type is encoded.
//
// ENSURE TO KEEP THIS TEST UP TO DATE WITH BOTH THE IMPLEMENTATION
// AS WELL AS THE EXTERNAL (/doc/**/*) DOCUMENTATION !!
func TestLegacyInputSigHash(t *testing.T) {
	// utility funcs
	hbs := func(str string) []byte { // hexStr -> byte slice
		bs, err := hex.DecodeString(str)
		if err != nil {
			t.Error(err)
		}
		return bs
	}
	hs := func(str string) (hash crypto.Hash) { // hbs -> crypto.Hash
		copy(hash[:], hbs(str))
		return
	}
	sliceToHex := func(b []byte) string {
		return hex.EncodeToString(b[:])
	}
	hashToHex := func(hash crypto.Hash) string {
		return sliceToHex(hash[:])
	}

	// deterministic private/public key pair
	var entropy [crypto.EntropySize]byte
	b, err := hex.DecodeString("5722127ae85b120d15d5998d7591a603f4cbe2ac47a69b4d16bd7c90e21aeedd")
	if err != nil {
		t.Fatal(err)
	}
	copy(entropy[:], b[:])
	sk, rpk := crypto.GenerateKeyPairDeterministic(entropy)
	pk := Ed25519PublicKey(rpk)

	// deterministic atomic swap secret
	var secret AtomicSwapSecret
	b, err = hex.DecodeString("8a1ed1737c0a84598bbf429ba351e03b5f4e00d7ab5d8635da832957f14bb272")
	if err != nil {
		t.Fatal(err)
	}
	copy(secret[:], b[:])
	hashedSecret := NewAtomicSwapHashedSecret(secret)

	// signature computation algo, isolated from all other logic
	sig := func(input string) crypto.Signature { // hexStr -> signature
		b := hbs(input)
		hash := crypto.HashBytes(b)
		return crypto.SignHash(hash, sk)
	}

	// transaction containing all possible fields,
	// as well as all valid unlock hashes and input locks.
	txn := Transaction{
		Version: 0,
		CoinInputs: []CoinInput{
			{
				ParentID: CoinOutputID(hs("a3331dc5ad8fef504e148e1608db0f05156dc72283dc2b1b083afa8bfdcdb6d9")),
				Fulfillment: UnlockFulfillmentProxy{
					Fulfillment: NewSingleSignatureFulfillment(pk),
				},
			},
			{
				ParentID: CoinOutputID(hs("fd166ead847ae80a5754700980130ee7bd23a94bd2a8fa14807955142719eb05")),
				Fulfillment: UnlockFulfillmentProxy{
					Fulfillment: &LegacyAtomicSwapFulfillment{
						Sender:       unlockHashFromHex("01fea3ae2854f6e497c92a1cdd603a0bc92ada717200e74f64731e86a923479883519804b18d9d"),
						Receiver:     NewPubKeyUnlockHash(pk),
						HashedSecret: hashedSecret,
						TimeLock:     2524608000,
						PublicKey:    pk,
						Secret:       secret,
					},
				},
			},
		},
		CoinOutputs: []CoinOutput{
			{
				Value: NewCurrency64(684465979878000000),
				Condition: UnlockConditionProxy{
					Condition: NewUnlockHashCondition(unlockHashFromHex("0118bc060e820e8ae35bafae2a3b85e600caa9cb2e50cfd46a2d98253d7d7780c940ea682e7b40")),
				},
			},
			{
				Value: NewCurrency64(50000000000000),
				Condition: UnlockConditionProxy{
					Condition: NewUnlockHashCondition(unlockHashFromHex("0296551728f6e1a244184dcad09b4e76debf16bd4721acd87b073404eb74d151aefff6ce6b7a86")),
				},
			},
		},
		BlockStakeInputs: []BlockStakeInput{
			{
				ParentID: BlockStakeOutputID(hs("dca04a18b64ba012218d28d265a852625416567932f5eda3bcb0a0bc66da8b09")),
				Fulfillment: UnlockFulfillmentProxy{
					Fulfillment: NewSingleSignatureFulfillment(pk),
				},
			},
		},
		BlockStakeOutputs: []BlockStakeOutput{
			{
				Value: NewCurrency64(390),
				Condition: UnlockConditionProxy{
					Condition: NewUnlockHashCondition(unlockHashFromHex("01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9aa9e7c6f58893")),
				},
			},
		},
		MinerFees: []Currency{
			NewCurrency64(100000000),
		},
		ArbitraryData: []byte("Nestle Pays $7.2 Billion to Sell Coffee With Starbucks Brand, 07/05/2018 07:13 CEST"),
	}

	var (
		signature         ByteSlice
		expectedSignature crypto.Signature
	)

	// BlockStake Fulfillment, input #0: SingleSignature
	err = txn.BlockStakeInputs[0].Fulfillment.Sign(FulfillmentSignContext{
		InputIndex:  0,
		Transaction: txn,
		Key:         sk,
	})
	if err != nil {
		t.Error(err)
	}
	// decode expected signature
	expectedSignature = sig(
		`0000000000000000` + // input index
			`a3331dc5ad8fef504e148e1608db0f05156dc72283dc2b1b083afa8bfdcdb6d9` + // CI#1
			`01` + hashToHex(crypto.HashObject(encoding.Marshal(pk))) +
			`fd166ead847ae80a5754700980130ee7bd23a94bd2a8fa14807955142719eb05` + // CI#2
			`02` + hashToHex(crypto.HashObject(encoding.MarshalAll(
			unlockHashFromHex("01fea3ae2854f6e497c92a1cdd603a0bc92ada717200e74f64731e86a923479883519804b18d9d"),
			NewPubKeyUnlockHash(pk),
			hashedSecret,
			Timestamp(2524608000),
		))) +
			`0200000000000000` + // length(coinOutputs), 2
			`0800000000000000097fb62ea777c580` + // CO#1
			`0118bc060e820e8ae35bafae2a3b85e600caa9cb2e50cfd46a2d98253d7d7780c9` +
			`06000000000000002d79883d2000` + // CO#2
			`0296551728f6e1a244184dcad09b4e76debf16bd4721acd87b073404eb74d151ae` +
			`dca04a18b64ba012218d28d265a852625416567932f5eda3bcb0a0bc66da8b09` + // BSI#1
			`01` + hashToHex(crypto.HashObject(encoding.Marshal(pk))) +
			`0100000000000000` + // length(blockStakeOutputs), 1
			`02000000000000000186` + // BSO#1
			`01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a` +
			`0100000000000000` + // length(minerFees), 1
			`040000000000000005f5e100` + // MF#1
			`5300000000000000` + // length(arbitraryData)
			`4e6573746c6520506179732024372e322042696c6c696f6e20746f2053656c6c20436f66666565205769746820537461726275636b73204272616e642c2030372f30352f323031382030373a31332043455354`, // arbitraryData
	)
	signature = txn.BlockStakeInputs[0].Fulfillment.Fulfillment.(*SingleSignatureFulfillment).Signature
	if bytes.Compare(expectedSignature[:], signature[:]) != 0 {
		t.Error("blockStake input #0: ",
			hex.EncodeToString(expectedSignature[:]), "!=",
			hex.EncodeToString(signature[:]))
	}

	// Coin Fulfillment, input #0: SingleSignature
	err = txn.CoinInputs[0].Fulfillment.Sign(FulfillmentSignContext{
		InputIndex:  0,
		Transaction: txn,
		Key:         sk,
	})

	if err != nil {
		t.Error(err)
	}
	signature = txn.CoinInputs[0].Fulfillment.Fulfillment.(*SingleSignatureFulfillment).Signature
	// signature is same as previous signature, so we can compare directly
	if bytes.Compare(expectedSignature[:], signature[:]) != 0 {
		t.Error("coin input #0: ",
			hex.EncodeToString(expectedSignature[:]), "!=",
			hex.EncodeToString(signature[:]))
	}

	// Coin Fulfillment, input #1: SingleSignature
	err = txn.CoinInputs[1].Fulfillment.Sign(FulfillmentSignContext{
		InputIndex:  1,
		Transaction: txn,
		Key:         sk,
	})
	if err != nil {
		t.Error(err)
	}
	signature = txn.CoinInputs[1].Fulfillment.Fulfillment.(*LegacyAtomicSwapFulfillment).Signature
	// decode expected signature
	expectedSignature = sig(
		`0100000000000000` + // input index
			sliceToHex(encoding.Marshal(pk)) + // AS: PublicKey
			sliceToHex(secret[:]) + // AS: Claim: secret
			`a3331dc5ad8fef504e148e1608db0f05156dc72283dc2b1b083afa8bfdcdb6d9` + // CI#1
			`01` + hashToHex(crypto.HashObject(encoding.Marshal(pk))) +
			`fd166ead847ae80a5754700980130ee7bd23a94bd2a8fa14807955142719eb05` + // CI#2
			`02` + hashToHex(crypto.HashObject(encoding.MarshalAll(
			unlockHashFromHex("01fea3ae2854f6e497c92a1cdd603a0bc92ada717200e74f64731e86a923479883519804b18d9d"),
			NewPubKeyUnlockHash(pk),
			hashedSecret,
			Timestamp(2524608000),
		))) +
			`0200000000000000` + // length(coinOutputs), 2
			`0800000000000000097fb62ea777c580` + // CO#1
			`0118bc060e820e8ae35bafae2a3b85e600caa9cb2e50cfd46a2d98253d7d7780c9` +
			`06000000000000002d79883d2000` + // CO#2
			`0296551728f6e1a244184dcad09b4e76debf16bd4721acd87b073404eb74d151ae` +
			`dca04a18b64ba012218d28d265a852625416567932f5eda3bcb0a0bc66da8b09` + // BSI#1
			`01` + hashToHex(crypto.HashObject(encoding.Marshal(pk))) +
			`0100000000000000` + // length(blockStakeOutputs), 1
			`02000000000000000186` + // BSO#1
			`01746677df456546d93729066dd88514e2009930f3eebac3c93d43c88a108f8f9a` +
			`0100000000000000` + // length(minerFees), 1
			`040000000000000005f5e100` + // MF#1
			`5300000000000000` + // length(arbitraryData)
			`4e6573746c6520506179732024372e322042696c6c696f6e20746f2053656c6c20436f66666565205769746820537461726275636b73204272616e642c2030372f30352f323031382030373a31332043455354`, // arbitraryData
	)
	if bytes.Compare(expectedSignature[:], signature[:]) != 0 {
		t.Error("coin input #1: ",
			hex.EncodeToString(expectedSignature[:]), "!=",
			hex.EncodeToString(signature[:]))
	}
}
