package explorerdb

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/types"

	mp "github.com/vmihailenco/msgpack"
)

// storm db models for Rivine types,
// sad work required as Rivine types do not support msgpack

const (
	StormAtomicSwapHashedSecretSize = sha256.Size
	StormAtomicSwapSecretSize       = sha256.Size
)

type (
	StormHash struct {
		crypto.Hash
	}

	StormUnlockHash struct {
		types.UnlockHash
	}

	StormBigInt struct {
		*big.Int
	}

	StormUnlockCondition struct {
		types.UnlockConditionProxy
	}

	StormAtomicSwapCondition struct {
		Sender       StormUnlockHash
		Receiver     StormUnlockHash
		HashedSecret types.AtomicSwapHashedSecret
		TimeLock     types.Timestamp
	}

	StormUnlockFulfillment struct {
		types.UnlockFulfillmentProxy
	}

	StormAtomicSwapFulfillment struct {
		PublicKeyAlgorithm uint8
		PublicKey          []byte
		Signature          []byte
		Secret             []byte
	}
)

var (
	_ mp.CustomEncoder = (*StormHash)(nil)
	_ mp.CustomDecoder = (*StormHash)(nil)

	_ mp.CustomEncoder = (*StormUnlockHash)(nil)
	_ mp.CustomDecoder = (*StormUnlockHash)(nil)

	_ mp.CustomEncoder = (*StormBigInt)(nil)
	_ mp.CustomDecoder = (*StormBigInt)(nil)

	_ mp.CustomEncoder = (*StormUnlockCondition)(nil)
	_ mp.CustomDecoder = (*StormUnlockCondition)(nil)

	_ mp.CustomEncoder = (*StormAtomicSwapCondition)(nil)
	_ mp.CustomDecoder = (*StormAtomicSwapCondition)(nil)

	_ mp.CustomEncoder = (*StormUnlockFulfillment)(nil)
	_ mp.CustomDecoder = (*StormUnlockFulfillment)(nil)

	_ mp.CustomEncoder = (*StormAtomicSwapFulfillment)(nil)
	_ mp.CustomDecoder = (*StormAtomicSwapFulfillment)(nil)
)

func StormHashFromHash(h crypto.Hash) StormHash {
	return StormHash{Hash: h}
}
func StormHashFromOutputID(id types.OutputID) StormHash {
	return StormHashFromHash(crypto.Hash(id))
}
func StormHashFromCoinOutputID(id types.CoinOutputID) StormHash {
	return StormHashFromHash(crypto.Hash(id))
}
func StormHashFromBlockStakeOutputID(id types.BlockStakeOutputID) StormHash {
	return StormHashFromHash(crypto.Hash(id))
}
func StormHashFromTransactionID(id types.TransactionID) StormHash {
	return StormHashFromHash(crypto.Hash(id))
}
func StormHashFromBlockID(id types.BlockID) StormHash {
	return StormHashFromHash(crypto.Hash(id))
}
func StormHashFromTarget(tgt types.Target) StormHash {
	return StormHashFromHash(crypto.Hash(tgt))
}

func (h *StormHash) AsCryptoHash() crypto.Hash {
	return h.Hash
}
func (h *StormHash) AsOutputID() types.OutputID {
	return types.OutputID(h.Hash)
}
func (h *StormHash) AsCoinOutputID() types.CoinOutputID {
	return types.CoinOutputID(h.Hash)
}
func (h *StormHash) AsBlockStakeutputID() types.BlockStakeOutputID {
	return types.BlockStakeOutputID(h.Hash)
}
func (h *StormHash) AsTransactionID() types.TransactionID {
	return types.TransactionID(h.Hash)
}
func (h *StormHash) AsBlockID() types.BlockID {
	return types.BlockID(h.Hash)
}
func (h *StormHash) AsTarget() types.Target {
	return types.Target(h.Hash)
}

func (h StormHash) EncodeMsgpack(enc *mp.Encoder) error {
	return enc.EncodeBytes(h.Hash[:])
}

func (h *StormHash) DecodeMsgpack(dec *mp.Decoder) error {
	b, err := dec.DecodeBytes()
	if err != nil {
		return err
	}
	if len(b) != crypto.HashSize {
		return fmt.Errorf("unexpected msgpack decoded hash size: expected %d, not %d", crypto.HashSize, len(b))
	}
	copy(h.Hash[:], b[:])
	return nil
}

