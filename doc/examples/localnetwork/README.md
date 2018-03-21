# Local testnet using docker

Run the [setup sell script](deploy_testnet.sh) to set up a small rivine test network using some dockers. It will first create the docker image
with the [provided dockerfile](Dockerfile_testnet). Note that this docker image will have a golang environment as well as the test binaries, making it unfit for prodution use. 

The setup script makes the very bold assumption that no dockers with the names `r1`, `r2` and `r3` do not exist yet.

The `rivine` network used for this example will run in dockers, with the following setup:
    - 2 nodes with all modules enabled, these will create blocks on the network
    - 1 node with the gateway and transactionpool module (can accept transactions), but without wallet and block creator modules. It also has the explorer module

Because the gateway node listens on none-localhost addresses, it requires API authentication. In the setup script, the API password is set to `test123`.
After running the setup script, you can verify that the gateway is indeed accessible by connecting with a rivinec binary from one of the 2 "mining" docker containers by adding the `-a $GATEWAY_IP:23110` flag (e.g. `docker exec -ti r1 rivinec gateway list -a 172.17.0.4:23110` to verify that the gateway is indeed connected to both other nodes, as we only connected it to 1 in the setup script). 

Like said earlier, this setup will by default run with the `test123` password. Changing the password can be done by editing it in both the [deploy_testnet.sh]() and [offline_transaction.go]() files. 
At the end of the script, some usefull info will be printed, e.g.: 

```bash
Gateway addr:
http://172.17.0.4:23110
Possible address to send coins: e5bd83a85e263817e2040054064575066874ee45a7697facca7a2721d4792af374ea35f549a1
```
