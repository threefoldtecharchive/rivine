package explorer

import (
	"encoding/hex"
	"testing"

	"github.com/rivine/rivine/encoding"
)

func TestDecodeLegacyCoinOutput(t *testing.T) {
	testCases := []string{
		`050000000000000002540be4000188adad4e890b57afdcea6ea7c7d8dbc42c939801a3a8015cfc4e779d57d34cf4`,
		`050000000000000002540be40001ec4904210f9868d3d879c513ca2b04e06a41a71ce20d374b8daaab15419ac9d9`,
	}
	for idx, testCase := range testCases {
		b, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(idx, err)
			continue
		}
		var output legacyOutput
		err = encoding.Unmarshal(b, &output)
		if err != nil {
			t.Error(idx, err)
		}
	}
}

func TestDecodeLegacyBlockStakeOutput(t *testing.T) {
	testCases := []string{
		`020000000000000003e801e34588bee49b2cbd53f2198cd5022fbbe78aecb8125a39efb8699720b946e84e`,
		`020000000000000003e80188adad4e890b57afdcea6ea7c7d8dbc42c939801a3a8015cfc4e779d57d34cf4`,
	}
	for idx, testCase := range testCases {
		b, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(idx, err)
			continue
		}
		var output legacyOutput
		err = encoding.Unmarshal(b, &output)
		if err != nil {
			t.Error(idx, err)
		}
	}
}