func StormHashSliceAsOutputIDSlice(hashes []StormHash) (ids []types.OutputID) {
	ids = make([]types.OutputID, 0, len(hashes))
	for _, hash := range hashes {
		ids = append(ids, hash.AsOutputID())
	}
	return
}

func OutputIDSliceAsStormHashSlice(ids []types.OutputID) (hashes []StormHash) {
	hashes = make([]StormHash, 0, len(ids))
	for _, id := range ids {
		hashes = append(hashes, StormHashFromOutputID(id))
	}
	return
}

func BlockIDSliceAsStormHashSlice(ids []types.BlockID) (hashes []StormHash) {
	hashes = make([]StormHash, 0, len(ids))
	for _, id := range ids {
		hashes = append(hashes, StormHashFromBlockID(id))
	}
	return
}

func TransactionIDSliceAsStormHashSlice(ids []types.TransactionID) (hashes []StormHash) {
	hashes = make([]StormHash, 0, len(ids))
	for _, id := range ids {
		hashes = append(hashes, StormHashFromTransactionID(id))
	}
	return
}

func StormHashSliceAsTransactionIDSlice(hashes []StormHash) (ids []types.TransactionID) {
	ids = make([]types.TransactionID, 0, len(hashes))
	for _, hash := range hashes {
		ids = append(ids, hash.AsTransactionID())
	}
	return
}

func StormHashSliceAsBlockIDSlice(hashes []StormHash) (ids []types.BlockID) {
	ids = make([]types.BlockID, 0, len(hashes))
	for _, hash := range hashes {
		ids = append(ids, hash.AsBlockID())
	}
	return
}

func StormUnlockHashFromUnlockHash(uh types.UnlockHash) StormUnlockHash {
	return StormUnlockHash{UnlockHash: uh}
}

func (uh *StormUnlockHash) AsUnlockHash() types.UnlockHash {
	return uh.UnlockHash
}

func (uh StormUnlockHash) EncodeMsgpack(enc *mp.Encoder) error {
	err := enc.EncodeUint8(uint8(uh.Type))
	if err != nil {
		return err
	}
	return enc.EncodeBytes(uh.Hash[:])
}

func (uh *StormUnlockHash) DecodeMsgpack(dec *mp.Decoder) error {
	t, err := dec.DecodeUint8()
	if err != nil {
		return err
	}
	uh.Type = types.UnlockType(t)
	b, err := dec.DecodeBytes()
	if err != nil {
		return err
	}
	if len(b) != crypto.HashSize {
		return fmt.Errorf("unexpected msgpack decoded unlockhash's hash size: expected %d, not %d", crypto.HashSize, len(b))
	}
	copy(uh.Hash[:], b[:])
	return nil
}

func StormUnlockHashSliceAsUnlockHashSlice(suhs []StormUnlockHash) (uhs []types.UnlockHash) {
	uhs = make([]types.UnlockHash, 0, len(suhs))
	for _, suh := range suhs {
		uhs = append(uhs, suh.AsUnlockHash())
	}
	return
}

func UnlockHashSliceAsStormUnlockHashSlice(uhs []types.UnlockHash) (suhs []StormUnlockHash) {
	suhs = make([]StormUnlockHash, 0, len(uhs))
	for _, uh := range uhs {
		suhs = append(suhs, StormUnlockHashFromUnlockHash(uh))
	}
	return
}

func StormBigIntFromInt(i *big.Int) StormBigInt {
	return StormBigInt{Int: i}
}
func StormBigIntFromCurrency(c types.Currency) StormBigInt {
	return StormBigIntFromInt(c.Big())
}
func StormBigIntFromDifficulty(d types.Difficulty) StormBigInt {
	return StormBigIntFromInt(d.Big())
}

func (bi *StormBigInt) AsBigInt() *big.Int {
	return bi.Int
}
func (bi *StormBigInt) AsCurrency() types.Currency {
	return types.NewCurrency(bi.Int)
}
func (bi *StormBigInt) AsDifficulty() types.Difficulty {
	return types.NewDifficulty(bi.Int)
}

