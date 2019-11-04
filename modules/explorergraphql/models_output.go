package explorergraphql

import (
	"context"
	"fmt"
	"sync"

	"github.com/threefoldtech/rivine/crypto"
	"github.com/threefoldtech/rivine/modules/explorergraphql/explorerdb"
	"github.com/threefoldtech/rivine/types"
)

type (
	outputData struct {
		Type      *OutputType
		Value     *BigInt
		Condition UnlockCondition
		ParentID  crypto.Hash
	}

	Output struct {
		id     types.OutputID
		child  *Input
		parent OutputParent
		db     explorerdb.DB

		onceData sync.Once
		data     *outputData
		dataErr  error
	}
)

// compile-time interface check
var (
	_ Object = (*Output)(nil)
)

func NewOutputParent(hash crypto.Hash, db explorerdb.DB) (OutputParent, error) {
	objInfo, err := db.GetObjectInfo(explorerdb.ObjectID(hash[:]))
	if err != nil {
		return nil, err
	}
	switch objInfo.Type {
	case explorerdb.ObjectTypeTransaction:
		return NewTransactionWithVersion(
			types.TransactionID(hash), types.TransactionVersion(objInfo.Version),
			nil, db)
	case explorerdb.ObjectTypeBlock:
		return NewBlock(types.BlockID(hash), db), nil
	default:
		return nil, fmt.Errorf(
			"unexpected explorer object %s type %d cannot be converted to OutputParent",
			hash.String(), objInfo.Type)
	}
}

func NewOutput(id types.OutputID, child *Input, parent OutputParent, db explorerdb.DB) *Output {
	return &Output{
		id:     id,
		child:  child,
		parent: parent,
		db:     db,
	}
}

func (output *Output) outputData(ctx context.Context) (*outputData, error) {
	output.onceData.Do(output._outputDataOnce)
	return output.data, output.dataErr
}

func (output *Output) _outputDataOnce() {
	defer func() {
		if e := recover(); e != nil {
			output.dataErr = fmt.Errorf("failed to fetch output %s data from DB: %v", output.id.String(), e)
		}
	}()

	data, err := output.db.GetOutput(output.id)
	if err != nil {
		output.dataErr = fmt.Errorf("failed to fetch output %s data from DB: %v", output.id.String(), err)
		return
	}

	// restructure all fetched data
	gqlCondition, err := dbConditionAsUnlockCondition(data.Condition)
	output.data = &outputData{
		Type:      dbOutputTypeAsGQL(data.Type),
		Value:     dbCurrencyAsBigIntRef(data.Value),
		Condition: gqlCondition,
		ParentID:  data.ParentID,
	}
	if output.child == nil && data.SpenditureData != nil {
		output.child = NewInput(output.id, output, output.db)
		// already resolve data, as we already have it anyhow
		output.child.resolveDataWithOutput(data.SpenditureData.Fulfillment)
	}
	if output.parent == nil {
		output.parent, err = NewOutputParent(data.ParentID, output.db)
		if err != nil {
			output.dataErr = fmt.Errorf("failed to fetch output parent %s data from DB: %v", data.ParentID.String(), err)
			return
		}
	}
}

// IsObject implements the GraphQL Object interface
func (output *Output) IsObject() {}

func (output *Output) ID(ctx context.Context) (crypto.Hash, error) {
	return crypto.Hash(output.id), nil
}

func (output *Output) Type(ctx context.Context) (*OutputType, error) {
	data, err := output.outputData(ctx)
	if err != nil {
		return nil, err
	}
	return data.Type, nil
}

func (output *Output) Value(ctx context.Context) (*BigInt, error) {
	data, err := output.outputData(ctx)
	if err != nil {
		return nil, err
	}
	return data.Value, nil
}

func (output *Output) Condition(ctx context.Context) (UnlockCondition, error) {
	data, err := output.outputData(ctx)
	if err != nil {
		return nil, err
	}
	return data.Condition, nil
}

