package main

import (
	"github.com/rivine/rivine/pkg/client"
	"github.com/rivine/rivine/types"
)

func main() {
	bchainInfo := types.DefaultBlockchainInfo()
	client.DefaultCLIClient("", bchainInfo.Name, nil)
}