func StormBigIntSliceAsCurrencies(bis []StormBigInt) (cus []types.Currency) {
	cus = make([]types.Currency, 0, len(bis))
	for _, bi := range bis {
		cus = append(cus, bi.AsCurrency())
	}
	return
}

func CurrencySliceAsStormBigIntSlice(cus []types.Currency) (bis []StormBigInt) {
	bis = make([]StormBigInt, 0, len(bis))
	for _, cu := range cus {
		bis = append(bis, StormBigIntFromCurrency(cu))
	}
	return
}

func (bi StormBigInt) EncodeMsgpack(enc *mp.Encoder) error {
	return enc.EncodeBytes(bi.Bytes())
}

func (bi *StormBigInt) DecodeMsgpack(dec *mp.Decoder) error {
	b, err := dec.DecodeBytes()
	if err != nil {
		return err
	}

	bi.Int = new(big.Int).SetBytes(b)
	return nil
}

func StormAtomicSwapConditionFromAtomicSwapCondition(c types.AtomicSwapCondition) StormAtomicSwapCondition {
	return StormAtomicSwapCondition{
		Sender:       StormUnlockHashFromUnlockHash(c.Sender),
		Receiver:     StormUnlockHashFromUnlockHash(c.Receiver),
		HashedSecret: c.HashedSecret,
		TimeLock:     c.TimeLock,
	}
}
func (sawc *StormAtomicSwapCondition) AsAtomicSwapCondition() types.AtomicSwapCondition {
	return types.AtomicSwapCondition{
		Sender:       sawc.Sender.AsUnlockHash(),
		Receiver:     sawc.Receiver.AsUnlockHash(),
		HashedSecret: sawc.HashedSecret,
		TimeLock:     sawc.TimeLock,
	}
}

func (sawc StormAtomicSwapCondition) EncodeMsgpack(enc *mp.Encoder) error {
	err := sawc.Sender.EncodeMsgpack(enc)
	if err != nil {
		return err
	}
	err = sawc.Receiver.EncodeMsgpack(enc)
	if err != nil {
		return err
	}
	err = enc.EncodeBytes(sawc.HashedSecret[:])
	if err != nil {
		return err
	}
	return enc.EncodeUint64(uint64(sawc.TimeLock))
}

func (sawc *StormAtomicSwapCondition) DecodeMsgpack(dec *mp.Decoder) error {
	err := sawc.Sender.DecodeMsgpack(dec)
	if err != nil {
		return err
	}
	err = sawc.Receiver.DecodeMsgpack(dec)
	if err != nil {
		return err
	}
	b, err := dec.DecodeBytes()
	if err != nil {
		return err
	}
	if len(b) != StormAtomicSwapHashedSecretSize {
		return fmt.Errorf("unexpected msgpack decoded atomic swap hashed secret size: expected %d, not %d", StormAtomicSwapHashedSecretSize, len(b))
	}
	copy(sawc.HashedSecret[:], b)
	t, err := dec.DecodeUint8()
	if err != nil {
		return err
	}
	sawc.TimeLock = types.Timestamp(t)
	return nil
}

func StormAtomicSwapFulfillmentFromAtomicSwapFulfillment(ff types.AtomicSwapFulfillment) StormAtomicSwapFulfillment {
	sff := StormAtomicSwapFulfillment{
		PublicKeyAlgorithm: uint8(ff.PublicKey.Algorithm),
		PublicKey:          ff.PublicKey.Key[:],
		Signature:          ff.Signature[:],
	}
	if ff.Secret != (types.AtomicSwapSecret{}) {
		sff.Secret = ff.Secret[:]
	}
	return sff
}
func (sawf *StormAtomicSwapFulfillment) AsAtomicSwapFulfillment() types.AtomicSwapFulfillment {
	ff := types.AtomicSwapFulfillment{
		PublicKey: types.PublicKey{
			Algorithm: types.SignatureAlgoType(sawf.PublicKeyAlgorithm),
			Key:       types.ByteSlice(sawf.PublicKey),
		},
		Signature: types.ByteSlice(sawf.Signature),
	}
	if len(sawf.Secret) > 0 {
		copy(ff.Secret[:], sawf.Secret)
	}
	return ff
}

