#! /usr/bin/env bash

set -e

testpass="test123"
genesisoutputseed="across knife thirsty puck itches hazard enmity fainted pebbles unzip echo queen rarest aphid bugs yanks okay abbey eskimos dove orange nouns august ailments inline rebel glass tyrant acumen"

# First ensure that we have the testnet dockers build
docker build -t rivine_testnet ../../../. -f ./Dockerfile_testnet

# Now run 2 which will (well could, only one will have the blockstakes to do so) create the blocks
docker run -d --name r1 rivine_testnet
docker run -d --name r2 rivine_testnet

r1_addr=$(docker inspect -f "{{ .NetworkSettings.IPAddress }}" r1)
r2_addr=$(docker inspect -f "{{ .NetworkSettings.IPAddress }}" r2)

# Connect the dockers
docker exec r1 rivinec gateway connect "$r2_addr:23112"
 

# Create a wallet
echo "$testpass" | docker exec -i r1 rivinec wallet init -p
echo "$testpass" | docker exec -i r1 rivinec wallet unlock

echo "$testpass" | docker exec -i r2 rivinec wallet init -p
echo "$testpass" | docker exec -i r2 rivinec wallet unlock

# Save an address for later
addr=$(docker exec r2 rivinec wallet address)
# Trim the "Created new address: prefix so we only have the hash"
addr=${addr#"Created new address: "}

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

# So piping the password to the docker to ensure that the http api listens on none localhost addresses causes some issues.
docker run -d -i --name r3 rivine_testnet --disable-api-security --authenticate-api --no-bootstrap -M cgte --api-addr :23110
# Do some serious monkey business to get the gateway running
echo $testpass | docker attach r3


# Give the gateway some time to initialize
sleep 1

# Connect the gateway to the network
echo $testpass | docker exec -i r3 rivinec gateway connect "$r1_addr:23112"

# So now we have a docker with a gateway running which accepts commands from non-localhost addresses.
# Echo the gateway ip for good measure
echo "Gateway addr:"
echo http://$(docker inspect -f "{{ .NetworkSettings.IPAddress }}" r3):23110
echo "Possible address to send coins: $addr"
