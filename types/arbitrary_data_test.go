package types

import (
	"encoding/hex"
	"testing"

	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
)

// validates that Arbitrary data registered on the chain prior to the
// optional type feature can still be sia-decoded properly.
// Rivine-encoding doesn't have to be tested as no transaction
// was registed using Rivine-encoding when the optional typing of such data became a thing.
func TestArbitraryDataLegacyRivine(t *testing.T) {
	testCases := []string{
		`0000000000000000`, // no arbitrary data
		`180000000000000068656c6c6f206f6c64206172626974726172792064617461`, // some arbitrary data
	}
	for idx, testCase := range testCases {
		b, err := hex.DecodeString(testCase)
		if err != nil {
			t.Error(idx, err)
			continue
		}
		var ad ArbitraryData
		err = siabin.Unmarshal(b, &ad)
		if err != nil {
			t.Error(idx, err)
			continue
		}
		if ad.Type != ArbitraryDataTypeBinary {
			t.Error(idx, "unexpected arbitrary data type:", ad.Type)
			continue
		}
		expectedDataHex := testCase[16:]
		dataHex := hex.EncodeToString(ad.Data)
		if dataHex != expectedDataHex {
			t.Error(idx, dataHex, "!=", expectedDataHex)
		}
	}
}