func (sawf StormAtomicSwapFulfillment) EncodeMsgpack(enc *mp.Encoder) error {
	err := enc.EncodeUint8(sawf.PublicKeyAlgorithm)
	if err != nil {
		return err
	}
	err = enc.EncodeBytes(sawf.PublicKey)
	if err != nil {
		return err
	}
	err = enc.EncodeBytes(sawf.Signature)
	if err != nil {
		return err
	}
	return enc.EncodeBytes(sawf.Secret)
}

func (sawf *StormAtomicSwapFulfillment) DecodeMsgpack(dec *mp.Decoder) (err error) {
	sawf.PublicKeyAlgorithm, err = dec.DecodeUint8()
	if err != nil {
		return
	}
	if types.SignatureAlgoType(sawf.PublicKeyAlgorithm) != types.SignatureAlgoEd25519 {
		err = fmt.Errorf("unexpected atomic swap public key algorithm %d, only known algorithm is %d", sawf.PublicKeyAlgorithm, types.SignatureAlgoEd25519)
		return
	}
	sawf.PublicKey, err = dec.DecodeBytes()
	if err != nil {
		return
	}
	if len(sawf.PublicKey) != crypto.PublicKeySize {
		err = fmt.Errorf("unexpected msgpack decoded atomic swap public key size: expected %d, not %d", crypto.PublicKeySize, len(sawf.PublicKey))
		return
	}
	sawf.Signature, err = dec.DecodeBytes()
	if err != nil {
		return
	}
	if len(sawf.Signature) != crypto.SignatureSize {
		err = fmt.Errorf("unexpected msgpack decoded atomic swap signature size: expected %d, not %d", crypto.SignatureSize, len(sawf.Signature))
		return
	}
	sawf.Secret, err = dec.DecodeBytes()
	if err != nil {
		return
	}
	if sl := len(sawf.Secret); sl > 0 && sl != StormAtomicSwapSecretSize {
		err = fmt.Errorf("unexpected msgpack decoded atomic swap secret size: expected 0 or %d, not %d", StormAtomicSwapSecretSize, len(sawf.Secret))
		return
	}
	return // nil
}

func StormUnlockConditionFromUnlockCondition(c types.UnlockConditionProxy) StormUnlockCondition {
	return StormUnlockCondition{UnlockConditionProxy: c}
}
func (suc *StormUnlockCondition) AsUnlockConditionProxy() types.UnlockConditionProxy {
	return suc.UnlockConditionProxy
}

func (suc StormUnlockCondition) EncodeMsgpack(enc *mp.Encoder) error {
	switch ct := suc.ConditionType(); ct {
	case types.ConditionTypeUnlockHash:
		uhc, ok := suc.Condition.(*types.UnlockHashCondition)
		if ok {
			return encodeMsgpackUnlockCondition(uhc, enc)
		}
	case types.ConditionTypeMultiSignature:
		msc, ok := suc.Condition.(types.MultiSignatureConditionOwnerInfoGetter)
		if ok {
			return encodeMsgpackMultiSignatureCondition(msc, enc)
		}
	case types.ConditionTypeAtomicSwap:
		asc, ok := suc.Condition.(*types.AtomicSwapCondition)
		if ok {
			err := enc.EncodeBool(true) // to indicate we recognised the uc
			if err != nil {
				return err
			}
			err = enc.EncodeUint8(uint8(types.ConditionTypeAtomicSwap))
			if err != nil {
				return err
			}
			sasc := StormAtomicSwapConditionFromAtomicSwapCondition(*asc)
			return sasc.EncodeMsgpack(enc)
		}
	case types.ConditionTypeTimeLock:
		tlc, ok := suc.Condition.(*types.TimeLockCondition)
		if ok {
			return encodeMsgpackTimeLockCondition(tlc, enc)
		}
	case types.ConditionTypeNil:
		err := enc.EncodeBool(true) // to indicate we recognised the uc
		if err != nil {
			return err
		}
		return enc.EncodeUint8(uint8(ct))
	}
	// (slow) fallback for custom types
	if suc.Condition == nil {
		err := enc.EncodeBool(true) // to indicate we recognised the uc
		if err != nil {
			return err
		}
		return enc.EncodeUint8(uint8(types.ConditionTypeNil))
	}
	return encodeMsgpackMarshableUnlockCondition(suc.Condition, enc)
}

