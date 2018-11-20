package client

import (
	"encoding/json"
	"testing"

	"github.com/threefoldtech/rivine/types"
)

func TestCompareNonMergeableTransactionData(t *testing.T) {
	testCases := []struct {
		Master, Other transactionInputs
		IsError       bool
		Description   string
	}{
		{transactionInputs{}, transactionInputs{}, false, "nil transactions"},
		{transactionInputs{Version: types.TransactionVersionOne}, transactionInputs{}, true, "different transaction versions"},
		{
			transactionInputs{Version: types.TransactionVersionOne},
			transactionInputs{Version: types.TransactionVersionOne},
			false, "nil transaction data for both txns",
		},
		{
			transactionInputs{
				Version: types.TransactionVersionOne,
			},
			transactionInputs{
				Version: types.TransactionVersionOne,
				Data: transactionInputData{
					ArbitraryData: []byte("hello"),
				},
			},
			true, "arbitrary data different",
		},
		{
			transactionInputs{
				Version: types.TransactionVersionOne,
			},
			transactionInputs{
				Version: types.TransactionVersionOne,
				Data: transactionInputData{
					MinerFees: json.RawMessage("42"),
				},
			},
			true, "miner fees are different",
		},
		{
			transactionInputs{
				Version: types.TransactionVersionOne,
				Data: transactionInputData{
					MinerFees: json.RawMessage("24"),
				},
			},
			transactionInputs{
				Version: types.TransactionVersionOne,
				Data: transactionInputData{
					MinerFees: json.RawMessage("42"),
				},
			},
			true, "miner fees are different",
		},
		{
			transactionInputs{
				Version: types.TransactionVersionOne,
				Data: transactionInputData{
					BlockStakeOutputs: json.RawMessage("different"),
				},
			},
			transactionInputs{
				Version: types.TransactionVersionOne,
			},
			true, "block stake outputs are different",
		},
		{
			transactionInputs{
				Version: types.TransactionVersionOne,
			},
			transactionInputs{
				Version: types.TransactionVersionOne,
				Data: transactionInputData{
					CoinOutputs: json.RawMessage("different"),
				},
			},
			true, "block stake outputs are different",
		},
		{
			transactionInputs{
				Version: types.TransactionVersionOne,
				Data: transactionInputData{
					CoinInputs: []types.CoinInput{
						types.CoinInput{},
					},
				},
			},
			transactionInputs{
				Version: types.TransactionVersionOne,
			},
			true, "coin input length is different",
		},
		{
			transactionInputs{
				Version: types.TransactionVersionOne,
			},
			transactionInputs{
				Version: types.TransactionVersionOne,
				Data: transactionInputData{
					BlockStakeInputs: []types.BlockStakeInput{
						types.BlockStakeInput{},
					},
				},
			},
			true, "blockstake input length is different",
		},
		{
			transactionInputs{
				Version: types.TransactionVersionOne,
				Data: transactionInputData{
					CoinInputs: []types.CoinInput{
						{
							ParentID: types.CoinOutputID{},
							Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
								PublicKey: types.PublicKey{
									Algorithm: types.SignatureAlgoEd25519,
									Key:       types.ByteSlice{},
								},
								Signature: types.ByteSlice{},
							}),
						},
					},
					CoinOutputs: json.RawMessage("[a,b]"),
					BlockStakeInputs: []types.BlockStakeInput{
						{
							ParentID: types.BlockStakeOutputID{},
							Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
								PublicKey: types.PublicKey{
									Algorithm: types.SignatureAlgoEd25519,
									Key:       types.ByteSlice{},
								},
								Signature: types.ByteSlice{},
							}),
						},
					},
					BlockStakeOutputs: json.RawMessage("[a,b]"),
					MinerFees:         json.RawMessage("1000"),
					ArbitraryData:     json.RawMessage("Hello, world!"),
				},
			},
			transactionInputs{
				Version: types.TransactionVersionOne,
				Data: transactionInputData{
					CoinInputs: []types.CoinInput{
						{
							ParentID: types.CoinOutputID{},
							Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
								PublicKey: types.PublicKey{
									Algorithm: types.SignatureAlgoEd25519,
									Key:       types.ByteSlice{},
								},
								Signature: types.ByteSlice{},
							}),
						},
					},
					CoinOutputs: json.RawMessage("[a,b]"),
					BlockStakeInputs: []types.BlockStakeInput{
						{
							ParentID: types.BlockStakeOutputID{},
							Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
								PublicKey: types.PublicKey{
									Algorithm: types.SignatureAlgoEd25519,
									Key:       types.ByteSlice{},
								},
								Signature: types.ByteSlice{},
							}),
						},
					},
					BlockStakeOutputs: json.RawMessage("[a,b]"),
					MinerFees:         json.RawMessage("1000"),
					ArbitraryData:     json.RawMessage("Hello, world!"),
				},
			},
			false, "all non mergeable data is equal",
		},
		{
			transactionInputs{
				Version: types.TransactionVersionOne,
				Data: transactionInputData{
					CoinInputs: []types.CoinInput{
						{
							ParentID: types.CoinOutputID{},
							Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
								PublicKey: types.PublicKey{
									Algorithm: types.SignatureAlgoEd25519,
									Key:       types.ByteSlice{},
								},
								Signature: types.ByteSlice{},
							}),
						},
					},
					CoinOutputs: json.RawMessage("[a,b]"),
					BlockStakeInputs: []types.BlockStakeInput{
						{
							ParentID: types.BlockStakeOutputID{},
							Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
								PublicKey: types.PublicKey{
									Algorithm: types.SignatureAlgoEd25519,
									Key:       types.ByteSlice{},
								},
								Signature: types.ByteSlice{},
							}),
						},
					},
					BlockStakeOutputs: json.RawMessage("[a,b]"),
					MinerFees:         json.RawMessage("1000"),
					ArbitraryData:     json.RawMessage("Hello, world!"),
				},
			},
			transactionInputs{
				Version: types.TransactionVersionOne,
				Data: transactionInputData{
					CoinInputs: []types.CoinInput{
						{
							ParentID: types.CoinOutputID{},
							Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
								PublicKey: types.PublicKey{
									Algorithm: types.SignatureAlgoEd25519,
									Key:       types.ByteSlice{},
								},
								Signature: types.ByteSlice{},
							}),
						},
					},
					CoinOutputs: json.RawMessage("[a,b]"),
					BlockStakeInputs: []types.BlockStakeInput{
						{
							ParentID: types.BlockStakeOutputID{4, 2},
							Fulfillment: types.NewFulfillment(&types.SingleSignatureFulfillment{
								PublicKey: types.PublicKey{
									Algorithm: types.SignatureAlgoEd25519,
									Key:       types.ByteSlice{},
								},
								Signature: types.ByteSlice{},
							}),
						},
					},
					BlockStakeOutputs: json.RawMessage("[a,b]"),
					MinerFees:         json.RawMessage("1000"),
					ArbitraryData:     json.RawMessage("Hello, world!"),
				},
			},
			false, "all non mergeable data is equal, inputs aren't checked beyond length",
		},
	}
	for idx, testCase := range testCases {
		err := compareNonMergeableTransactionData(testCase.Master, testCase.Other)
		if testCase.IsError && err == nil {
			t.Errorf("expected error for testCase #%d (%s), but none was received", idx, testCase.Description)
		} else if !testCase.IsError && err != nil {
			t.Errorf("expected no error for testCase #%d (%s), but one was received: %v", idx, testCase.Description, err)
		}
	}
}

