#!/bin/bash
set -e

echo "--> Deploy asset contract... "

#prereqs/assumptions
#	OBC dev environment setup
#	locally running OBC peer
#	running within vagrant
#	registered user
#	Put registry address in file in the local directory called registry.address
#input arguments
#	username
#	directory containing chaincode source
#	chain code ID 
#	constructor name
#	security identifier	
#EXAMPLE:
#	./deploySecurityContract.sh chrisP myChainCodeID github.com/openblockchain/obc-peer/openchain/example/chaincode/simpleFinancialSecurity/commonstock.go init "IBM"


HOST=0.0.0.0
PORT=30303

chainCodeUser=$1
chainCodeID=$2
chainCodePath=$3
constructorName=$4
securityName=$5
pid=-1

registryName=`cat registry.address`

#Login to the local OBC peer
echo "Logging in $chainCodeUser..."
cd $GOPATH/src/github.com/hyperledger-incubator/obc-peer
./obc-peer login $chainCodeUser 

#Register the chaincode and put the process in the background
#This is only really needed if running the OBC peer in dev mode
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
cd $GOPATH/src/github.com/hyperledger-incubator/obc-peer
echo "Deploying contract $chainCodeID with constructor $constructorName..."
constructorJSON+='{"Function":"'
constructorJSON+=$constructorName
constructorJSON+='", "Args": ["'
constructorJSON+=$securityName
constructorJSON+='"]}'
echo "Constructor args: $constructorJSON"
prefix=$GOPATH/src
deployPath=${directoryName#*"$prefix"}
./obc-peer chaincode deploy -u $chainCodeUser -p $deployPath --ctor="$constructorJSON" > $directoryName/$chainCodeID
chainCodeAddress=$(cat $directoryName/$chainCodeID)
rm $directoryName/$chainCodeID

#Capture the contract hash and register it in the asset registry
echo "Contract name: $chainCodeAddress"
invokeConstructor='{"Function": "register", "Args": ["'$securityName'", "'$chainCodeAddress'"]}'
./obc-peer chaincode invoke -u $chainCodeUser -n $registryName -c "$invokeConstructor"

#kill the local chain code that we started earlier if it is still running
if [[ $pid > 0 ]] &&  ps -p $pid > /dev/null;
then
	kill $pid
fi

echo "Chaincode $chainCodeID has been compiled, registered and deployed!"

