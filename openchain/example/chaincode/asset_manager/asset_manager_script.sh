#!/usr/bin/env bash

OPENCHAIN_PEER_ADDRESS=127.0.0.1:30303
export OPENCHAIN_PEER_ADDRESS

USER=stan
LOG_FILE=trial.log
ADMIN=admin_user

cd $GOPATH/src/github.com/hyperledger-incubator/obc-peer

if [[ $1 != "nodeploy" ]]; then
CHAINCODE_NAME=`./obc-peer chaincode deploy -p github.com/hyperledger-incubator/obc-peer/openchain/example/chaincode/asset_manager \
	-c '{"Function":"init", "Args": ["$ADMIN", "HLP"]}'`

echo $CHAINCODE_NAME > asset_manager.cc

echo "Deployed chaincode with name: $CHAINCODE_NAME"

./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
	-c '{"Function": "admin", "Args": ["$ADMIN", "issuer_user", "110"]}'

./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
	-c '{"Function": "admin", "Args": ["$ADMIN", "user1", "010"]}'

./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
	-c '{"Function": "admin", "Args": ["$ADMIN", "user2", "010"]}'

./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
	-c '{"Function": "admin", "Args": ["$ADMIN", "other_admin_user", "011"]}'

./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
	-c '{"Function": "admin", "Args": ["$ADMIN", "auditor_user", "000"]}' 
else
	CHAINCODE_NAME=`cat asset_manager.cc`
fi

echo 'Running ISSUE command'
./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
	-c '{"Function": "issue", "Args": ["issuer_user", "10000"]}'

./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
	-c '{"Function": "transact", "Args": ["issuer_user", "user1", "1"]}'

# Negative cases
./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
	-c '{"Function": "issue", "Args": ["user1", "10000"]}'

