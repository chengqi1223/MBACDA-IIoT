package main

import (
	"fmt"
	"encoding/json"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type DomainInfo struct {
	Did string `json:"did"`
	Gid []string `json:"gid"`
	PK string `json:"pk"`
	Y1 string `json:"y_1"`
	Y2 string `json:"y_2"`
	Acc string `json:"acc"`
	Wit string `json:"wit"`
	Id  map[string]struct{} `json:"id"`
	Pid map[string]map[string]struct{} `json:"pid"`
}

func (s *SmartContract) DomainRegister(ctx contractapi.TransactionContextInterface, did string, gidJSON string, pk string, y1 string, y2 string) error {
	existingDomainJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return fmt.Errorf("error occurred while checking domain existence: %v", err)
	}
	if existingDomainJSON != nil {
		return fmt.Errorf("the domain with ID '%s' already exists", did)
	}
	var gid []string
	if err := json.Unmarshal([]byte(gidJSON), &gid); err != nil {
		return fmt.Errorf("failed to unmarshal gid JSON: %v", err)
	}
	domainInfo := DomainInfo{
		Did: did,
		Gid: gid,
		PK: pk,
		Y1: y1,
		Y2: y2,
		Id: make(map[string]struct{}),
		Pid: make(map[string]map[string]struct{}),
	}
	domainInfoJSON, err := json.Marshal(domainInfo)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(did, domainInfoJSON)
}

func (s *SmartContract) DomainPreAuth(ctx contractapi.TransactionContextInterface, did1 string, did2 string, acc string, wit string) (string, error) {
	existingDomainJSON, err := ctx.GetStub().GetState(did1)
	if err != nil {
		return "", fmt.Errorf("error occurred while checking domain existence: %v", err)
	}
	if existingDomainJSON == nil {
		return "", fmt.Errorf("the domain with ID '%s' does not exist", did1)
	}
	var domainInfo DomainInfo
	err = json.Unmarshal(existingDomainJSON, &domainInfo)
	if err != nil {
		return "", err
	}
	domainInfo.Acc = acc
	domainInfo.Wit = wit
	updatedDomainJSON, err := json.Marshal(domainInfo)
	if err != nil {
		return "", err
	}
	err = ctx.GetStub().PutState(did1, updatedDomainJSON)
	if err != nil {
		return "", fmt.Errorf("failed to update domain: %v", err)
	}

	existingDomainJSON2, err := ctx.GetStub().GetState(did2)
	if err != nil {
		return "", fmt.Errorf("error occurred while checking domain existence: %v", err)
	}
	if existingDomainJSON2 == nil {
		return "", fmt.Errorf("the domain with ID '%s' does not exist", did2)
	}
	err = json.Unmarshal(existingDomainJSON2, &domainInfo)
	if err != nil {
		return "", err
	}
	return domainInfo.PK, nil
}

func (c *DomainInfo) AddId(id string) bool {
	if _, exists := c.Id[id]; exists {
		return false
	}
	c.Id[id] = struct{}{}
	return true
}

func (c *DomainInfo) AddPid(id, pid string) bool {
	if _, exists := c.Pid[id][pid]; exists {
		return false
	}

	if c.Pid[id] == nil {
		c.Pid[id] = make(map[string]struct{})
	}

	c.Pid[id][pid] = struct{}{}
	return true
}

func (c *DomainInfo) IsPidRevoked(pid string) bool {
	for _, pidsForId := range c.Pid {
		if _, exists := pidsForId[pid]; exists {
			return true
		}
	}
	return false
}

// Contains checks if a string is present in a slice
func Contains(slice []string, str string) bool {
    for _, v := range slice {
        if v == str {
            return true
        }
    }
    return false
}

func (s *SmartContract) DomainAuth(ctx contractapi.TransactionContextInterface, did string, gid string, pid string) (string, error) {
	existingDomainJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return "", fmt.Errorf("error occurred while checking domain existence: %v", err)
	}
	if existingDomainJSON == nil {
		return "", fmt.Errorf("the domain with ID '%s' does not exist", did)
	}
	var domainInfo DomainInfo
	err = json.Unmarshal(existingDomainJSON, &domainInfo)
	if err != nil {
		return "", err
	}
	if !Contains(domainInfo.Gid, gid) {
		return "", fmt.Errorf("the gateway '%s' does not exist", gid)
	}
	if domainInfo.IsPidRevoked(pid) {
		return "", fmt.Errorf("the pid '%s' has been revoked", pid)
	}

	result := fmt.Sprintf("%s/%s", domainInfo.Acc, domainInfo.Wit)
	return result, nil
}

func (s *SmartContract) DomainRevoke(ctx contractapi.TransactionContextInterface, did string, id string, pidJSON string, N int) error {
	var pid []string
	if err := json.Unmarshal([]byte(pidJSON), &pid); err != nil {
		return fmt.Errorf("failed to unmarshal pid JSON: %v", err)
	}
	existingDomainJSON, err := ctx.GetStub().GetState(did)
	if err != nil {
		return fmt.Errorf("error occurred while checking domain existence: %v", err)
	}
	if existingDomainJSON == nil {
		return fmt.Errorf("the domain with ID '%s' does not exist", did)
	}
	var domainInfo DomainInfo
	err = json.Unmarshal(existingDomainJSON, &domainInfo)
	if err != nil {
		return err
	}
	if !domainInfo.AddId(id) {
		return fmt.Errorf("the device with ID '%s' already exists", id)
	}
	if len(domainInfo.Pid[id]) >= N {
		return fmt.Errorf("the device with ID '%s' has enough pid", id)
	}
	for _, pid := range pid {
		domainInfo.AddPid(id, pid)
	}
	updatedDomainJSON, err := json.Marshal(domainInfo)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(did, updatedDomainJSON)
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
