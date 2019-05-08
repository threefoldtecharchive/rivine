package explorer

import (
	"fmt"
	"os"

	"github.com/threefoldtech/rivine/types"

	persist "github.com/threefoldtech/rivine/tarantool-persist"
)

// initPersist initializes the persistent structures of the explorer module.
func (e *Explorer) initPersist() error {
	// Make the persist directory
	err := os.MkdirAll(e.persistDir, 0700)
	if err != nil {
		return err
	}

	// Open the database
	e.client = persist.NewTarantoolClient()

	// err = SetupExplorerDatabase(e.client)
	// if err != nil {
	// 	return err
	// }

	err = SetupExplorerDatabaseOperations(e.client)
	if err != nil {
		return err
	}

	// set default values for the spaceInternal
	internalDefaults := []struct {
		key string
		val interface{}
	}{
		{"initial", types.BlockHeight(0)},
	}
	// spaceIntern := e.client.Schema.Spaces[spaceInternal]

	for _, d := range internalDefaults {
		data, err := e.client.Call("get_consensus_changeid", []interface{}{})
		// data, err := e.client.Get(InternalSpace, "key", 0, 1, tarantool.IterEq, d.key)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			fmt.Print(d)
			// // data, err = e.client.Insert(InternalSpace, []interface{}{d.key, d.val})
			// _, err = e.client.Call("insert_info", []interface{}{d.key, d.val})
			// if err != nil {
			// 	return err
			// }
		}
	}

	return nil
}
