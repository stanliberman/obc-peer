#!/bin/bash
set -e

if [[ $# -lt 3 ]]; then
	echo "ERROR: Insufficient arguments to script"
	echo "Usage: $0 <chaincode user> <chaincode path> <security name> [<dev mode (0|1)>]"
	exit
fi

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
#	security identifier	
#	dev mode (1=YES, 0=NO)
#EXAMPLE:
#	./deploySecurityContract.sh chrisP github.com/openblockchain/obc-peer/openchain/example/chaincode/simpleFinancialSecurity/ "IBM" 1


HOST=0.0.0.0
PORT=30303

#Incoming args
chainCodeUser=$1
chainCodePath=$2
securityName=$3
devMode=$4

registryNameFile=registry.address
deployMode="-p"
deployName=$chainCodePath
pid=-1

if [[ ! -e $registryNameFile ]]; then
	echo Registry chaincode is missing... Will proceeed without
else
	registryName=`cat $registryNameFile`
fi

#Login to the local OBC peer
echo "Logging in $chainCodeUser..."
cd $GOPATH/src/github.com/hyperledger-incubator/obc-peer
./obc-peer login $chainCodeUser 

OPENCHAIN_PEER_ADDRESS=$HOST:$PORT
export OPENCHAIN_PEER_ADDRESS

directoryName=$GOPATH/src/$chainCodePath
if [[ $devMode -eq 1 ]]; then
	#Register the chaincode and put the process in the background
	#This is only really needed if running the OBC peer in dev mode
	fileExtension=".go"
	fileName=`ls $directoryName/*.go`	
	fileName=${fileName%$fileExtension}
	echo "Move to $directoryName..."
	cd $directoryName
	echo "Building GO file $fileName..."
	go build
	OPENCHAIN_CHAINCODE_ID_NAME=$securityName $fileName &
	pid=$!
	deployMode="-n"
	deployName=$securityName
fi

#Deploy the chaincode
cd $GOPATH/src/github.com/hyperledger-incubator/obc-peer
echo "Deploying contract $securityName with constructor init..."
constructorJSON="{\"Function\":\"init\", \"Args\": [\"$chainCodeUser\", \"$securityName\"]}"
echo "Constructor args: $constructorJSON"
./obc-peer chaincode deploy -u $chainCodeUser $deployMode $deployName --ctor="$constructorJSON" > $directoryName/$securityName
chainCodeAddress=$(cat $directoryName/$securityName)
rm $directoryName/$securityName

if [[ $registryName -ne "" ]]; then
	#Capture the contract hash and register it in the asset registry
	echo "Contract name: $chainCodeAddress"
	invokeConstructor='{"Function": "register", "Args": ["'$securityName'", "'$chainCodeAddress'"]}'
	./obc-peer chaincode invoke -u $chainCodeUser -n $registryName -c "$invokeConstructor"
fi

if [[ $devMode -eq 1 ]]; then
	#kill the local chain code that we started earlier if it is still running
	if [[ $pid -eq 0 ]] &&  ps -p $pid > /dev/null ;
	then
		kill $pid
	fi
fi

echo "Chaincode $securityName has been compiled, registered and deployed!"

