package dash

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strings"
	"time"
//		"vermi/models"
)

type SmartContract struct {
        contractapi.Contract
}

type Record struct {
	ObjectType  string `json:"objectType"`  // The type of the object, used as a key in the composite key
	Key         string `json:"key"`         // The unique key for the record, used as part of the composite key
	Timestamp   int64  `json:"timestamp"`   // The timestamp when the record was created
	Description string `json:"description"` // A description of the record
}
type QueryResponse struct {
	Results  []Record `json:"results"`
	Metadata struct {
		FetchedRecords int32  `json:"fetchedRecords"`
		Bookmark       string `json:"bookmark"`
	} `json:"metadata"`
}

type QuerySelector struct {
	DocType string `json:"docType"`
}

type Query struct {
	Selector QuerySelector `json:"selector"`
}

type Metadata struct {
	Count int32 `json:"count"`
}

const recordIndexName = "record~docType~ownerIndex"
const recordDocType = "Record"

func (sc *SmartContract) GetHistoryForRecord(ctx contractapi.TransactionContextInterface, recordKey string) ([]byte, error) {
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(recordKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get history for record key %s: %v", recordKey, err)
	}
	defer resultsIterator.Close()

	var history []string
	for resultsIterator.HasNext() {
		result, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next history data for record key %s: %v", recordKey, err)
		}

		txTimestamp := time.Unix(result.Timestamp.GetSeconds(), int64(result.Timestamp.GetNanos()))
		txAction := string(result.Value)

		history = append(history, fmt.Sprintf("[%s] %s", txTimestamp.String(), txAction))
	}

	return json.Marshal(history)
}
func (t *SmartContract) QueryAllRecords(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string) ([]byte, error) {
	// Set up the query selector
	queryString := fmt.Sprintf(`{"selector": {"docType": "%s"}}`, recordDocType)
	queryResults, metadata, err := ctx.GetStub().GetQueryResultWithPagination(queryString, pageSize, bookmark)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	defer queryResults.Close()

	// Create a slice to hold the results
	var results []Record

	// Iterate through the query results and unmarshal each record into a Record struct
	for queryResults.HasNext() {
		recordBytes, err := queryResults.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to read record from query results: %v", err)
		}

		var record Record
		if err := json.Unmarshal(recordBytes.Value, &record); err != nil {
			return nil, fmt.Errorf("failed to unmarshal record from query results: %v", err)
		}
		results = append(results, record)
	}
	 metadataBytes, err := json.Marshal(metadata)
        if err != nil {
                return nil, fmt.Errorf("failed to marshal metadata: %v", err)
        }
        metadataStr := string(metadataBytes)

	// Create a QueryResponse struct to hold the results and metadata
	type QueryResponse struct {
		Results        []Record `json:"results"`
		Metadata       string   `json:"metadata"`
		FetchedRecords int32    `json:"fetched_records"`
	}

	response := QueryResponse{
		Results:        results,
		Metadata:       metadataStr,
		FetchedRecords: int32(len(results)),
	}

	// Marshal the QueryResponse struct into JSON and return it
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %v", err)
	}
	return responseBytes, nil
}

func (sc *SmartContract) QueryRecordsByRange(ctx contractapi.TransactionContext, objectType string, startKey string, endKey string) ([]byte, error) {
	// Create a selector for the given range
	selector := map[string]interface{}{
		"selector": map[string]interface{}{
			"objectType": objectType,
			"recordKey": map[string]interface{}{
				"$gte": startKey,
				"$lte": endKey,
			},
		},
	}

	// Execute the query and retrieve the results
	resultsIterator, err := ctx.GetStub().GetQueryResult(JSONToString(selector))
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer resultsIterator.Close()

	// Convert the results to a slice of Record structs
	var records []Record
	for resultsIterator.HasNext() {
		recordBytes, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to read query result: %v", err)
		}

		var record Record
		err = json.Unmarshal(recordBytes.Value, &record)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal query result into Record struct: %v", err)
		}

		records = append(records, record)
	}
	// Marshal the results into JSON and return them
	return json.Marshal(records)
}

func getKeyForObjectType(objectType string, key string) (string, error) {
    if len(objectType) == 0 {
        return "", fmt.Errorf("object type cannot be empty")
    }

    if len(key) == 0 {
        return "", fmt.Errorf("key cannot be empty")
    }

    // Combine the objectType and key to form a composite key
    return fmt.Sprintf("%s~%s", objectType, key), nil
}

func GetQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) (shim.StateQueryIteratorInterface, error) {
    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
        return nil, fmt.Errorf("failed to execute range query: %v", err)
    }
    defer resultsIterator.Close()

    return resultsIterator, nil
}

func (sc *SmartContract) QueryRecordsByObjectType(ctx contractapi.TransactionContextInterface, objectType string) ([]byte, error) {
	// Create a selector for the given object type
	queryString := fmt.Sprintf(`{
		"selector": {
			"docType": "record",
			"objectType": "%s"
		}
	}`, objectType)

	// Execute the query and retrieve the results
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer resultsIterator.Close()

	// Convert the results to a JSON array
	var results []Record
	for resultsIterator.HasNext() {
		recordBytes, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to read query result: %v", err)
		}

		var record Record
		err = json.Unmarshal(recordBytes.Value, &record)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal query result: %v", err)
		}

		results = append(results, record)
	}
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return nil, fmt.Errorf("failed to convert query results to JSON: %v", err)
	}

	return resultsJSON, nil
}

func JSONToString(data map[string]interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(b)
}

func (sc *SmartContract) QueryRecordsByObjectAndTime(ctx contractapi.TransactionContextInterface, objectType string, startTimestamp int64, endTimestamp int64) ([]byte, error) {
	startKey := getKeyForTimestamp(startTimestamp)
	endKey := getKeyForTimestamp(endTimestamp)

	// Create the composite key prefix
	objectTimePrefix, err := ctx.GetStub().CreateCompositeKey(recordIndexName, []string{objectType})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	// Get the results iterator
	resultsIterator, err := ctx.GetStub().GetStateByRange(objectTimePrefix+startKey, objectTimePrefix+endKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %v", err)
	}
	defer resultsIterator.Close()

	// Unmarshal the records
	records, err := unmarshalRecords(resultsIterator)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal records: %v", err)
	}

	// Marshal the records into JSON and return them
	recordsJSON, err := json.Marshal(records)
	if err != nil {
		return nil, fmt.Errorf("failed to convert records to JSON: %v", err)
	}
	return recordsJSON, nil
}

func getKeyForTimestamp(timestamp int64) string {
	return fmt.Sprintf("%010d", timestamp)
}
func unmarshalRecords(resultsIterator shim.StateQueryIteratorInterface) ([]Record, error) {
	var records []Record

	for resultsIterator.HasNext() {
		recordValue, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next result: %v", err)
		}
		var record Record
		err = json.Unmarshal(recordValue.Value, &record)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal record: %v", err)
		}

		record.ObjectType, record.Key, err = getCompositeKeyParts(recordValue.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to get composite key parts: %v", err)
		}
		records = append(records, record)
	}

	return records, nil
}

func getCompositeKeyParts(compositeKey string) (string, string, error) {
	parts := strings.Split(compositeKey, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid composite key: %s", compositeKey)
	}
	return parts[0], parts[1], nil
}
