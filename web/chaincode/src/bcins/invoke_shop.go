package main

import (
	"encoding/json"

	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func createContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Invalid argument count.")
	}

	dto := struct {
		UUID             string    `json:"uuid"`
		ContractTypeUUID string    `json:"contract_type_uuid"`
		Username         string    `json:"username"`
		Password         string    `json:"password"`
		FirstName        string    `json:"first_name"`
		LastName         string    `json:"last_name"`
		Item             item      `json:"item"`
		StartDate        time.Time `json:"start_date"`
		EndDate          time.Time `json:"end_date"`
		// why does it not work when using `json:"specialRate"` ?!
		SpecialRate float32 `json:"special_rate"`
	}{}

	err := json.Unmarshal([]byte(args[0]), &dto)
	if err != nil {
		return shim.Error(err.Error())
	}

	logger.Infof("createContract, dto: %+v\n", dto)

	// Create new user if necessary
	var u user
	requestUserCreate := len(dto.Username) > 0 && len(dto.Password) > 0
	userKey, err := stub.CreateCompositeKey(prefixUser, []string{dto.Username})
	if requestUserCreate {
		// Check if a user with the same username exists
		if err != nil {
			return shim.Error(err.Error())
		}
		userAsBytes, _ := stub.GetState(userKey)
		if userAsBytes == nil {
			// Create new user
			u = user{
				Username:  dto.Username,
				Password:  dto.Password,
				FirstName: dto.FirstName,
				LastName:  dto.LastName,
			}
			// Persist the new user
			userAsBytes, err := json.Marshal(u)
			if err != nil {
				return shim.Error(err.Error())
			}
			err = stub.PutState(userKey, userAsBytes)
			if err != nil {
				return shim.Error(err.Error())
			}
		} else {
			// Parse the existing user
			err := json.Unmarshal(userAsBytes, &u)
			if err != nil {
				return shim.Error(err.Error())
			}
		}
	} else {
		// Validate if the user with the provided username exists
		userAsBytes, _ := stub.GetState(userKey)
		if userAsBytes == nil {
			return shim.Error("User with this username does not exist.")
		}
	}

	contract := contract{
		Username:         dto.Username,
		ContractTypeUUID: dto.ContractTypeUUID,
		Item:             dto.Item,
		StartDate:        dto.StartDate,
		EndDate:          dto.EndDate,
		SpecialRate:      dto.SpecialRate,
		Void:             false,
		ClaimIndex:       []string{},
	}

	logger.Infof("createContract, contract: %+v\n", contract)

	contractKey, err := stub.CreateCompositeKey(prefixContract, []string{dto.Username, dto.UUID})
	if err != nil {
		return shim.Error(err.Error())
	}
	contractAsBytes, err := json.Marshal(contract)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(contractKey, contractAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Return success, if the new user has been created
	// (the user variable "u" should be blank)
	if !requestUserCreate {
		return shim.Success(nil)
	}

	response := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: u.Username,
		Password: u.Password,
	}
	responseAsBytes, err := json.Marshal(response)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(responseAsBytes)
}

func createUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Invalid argument count.")
	}

	user := user{}
	err := json.Unmarshal([]byte(args[0]), &user)
	if err != nil {
		return shim.Error(err.Error())
	}

	key, err := stub.CreateCompositeKey(prefixUser, []string{user.Username})
	if err != nil {
		return shim.Error(err.Error())
	}

	// Check if the user already exists
	userAsBytes, _ := stub.GetState(key)
	// User does not exist, attempting creation
	if len(userAsBytes) == 0 {
		userAsBytes, err = json.Marshal(user)
		if err != nil {
			return shim.Error(err.Error())
		}

		err = stub.PutState(key, userAsBytes)
		if err != nil {
			return shim.Error(err.Error())
		}

		// Return nil, if user is newly created
		return shim.Success(nil)
	}

	err = json.Unmarshal(userAsBytes, &user)
	if err != nil {
		return shim.Error(err.Error())
	}

	userResponse := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: user.Username,
		Password: user.Password,
	}

	userResponseAsBytes, err := json.Marshal(userResponse)
	if err != nil {
		return shim.Error(err.Error())
	}
	// Return the username and the password of the already existing user
	return shim.Success(userResponseAsBytes)
}

func listShopContracts(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	input := struct {
		Username string `json:"username"`
	}{}
	if len(args) == 1 {
		err := json.Unmarshal([]byte(args[0]), &input)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	var resultsIterator shim.StateQueryIteratorInterface
	var err error
	resultsIterator, err = stub.GetStateByPartialCompositeKey(prefixContract, []string{input.Username})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	results := []interface{}{}
	// Iterate over the results
	for resultsIterator.HasNext() {
		kvResult, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// Construct response struct
		result := struct {
			UUID string `json:"uuid"`
			*contract
			Claims []claim `json:"claims,omitempty"`
		}{}

		err = json.Unmarshal(kvResult.Value, &result)
		if err != nil {
			return shim.Error(err.Error())
		}

		// Fetch key
		prefix, keyParts, err := stub.SplitCompositeKey(kvResult.Key)
		if len(keyParts) == 2 {
			result.UUID = keyParts[1]
		} else {
			result.UUID = prefix
		}

		results = append(results, result)
	}

	resultsAsBytes, err := json.Marshal(results)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(resultsAsBytes)
}
