#! /usr/bin/env bash

testpass="test123"
genesisoutputseed="recall view document apology stone tattoo job farm pilot favorite mango topic thing dilemma dawn width marble proud pen meadow sing museum lucky present"
network_name="rivnet"

# Check if the test network exists and create it if it does not
docker network inspect $network_name &> /dev/null
if [ $? -ne 0 ]; then
    echo "$network_name network does not exist yet, creating it"
    docker network create $network_name
else 
    echo "$network_name network already exists"
fi

# Ensure that we have the testnet dockers build
docker build -t rivine_testnet ../../../. -f ./Dockerfile_testnet


# Now run 2 which will (well could, only one will have the blockstakes to do so) create the blocks
docker rm -f r1
docker run -d --name r1 --net=$network_name rivine_testnet
docker rm -f r2
docker run -d --name r2 --net=$network_name rivine_testnet

# Connect the dockers
docker exec r1 rivinec gateway connect "r2:23112"
 

# Create a wallet
docker exec -i r1 rivinec wallet init << EOF
$testpass
$testpass
EOF
echo "$testpass" | docker exec -i r1 rivinec wallet unlock

echo "$testpass" | docker exec -i r2 rivinec wallet init << EOF
$testpass
$testpass
EOF
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
docker rm -f r3
docker run -d -i --name r3 --net=$network_name rivine_testnet --disable-api-security --authenticate-api --no-bootstrap -M cgte --api-addr :23110
# Do some serious monkey business to get the gateway running
echo $testpass | docker attach r3


# Give the gateway some time to initialize
sleep 1

# Connect the gateway to the network
echo $testpass | docker exec -i r3 rivinec gateway connect "r1.$network_name:23112"

# So now we have a docker with a gateway running which accepts commands from non-localhost addresses.
# Echo the gateway ip for good measure
echo "Gateway addr:"
echo http://$(docker inspect -f "{{ .NetworkSettings.Networks.$network_name.IPAddress }}" r3):23110
echo "Possible address to send coins: $addr"
