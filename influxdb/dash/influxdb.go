package dash

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	api "github.com/influxdata/influxdb-client-go/v2/api"
	"log"
	"strconv"
	"time"
	"vermi/models"
)

var (
	InfluxdbURL           string
	InfluxdbOrg           string
	InfluxdbBucket        string
	InfluxdbToken         string
	InfluxdbBatchSize     int
	InfluxdbFlushInterval time.Duration
)

// The connection's struct
type InfluxDB struct {
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
	queryAPI api.QueryAPI
}

// The struct into which the received points (from the chaincode call) are unmarshalled (i.e. json -> golang struct)
type PointsJson struct {
	Points []struct {
		Measurement string `json:"measurement"`
		Tags        []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"tags"`
		Fields []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"fields"`
		Timestamp string `json:"timestamp"`
	} `json:"points"`
}

// for retrieving data from influx
type PointInflux struct {
	deviceID  string
	timestamp int64
	value     interface{}
}

func (p *PointInflux) String() string {
	return fmt.Sprintf(`{ 'deviceID' : '%s' , 'timestamp' : '%v' , 'value' : '%v' }`, p.deviceID, p.timestamp, p.value)
}
func (infdb *InfluxDB) InitConnection(url, bucket, token, org string, batchSize uint, influxdbFlushInterval time.Duration) error {
	// Create a new client if it doesn't exist
	if infdb.client == nil {
		// Create a new InfluxDB client
		client := influxdb2.NewClient(url, token)
		infdb.client = client
	}

	// Create a new write API if it doesn't exist
	if infdb.writeAPI == nil {
		infdb.writeAPI = infdb.client.WriteAPIBlocking(org, bucket)
	}

	// Convert the flush interval to nanoseconds
	flushInterval := uint(influxdbFlushInterval.Nanoseconds())

	// Configure the write options
	wo := influxdb2.DefaultOptions().
		SetBatchSize(batchSize).
		SetFlushInterval(flushInterval)

	// Create the InfluxDB client and write API
	client := influxdb2.NewClientWithOptions(url, token, wo)
	writeAPI := client.WriteAPIBlocking(org, bucket)

	// Reconfigure the existing write API with the new options
	writeAPIBlocking, ok := infdb.writeAPI.(api.WriteAPIBlocking)
	if !ok {
		return fmt.Errorf("unable to reconfigure write API with new options: invalid write API type")
	}

	writeAPI = writeAPIBlocking
	infdb.writeAPI = writeAPI

	// Return nil to indicate success
	return nil

}

func (infdb *InfluxDB) TerminateConnection() error {

	infdb.client.Close()

	return nil
}

// Retrieves the points for the specified time period, meausurement and bucket.
func (sc *SmartContract) ReadFromInflux(ctx contractapi.TransactionContextInterface, database string, retentionPolicy string, start string, stop string, aux string, queryType string) (string, error) {
	// Check if start and stop times are valid timestamps
	_, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return "", fmt.Errorf("invalid start time: %s", err.Error())
	}
	_, err = time.Parse(time.RFC3339, stop)
	if err != nil {
		return "", fmt.Errorf("invalid stop time: %s", err.Error())
	}

	c := InfluxDB{}
	_ = c.InitConnection(InfluxdbURL, InfluxdbBucket, InfluxdbToken, InfluxdbOrg, uint(InfluxdbBatchSize), InfluxdbFlushInterval)

	// create flux query
	var fq string
	if queryType == "true" {
		fq = `from(bucket: "` + database + `/` + retentionPolicy + `")  |> range(start: ` + start + `, stop: ` + stop + `) |> filter(fn: (r) => r._measurement == "` + aux + `")`
	} else {
		fq = `from(bucket: "` + database + `/` + retentionPolicy + `")  |> range(start: ` + start + `, stop: ` + stop + `) |> filter(fn: (r) => r.location == "` + aux + `")`
	}

	// get QueryTableResult
	result, err := c.queryAPI.Query(context.Background(), fq)
	if err != nil {
		fmt.Printf("Query Error: %s\n", err.Error())
		return "", err
	}

	// create slice of PointInflux structs
	var recSet []PointInflux

	// iterate over result and process data one record at a time
	for result.Next() {
		tuple := result.Record()
		var rec PointInflux
		rec = PointInflux{
			deviceID:  tuple.Measurement(),
			timestamp: tuple.Time().Unix(),
			value:     tuple.Value(),
		}
		recSet = append(recSet, rec)
	}

	// check if any error occurred during iteration
	if result.Err() != nil {
		return "", result.Err()
	}

	// convert recSet to JSON string
	resultStr, err := json.Marshal(recSet)
	if err != nil {
		return "", err
	}

	// ensures background processes finishes
	c.TerminateConnection()

	return string(resultStr), nil
}

// This function writes to influx; auxiliary. //TESTED
// Does not return an error if the database it attempts to write in does not exist. //PENDING
func WriteToInflux(pts models.PointsJson) error {

	infdb := InfluxDB{}
	_ = infdb.InitConnection(InfluxdbURL, InfluxdbBucket, InfluxdbToken, InfluxdbOrg, uint(InfluxdbBatchSize), InfluxdbFlushInterval)
	for _, point := range pts.Points { //for each point

		tags_map := make(map[string]string)
		for _, tag := range point.Tags { //create the tags
			tags_map[tag.Key] = tag.Value
		}

		fields_map := make(map[string]interface{})
		for _, field := range point.Fields { // create the fields
			fields_map[field.Key] = field.Value
		}

		//parse the timestamp (received as an integer contained in a string) into the biggest int available
		tm, err := strconv.ParseInt(point.Timestamp, 10, 64)
		if err != nil {
			panic(err)
		}

		point_time := time.Unix(tm, 0)

		p := influxdb2.NewPoint(
			point.Measurement,
			tags_map,
			fields_map,
			point_time)

		ctx := context.Background()
		infdb.writeAPI.WritePoint(ctx, p)

		// write asynchronously
		//	infdb.writeAPI.WritePoint(p)
	}

	err := infdb.writeAPI.Flush(context.Background())
	if err != nil {
		log.Println("Error flushing writes:", err)
	}

	return nil
}

// This function writes to influx with a chaincode call; defined on the contract. // TESTED
func (sc *SmartContract) WriteToInflux(ctx contractapi.TransactionContextInterface, dataJson string) error {

	//data pre-processing
	var pts models.PointsJson
	err := json.Unmarshal([]byte(dataJson), &pts)
	if err != nil {
		return err
	}

	return WriteToInflux(pts)

}
