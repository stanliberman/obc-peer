#!/usr/bin/env bash

OPENCHAIN_PEER_ADDRESS=127.0.0.1:30303
USER=stan
export USER OPENCHAIN_PEER_ADDRESS

cd $GOPATH/src/github.com/hyperledger-incubator/obc-peer

CCN=`cat asset_manager.cc`

echo "Current chaincode is $CCN"

./obc-peer chaincode query -n $CCN -c '{"Function": "query", "Args": ["XX:user1"]}'
./obc-peer chaincode query -n $CCN -c '{"Function": "query", "Args": ["XX:admin_user"]}'
./obc-peer chaincode query -n $CCN -c '{"Function": "query", "Args": ["XX:issuer_user"]}'
./obc-peer chaincode query -n $CCN -c '{"Function": "query", "Args": ["user1"]}'
./obc-peer chaincode query -n $CCN -c '{"Function": "query", "Args": ["issuer_user"]}'

