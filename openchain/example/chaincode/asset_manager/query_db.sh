#!/usr/bin/env bash

OPENCHAIN_PEER_ADDRESS=127.0.0.1:30303
USER=stan
export USER OPENCHAIN_PEER_ADDRESS

CCN=`cat $1`

cd $GOPATH/src/github.com/hyperledger-incubator/obc-peer

if [[ $# != 1 ]]; then
	echo "Usage: $0 <chaincode name file>"
	exit
fi

echo "Current chaincode is $CCN"

./obc-peer chaincode query -n $CCN -c '{"Function": "permissions", "Args": ["stan"]}'
./obc-peer chaincode query -n $CCN -c '{"Function": "permissions", "Args": ["user1"]}'
./obc-peer chaincode query -n $CCN -c '{"Function": "permissions", "Args": ["admin_user"]}'
./obc-peer chaincode query -n $CCN -c '{"Function": "permissions", "Args": ["issuer_user"]}'
./obc-peer chaincode query -n $CCN -c '{"Function": "balance", "Args": ["user1"]}'
./obc-peer chaincode query -n $CCN -c '{"Function": "balance", "Args": ["issuer_user"]}'

