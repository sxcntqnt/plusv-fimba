package dash

import (
	"encoding/json"
	"vermi/models"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// This function retrieves the metadata of all the ledger entries for the device at reference //TESTED
func (sc *SmartContract) QueryDeviceMetadata(ctx contractapi.TransactionContextInterface, id string) (*models.Auxiliary, error) {

	var metaData []models.Metadata

	// Query the deviceID~fromToTimestamp index by deviceID
	// This will execute a key range query on all keys starting with the identifier of the device
	deviceResultIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("deviceID~fromToTimestamp", []string{id})
	if err != nil {
		return nil, err
	}

	// Iterate through result set and for each device found
	var i int
	for i = 0; deviceResultIterator.HasNext(); i++ {

		// Get the ledger entry
		deviceEntry, err := deviceResultIterator.Next()
		if err != nil {
			return nil, err
		}

		// split key, and ensure the "0_0" (i.e. init) ledger entry is not processed
		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(deviceEntry.Key)
		if err != nil {
			return nil, err
		}

		//returnedDeviceName := compositeKeyParts[0]
		returnedfromToTimestamp := compositeKeyParts[1]
		//fmt.Printf("deviceID= %s, fromToTimestamp= %s\n", returnedDeviceName, returnedfromToTimestamp)

		if returnedfromToTimestamp == "0_0" {
			continue
		}

		// Get the value
		temp := deviceEntry.GetValue()
		var metaEntry models.Metadata
		err = json.Unmarshal(temp, &metaEntry)
		if err != nil {
			return nil, err
		}

		metaData = append(metaData, metaEntry)

	}

	data := new(models.Auxiliary)
	if len(metaData) == 0 { //escape null pointer error in case no ledger entries in array
		return nil, nil
	}
	data.MetaArray = metaData

	return data, nil // must be converted to return an array of Metadata
}
