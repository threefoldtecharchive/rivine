package main

import "github.com/rivine/rivine/pkg/client"

func main() {
	// The name defaults to rivine if it isn't specified but set it again to make sure
	client.ClientName = "rivine"
	client.DefaultClient()
}
