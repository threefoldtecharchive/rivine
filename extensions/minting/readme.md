# Minting Extension

This extension powers minting functionality, it basically subscribes to the consensusset.

## Usage

Import this minting extension into an existing project.

`minting "github.com/threefoldtech/rivine/extensions/minting"`

Extend the current `TransactionDB` struct with the one in this Extension like this:

```golang
TransactionDB struct {
    tg rivinesync.ThreadGroup

    db    *persist.BoltDatabase
    stats transactionDBStats

    // This is the extending minting functionality
    *minting.TransactionDB
}
```

Declare the `dbMetadata` in the `TransactionDB`, example:

```golang
var (
	dbMetadata = persist.Metadata{
		Header:  "Any Transaction Database",
		Version: "1",
	}
)
```

Write a new function that extends `TransactionDB`: 

```golang
// ExtendTxdb extends the current TransactionDB with minting functionality
// This function will call Rivine TransactionDB
func (txdb *TransactionDB) ExtendTxdb(genesisMintCondition rivinetypes.UnlockConditionProxy) {
	txdb.ExtendTransactionDB(genesisMintCondition, TransactionDBFilename, dbMetadata)
}
```

After creating `TransactionDB` which can be anywhere you can call

```golang
// Creating the txdb
txdb, err := persist.NewTransactionDB(cfg.RootPersistentDir, GenesisMintCondition)

// Extends txdb with minting functionality
txdb.ExtendTxdb(GenesisMintCondition)
```