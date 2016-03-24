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
	"errors"
	"fmt"
	"strconv"

	"github.com/op/go-logging"

	"github.com/hyperledger-incubator/obc-peer/openchain/chaincode/shim"
)

const (
	ADMIN_PERMISSION_BIT = 1
	TRANSACT_PERMISSION_BIT = 2
	DESTROY_PERMISSION_BIT = 4
	ISSUE_PERMISSION_BIT = 8
)

var log = logging.MustGetLogger("asset_manager")


// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

/*
Args:
	- initial admin user
	- asset ID
*/
func (t *SimpleChaincode) init(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error

	log.Info("Initializing")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	initialAdmin := args[0]

	// TODO check this user doesn't yet exist or has different permissions
	err = stub.PutState(initialAdmin, []byte(strconv.Itoa(ADMIN_PERMISSION_BIT | TRANSACT_PERMISSION_BIT)))
	if err != nil {
		return nil, err
	}

	err = stub.PutState("ASSET_ID", []byte(args[1]))

	return nil, err
}

// Helper function to check the pemissions against the database
func (t *SimpleChaincode) checkPermission(stub *shim.ChaincodeStub, user string, permissionBit int) (bool, error) {
		
	permBytes, err := stub.GetState(user)
	if err != nil {
		return false, errors.New("Failed to get permissions for " + user)
	}

	if permBytes == nil {
		return false, errors.New("Nil permissions for " + user)
	}

	permissionsMask,_ := strconv.Atoi(string(permBytes))

	return (1 == permissionsMask & permissionBit), nil
}

/*
Admin function will only currently act to add a new user permissions. This should be invoked on logon. Maybe a better approach is to have dedicated logon/logoff functions.

 Args:
	- currentUser
	- user to add
	- permission mask
*/
func (t *SimpleChaincode) admin(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	currentUser := args[0]
	allowed,_ := t.checkPermission(stub, currentUser, ADMIN_PERMISSION_BIT)
	if !allowed {
		return nil, errors.New("User " + currentUser + " is not permissioned for ADMIN action")
	} 

	log.Info("Admin stuff permitted")

	permissions,err := strconv.ParseInt(args[2], 2, 10)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, errors.New("Failed to parse permissions mask: " + args[2])
	}

	user := args[1]
	log.Info(fmt.Sprintf("===> adding user:%s with permission mask of %d(%s)\n", user, permissions, args[2]))
	err = stub.PutState(user, []byte(strconv.Itoa(int(permissions))))

	return nil, err
}

/*
Args:
	- currentUser
	- asset quantity
*/
func (t *SimpleChaincode) issue(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	currentUser := args[0]
	allowed,err := t.checkPermission(stub, currentUser, ISSUE_PERMISSION_BIT)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, errors.New("User " + currentUser + " is not permissioned for ISSUE action")
	} 

	log.Info("Issuing assets permitted")

	assetQty,_ := strconv.ParseInt(args[1], 10, 10)

	log.Info(fmt.Sprintf("===> issuing %d to %s\n", assetQty, currentUser))
	err = stub.PutState(currentUser, []byte(strconv.Itoa(int(assetQty))))

	return nil, err
}

/*
Args:
	- currentUser
	- to user
	- asset quantity to transfer
*/
func (t *SimpleChaincode) transact(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	currentUser := args[0]
	allowed,err := t.checkPermission(stub, currentUser, TRANSACT_PERMISSION_BIT)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, errors.New("User " + currentUser + " is not permissioned for TRANSACT action")
	} 

	log.Info("Transacting permitted")

	return nil, nil
}

/*

Only can destroy own assets

Args:
	- currentUser
	- asset quantity
*/
func (t *SimpleChaincode) destroy(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	log.Info("Destoying stuff")

	return nil, nil
}

// Run callback representing the invocation of a chaincode
// This chaincode will manage two accounts A and B and will transfer X units from A to B upon invoke
func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	log.Info("Received command: " + function)

	// Handle different functions
	if function == "init" {
		return t.init(stub, args)
	} else if function == "admin" {
		return t.admin(stub, args)
	} else if function == "issue" {
		return t.issue(stub, args)
	} else if function == "destroy" {
		return t.destroy(stub, args)
	} else if function == "transact" {
		return t.transact(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if function != "query" {
		return nil, errors.New("Invalid query function name. Expecting \"query\"")
	}
	var A string // Entities
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return nil, errors.New(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return nil, errors.New(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return Avalbytes, nil
}

func main() {

	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
