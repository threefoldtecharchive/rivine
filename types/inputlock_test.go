package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/rivine/rivine/crypto"
	"github.com/rivine/rivine/encoding"
)

func TestSingleSignatureUnlocker(t *testing.T) {
	sk, pk := crypto.GenerateKeyPair()
	ul := NewSingleSignatureInputLock(Ed25519PublicKey(pk))

	err := ul.StrictCheck()
	if err == nil {
		t.Error("error was expected, nil received")
	}

	uh1 := ul.UnlockHash()
	uh2 := ul.UnlockHash()
	if uh1.String() != uh2.String() {
		t.Error("inconsistent unlock hashes:", uh1, uh2)
	}

	err = ul.Unlock(0, Transaction{})
	if err == nil {
		t.Error("error was expected, nil received")
	}

	err = ul.Lock(0, Transaction{}, sk)
	if err != nil {
		t.Errorf("failed to lock transaction: %v", err)
	}

	err = ul.StrictCheck()
	if err != nil {
		t.Errorf("strict check failed while it was expected to succeed: %v", err)
	}
	err = ul.Unlock(0, Transaction{})
	if err != nil {
		t.Errorf("unlock failed while it was expected to succeed: %v", err)
	}

	uh3 := ul.UnlockHash()
	if uh1.String() != uh3.String() {
		t.Error("inconsistent unlock hashes:", uh1, uh3)
	}
}

func TestSingleSignatureUnlockerBadTransaction(t *testing.T) {
	sk, pk := crypto.GenerateKeyPair()
	ul := NewSingleSignatureInputLock(Ed25519PublicKey(pk))

	tx := Transaction{}

	err := ul.Lock(0, tx, sk)
	if err != nil {
		t.Errorf("failed to lock transaction: %v", err)
	}
	err = ul.Unlock(0, tx)
	if err != nil {
		t.Errorf("unlock failed while it was expected to succeed: %v", err)
	}

	tx.CoinInputs = append(tx.CoinInputs, CoinInput{
		Unlocker: NewSingleSignatureInputLock(Ed25519PublicKey(pk)),
	})
	ul.il.(*SingleSignatureInputLock).Signature = nil
	err = ul.Lock(0, tx, sk)
	if err != nil {
		t.Errorf("failed to lock transaction: %v", err)
	}
	err = ul.Unlock(0, tx)
	if err != nil {
		t.Errorf("unlock failed while it was expected to succeed: %v", err)
	}
	err = ul.Unlock(0, Transaction{})
	if err == nil {
		t.Errorf("unlock should fail as transaction is wrong")
	}

	tx.CoinInputs = append(tx.CoinInputs, tx.CoinInputs[0])
	ul.il.(*SingleSignatureInputLock).Signature = nil
	err = ul.Lock(0, tx, sk)
	if err != nil {
		t.Errorf("failed to lock transaction: %v", err)
	}
	err = ul.Unlock(0, tx)
	if err != nil {
		t.Errorf("unlock failed while it was expected to succeed: %v", err)
	}
}

