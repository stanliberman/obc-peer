#!/usr/bin/env bash

OPENCHAIN_PEER_ADDRESS=127.0.0.1:30303
USER=stan
LOG_FILE=trial.log
export USER OPENCHAIN_PEER_ADDRESS LOG_FILES

cd $GOPATH/src/github.com/hyperledger-incubator/obc-peer

CHAINCODE_NAME=`./obc-peer chaincode deploy -p github.com/hyperledger-incubator/obc-peer/openchain/example/chaincode/asset_manager \
	-c '{"Function":"init", "Args": ["admin_user", "HLP"]}'`

#CHAINCODE_NAME="f1c5e224a86e0296db26a210c1fcab30478cfc6df2c06add2211d12e00baceacc81d485ab9fae9ca008259942d73e684db8828702f050e7c3d055d4c8b5bb4ad"

echo "Deployed chaincode with name: $CHAINCODE_NAME"

./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
	-c '{"Function": "admin", "Args": ["admin_user", "issuer_user", "110"]}'

./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
	-c '{"Function": "admin", "Args": ["admin_user", "user1", "010"]}'

#./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
#	-c '{"Function": "admin", "Args": ["admin_user", "user2", "010"]}'

#./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
#	-c '{"Function": "admin", "Args": ["admin_user", "other_admin_user", "011"]}'

#./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
#	-c '{"Function": "admin", "Args": ["admin_user", "auditor_user", "000"]}' 

./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
	-c '{"Function": "issue", "Args": ["issuer_user", "10000"]}'

./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
	-c '{"Function": "transact", "Args": ["issuer_user", "user1", "1"]}'

#./obc-peer chaincode query -n $CHAINCODE_NAME -c '{"Function": "query", "Args": ["user1"]}'
./obc-peer chaincode query -n $CHAINCODE_NAME -c '{"Function": "query", "Args": ["admin_user"]}'
./obc-peer chaincode query -n $CHAINCODE_NAME -c '{"Function": "query", "Args": ["issuer_user"]}'


# ./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
#	-c '{"Function": "inituser", "Args": ["admin_user", "issuer_user", "Will", "Smith", "110"]}' 2>&1 >$LOG_FILE

#./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
#	-c '{"Function": "inituser", "Args": ["admin_user", "transactor_user", "Joe", "Smith", "010"]}' 2>&1 >$LOG_FILE

#./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
#	-c '{"Function": "inituser", "Args": ["admin_user", "other_admin_user", "Jane", "Doe", "011"]}' 2>&1 >$LOG_FILE

#./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
#	-c '{"Function": "inituser", "Args": ["admin_user", "auditor_user", "Michael", "Moore", "000"]}' 2>&1 >$LOG_FILE


# This must fail
#./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
#	-c '{"Function": "inituser", "Args": ["transactor_user", "auditor_user", "James", "Hook", "001"]}' 2>&1 >$LOG_FILE


#./obc-peer chaincode invoke -u $USER -n $CHAINCODE_NAME \
#	-c '{"Function": "issue", "Args": ["issuer_user", "Asset Name", "100"]}' 2>&1 >$LOG_FILE
