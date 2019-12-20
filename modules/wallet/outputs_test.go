package wallet

import (
	"sort"
	"testing"

	"github.com/threefoldtech/rivine/types"
)

// TestSortedOutputsSorting checks that the outputs are being correctly sorted
// by the currency value.
func TestSortedOutputsSorting(t *testing.T) {
	so := sortedOutputs{
		ids: []types.CoinOutputID{{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7}},
		outputs: []types.CoinOutput{
			{Value: types.NewCurrency64(2)},
			{Value: types.NewCurrency64(3)},
			{Value: types.NewCurrency64(4)},
			{Value: types.NewCurrency64(7)},
			{Value: types.NewCurrency64(6)},
			{Value: types.NewCurrency64(0)},
			{Value: types.NewCurrency64(1)},
			{Value: types.NewCurrency64(5)},
		},
	}
	sort.Sort(so)

	expectedIDSorting := []types.CoinOutputID{{5}, {6}, {0}, {1}, {2}, {7}, {4}, {3}}
	for i := uint64(0); i < 8; i++ {
		if so.ids[i] != expectedIDSorting[i] {
			t.Error("an id is out of place: ", i)
		}
		if !so.outputs[i].Value.Equals64(i) {
			t.Error("a value is out of place: ", i)
		}
	}
}