func TestAtomicSwapUnlocker(t *testing.T) {
	sk1, pk1 := crypto.GenerateKeyPair()
	sk2, pk2 := crypto.GenerateKeyPair()

	secret := sha256.Sum256([]byte{1, 2, 3, 4})
	hashedSecret := sha256.Sum256(secret[:])

	ul := NewAtomicSwapInputLock(AtomicSwapCondition{
		Sender:       NewSingleSignatureInputLock(Ed25519PublicKey(pk1)).UnlockHash(),
		Receiver:     NewSingleSignatureInputLock(Ed25519PublicKey(pk2)).UnlockHash(),
		HashedSecret: AtomicSwapHashedSecret(hashedSecret),
		TimeLock:     CurrentTimestamp() + 50000,
	})

	err := ul.StrictCheck()
	if err == nil {
		t.Error("error was expected, nil received")
	}

	uh1 := ul.UnlockHash()
	uh2 := ul.UnlockHash()
	if uh1.String() != uh2.String() {
		t.Error("inconsistent unlock hashes:", uh1, uh2)
	}

	err = ul.Unlock(0, Transaction{})
	if err == nil {
		t.Error("error was expected, nil received")
	}

	err = ul.Lock(0, Transaction{}, sk2)
	if err == nil {
		t.Errorf("should fail, as reclaim cannot be done, and especially not by sk2")
	}
	ul.il.(*AtomicSwapInputLock).Signature = nil
	err = ul.Lock(0, Transaction{}, sk1)
	if err == nil {
		t.Errorf("should fail, as reclaim cannot be done, and especially not by sk1")
	}
	ul.il.(*AtomicSwapInputLock).Signature = nil
	err = ul.Lock(0, Transaction{}, AtomicSwapClaimKey{
		PublicKey: Ed25519PublicKey(pk1),
		SecretKey: sk1[:],
		Secret:    AtomicSwapSecret{},
	})
	if err == nil {
		t.Errorf("should fail, as reclaim cannot be done, and especially not by sk1")
	}
	ul.il.(*AtomicSwapInputLock).Signature = nil
	err = ul.Lock(0, Transaction{}, AtomicSwapClaimKey{
		PublicKey: Ed25519PublicKey(pk2),
		SecretKey: sk2[:],
		Secret:    AtomicSwapSecret{},
	})
	if err == nil {
		t.Errorf("should fail, as secret is missing/wrong")
	}
	err = ul.Lock(0, Transaction{}, AtomicSwapClaimKey{
		PublicKey: Ed25519PublicKey(pk2),
		SecretKey: sk2[:],
		Secret:    AtomicSwapSecret{},
	})
	ul.il.(*AtomicSwapInputLock).Signature = nil
	if err == nil {
		t.Errorf("should fail, as secret is missing/wrong")
	}

	err = ul.Lock(0, Transaction{}, AtomicSwapClaimKey{
		PublicKey: Ed25519PublicKey(pk2),
		SecretKey: sk2[:],
		Secret:    secret,
	})
	if err != nil {
		t.Errorf("should succeed, as secret is correct: %v", err)
	}

	err = ul.StrictCheck()
	if err != nil {
		t.Errorf("strict check failed while it was expected to succeed: %v", err)
	}
	err = ul.Unlock(0, Transaction{})
	if err != nil {
		t.Errorf("unlock failed while it was expected to succeed: %v", err)
	}

	uh3 := ul.UnlockHash()
	if uh1.String() != uh3.String() {
		t.Error("inconsistent unlock hashes:", uh1, uh3)
	}
}

func TestInputLockTypeMarshaling(t *testing.T) {
	var ut UnlockType
	buffer := bytes.NewBuffer(nil)
	err := ut.MarshalSia(buffer)
	if err != nil {
		t.Errorf("error wile marshalling: %v", err)
	}
	if bytes.Compare([]byte{0}, buffer.Bytes()) != 0 {
		t.Errorf("invalid buffer: %v", buffer.Bytes())
	}
	err = ut.UnmarshalSia(buffer)
	if err != nil {
		t.Errorf("error wile unmarshalling: %v", err)
	}

	ut = UnlockTypeAtomicSwap
	err = ut.MarshalSia(buffer)
	if err != nil {
		t.Errorf("error wile marshalling: %v", err)
	}
	if bytes.Compare([]byte{byte(UnlockTypeAtomicSwap)}, buffer.Bytes()) != 0 {
		t.Errorf("invalid buffer: %v", buffer.Bytes())
	}
	err = ut.UnmarshalSia(buffer)
	if err != nil {
		t.Errorf("error wile unmarshalling: %v", err)
	}

	err = ut.UnmarshalSia(buffer)
	if err != io.EOF {
		t.Errorf("expected EOF but received: %v", err)
	}
}

func TestSingleSignatureInputLockEncodeDecode(t *testing.T) {
	sk, pk := crypto.GenerateKeyPair()
	ul := NewSingleSignatureInputLock(Ed25519PublicKey(pk))
	err := ul.Lock(0, Transaction{}, sk)
	if err != nil {
		t.Errorf("couldn't lock sslock: %v", err)
	}
	testInputLockEncodeDecode(t, ul.il)
}