func encodeMsgpackMarshableUnlockCondition(muc types.MarshalableUnlockCondition, enc *mp.Encoder) error {
	err := enc.EncodeBool(false) // to indicate we didn't recognise the uc
	if err != nil {
		return err
	}
	c := types.NewCondition(muc)
	buf := bytes.NewBuffer(nil)
	err = c.MarshalRivine(buf)
	if err != nil {
		return err
	}
	return enc.EncodeBytes(buf.Bytes())
}

func encodeMsgpackUnlockCondition(uc *types.UnlockHashCondition, enc *mp.Encoder) error {
	err := enc.EncodeBool(true) // to indicate we recognised the uc
	if err != nil {
		return err
	}
	err = enc.EncodeUint8(uint8(types.ConditionTypeUnlockHash))
	if err != nil {
		return err
	}
	suc := StormUnlockHashFromUnlockHash(uc.UnlockHash())
	return suc.EncodeMsgpack(enc)
}

func encodeMsgpackMultiSignatureCondition(msc types.MultiSignatureConditionOwnerInfoGetter, enc *mp.Encoder) error {
	err := enc.EncodeBool(true) // to indicate we recognised the uc
	if err != nil {
		return err
	}
	err = enc.EncodeUint8(uint8(types.ConditionTypeMultiSignature))
	if err != nil {
		return err
	}
	slice := msc.UnlockHashSlice()
	err = enc.EncodeInt(int64(len(slice)))
	if err != nil {
		return err
	}
	for _, uh := range slice {
		suh := StormUnlockHashFromUnlockHash(uh)
		err = suh.EncodeMsgpack(enc)
		if err != nil {
			return err
		}
	}
	count := msc.GetMinimumSignatureCount()
	return enc.EncodeUint64(count)
}

func encodeMsgpackTimeLockCondition(tlc *types.TimeLockCondition, enc *mp.Encoder) error {
	err := enc.EncodeBool(true) // to indicate we recognised the uc
	if err != nil {
		return err
	}
	err = enc.EncodeUint8(uint8(types.ConditionTypeTimeLock))
	if err != nil {
		return err
	}

	err = enc.EncodeUint64(tlc.LockTime)
	if err != nil {
		return err
	}

	// encode internal type
	switch ct := tlc.Condition.ConditionType(); ct {
	case types.ConditionTypeUnlockHash:
		uhc, ok := tlc.Condition.(*types.UnlockHashCondition)
		if ok {
			return encodeMsgpackUnlockCondition(uhc, enc)
		}
	case types.ConditionTypeMultiSignature:
		msc, ok := tlc.Condition.(types.MultiSignatureConditionOwnerInfoGetter)
		if ok {
			return encodeMsgpackMultiSignatureCondition(msc, enc)
		}
	case types.ConditionTypeNil:
		err := enc.EncodeBool(true) // to indicate we recognised the uc
		if err != nil {
			return err
		}
		return enc.EncodeUint8(uint8(ct))
	}
	// (slow) fallback for custom types
	if tlc.Condition == nil {
		err := enc.EncodeBool(true) // to indicate we recognised the uc
		if err != nil {
			return err
		}
		return enc.EncodeUint8(uint8(types.ConditionTypeNil))
	}
	return encodeMsgpackMarshableUnlockCondition(tlc.Condition, enc)
}

