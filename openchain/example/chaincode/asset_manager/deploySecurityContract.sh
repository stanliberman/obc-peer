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


HOST=127.0.0.1
PORT=30303

#Incoming args
chainCodeUser=$1
chainCodePath=$2
securityName=$3
devMode=$4

registryNameFile=registry.address

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
	pidFile=$directoryName/$securityName.pid
	if [[ -e $pidFile ]]; then
		oldPid=`cat $pidFile`
		echo "Existing process running chaincode found at $oldPid. Killing..."
		if kill -0 $oldPid 2>/dev/null; then
			kill $oldPid 2>&1 >/dev/null
		else
			echo "$oldPid is not running"
		fi
	fi
	OPENCHAIN_CHAINCODE_ID_NAME=$securityName nohup $fileName >$directoryName/$securityName.log 2>&1 &
	pid=$!
	echo $pid > $directoryName/$securityName.pid
	echo "Started chaincode in dev mode with PID $pid"
	deployMode="-n $securityName"
else
	deployMode="-p $chainCodePath"
fi

#Deploy the chaincode
cd $GOPATH/src/github.com/hyperledger-incubator/obc-peer
echo "Deploying contract $securityName with constructor init..."

constructorJSON="{\"Function\":\"init\", \"Args\": [\"$chainCodeUser\", \"$securityName\"]}"
echo "Constructor args: $constructorJSON"

chainCodeAddress=`./obc-peer chaincode deploy -u $chainCodeUser $deployMode --ctor="$constructorJSON"`
echo "Deployed chaincode with name: $chainCodeAddress"

if [[ $registryName -ne "" ]]; then
	#Capture the contract hash and register it in the asset registry
	echo "Registering Contract name $securityName as $chainCodeAddress"
	invokeConstructor='{"Function": "register", "Args": ["'$securityName'", "'$chainCodeAddress'"]}'
	./obc-peer chaincode invoke -u $chainCodeUser -n $registryName -c "$invokeConstructor"
else
	# In the absence of the registry store the generated chaincode name into a local file
	echo $chainCodeAddress > $directoryName/$securityName
fi

echo "Chaincode $securityName has been compiled, registered and deployed!"
