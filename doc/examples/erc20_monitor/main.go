package main

import (
	"encoding/json"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/nanmu42/etherscan-api"
)

var (
	// client is thread safe
	client *etherscan.Client
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Wrong number of arguments.\nUsage:\n\t%s [apikey] [server address]", os.Args[0])
	}
	apiKey := os.Args[1]
	serverAddr := os.Args[2]

	client = etherscan.New(etherscan.Rinkby, apiKey)

	server := setupServer(serverAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Error while serving requests:", err)
	}
}

func setupServer(serverAddr string) *http.Server {
	server := http.Server{Addr: serverAddr}

	mux := http.NewServeMux()
	mux.HandleFunc("/tokenbalance", getBalance)
	mux.Handle("/", http.FileServer(http.Dir("./public")))

	server.Handler = mux

	return &server
}

// sets Cors header for development
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

// getBalance is the handler for GET /tokenbalance. This endpoint
// returns the token balance on the given address.
func getBalance(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	contractAddres := q.Get("contractaddress")
	address := q.Get("address")
	if address == "" || contractAddres == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	balance, err := client.TokenBalance(contractAddres, address)
	if err != nil {
		log.Error("Unable to get token balance:", err)
		// TODO: Figure out the right error code for this case
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}