func (output *Output) ChildInput(ctx context.Context) (*Input, error) {
	_, err := output.outputData(ctx)
	if err != nil {
		return nil, err
	}
	// data is not used, but the internal function does resolve
	// a non-nil child if not given and if stored in DB
	// (if output isn't spent, it won't be found in DB and will still be nil at this point)
	return output.child, nil
}

func (output *Output) ParentID(ctx context.Context) (crypto.Hash, error) {
	data, err := output.outputData(ctx)
	if err != nil {
		return crypto.Hash{}, err
	}
	return data.ParentID, nil
}

func (output *Output) Parent(ctx context.Context) (OutputParent, error) {
	_, err := output.outputData(ctx)
	if err != nil {
		return nil, err
	}

	// data is not used, but the internal function does resolve
	// a non-nil parent if not given

	return output.parent, nil
}

type (
	inputData struct {
		Type        *OutputType
		Value       *BigInt
		Fulfillment types.UnlockFulfillmentProxy
	}

	Input struct {
		id     types.OutputID
		parent *Output
		db     explorerdb.DB

		onceData sync.Once
		data     *inputData
		dataErr  error
	}
)

func NewInput(id types.OutputID, parent *Output, db explorerdb.DB) *Input {
	return &Input{
		id:     id,
		parent: parent,
		db:     db,
	}
}

func (input *Input) inputData(ctx context.Context) (*inputData, error) {
	input.onceData.Do(input._inputDataOnce)
	return input.data, input.dataErr
}

func (input *Input) _inputDataOnce() {
	defer func() {
		if e := recover(); e != nil {
			input.dataErr = fmt.Errorf("failed to fetch output %s (used as input) data from DB: %v", input.id.String(), e)
		}
	}()

	data, err := input.db.GetOutput(input.id)
	if err != nil {
		input.dataErr = fmt.Errorf("failed to fetch output %s data (used as input) from DB: %v", input.id.String(), err)
		return
	}

	// restructure all fetched data
	if data.SpenditureData == nil {
		input.dataErr = fmt.Errorf("failed to convert output %s data (used as input): spenditure data of output is not defined and can as such not be used as a GQL input", input.id.String())
		return
	}
	if input.parent == nil {
		input.parent = NewOutput(input.id, input, nil, input.db)
	}
	input.data = &inputData{
		Type:        dbOutputTypeAsGQL(data.Type),
		Value:       dbCurrencyAsBigIntRef(data.Value),
		Fulfillment: data.SpenditureData.Fulfillment,
	}
}

func (input *Input) resolveDataWithOutput(fulfillment types.UnlockFulfillmentProxy) {
	if input.parent == nil {
		panic("BUG: parent should never be nil at this point")
	}
	if input.parent.data == nil {
		panic("BUG: parent output data should never be nil at this point")
	}
	input.onceData.Do(func() {
		input.data = &inputData{
			Type:        input.parent.data.Type,
			Value:       input.parent.data.Value,
			Fulfillment: fulfillment,
		}
	})
}

func (input *Input) ID(ctx context.Context) (crypto.Hash, error) {
	return crypto.Hash(input.id), nil
}
func (input *Input) Type(ctx context.Context) (*OutputType, error) {
	data, err := input.inputData(ctx)
	if err != nil {
		return nil, err
	}
	return data.Type, nil
}
func (input *Input) Value(ctx context.Context) (*BigInt, error) {
	data, err := input.inputData(ctx)
	if err != nil {
		return nil, err
	}
	return data.Value, nil
}
func (input *Input) Fulfillment(ctx context.Context) (UnlockFulfillment, error) {
	output, err := input.ParentOutput(ctx)
	if err != nil {
		return nil, err
	}
	parentCondition, err := output.Condition(ctx)
	if err != nil {
		return nil, err
	}
	// input.data is once resolved as a child call of the initial
	// `input.ParentOutput(ctx)` call
	return dbFulfillmentAsUnlockFulfillment(input.data.Fulfillment, parentCondition)
}
func (input *Input) ParentOutput(ctx context.Context) (*Output, error) {
	_, err := input.inputData(ctx)
	if err != nil {
		return nil, err
	}

	// data is not used, but the internal function does resolve
	// a non-nil parent if not given

	return input.parent, nil
}
