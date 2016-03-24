/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger-incubator/obc-peer/openchain/chaincode/shim"
)

// This is a chaincode that will hold asset registration information for all assets on a blockchain
// It is a convenience mechanism so that financial market participants will only need to recall the
// chaincode identifier for the registry. The registry can be queried to retrieve the addresses of 
// chaincodes used to represent a particular asset. For instance, the key "IBM" could be stored in 
// the registry and point to the chaincode that represent IBM Corp common stock. Effectively, this
// chaincode is just a fancy map for storing asset chaincode addresses

// Invoke operations
// 	register - Takes two paramaeters:
//		1) An appropriate asset identitifer (ticker, CUSIP, etc.) 
//		2) The address of the chaincode used to represent the asset
// 	deactivate - Takes a key that was previously register and removes it from the registry

// Query operations
// 	get - requires one argument, an asset identifier and returns the address of the chaincode for this asset
// 	keys - requires no arguments, returns all asset keys

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// Run callback representing the invocation of a chaincode
func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	switch function {

	case "init":
		// Do nothing

	case "register":
		if len(args) < 2 {
			return nil, errors.New("Please supply an asset identifier and a chaincode address for the asset")
		}
		key := args[0]
		value := args[1]

		existingKey, err := stub.GetState(key)
		if existingKey != nil {
			return nil, fmt.Errorf("This asset is already registered")
		} 

		err := stub.PutState(key, []byte(value))
		if err != nil {
			fmt.Printf("Error putting state %s", err)
			return nil, fmt.Errorf("put operation failed. Error updating state: %s", err)
		}
		return nil, nil

	case "deactivate":
		if len(args) < 1 {
			return nil, errors.New("Please supply an asset identifieir")
		}
		key := args[0]

		err := stub.DelState(key)
		if err != nil {
			return nil, fmt.Errorf("remove operation failed. Error updating state: %s", err)
		}
		return nil, nil

	default:
		return nil, errors.New("Unsupported operation")
	}

	return nil, nil
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	switch function {

	case "getAsset":
		if len(args) < 1 {
			return nil, errors.New("getAsset operation must include one argument, a key")
		}
		key := args[0]
		value, err := stub.GetState(key)
		if err != nil {
			return nil, fmt.Errorf("get operation failed. Error accessing state: %s", err)
		}
		return value, nil

	case "getAllAssetKeys":

		keysIter, err := stub.RangeQueryState("", "")
		if err != nil {
			return nil, fmt.Errorf("getAllAssetKeys operation failed. Error accessing state: %s", err)
		}
		defer keysIter.Close()

		var keys []string
		for keysIter.HasNext() {
			key, _, err := keysIter.Next()
			if err != nil {
				return nil, fmt.Errorf("getAllAssetKeys operation failed. Error accessing state: %s", err)
			}
			keys = append(keys, key)
		}

		jsonKeys, err := json.Marshal(keys)
		if err != nil {
			return nil, fmt.Errorf("keys operation failed. Error marshaling JSON: %s", err)
		}

		return jsonKeys, nil

	default:
		return nil, errors.New("Unsupported operation")
	}
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
