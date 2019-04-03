# Rivine minting extensions

## Usage

### Daemon

After creating the consensus module and registering the http handlers you can use following snippet

```golang
condition := AnyUnlockCondition
plugin := minting.NewMintingPlugin(types.UnlockConditionProxy{Condition: condition})
    err = cs.RegisterPlugin("minting", plugin, cancel)
    if err != nil {
        return err
    }
```

This will register the minting plugin on the consensus.

### Client

After creating the command line client you can use following snippet

```golang
import (
    mintingcli "github.com/threefoldtech/rivine/extensions/minting/client"
)

// Will create the explore mintcondtion command
// * rivinec explore mintcondition [height] [flags]
mintingcli.CreateExploreCmd(cliClient)

// Will create the createCoinTransaction and createMinterDefinitionTransaction command
// * rivinec wallet create minterdefinitiontransaction
// * rivinec wallet create coincreationtransaction
mintingcli.CreateWalletCmds(cliClient)

mintingReader := mintingcli.NewPluginExplorerClient(cliClient)

// Register the transaction types
types.RegisterTransactionVersion(minting.TransactionVersionMinterDefinition, minting.MinterDefinitionTransactionController{
    MintConditionGetter: mintingReader,
})
types.RegisterTransactionVersion(minting.TransactionVersionCoinCreation, minting.CoinCreationTransactionController{
    MintConditionGetter: mintingReader,
})
```