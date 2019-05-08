package explorer

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"runtime"

	"github.com/threefoldtech/rivine/pkg/encoding/siabin"
	persist "github.com/threefoldtech/rivine/tarantool-persist"
	"github.com/threefoldtech/rivine/types"
)

var (
	errNotExist = errors.New("entry does not exist")

	BlockSpace            = "Block"
	TransactionSpace      = "Transaction"
	CoinOutputSpace       = "CoinOutput"
	CoinInputSpace        = "CoinInput"
	BlockStakeOutputSpace = "BlockstakeOutput"
	BlockStakeInputSpace  = "BlockStakeInput"
	FullfillmentSpace     = "Fullfillments"
	UnlockConditionSpace  = "UnlockConditions"
	WalletToMultiSigSpace = "WalletToMultiSig"
	InternalSpace         = "Internal"

	BlockIDIndex     = "BlockID"
	BlockHeightIndex = "BlockHeight"

	TransactionIDIndex = "Txid"
	// FK BlockIDIndex

	CoinOutputIDIndex = "ID"
	// FK TXID

	CoinInputIDIndex = "ParentID"
	// FK TXID

	BlockstakeOutputIDIndex = "ID"
	// FK TXID

	BlockstakeInputIDIndex = "ParentID"
	// FK TXID

	UnlockConditionIDIndex = "ID"
	FullfillmentIDIndex    = "ID"
	InternalSpaceIDIndex   = "ConsensusChangeID"

	WalletToMultiSigWalletIndex   = "WalletAddress"
	WalletToMultiSigMultisigIndex = "MultisigAddress"
)

// SetupExplorerDatabase reads in the Lua database creation file and passes these expressions to the client for evaluation
func SetupExplorerDatabase(client *persist.TarantoolClient) error {
	_, err := client.Eval(`box.session.su('admin')`, []interface{}{})
	if err != nil {
		return fmt.Errorf("Error switching to admin user: , %s", err.Error())
	}

	_, filename, _, ok := runtime.Caller(1)
	filepath := path.Join(path.Dir(filename), "./database.lua")
	if !ok {
		return fmt.Errorf("Error reading Lua database file, %s")
	}

	expr, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("Error reading Lua database file, %s", err.Error())
	}

	_, err = client.Eval(string(expr), []interface{}{})
	if err != nil {
		return fmt.Errorf("Error parsing database expression, %s", err.Error())
	}

	_, err = client.Call("createExplorerDatabase", []interface{}{})
	if err != nil {
		return fmt.Errorf("Error creating database, %s", err.Error())
	}
	return nil
}

// SetupExplorerDatabaseOperations initializes database operation so they can later be called with persist.Tarantoolclient.call(..operation)
// Operations can be selects, inserts, updates, ...
func SetupExplorerDatabaseOperations(client *persist.TarantoolClient) error {
	_, err := client.Eval(`box.session.su('admin')`, []interface{}{})
	if err != nil {
		return fmt.Errorf("Error switching to admin user: , %s", err.Error())
	}

	_, filename, _, ok := runtime.Caller(1)
	filepath := path.Join(path.Dir(filename), "./functions.lua")
	if !ok {
		return fmt.Errorf("Error reading Lua database file, %s")
	}

	expr, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("Error reading Lua database file, %s", err.Error())
	}

	_, err = client.Eval(string(expr), []interface{}{})
	if err != nil {
		return fmt.Errorf("Error parsing database expression, %s", err.Error())
	}
	return nil
}

// dbGetAndDecode returns a 'func(*bolt.Tx) error' that retrieves and decodes
// a value from the specified bucket. If the value does not exist,
// dbGetAndDecode returns errNotExist.
func dbGetAndDecode(function string, key, val interface{}, client *persist.TarantoolClient) error {
	// valBytes, err := client.Get(space, "id", 0, 0, tarantool.IterEq, key)
	valBytes, err := client.Call(function, []interface{}{key})
	if valBytes == nil {
		return errNotExist
	}
	if err != nil {
		return err
	}
	for _, interf := range valBytes {
		bytes, ok := interf.([]byte)
		if ok {
			return siabin.Unmarshal(bytes, val)
		}
	}
	return nil
}

