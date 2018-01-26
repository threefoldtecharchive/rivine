#! /usr/bin/env bash

set -e

testpass="test123"
genesisoutputseed="across knife thirsty puck itches hazard enmity fainted pebbles unzip echo queen rarest aphid bugs yanks okay abbey eskimos dove orange nouns august ailments inline rebel glass tyrant acumen"

# First ensure that we have the testnet dockers build
docker build -t rivine_testnet ../../../. -f ./Dockerfile_testnet

# Now run 2 which will create the blocks
docker run -d --name r1 rivine_testnet
docker run -d --name r2 rivine_testnet

r1_addr=$(docker inspect -f "{{ .NetworkSettings.IPAddress }}" r1)
r2_addr=$(docker inspect -f "{{ .NetworkSettings.IPAddress }}" r2)

# Connect the dockers
docker exec r1 rivinec gateway connect "$r2_addr:23112"
 
echo "$testpass"; echo "$genesisoutputseed"

# Create a wallet
echo "$testpass" | docker exec -i r1 rivinec wallet init -p
echo "$testpass" | docker exec -i r1 rivinec wallet unlock

# Load the seed
docker exec -i r1 rivinec wallet load seed << EOF
$testpass
$genesisoutputseed
EOF

# restart and unlock the wallet
docker restart r1
echo "$testpass" | docker exec -i r1 rivinec wallet unlock 

# The r1 daemon's wallet now has controll of all the blockstakes, thus r1 is creating blocks +in the network 

# Start a gateway daemon, without wallet or blockcreator modules

# Funny story: if we specify `--entrypoint "rivined ..." with all the required arguments, the `rivined` process will spawn as the init process
# in the docker (PID 1). Which seems to prevent it from binding ports, thus doing anything usefull... Whoever would have known. By specifying 
# rivined like this allong with all of the needed arguments, the docker actually spawn with a `/bin/sh` as init process, which in turn
# spawn the rivined process. Anyway
echo "$testpass" | docker run -d -i --name r3 rivine_testnet rivined --disable-api-security --authenticate-api --no-bootstrap -M cgt

# Connect the gateway to the network
docker exec r3 rivinec gateway connect "$r1_addr:23112"

# So now we have a docker with a gateway running which accepts commands from non-localhost addresses.
# Echo the gateway ip for good measure
echo "gateway ip:"
echo $(docker inspect -f "{{ .NetworkSettings.IPAddress }}" r3)