func TestAtomicSwapInputLockEncodeDecode(t *testing.T) {
	_, pk := crypto.GenerateKeyPair()
	sk2, pk2 := crypto.GenerateKeyPair()

	var secret, hashedSecret [sha256.Size]byte
	hashedSecret = sha256.Sum256(secret[:])

	ul := NewAtomicSwapInputLock(AtomicSwapCondition{
		Sender:       NewSingleSignatureInputLock(Ed25519PublicKey(pk)).UnlockHash(),
		Receiver:     NewSingleSignatureInputLock(Ed25519PublicKey(pk2)).UnlockHash(),
		HashedSecret: AtomicSwapHashedSecret(hashedSecret),
		TimeLock:     CurrentTimestamp() + 500000,
	})
	err := ul.Lock(0, Transaction{}, AtomicSwapClaimKey{
		PublicKey: Ed25519PublicKey(pk2),
		Secret:    secret,
		SecretKey: sk2[:],
	})
	if err != nil {
		t.Errorf("error while locking InputLock: %v", err)
	}
	testInputLockEncodeDecode(t, ul.il)
}

func testInputLockEncodeDecode(t *testing.T, il InputLock) {
	cond := il.EncodeCondition()
	err := il.Decode(RawInputLockFormat{
		Condition:   cond,
		Fulfillment: il.EncodeFulfillment(),
	})
	if err != nil {
		t.Errorf("couldn't encode->decode: %v", err)
	}
	cond2 := il.EncodeCondition()
	if bytes.Compare(cond, cond2) != 0 {
		t.Errorf("pre-cond (%v) != (%v) post-cond", cond, cond2)
	}
}

func TestSingleSignatureUnknownEncoding(t *testing.T) {
	sk, pk := crypto.GenerateKeyPair()
	ul := NewSingleSignatureInputLock(Ed25519PublicKey(pk))
	err := ul.Lock(0, Transaction{}, sk)
	if err != nil {
		t.Errorf("couldn't lock sslock: %v", err)
	}

	// serialize using specific struct made for it
	buf := bytes.NewBuffer(nil)
	err = ul.MarshalSia(buf)
	if err != nil {
		t.Errorf("couldn't marshal sslock: %v", err)
	}
	var ul2 InputLockProxy
	err = ul2.UnmarshalSia(buf)
	if err != nil {
		t.Errorf("couldn't unmarshal sslock: %v", err)
	}
	if !reflect.DeepEqual(ul, ul2) {
		t.Errorf("%v != %v", ul, ul2)
	}
	if !reflect.DeepEqual(ul.UnlockHash(), ul2.UnlockHash()) {
		t.Errorf("UH(%v) != UH(%v)", ul, ul2)
	}

	ul2.t = 42
	err = ul.MarshalSia(buf)
	if err != nil {
		t.Errorf("couldn't marshal ?lock: %v", err)
	}

	var ul3 InputLockProxy
	err = ul3.UnmarshalSia(buf)
	if err != nil {
		t.Errorf("couldn't unmarshal sslock: %v", err)
	}
	if !reflect.DeepEqual(ul, ul3) {
		t.Errorf("%v != %v", ul, ul3)
	}
	if !reflect.DeepEqual(ul.UnlockHash(), ul3.UnlockHash()) {
		t.Errorf("UH(%v) != UH(%v)", ul, ul3)
	}
	ul3.t = UnlockTypeSingleSignature
	err = ul3.MarshalSia(buf)
	if err != nil {
		t.Errorf("couldn't marshal ?lock: %v", err)
	}

	var ul4 InputLockProxy
	err = ul4.UnmarshalSia(buf)
	if err != nil {
		t.Errorf("couldn't unmarshal sslock: %v", err)
	}
	if ul4.t != UnlockTypeSingleSignature {
		t.Errorf("wrong input lock type: %v", ul4.t)
	}
	if !reflect.DeepEqual(ul, ul4) {
		t.Errorf("%v != %v", ul, ul4)
	}
	if !reflect.DeepEqual(ul.UnlockHash(), ul4.UnlockHash()) {
		t.Errorf("UH(%v) != UH(%v)", ul, ul4)
	}
}