func (suc *StormUnlockCondition) DecodeMsgpack(dec *mp.Decoder) error {
	knownEncoding, err := dec.DecodeBool()
	if err != nil {
		return err
	}
	if !knownEncoding {
		suc.UnlockConditionProxy, err = decodeMsgpackUnlockConditionProxy(dec)
		return err
	}
	ctx, err := dec.DecodeUint8()
	if err != nil {
		return err
	}
	switch ct := types.ConditionType(ctx); ct {
	case types.ConditionTypeUnlockHash:
		suc.UnlockConditionProxy.Condition, err = decodeMsgpackUnlockHashCondition(dec)
		if err != nil {
			return err
		}
	case types.ConditionTypeTimeLock:
		suc.UnlockConditionProxy.Condition, err = decodeMsgpackTimeLockCondition(dec)
		if err != nil {
			return err
		}
	case types.ConditionTypeMultiSignature:
		suc.UnlockConditionProxy.Condition, err = decodeMsgpackMultiSignatureCondition(dec)
		if err != nil {
			return err
		}
	case types.ConditionTypeAtomicSwap:
		var asc StormAtomicSwapCondition
		err = asc.DecodeMsgpack(dec)
		if err != nil {
			return err
		}
		c := asc.AsAtomicSwapCondition()
		suc.UnlockConditionProxy.Condition = &c
	case types.ConditionTypeNil:
		suc.UnlockConditionProxy.Condition = &types.NilCondition{} // nothing extra to decode
	default:
		return fmt.Errorf("unknown recognised condition type %d", ctx)
	}
	// all good
	return nil
}

func decodeMsgpackUnlockConditionProxy(dec *mp.Decoder) (types.UnlockConditionProxy, error) {
	b, err := dec.DecodeBytes()
	if err != nil {
		return types.UnlockConditionProxy{}, err
	}
	var c types.UnlockConditionProxy
	err = c.UnmarshalRivine(bytes.NewReader(b))
	return c, err
}

func decodeMsgpackMarshalableUnlockCondition(dec *mp.Decoder) (types.MarshalableUnlockCondition, error) {
	c, err := decodeMsgpackUnlockConditionProxy(dec)
	return c.Condition, err
}

func decodeMsgpackUnlockHashCondition(dec *mp.Decoder) (*types.UnlockHashCondition, error) {
	var uh StormUnlockHash
	err := uh.DecodeMsgpack(dec)
	if err != nil {
		return nil, err
	}
	return types.NewUnlockHashCondition(uh.AsUnlockHash()), nil
}

func decodeMsgpackMultiSignatureCondition(dec *mp.Decoder) (*types.MultiSignatureCondition, error) {
	uhl, err := dec.DecodeInt()
	if err != nil {
		return nil, err
	}
	uhs := make([]types.UnlockHash, 0, uhl)
	for i := 0; i < uhl; i++ {
		var uh StormUnlockHash
		err = uh.DecodeMsgpack(dec)
		if err != nil {
			return nil, err
		}
		uhs = append(uhs, uh.AsUnlockHash())
	}
	count, err := dec.DecodeUint64()
	if err != nil {
		return nil, err
	}
	return types.NewMultiSignatureCondition(types.UnlockHashSlice(uhs), count), nil
}

func decodeMsgpackTimeLockCondition(dec *mp.Decoder) (*types.TimeLockCondition, error) {
	tl, err := dec.DecodeUint64()
	if err != nil {
		return nil, err
	}

	// encode internal type
	knownEncoding, err := dec.DecodeBool()
	if err != nil {
		return nil, err
	}
	var internalCondition types.MarshalableUnlockCondition
	if !knownEncoding {
		internalCondition, err = decodeMsgpackMarshalableUnlockCondition(dec)
		if err != nil {
			return nil, err
		}
	} else {
		ctx, err := dec.DecodeUint8()
		if err != nil {
			return nil, err
		}
		switch ct := types.ConditionType(ctx); ct {
		case types.ConditionTypeUnlockHash:
			internalCondition, err = decodeMsgpackUnlockHashCondition(dec)
			if err != nil {
				return nil, err
			}
		case types.ConditionTypeMultiSignature:
			internalCondition, err = decodeMsgpackMultiSignatureCondition(dec)
			if err != nil {
				return nil, err
			}
		case types.ConditionTypeNil:
			// nothing extra to be decoded for a nil condition
			internalCondition = &types.NilCondition{}
		default:
			return nil, fmt.Errorf("unknown internal time lock condition type %d", ctx)
		}
	}
	return &types.TimeLockCondition{
		LockTime:  tl,
		Condition: internalCondition,
	}, nil
}

func StormUnlockFulfillmentFromUnlockFulfillment(ff types.UnlockFulfillmentProxy) StormUnlockFulfillment {
	return StormUnlockFulfillment{UnlockFulfillmentProxy: ff}
}
func (suf *StormUnlockFulfillment) AsUnlockFulfillmentProxy() types.UnlockFulfillmentProxy {
	return suf.UnlockFulfillmentProxy
}

