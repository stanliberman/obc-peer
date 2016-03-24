#!/bin/bash
set -e

echo "--> Deploy asset contract... "

#prereqs/assumptions
#	OBC dev environment setup
#	locally running OBC peer
#	running within vagrant
#	registered user
#input arguments
#	username
#	directory containing chaincode source
#	chain code ID 
#	constructor name
#	constructor arugements
#		format: '["a":"aVal","b":"bVal"]'
#EXAMPLE:
#	./deploySecurityContract.sh chrisP myChainCodeID github.com/openblockchain/obc-peer/openchain/example/chaincode/simpleFinancialSecurity/commonstock.go init '["IBM", "100"]'


HOST=0.0.0.0
PORT=30303

chainCodeUser=$1
chainCodeID=$2
chainCodePath=$3
constructorName=$4
constructorArgs=$5

#Login to the local OBC peer
echo "Logging in $chainCodeUser..."
cd $GOPATH/src/github.com/openblockchain/obc-peer
./obc-peer login $chainCodeUser 

#Register the chaincode
fileName="$(basename $chainCodePath)" 
directoryName=$GOPATH/src/$(dirname $chainCodePath)
compiled=$(echo $fileName | cut -f 1 -d '.')
echo "Move to $directoryName..."
cd $directoryName
echo "Building GO file $fileName..."
go build
OPENCHAIN_CHAINCODE_ID_NAME=$chainCodeID OPENCHAIN_PEER_ADDRESS=$HOST:$PORT $directoryName/$compiled &
pid=$!

#Deploy the chaincode
cd $GOPATH/src/github.com/openblockchain/obc-peer
echo "Deploying contract $chainCodeID with constructor $constructorName..."
constructorJSON+='{"Function":"'
constructorJSON+=$constructorName
constructorJSON+='", "Args": '
constructorJSON+=$constructorArgs
constructorJSON+='}'
echo "Constructor args: $constructorJSON"
prefix=$GOPATH/src
deployPath=${directoryName#*"$prefix"}
./obc-peer chaincode deploy -u $chainCodeUser -p $deployPath --ctor="$constructorJSON"

kill $pid
echo "Chaincode $chainCodeID has been compiled, registered and deployed!"