func TestAtomicSwapUnknownEncoding(t *testing.T) {
	_, pk := crypto.GenerateKeyPair()
	sk2, pk2 := crypto.GenerateKeyPair()

	var secret, hashedSecret [sha256.Size]byte
	hashedSecret = sha256.Sum256(secret[:])

	ul := NewAtomicSwapInputLock(AtomicSwapCondition{
		Sender:       NewSingleSignatureInputLock(Ed25519PublicKey(pk)).UnlockHash(),
		Receiver:     NewSingleSignatureInputLock(Ed25519PublicKey(pk2)).UnlockHash(),
		HashedSecret: AtomicSwapHashedSecret(hashedSecret),
		TimeLock:     CurrentTimestamp() + 500000,
	})
	err := ul.Lock(0, Transaction{}, AtomicSwapClaimKey{
		PublicKey: Ed25519PublicKey(pk2),
		Secret:    secret,
		SecretKey: sk2[:],
	})
	if err != nil {
		t.Errorf("error while locking InputLock: %v", err)
	}

	// serialize using specific struct made for it
	buf := bytes.NewBuffer(nil)
	err = ul.MarshalSia(buf)
	if err != nil {
		t.Errorf("couldn't marshal sslock: %v", err)
	}
	var ul2 InputLockProxy
	err = ul2.UnmarshalSia(buf)
	if err != nil {
		t.Errorf("couldn't unmarshal sslock: %v", err)
	}
	if !reflect.DeepEqual(ul, ul2) {
		fmt.Println(ul.il, ul2.il)
		t.Errorf("%v != %v", ul, ul2)
	}
	if !reflect.DeepEqual(ul.UnlockHash(), ul2.UnlockHash()) {
		t.Errorf("UH(%v) != UH(%v)", ul, ul2)
	}

	ul2.t = 42
	err = ul.MarshalSia(buf)
	if err != nil {
		t.Errorf("couldn't marshal ?lock: %v", err)
	}

	var ul3 InputLockProxy
	err = ul3.UnmarshalSia(buf)
	if err != nil {
		t.Errorf("couldn't unmarshal sslock: %v", err)
	}
	if !reflect.DeepEqual(ul, ul3) {
		t.Errorf("%v != %v", ul, ul3)
	}
	if !reflect.DeepEqual(ul.UnlockHash(), ul3.UnlockHash()) {
		t.Errorf("UH(%v) != UH(%v)", ul, ul3)
	}
	ul3.t = UnlockTypeAtomicSwap
	err = ul3.MarshalSia(buf)
	if err != nil {
		t.Errorf("couldn't marshal ?lock: %v", err)
	}

	var ul4 InputLockProxy
	err = ul4.UnmarshalSia(buf)
	if err != nil {
		t.Errorf("couldn't unmarshal sslock: %v", err)
	}
	if ul4.t != UnlockTypeAtomicSwap {
		t.Errorf("wrong input lock type: %v", ul4.t)
	}
	if !reflect.DeepEqual(ul, ul4) {
		t.Errorf("%v != %v", ul, ul4)
	}
	if !reflect.DeepEqual(ul.UnlockHash(), ul4.UnlockHash()) {
		t.Errorf("UH(%v) != UH(%v)", ul, ul4)
	}
}

