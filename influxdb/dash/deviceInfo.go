package dash

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"vermi/models"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// This function adds a new Device to the world state. //TESTED
// id= the device's unique identifier
// location= the device's location
// tmstamp= the timestamp on which the
func (sc *SmartContract) CreateDevice(ctx contractapi.TransactionContextInterface, id string, location string, tmstamp string) (*models.DeviceInfo, error) {

	// Create the composite keys
	indexName := "deviceID~fromToTimestamp"
	devMeta, err := ctx.GetStub().CreateCompositeKey(indexName, []string{id, "0_0"}) //fromTimestamp_toTimestamp
	if err != nil {
		return nil, err
	}
	indexName = "location~deviceID"
	devLocIndex, err := ctx.GetStub().CreateCompositeKey(indexName, []string{location, id})
	if err != nil {
		return nil, err
	}
	// empty value for initializing
	value := []byte{0x00}

	//Create the ordinary keys
	deviceInfokey := id
	tm, err := strconv.ParseInt(tmstamp, 10, 64)
	if err != nil {
		return nil, err
	}
	deviceInfo := models.DeviceInfo{Location: location, LastWriteTimestamp: tm}
	deviceInfoBytes, err := json.Marshal(deviceInfo)
	if err != nil {
		return nil, err
	}

// Update the world state -> must update the world state in a transaction manner; either both go in, or none... : PENDING
err = ctx.GetStub().PutState(devMeta, value)
if err != nil {
    return nil, err
}
err = ctx.GetStub().PutState(devLocIndex, value)
if err != nil {
    return nil, err
}
err = ctx.GetStub().PutState(deviceInfokey, deviceInfoBytes)
if err != nil {
    return nil, err
}

// Add an event to indicate that the device has been updated
eventPayload := []byte(fmt.Sprintf("Device %s updated", deviceInfo.DeviceID))
err = ctx.GetStub().SetEvent("DeviceUpdated", eventPayload)
if err != nil {
    return nil, err
}

return &deviceInfo, nil //success
}

// Computes the digest of a PointsJson struct; uses its string representation
func CalculateDigest(pts string) string { //PENDING
	return pts
}

// This function stores new metadata on blockchain // TESTED
func (sc *SmartContract) WriteBatch(ctx contractapi.TransactionContextInterface, id, fromTimestamp, toTimestamp string, dataJson string) (*models.Metadata, error) {

	//Check if the device is registered/created; if not, return error, else get its details.
	deviceInfo, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, errors.New("Device does not exist!") //err
	}
	var dInfo models.DeviceInfo
	err = json.Unmarshal([]byte(deviceInfo), &dInfo)
	if err != nil {
		return nil, errors.New("DeviceInfo Unmarshal failed!") //err
	}

	//update device info struct
	tm, err := strconv.ParseInt(toTimestamp, 10, 64)
	if err != nil {
		return nil, err
	}
	dInfo.LastWriteTimestamp = tm

	//Check if the data batch is empty
	var pts models.PointsJson
	err = json.Unmarshal([]byte(dataJson), &pts)
	if err != nil {
		return nil, errors.New("PointsJson unmarshal failed!") //err
	}
	isEmpty := false
	if len(pts.Points) == 0 {
		isEmpty = true
	} else {
		//Since device exists and there are points to write; attempt to write points in influx
		err = WriteToInflux(pts)

		// if writing to influx fails
		if err != nil {
			// trigger an empty entry in blockchain
			pts.Points = nil
			isEmpty = true
		}
	}

	// If no data in the data batch skip the digest computation,
	// else set it to false and calculate the digest.
	digest := ""
	if !isEmpty {
		digest = CalculateDigest(pts.String())
	}

	// Prepare the value to be added to the world state
	metaEntry := models.Metadata{Digest: digest, IsEmpty: isEmpty}
	metaEntryBytes, err := json.Marshal(metaEntry)
	if err != nil {
		return nil, errors.New("Metadata marshal failed!") //err
	}

	// Create the composite key
	indexName := "deviceID~fromToTimestamp"
	devMeta, err := ctx.GetStub().CreateCompositeKey(indexName, []string{id, fromTimestamp + `_` + toTimestamp})
	if err != nil {
		return nil, errors.New("Creation of composite key failed!") //err
	}

	//Update the world state
	//device info
	deviceInfo, err = json.Marshal(dInfo)
	if err != nil {
		return nil, errors.New("World state update: DeviceInfo Marshal failed!") //err
	}
	err = ctx.GetStub().PutState(id, deviceInfo)
	if err != nil {
		return nil, errors.New("World state update: Updating the world state with DeviceInfo failed!") //err
	}

	//metadata
	err = ctx.GetStub().PutState(devMeta, metaEntryBytes)
	if err != nil {
		return nil, errors.New("World state update: Updating the world state with metadata failed!") //err
	}

	return &metaEntry, nil

}
