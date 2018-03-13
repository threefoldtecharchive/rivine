# go-bip39

[![Build Status](https://travis-ci.org/rivine/go-bip39.svg?branch=master)](https://travis-ci.org/rivine/go-bip39) [![GoDoc](https://godoc.org/github.com/rivine/go-bip39?status.svg)](https://godoc.org/github.com/rivine/go-bip39) [![Go Report Card](https://goreportcard.com/badge/github.com/rivine/go-bip39)](https://goreportcard.com/report/github.com/rivine/go-bip39)

A golang implementation of the BIP0039 spec for mnemonic seeds


## Credits

English wordlist and test vectors are from the standard Python BIP0039 implementation
from the Trezor guys: [https://github.com/trezor/python-mnemonic](https://github.com/trezor/python-mnemonic)

## Example

```go
package main

import (
  "fmt"
  "bytes"

  "github.com/rivine/go-bip39"
)

func main(){
  // Generate a mnemonic for memorization or user-friendly seeds
  entropy, _ := bip39.NewEntropy(256)
  mnemonic, _ := bip39.NewMnemonic(entropy)

  // generate the seed from the mnemonic
  seed := bip39.NewSeed(mnemonic, "Secret Passphrase")

  // Display mnemonic and seed
  fmt.Println("Mnemonic: ", mnemonic)
  fmt.Println("Seed: ", seed)

  // you can also get the original entropy from a mnemonic
  origEntropy, _ := bip39.EntropyFromMnemonic(mnemonic)
  if bytes.Compare(entropy, origEntropy) != 0 {
    panic("oops")
  }
}
```