func TestInputLockProxyEncoding(t *testing.T) {
	var ul InputLockProxy
	err := encoding.NewDecoder(bytes.NewReader([]byte{0})).Decode(&ul)
	if err != nil {
		t.Errorf("error while decoding a nil-input-porxy: %v", err)
	}
	if ul.t != UnlockTypeNil {
		t.Errorf("wrong inputLock type: %v", ul.t)
	}

	var entropy [crypto.EntropySize]byte
	sk, pk := crypto.GenerateKeyPairDeterministic(entropy)
	ul = NewSingleSignatureInputLock(Ed25519PublicKey(pk))
	err = ul.Lock(0, Transaction{}, sk)
	if err != nil {
		t.Errorf("error while locking InputLock: %v", err)
	}

	buf := bytes.NewBuffer(nil)
	err = ul.MarshalSia(buf)
	if err != nil {
		t.Errorf("error while marshaling InputLock: %v", err)
	}

	expectedBytes := []byte{1, 56, 0, 0, 0, 0, 0, 0, 0, 101, 100, 50, 53, 53, 49, 57, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 59, 106, 39, 188, 206, 182, 164, 45, 98, 163, 168, 208, 42, 111, 13, 115, 101, 50, 21, 119, 29, 226, 67, 166, 58, 192, 72, 161, 139, 89, 218, 41, 64, 0, 0, 0, 0, 0, 0, 0, 66, 25, 213, 18, 135, 159, 37, 24, 230, 55, 197, 175, 69, 230, 179, 61, 126, 220, 255, 213, 2, 30, 181, 202, 91, 195, 38, 154, 136, 120, 151, 38, 228, 28, 153, 72, 152, 85, 249, 167, 220, 161, 15, 61, 4, 55, 21, 134, 208, 142, 59, 180, 111, 50, 50, 126, 97, 12, 116, 246, 162, 235, 93, 9}
	if bytes.Compare(expectedBytes, buf.Bytes()) != 0 {
		t.Errorf("wrong marshaling: %v", buf.Bytes())
	}

	var ul2 InputLockProxy
	err = ul2.UnmarshalSia(buf)
	if err != nil {
		t.Errorf("error while unmarshaling InputLock: %v", err)
	}
	if !reflect.DeepEqual(ul, ul2) {
		t.Errorf("%v != %v", ul, ul2)
	}

	var secret, hashedSecret [sha256.Size]byte
	hashedSecret = sha256.Sum256(secret[:])

	var entropy2 [crypto.EntropySize]byte
	for i := range entropy2 {
		entropy2[i]++
	}
	sk2, pk2 := crypto.GenerateKeyPairDeterministic(entropy2)
	ul = NewAtomicSwapInputLock(AtomicSwapCondition{
		Sender:       NewSingleSignatureInputLock(Ed25519PublicKey(pk)).UnlockHash(),
		Receiver:     NewSingleSignatureInputLock(Ed25519PublicKey(pk2)).UnlockHash(),
		HashedSecret: AtomicSwapHashedSecret(hashedSecret),
		TimeLock:     CurrentTimestamp() + 500000,
	})
	err = ul.Lock(0, Transaction{}, AtomicSwapClaimKey{
		PublicKey: Ed25519PublicKey(pk2),
		Secret:    secret,
		SecretKey: sk2[:],
	})
	if err != nil {
		t.Errorf("error while locking InputLock: %v", err)
	}

	err = ul.MarshalSia(buf)
	if err != nil {
		t.Errorf("error while marshaling InputLock: %v", err)
	}
	var ul3 InputLockProxy
	err = ul3.UnmarshalSia(buf)
	if err != nil {
		t.Errorf("error while unmarshaling InputLock: %v", err)
	}
	if !reflect.DeepEqual(ul, ul3) {
		t.Errorf("%v != %v", ul, ul3)
	}
}

func TestAtomicSwapClaim(t *testing.T) {
	_, pk := crypto.GenerateKeyPair()
	sk2, pk2 := crypto.GenerateKeyPair()

	var secret, hashedSecret [sha256.Size]byte
	hashedSecret = sha256.Sum256(secret[:])

	ul := NewAtomicSwapInputLock(AtomicSwapCondition{
		Sender:       NewSingleSignatureInputLock(Ed25519PublicKey(pk)).UnlockHash(),
		Receiver:     NewSingleSignatureInputLock(Ed25519PublicKey(pk2)).UnlockHash(),
		HashedSecret: AtomicSwapHashedSecret(hashedSecret),
		TimeLock:     CurrentTimestamp() + 500000,
	})
	err := ul.Lock(0, Transaction{}, AtomicSwapClaimKey{
		PublicKey: Ed25519PublicKey(pk2),
		Secret:    secret,
		SecretKey: sk2[:],
	})
	if err != nil {
		t.Errorf("error while locking InputLock: %v", err)
	}
	err = ul.Unlock(0, Transaction{})
	if err != nil {
		t.Errorf("failed to claim (redeem) input: %v", err)
	}
}

func TestAtomicSwapRefund(t *testing.T) {
	sk, pk := crypto.GenerateKeyPair()
	_, pk2 := crypto.GenerateKeyPair()

	var secret, hashedSecret [sha256.Size]byte
	hashedSecret = sha256.Sum256(secret[:])

	ul := NewAtomicSwapInputLock(AtomicSwapCondition{
		Sender:       NewSingleSignatureInputLock(Ed25519PublicKey(pk)).UnlockHash(),
		Receiver:     NewSingleSignatureInputLock(Ed25519PublicKey(pk2)).UnlockHash(),
		HashedSecret: AtomicSwapHashedSecret(hashedSecret),
		TimeLock:     CurrentTimestamp() - 100,
	})
	err := ul.Lock(0, Transaction{},
		AtomicSwapRefundKey{
			PublicKey: Ed25519PublicKey(pk),
			SecretKey: sk[:],
		})
	if err != nil {
		t.Errorf("error while locking InputLock: %v", err)
	}
	err = ul.Unlock(0, Transaction{})
	if err != nil {
		t.Errorf("failed to refund input: %v", err)
	}
}

