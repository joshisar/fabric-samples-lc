/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"sample.com/lc/chaincode"
)

func main() {
	locChaincode, err := contractapi.NewChaincode(&chaincode.LocContract{})
	if err != nil {
		log.Panicf("Error creating loc chaincode: %v", err)
	}
	if err := locChaincode.Start(); err != nil {
		log.Panicf("Error starting loc chaincode: %v", err)
	}
}
