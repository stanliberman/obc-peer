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
	ISSUE_PERMISSION_BIT = 4
)

var log = logging.MustGetLogger("asset_manager")


/*
	AssetManagerChaincode example.
*/
type AssetManagerChaincode struct {
}

/*
This function should additionally have a guard against multiple invocations, to prevent run-time addition
of admin users.

Args:
	- initial admin user
	- asset ID
*/
func (t *AssetManagerChaincode) init(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error

	log.Info("Initializing")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	initialAdmin := "XX:" + args[0]

	// TODO check this user doesn't yet exist or has different permissions
	err = stub.PutState(initialAdmin, []byte(strconv.Itoa(ADMIN_PERMISSION_BIT | TRANSACT_PERMISSION_BIT)))
	if err != nil {
		return nil, err
	}

	err = stub.PutState("ASSET_ID", []byte(args[1]))

	return nil, err
}

func (t *AssetManagerChaincode) getPermissions(stub *shim.ChaincodeStub, user string) (int, error) {
	permBytes, err := stub.GetState("XX:" + user)
	if err != nil {
		return 0, errors.New("Failed to get permissions for " + user)
	}

	if permBytes == nil {
		return 0, errors.New("Nil permissions for " + user)
	}

	var permissionsMask int
	permissionsMask,err = strconv.Atoi(string(permBytes))
	log.Info("Retrieved permission mask %d", permissionsMask)

	if err != nil {
		return 0, errors.New("Failed to interpret permissions mask: " + strconv.Itoa(permissionsMask))
	}

	return permissionsMask, nil
}

// Helper function to check the pemissions against the database
func (t *AssetManagerChaincode) checkPermission(stub *shim.ChaincodeStub, user string, permissionBit int) (bool, error) {
		
	permissionsMask, err := t.getPermissions(stub, user)
	log.Info("Retreived permission mask %d. Checking against %d", permissionsMask, permissionBit)

	return (permissionBit == permissionsMask & permissionBit), err
}

/*
Admin function will only currently act to add a new user permissions. This should be invoked on logon. Maybe a better approach is to have dedicated logon/logoff functions.

 Args:
	- currentUser
	- user to add
	- permission mask
*/
func (t *AssetManagerChaincode) admin(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
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
	prefixedUser := "XX:" + user

	permBytes, err := stub.GetState(prefixedUser)
	if permBytes != nil {
		log.Warning("User " + prefixedUser + " already exists");
		return nil, err
	}

	log.Info(fmt.Sprintf("===> adding user:%s with permission mask of %d(%s)\n", user, permissions, args[2]))
	permString := strconv.Itoa(int(permissions))
	err = stub.PutState(prefixedUser, []byte(permString))
	if err != nil {
		log.Error("Failed to store " + permString + " for " + prefixedUser)
		return nil, err
	}

	// Establish initial 0 balance
	err = stub.PutState(user, []byte(strconv.Itoa(0)))

	return nil, err
}

/*
Args:
	- currentUser
	- asset quantity
*/
func (t *AssetManagerChaincode) issue(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
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

	assetQty,err := strconv.Atoi(args[1])
	if err != nil {
		log.Error("Failed to parse issue amount: " + args[1])
		return nil, err
	}

	// Get current balance
	balBytes,err := stub.GetState(currentUser)
	if err != nil {
		log.Error("Failed to get current balance for " + currentUser)
		return nil, err	
	}
	balance, err := strconv.Atoi(string(balBytes))
	log.Info("Current balance for " + currentUser + " is " + string(balBytes))

	log.Info(fmt.Sprintf("===> issuing %d to %s\n", assetQty, currentUser))
	newAmt := strconv.Itoa(int(assetQty) + balance)
	err = stub.PutState(currentUser, []byte(newAmt))
	if err != nil {
		log.Error("Failed to store new balance of " + newAmt)
	}

	return nil, err
}

/*
Args:
	- currentUser
	- to user
	- asset quantity to transfer
*/
func (t *AssetManagerChaincode) transact(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
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

	log.Debug("Transacting permitted")

	// Check whether the recepient has TRANSACT permissions
	user := args[1]
	quantity,err := strconv.Atoi(args[2])
	allowed,err = t.checkPermission(stub, user, TRANSACT_PERMISSION_BIT)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, errors.New("Recepient user " + user + " is not permissioned for TRANSACT action")
	} 

	// Check balance
	balBytes, err := stub.GetState(currentUser)
	if err != nil {
		return nil, err	
	}
	balance, err := strconv.Atoi(string(balBytes))

	if balance < quantity {
		return nil, errors.New("Insufficient balance for " + currentUser +
			". Available: " + string(balBytes) + "; requested: " + args[2])	
	}

	// Subtract from currentUser balance
	err = stub.PutState(currentUser, []byte(strconv.Itoa(balance - quantity)))

	// Get recepient balance
	recvBalBytes, err := stub.GetState(user)
	if err != nil {
		return nil, errors.New("Failed to get balance for " + user)
	}
	oldBalance,_ := strconv.Atoi(string(recvBalBytes))

	// Store new balance
	err = stub.PutState(user, []byte(strconv.Itoa(oldBalance + quantity)))

	return nil, err
}

/*

Only can destroy own assets

Args:
	- currentUser
	- asset quantity
*/
func (t *AssetManagerChaincode) destroy(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
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

	log.Info("Destroying permitted")

	// Check balance
	balBytes, err := stub.GetState(currentUser)
	if err != nil {
		return nil, err	
	}
	balance, err := strconv.Atoi(string(balBytes))

	quantity,err := strconv.Atoi(args[1])
	if balance < quantity {
		return nil, errors.New("Insufficient balance for " + currentUser +
			". Available: " + string(balBytes) + "; requested: " + args[2])	
	}

	log.Info("Destroying " + args[1] + " of " + string(balBytes) + " for " + currentUser)

	err = stub.PutState(currentUser, []byte(strconv.Itoa(balance - quantity)))

	return nil, err
}

// Run callback representing the invocation of a chaincode
// This chaincode will manage two accounts A and B and will transfer X units from A to B upon invoke
func (t *AssetManagerChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

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
func (t *AssetManagerChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting user name to query")
	}

	user := args[0]

	if function == "permissions" {
		permissionsMask, err := t.getPermissions(stub, user)
		fmt.Printf("Query Response: user: %s, permissions: %d\n", user, permissionsMask)
		return []byte(strconv.Itoa(permissionsMask)), err
	} else if function == "balance" {
		var err error

		// Get the state from the ledger
		balbytes, err := stub.GetState(user)
		if err != nil {
			return nil, errors.New("Failed to get state for " + user)
		}

		if balbytes == nil {
			return nil, errors.New("Nil amount for " + user)
		}

		bytes,err := stub.GetState("ASSET_ID")
		fmt.Printf("Asset ID: %d\n", string(bytes))
		fmt.Printf("Query Response: user: %s, permissions: %d\n", user, balbytes)
		return balbytes, nil
	} else {
		return nil, errors.New("Invalid query function name: " + function)
	}

	return nil, nil
}

func main() {

	err := shim.Start(new(AssetManagerChaincode))
	if err != nil {
		fmt.Printf("Error starting AssetManager chaincode: %s", err)
	}
}