func TestInputLockProxyJSONEncoding(t *testing.T) { // utility funcs
	hbs := func(str string) []byte { // hexStr -> byte slice
		bs, _ := hex.DecodeString(str)
		return bs
	}
	hs := func(str string) (hash crypto.Hash) { // hbs -> crypto.Hash
		copy(hash[:], hbs(str))
		return
	}

	// test cases
	testCases := []struct {
		JSONEncoded            string
		ExpectedInputLockProxy InputLockProxy
		// alternative JSON Output
		OutputJSONEncoded string
	}{
		// nil input lock
		{`{}`, InputLockProxy{}, ``},
		{`{"type":0}`, InputLockProxy{}, `{}`},
		{`{"type":0,"condition":null,"fulfillment":null}`, InputLockProxy{}, `{}`},
		// single signature input lock
		{
			`{"type":1,"condition":{"publickey":"ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},"fulfillment":{"signature":"abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"}}`,
			InputLockProxy{
				t: UnlockTypeSingleSignature,
				il: &SingleSignatureInputLock{
					PublicKey: SiaPublicKey{
						Algorithm: SignatureEd25519,
						Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
					},
					Signature: hbs("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"),
				},
			},
			``,
		},
		// atomic swap input lock
		{
			`{"type":2,"condition":{"sender":"010123456789012345678901234567890101234567890123456789012345678901dec8f8544d34","receiver":"01abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc0efb39211ea2a","hashedsecret":"abc543defabc543defabc543defabc543defabc543defabc543defabc543defa","timelock":1522068743},"fulfillment":{"publickey":"ed25519:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","signature":"abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab","secret":"def789def789def789def789def789dedef789def789def789def789def789de"}}`,
			InputLockProxy{
				t: UnlockTypeAtomicSwap,
				il: &AtomicSwapInputLock{
					Sender: UnlockHash{
						Type: UnlockTypeSingleSignature,
						Hash: hs("0123456789012345678901234567890101234567890123456789012345678901"),
					},
					Receiver: UnlockHash{
						Type: UnlockTypeSingleSignature,
						Hash: hs("abc0123abc0123abc0123abc0123abc0abc0123abc0123abc0123abc0123abc0"),
					},
					HashedSecret: AtomicSwapHashedSecret(hs("abc543defabc543defabc543defabc543defabc543defabc543defabc543defa")),
					TimeLock:     1522068743,
					PublicKey: SiaPublicKey{
						Algorithm: SignatureEd25519,
						Key:       hbs("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
					},
					Signature: hbs("abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefab"),
					Secret:    AtomicSwapSecret(hs("def789def789def789def789def789dedef789def789def789def789def789de")),
				},
			},
			``,
		},
		// unknown input lock
		{
			`{"type":42,"condition":"Y29uZGl0aW9u","fulfillment":"ZnVsZmlsbG1lbnQ="}`,
			InputLockProxy{
				t: 42,
				il: &UnknownInputLock{
					Condition:   []byte("condition"),
					Fulfillment: []byte("fulfillment"),
				},
			},
			``,
		},
	}
	for idx, testCase := range testCases {
		var il InputLockProxy
		err := json.Unmarshal([]byte(testCase.JSONEncoded), &il)
		if err != nil {
			t.Error(idx, err)
			continue
		}
		if !reflect.DeepEqual(testCase.ExpectedInputLockProxy, il) {
			t.Errorf("#%d: %v != %v", idx, testCase.ExpectedInputLockProxy, il)
			continue
		}
		b, err := json.Marshal(il)
		if err != nil {
			t.Error(idx, err)
		}
		jsonEncoded := string(b)
		if testCase.OutputJSONEncoded == "" {
			if testCase.JSONEncoded != jsonEncoded {
				t.Errorf("#%d: %v != %v", idx, testCase.JSONEncoded, jsonEncoded)
			}
		} else {
			if testCase.OutputJSONEncoded != jsonEncoded {
				t.Errorf("#%d: %v != %v", idx, testCase.OutputJSONEncoded, jsonEncoded)
			}
		}
	}
}