func (suf StormUnlockFulfillment) EncodeMsgpack(enc *mp.Encoder) error {
	switch ft := suf.FulfillmentType(); ft {
	case types.FulfillmentTypeSingleSignature:
		ssf, ok := suf.Fulfillment.(*types.SingleSignatureFulfillment)
		if ok {
			return encodeMsgpackSingleSignatureFulfillment(ssf, enc)
		}
	case types.FulfillmentTypeMultiSignature:
		msf, ok := suf.Fulfillment.(*types.MultiSignatureFulfillment)
		if ok {
			return encodeMsgpackMultiSignatureFulfillment(msf, enc)
		}
	case types.FulfillmentTypeAtomicSwap:
		asf, ok := suf.Fulfillment.(*types.AtomicSwapFulfillment)
		if ok {
			err := enc.EncodeBool(true) // to indicate we recognised the uc
			if err != nil {
				return err
			}
			err = enc.EncodeUint8(uint8(types.FulfillmentTypeAtomicSwap))
			if err != nil {
				return err
			}
			sasf := StormAtomicSwapFulfillmentFromAtomicSwapFulfillment(*asf)
			return sasf.EncodeMsgpack(enc)
		}
	}
	// (slow) fallback for custom types
	return encodeMsgpackUnlockFulfillmentProxy(suf.UnlockFulfillmentProxy, enc)
}

func encodeMsgpackUnlockFulfillmentProxy(ff types.UnlockFulfillmentProxy, enc *mp.Encoder) error {
	err := enc.EncodeBool(false) // to indicate we didn't recognise the uc
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	err = ff.MarshalRivine(buf)
	if err != nil {
		return err
	}
	return enc.EncodeBytes(buf.Bytes())
}

func encodeMsgpackSingleSignatureFulfillment(ss *types.SingleSignatureFulfillment, enc *mp.Encoder) error {
	err := enc.EncodeBool(true) // to indicate we recognised the uc
	if err != nil {
		return err
	}
	err = enc.EncodeUint8(uint8(types.FulfillmentTypeSingleSignature))
	if err != nil {
		return err
	}
	err = enc.EncodeUint8(uint8(ss.PublicKey.Algorithm))
	if err != nil {
		return err
	}
	err = enc.EncodeBytes(ss.PublicKey.Key[:])
	if err != nil {
		return err
	}
	return enc.EncodeBytes(ss.Signature[:])
}

func encodeMsgpackMultiSignatureFulfillment(ms *types.MultiSignatureFulfillment, enc *mp.Encoder) error {
	err := enc.EncodeBool(true) // to indicate we recognised the uc
	if err != nil {
		return err
	}
	err = enc.EncodeUint8(uint8(types.FulfillmentTypeMultiSignature))
	if err != nil {
		return err
	}
	err = enc.EncodeInt(int64(len(ms.Pairs)))
	if err != nil {
		return err
	}
	for _, pair := range ms.Pairs {
		err = enc.EncodeUint8(uint8(pair.PublicKey.Algorithm))
		if err != nil {
			return err
		}
		err = enc.EncodeBytes(pair.PublicKey.Key[:])
		if err != nil {
			return err
		}
		return enc.EncodeBytes(pair.Signature[:])
	}
	// all good
	return nil
}

func (suf *StormUnlockFulfillment) DecodeMsgpack(dec *mp.Decoder) error {
	knownEncoding, err := dec.DecodeBool()
	if err != nil {
		return err
	}
	if !knownEncoding {
		suf.UnlockFulfillmentProxy, err = decodeMsgpackUnlockFulfillmentProxy(dec)
		return err
	}
	ftx, err := dec.DecodeUint8()
	if err != nil {
		return err
	}
	switch ft := types.FulfillmentType(ftx); ft {
	case types.FulfillmentTypeSingleSignature:
		suf.UnlockFulfillmentProxy.Fulfillment, err = decodeMsgpackSingleSignatureFulfillment(dec)
		if err != nil {
			return err
		}
	case types.FulfillmentTypeMultiSignature:
		suf.UnlockFulfillmentProxy.Fulfillment, err = decodeMsgpackMultiSignatureFulfillment(dec)
		if err != nil {
			return err
		}
	case types.FulfillmentTypeAtomicSwap:
		var asf StormAtomicSwapFulfillment
		err = asf.DecodeMsgpack(dec)
		if err != nil {
			return err
		}
		ff := asf.AsAtomicSwapFulfillment()
		suf.UnlockFulfillmentProxy.Fulfillment = &ff
	default:
		return fmt.Errorf("unknown recognised fulfillment type %d", ftx)
	}
	// all good
	return nil
}