// dbGetTransactionIDSet returns a 'func(*bolt.Tx) error' that decodes a
// bucket of transaction IDs into a slice. If the bucket is nil,
// dbGetTransactionIDSet returns errNotExist.
func dbGetTransactionIDSet(function string, key interface{}, ids *[]types.TransactionID, client *persist.TarantoolClient) error {
	// valBytes, err := client.Get(space, "id", 0, 0, tarantool.IterAll, siabin.Marshal(key))
	valBytes, err := client.Call(function, []interface{}{key})
	// b := tx.Bucket(bucket).Bucket(siabin.Marshal(key))
	if valBytes == nil {
		return errNotExist
	}
	if err != nil {
		return err
	}
	// decode into a local slice
	var txids []types.TransactionID
	for _, interf := range valBytes {
		var id types.TransactionID

		bytes, ok := interf.([]byte)
		if ok {
			err := siabin.Unmarshal(bytes, &id)
			if err != nil {
				return err
			}
			txids = append(txids, id)
			return nil
		}
	}
	// set pointer
	*ids = txids
	return nil

}

// dbGetBlockFacts returns a 'func(*bolt.Tx) error' that decodes
// the block facts for `height` into blockfacts
func (e *Explorer) dbGetBlockFacts(height types.BlockHeight, bf *blockFacts) error {
	block, exists := e.cs.BlockAtHeight(height)
	if !exists {
		return errors.New("requested block facts for a block that does not exist")
	}
	return dbGetAndDecode("get_block", block.ID().String(), bf, e.client)
}

// dbSetConsensusChangeID sets the specified key of bucketInternal to the encoded value.
func dbSetConsensusChangeID(val interface{}, client *persist.TarantoolClient) error {
	// _, err := client.Upsert(key, key, siabin.Marshal(val))
	_, err := client.Call("insert_internal", []interface{}{val})
	if err != nil {
		return fmt.Errorf("setting the internal key failed %s", err.Error())
	}
	return nil
}

// dbGetConsensusChangeID gets the latest consensusChangeId.
func dbGetConsensusChangeID(val interface{}, client *persist.TarantoolClient) error {
	// valBytes, err := client.Get(InternalSpace, InternalSpaceIDIndex, 0, 1, tarantool.IterEq, key)
	valBytes, err := client.Call("get_consensus_changeid", []interface{}{})
	if valBytes == nil {
		return errNotExist
	}
	if err != nil {
		return err
	}
	if len(valBytes) > 0 {
		data := valBytes[0].([]interface{})
		value := data[1].([]byte)
		return siabin.Unmarshal(value, val)
	}
	return nil
}

// dbGetBlockheight decodes the specified key of bucketInternal into the supplied pointer.
func dbGetBlockheight(val interface{}, client *persist.TarantoolClient) error {
	valBytes, err := client.Call("get_blockheight", []interface{}{})
	if valBytes == nil {
		return errNotExist
	}
	if err != nil {
		return err
	}
	if len(valBytes) > 0 {
		data := valBytes[0].([]interface{})
		value := data[1].([]byte)
		return siabin.Unmarshal(value, val)
	}
	return nil
}

// dbGetBlockheight decodes the specified key of bucketInternal into the supplied pointer.
func dbGetStartingBlockheight(val interface{}, client *persist.TarantoolClient) error {
	valBytes, err := client.Call("get_starting_blockheight", []interface{}{})
	if valBytes == nil {
		return errNotExist
	}
	if err != nil {
		return err
	}
	if len(valBytes) > 0 {
		data := valBytes[0].([]interface{})
		value := data[1].([]byte)
		return siabin.Unmarshal(value, val)
	}
	return nil
}
