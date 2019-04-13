# Rivine minting extensions

## Usage

### Daemon

After creating the consensus module and registering the http handlers you can use following snippet

```golang
// This is an unlockcondition for testing purposes.
uhString := "01fbaa166912244082784a28ec8756d5d2126f789ed25d93074b2afa05f984e32c0ae996e398e5"
var uh types.UnlockHash
if err := uh.LoadString(uhString); err != nil {
    panic(err)
}
condition := types.NewUnlockHashCondition(uh)

/* any condition that is defined types.unlockconditions.go can be passed to NewMintingPlugin
    for example this can be one of following:
    * singlesignature condition, which is an unlockhash
    * timelock condition
    * multisignature condition
*/

// Pass the condition the NewMintingPlugin
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