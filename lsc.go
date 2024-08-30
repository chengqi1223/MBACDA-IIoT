package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a device
type SmartContract struct {
	contractapi.Contract
}

// DeviceInfo describes basic details of a device
type DeviceInfo struct {
	ID   string `json:"id"`
	PK   string `json:"pk"`
	Root string `json:"root"`
	N    int    `json:"n"`
}

// DeviceRegi registers a new device in the world state
func (s *SmartContract) DeviceRegi(ctx contractapi.TransactionContextInterface, id string, pk string, root string, n int) error {
	// Check if the Device is already registered
	existingDeviceJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("error occurred while checking asset existence: %v", err)
	}
	if existingDeviceJSON != nil {
		return fmt.Errorf("the asset with ID '%s' already exists", id)
	}

	// Proceed to store the new Device info as it does not exist yet
	deviceInfo := DeviceInfo{
		ID:   id,
		PK:   pk,
		Root: root,
		N:    n,
	}
	deviceInfoJSON, err := json.Marshal(deviceInfo)
	if err != nil {
		return fmt.Errorf("error marshalling device info: %v", err)
	}

	return ctx.GetStub().PutState(id, deviceInfoJSON)
}

// DeviceSearch retrieves a device info by ID from the world state
func (s *SmartContract) DeviceSearch(ctx contractapi.TransactionContextInterface, id string) (string, error) {
	deviceInfoJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}
	if deviceInfoJSON == nil {
		return "", fmt.Errorf("the asset %s does not exist", id)
	}

	var deviceInfo DeviceInfo
	err = json.Unmarshal(deviceInfoJSON, &deviceInfo)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling device info: %v", err)
	}

	result := fmt.Sprintf("%s/%s/%d", deviceInfo.PK, deviceInfo.Root, deviceInfo.N)
	return result, nil
}

// DeviceRevoke deletes a device from the world state by ID
func (s *SmartContract) DeviceRevoke(ctx contractapi.TransactionContextInterface, id string) error {
	// Check if the device actually exists before attempting revoke
	existingDeviceJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("error occurred while checking asset existence: %v", err)
	}
	if existingDeviceJSON == nil {
		return fmt.Errorf("the asset with ID '%s' does not exist", id)
	}

	// Delete the device from state
	err = ctx.GetStub().DelState(id)
	if err != nil {
		return fmt.Errorf("failed to delete asset '%s': %v", id, err)
	}

	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error creating asset chaincode: %s", err.Error())
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting asset chaincode: %s", err.Error())
	}
}