func TestCompareAndMergeFulfillmentsIfNeeded(t *testing.T) {
	var (
		testPairA = types.PublicKeySignaturePair{
			PublicKey: types.PublicKey{
				Algorithm: types.SignatureAlgoEd25519,
				Key: types.ByteSlice{
					1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				},
			},
			Signature: types.ByteSlice{
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			},
		}
		testPairB = types.PublicKeySignaturePair{
			PublicKey: types.PublicKey{
				Algorithm: types.SignatureAlgoEd25519,
				Key: types.ByteSlice{
					2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
				},
			},
			Signature: types.ByteSlice{
				2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
				2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
			},
		}
		testPairC = types.PublicKeySignaturePair{
			PublicKey: types.PublicKey{
				Algorithm: types.SignatureAlgoEd25519,
				Key: types.ByteSlice{
					3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
				},
			},
			Signature: types.ByteSlice{
				3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
				3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
			},
		}
	)
	testCases := []struct {
		MasterFulfillment   types.UnlockFulfillmentProxy
		OtherFulfillment    types.UnlockFulfillmentProxy
		ExpectedFulfillment types.UnlockFulfillmentProxy
		IsError             bool
		Description         string
	}{
		{
			types.UnlockFulfillmentProxy{},
			types.UnlockFulfillmentProxy{},
			types.UnlockFulfillmentProxy{},
			false,
			"nil fulfillments",
		},
		{
			types.UnlockFulfillmentProxy{},
			types.NewFulfillment(new(types.SingleSignatureFulfillment)),
			types.UnlockFulfillmentProxy{},
			true,
			"different fulfillment type",
		},
		{
			types.NewFulfillment(new(types.SingleSignatureFulfillment)),
			types.NewFulfillment(&types.SingleSignatureFulfillment{
				Signature: types.ByteSlice{4, 2},
			}),
			types.UnlockFulfillmentProxy{},
			true,
			"different non-mergeable fulfillment type",
		},
		{
			types.NewFulfillment(&types.MultiSignatureFulfillment{}),
			types.NewFulfillment(&types.MultiSignatureFulfillment{}),
			types.NewFulfillment(&types.MultiSignatureFulfillment{}),
			false,
			"nothing to merge, but all is good non the less",
		},
		{
			types.NewFulfillment(&types.MultiSignatureFulfillment{
				Pairs: []types.PublicKeySignaturePair{testPairA},
			}),
			types.NewFulfillment(&types.MultiSignatureFulfillment{}),
			types.NewFulfillment(&types.MultiSignatureFulfillment{
				Pairs: []types.PublicKeySignaturePair{testPairA},
			}),
			false,
			"good and merged",
		},
		{
			types.NewFulfillment(&types.MultiSignatureFulfillment{
				Pairs: []types.PublicKeySignaturePair{
					testPairA,
					testPairA,
				},
			}),
			types.NewFulfillment(&types.MultiSignatureFulfillment{}),
			types.NewFulfillment(&types.MultiSignatureFulfillment{
				Pairs: []types.PublicKeySignaturePair{
					testPairA,
					testPairA,
				},
			}),
			false,
			"good and merged",
		},
		{
			types.NewFulfillment(&types.MultiSignatureFulfillment{
				Pairs: []types.PublicKeySignaturePair{
					testPairA,
					testPairA,
				},
			}),
			types.NewFulfillment(&types.MultiSignatureFulfillment{
				Pairs: []types.PublicKeySignaturePair{
					testPairA,
				},
			}),
			types.NewFulfillment(&types.MultiSignatureFulfillment{
				Pairs: []types.PublicKeySignaturePair{
					testPairA,
					testPairA,
				},
			}),
			false,
			"good and merged",
		},
		{
			types.NewFulfillment(&types.MultiSignatureFulfillment{
				Pairs: []types.PublicKeySignaturePair{
					testPairA,
					testPairA,
					testPairC,
					testPairC,
				},
			}),
			types.NewFulfillment(&types.MultiSignatureFulfillment{
				Pairs: []types.PublicKeySignaturePair{
					testPairA,
					testPairA,
					testPairB,
					testPairB,
				},
			}),
			types.NewFulfillment(&types.MultiSignatureFulfillment{
				Pairs: []types.PublicKeySignaturePair{
					testPairA,
					testPairA,
					testPairC,
					testPairC,
					testPairB,
					testPairB,
				},
			}),
			false,
			"good and merged",
		},
	}
	for idx, testCase := range testCases {
		err := compareAndMergeFulfillmentsIfNeeded(&testCase.MasterFulfillment, testCase.OtherFulfillment)
		if testCase.IsError && err == nil {
			t.Errorf("expected error for testCase #%d (%s), but none was received",
				idx, testCase.Description)
		} else if !testCase.IsError && err != nil {
			t.Errorf("expected no error for testCase #%d (%s), but one was received: %v",
				idx, testCase.Description, err)
		} else if err == nil && !testCase.MasterFulfillment.Equal(testCase.ExpectedFulfillment) {
			t.Errorf("expected master transaction for testCase #%d (%s) to equal %v, while it actually is %v",
				idx, testCase.Description, testCase.ExpectedFulfillment, testCase.MasterFulfillment)
		}
	}
}
