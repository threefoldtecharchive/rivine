# Offline transaction example

This directory contains a code sample about creating and sending a transaction to a remote `rivine` node, which then processes it and broadcasts it on the network.

In order to run the example, first create a  [Local testnet using docker](../localnetwork/README.md). This will set up a small docker test network to test against. At the end of the script, some usefull info will be printed, e.g.: 

```bash
Gateway addr:
http://172.17.0.4:23110
Possible address to send coins: e5bd83a85e263817e2040054064575066874ee45a7697facca7a2721d4792af374ea35f549a1
```

The example can the by run with:
```bash
go run offline_transaction.go http://172.17.0.4:23110 100000000000000000000000000 e5bd83a85e263817e2040054064575066874ee45a7697facca7a2721d4792af374ea35f549a1
```

Where the first argument is the address of the gateway, the second is the amount of Hastings to transfer, and the third is the address of the recipient. The inputs are selected automatically. By default, we use the seed associated with the address to which the genesis block sends the initial coins. This can be changed if desired, but by doing it like this the example can be run without requiring manual intervention or more complex setup. From this seed, 50 addresses are derived, which are then checked for unspent coin outputs on them. When the transaction is created, inputs are randomly added until sufficient funds are present. A fixed miner fee of 10 hastings is also created. Any leftover funds are transfered back to the seed by taking a random address from the ones generated earlier.