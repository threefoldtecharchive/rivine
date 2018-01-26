# Offline transaction example

This directory contains a code sample about creating and sending a transaction to a remote `rivine` node, which then processes it and broadcasts it on the network.

## Setup

First run the [setup sell script](deploy_sample_setup.sh) to set up a small rivine test network using some dockers. It will first create the docker image
with the [provided dockerfile](Dockerfile_testnet). Note that this docker image will have a golang environment as well as the test binaries, making it unfit for prodution use. Then again the main objective of this example is to demonstrate how an offline transaction can be created and registered, and not to teach about using dockers.

The setup script makes the very bold assumption that no dockers with the names `r1`, `r2` and `r3` do not exist yet.

The `rivine` network used for this example will run in dockers, with the following setup:
    - 2 nodes with all modules enabled, these will create blocks on the network
    - 1 node with the gateway and transactionpool module (can accept transactions), but without wallet and block creator modules

Because the gateway node listens on none-localhost addresses, it requires API authentication. In the setup script, the API password is set to `test123`.
After running the setup script, you can verify that the gateway is indeed accessible by connecting with a rivinec binary from one of the 2 "mining" docker containers by adding the `-a $GATEWAY_IP:23110` flag (e.g. `docker exec -ti r1 rivinec gateway list -a 172.17.0.4:23110` to verify that the gateway is indeed connected to both other nodes, as we only connected it to 1 in the setup script). 