# Offline transaction example

This directory contains a code sample about creating and sending a transaction to a remote `rivine` node, which then processes it and broadcasts it on the network.

## Setup

First run the [setup sell script](deploy_sample_setup.sh) to set up a small rivine test network using some dockers. It will first create the docker image
with the [provided dockerfile](Dockerfile_testnet). Note that this docker image will have a golang environment as well as the test binaries, making it unfit for prodution use. Then again the main objective of this example is to demonstrate how an offline transaction can be created and registered, and not to teach about using dockers.

The setup script makes the very bold assumption that no dockers with the names `r1`, `r2` and `r3` do not exist yet.

The `rivine` network used for this example will run in dockers, with the following setup:
    - 2 nodes with all modules enabled, these will create blocks on the network
    - 1 node with the gateway and transactionpool module (can accept transactions), but without wallet and block creator modules. It also has the explorer module

Because the gateway node listens on none-localhost addresses, it requires API authentication. In the setup script, the API password is set to `test123`.
After running the setup script, you can verify that the gateway is indeed accessible by connecting with a rivinec binary from one of the 2 "mining" docker containers by adding the `-a $GATEWAY_IP:23110` flag (e.g. `docker exec -ti r1 rivinec gateway list -a 172.17.0.4:23110` to verify that the gateway is indeed connected to both other nodes, as we only connected it to 1 in the setup script). 

Like said earlier, this setup will by default run with the `test123` password. Changing the password can be done by editing it in both the [deploy_sample_setup.sh]() and [offline_transaction.go]() files. 

In order to run the example, first execute [deploy_sample_setup.sh](). This will set up a small docker test network to test against. At the end of the script, some usefull info will be printed, e.g.: 

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