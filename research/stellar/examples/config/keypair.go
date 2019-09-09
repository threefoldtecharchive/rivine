package config

import (
	"fmt"

	"github.com/pelletier/go-toml"
	"github.com/stellar/go/keypair"
)

//GetKeyPairFromConfig loads a named keypair from the `config.toml` file  in thre current working directory
func GetKeyPairFromConfig(accountname string) (pair *keypair.Full, err error) {
	config, err := toml.LoadFile("config.toml")

	if err != nil {
		return
	}
	seed := config.Get(accountname + ".seed")
	if seed == nil {
		err = fmt.Errorf("account %s not found", accountname)
		return
	}
	newPK, err := keypair.Parse(seed.(string))
	pair, _ = newPK.(*keypair.Full)
	return
}
