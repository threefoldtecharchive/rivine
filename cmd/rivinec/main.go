package main

import "github.com/rivine/rivine/rivinec"

func main() {
	// The name defaults to rivine if it isn't specified but set it again to make sure
	rivinec.ClientName = "rivine"
	rivinec.DefaultClient()
}
