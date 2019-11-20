package basedb

import (
	"github.com/threefoldtech/rivine/types"
)

type (
	TimestampFilterRange struct {
		Begin *types.Timestamp
		End   *types.Timestamp
	}

	BlockHeightFilterRange struct {
		Begin *types.BlockHeight
		End   *types.BlockHeight
	}

	IntFilterRange struct {
		Min *int
		Max *int
	}

	BlocksFilter struct {
		BlockHeight       *BlockHeightFilterRange
		Timestamp         *TimestampFilterRange
		TransactionLength *IntFilterRange
	}
)

func NewTimestampFilterRange(begin, end *types.Timestamp) *TimestampFilterRange {
	if begin == nil && end == nil {
		return nil
	}
	return &TimestampFilterRange{
		Begin: begin,
		End:   end,
	}
}

func NewBlockHeightFilterRange(begin, end *types.BlockHeight) *BlockHeightFilterRange {
	if begin == nil && end == nil {
		return nil
	}
	return &BlockHeightFilterRange{
		Begin: begin,
		End:   end,
	}
}

func NewIntFilterRange(min, max *int) *IntFilterRange {
	if min == nil && max == nil {
		return nil
	}
	return &IntFilterRange{
		Min: min,
		Max: max,
	}
}
