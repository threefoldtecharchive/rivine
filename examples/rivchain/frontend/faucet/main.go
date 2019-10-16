package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	rivchaintypes "github.com/threefoldtech/rivine/examples/rivchain/pkg/types"
	"github.com/threefoldtech/rivine/extensions/authcointx"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/api"
	"github.com/threefoldtech/rivine/pkg/daemon"
	"github.com/threefoldtech/rivine/types"
)

type faucet struct {
	// cts is a cached version of daemon constants
	// caching here avoids requiring a call to the daemon even if it is local
	cts *modules.DaemonConstants
	// coinsToGive is the amount of coins given in a single transaction
	coinsToGive types.Currency

	// lock to protect the fund endpoints. This ensures the wallet
	// we talk to only has 1 tx in progress at the same time
	mu sync.Mutex
}

var (
	websitePort int
	httpClient  = &api.HTTPClient{
		RootURL:   "http://localhost:23110",
		Password:  "",
		UserAgent: daemon.RivineUserAgent,
	}
	coinsToGive uint64 = 300
)

func getDaemonConstants() (*modules.DaemonConstants, error) {
	var constants modules.DaemonConstants
	err := httpClient.GetWithResponse("/daemon/constants", &constants)
	if err != nil {
		return nil, err
	}
	return &constants, nil
}

func main() {
	log.Println("[INFO] Starting faucet")
	log.Println("[INFO] Loading daemon constants")
	cts, err := getDaemonConstants()
	if err != nil {
		panic(err)
	}

	f := faucet{
		cts:         cts,
		coinsToGive: cts.OneCoin.Mul64(coinsToGive),
	}

	log.Println("[INFO] Faucet listening on port", websitePort)

	http.HandleFunc("/", f.requestFormHandler)
	http.HandleFunc("/request/tokens", f.requestTokensHandler)
	http.HandleFunc("/request/authorize", f.requestAuthorizationHandler)

	// register API endpoint
	http.HandleFunc("/api/v1/coins", f.requestCoins)
	http.HandleFunc("/api/v1/authorize", f.requestAuthorization)
	http.HandleFunc("/api/v1/deauthorize", f.requestDeauthorization)

	log.Println("[INFO] Faucet ready to serve")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", websitePort), nil))
}

func init() {
	flag.IntVar(&websitePort, "port", 2020, "local port to expose this web faucet on")
	flag.StringVar(&httpClient.Password, "daemon-password", httpClient.Password, "optional password, should the used daemon require it")
	flag.StringVar(&httpClient.RootURL, "daemon-address", httpClient.RootURL, "address of the daemon (with unlocked wallet) to talk to")
	flag.Uint64Var(&coinsToGive, "fund-amount", coinsToGive, "amount of coins to give per drip of the faucet")
	flag.Parse()
	// register tx versions for authentication
	types.RegisterTransactionVersion(rivchaintypes.TransactionVersionAuthAddressUpdate, authcointx.AuthAddressUpdateTransactionController{
		AuthInfoGetter:     nil,
		TransactionVersion: rivchaintypes.TransactionVersionAuthAddressUpdate,
	})
	types.RegisterTransactionVersion(rivchaintypes.TransactionVersionAuthConditionUpdate, authcointx.AuthConditionUpdateTransactionController{
		AuthInfoGetter:     nil,
		TransactionVersion: rivchaintypes.TransactionVersionAuthConditionUpdate,
	})
}
