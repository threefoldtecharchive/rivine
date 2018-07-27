package wallet

import (
	"fmt"

	"github.com/rivine/rivine/build"
	"github.com/rivine/rivine/types"
)

func getMultisigConditionProperties(condition types.MarshalableUnlockCondition) ([]types.UnlockHash, uint64) {
	ct := condition.ConditionType()
	if ct == types.ConditionTypeTimeLock {
		cg, ok := condition.(types.MarshalableUnlockConditionGetter)
		if !ok {
			if build.DEBUG {
				panic(fmt.Sprintf("unexpected Go-type for TimeLockCondition: %T", condition))
			}
			return nil, 0
		}
		return getMultisigConditionProperties(cg.GetMarshalableUnlockCondition())
	}
	if ct != types.ConditionTypeMultiSignature {
		return nil, 0
	}
	type multisigCondition interface {
		types.UnlockHashSliceGetter
		GetMinimumSignatureCount() uint64
	}
	switch c := condition.(type) {
	case multisigCondition:
		return c.UnlockHashSlice(), c.GetMinimumSignatureCount()
	default:
		return nil, 0
	}
}