func decodeMsgpackUnlockFulfillmentProxy(dec *mp.Decoder) (types.UnlockFulfillmentProxy, error) {
	b, err := dec.DecodeBytes()
	if err != nil {
		return types.UnlockFulfillmentProxy{}, err
	}
	var ff types.UnlockFulfillmentProxy
	err = ff.UnmarshalRivine(bytes.NewReader(b))
	return ff, err
}

func decodeMsgpackSingleSignatureFulfillment(dec *mp.Decoder) (*types.SingleSignatureFulfillment, error) {
	algorithm, err := dec.DecodeUint8()
	if err != nil {
		return nil, err
	}
	if types.SignatureAlgoType(algorithm) != types.SignatureAlgoEd25519 {
		return nil, fmt.Errorf("unexpected single signature public key algorithm %d, only known algorithm is %d", algorithm, types.SignatureAlgoEd25519)
	}
	pkBytes, err := dec.DecodeBytes()
	if err != nil {
		return nil, err
	}
	if len(pkBytes) != crypto.PublicKeySize {
		return nil, fmt.Errorf("unexpected msgpack decoded single signature public key size: expected %d, not %d", crypto.PublicKeySize, len(pkBytes))
	}
	sigBytes, err := dec.DecodeBytes()
	if err != nil {
		return nil, err
	}
	if len(sigBytes) != crypto.SignatureSize {
		return nil, fmt.Errorf("unexpected msgpack decoded single signature signature size: expected %d, not %d", crypto.SignatureSize, len(sigBytes))
	}
	return &types.SingleSignatureFulfillment{
		PublicKey: types.PublicKey{
			Algorithm: types.SignatureAlgoType(algorithm),
			Key:       types.ByteSlice(pkBytes),
		},
		Signature: types.ByteSlice(sigBytes),
	}, nil
}

func decodeMsgpackMultiSignatureFulfillment(dec *mp.Decoder) (*types.MultiSignatureFulfillment, error) {
	pl, err := dec.DecodeInt()
	if err != nil {
		return nil, err
	}
	pairs := make([]types.PublicKeySignaturePair, 0, pl)
	for i := 0; i < pl; i++ {
		algorithm, err := dec.DecodeUint8()
		if err != nil {
			return nil, err
		}
		if types.SignatureAlgoType(algorithm) != types.SignatureAlgoEd25519 {
			return nil, fmt.Errorf("unexpected multi signature pair %d public key algorithm %d, only known algorithm is %d", i+1, algorithm, types.SignatureAlgoEd25519)
		}
		pkBytes, err := dec.DecodeBytes()
		if err != nil {
			return nil, err
		}
		if len(pkBytes) != crypto.PublicKeySize {
			return nil, fmt.Errorf("unexpected msgpack decoded multi signature pair %d public key size: expected %d, not %d", i+1, crypto.PublicKeySize, len(pkBytes))
		}
		sigBytes, err := dec.DecodeBytes()
		if err != nil {
			return nil, err
		}
		if len(sigBytes) != crypto.SignatureSize {
			return nil, fmt.Errorf("unexpected msgpack decoded multi signature pair %d signature size: expected %d, not %d", i+1, crypto.SignatureSize, len(sigBytes))
		}
		pairs = append(pairs, types.PublicKeySignaturePair{
			PublicKey: types.PublicKey{
				Algorithm: types.SignatureAlgoType(algorithm),
				Key:       types.ByteSlice(pkBytes),
			},
			Signature: types.ByteSlice(sigBytes),
		})

	}
	return types.NewMultiSignatureFulfillment(pairs), nil
}
