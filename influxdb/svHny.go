package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"time"

	//       api "github.com/influxdata/influxdb-client-go/v2/api"
	influxdb2 "github.com/influxdata/influxdb-client-go"

	"log"
	"strconv"
	"vermi/dash"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/spf13/viper"
	//	"vermi/models"
)

var (
	InfluxdbURL           string
	InfluxdbOrg           string
	InfluxdbBucket        string
	InfluxdbToken         string
	InfluxdbBatchSize     int
	InfluxdbFlushInterval time.Duration
)
var cash = "./dash/sttQryz.go"
type SmartContract struct {
	contractapi.Contract
	SmartContractFunctions map[string]func(contractapi.TransactionContextInterface, []string) pb.Response
}
type serverConfig struct {
	CCID    string
	Address string
}

/*
	var SmartContractFunctions = map[string]func(contractapi.TransactionContextInterface, []string) pb.Response{
		"GetHistoryForRecord":          dash.GetHistoryForRecord,
		"QueryAllRecords":              dash.QueryAllRecordsHandler,
		"QueryRecordsByRange":          dash.QueryRecordsByRangeHandler,
		"getKeyForObjectType":          dash.getKeyForObjectTypeHandler,
		"GetQueryResultForQueryString": dash.GetQueryResultForQueryStringHandler,
		"QueryRecordsByObjectType":     dash.QueryRecordsByObjectTypeHandler,
	}
*/
func (sc *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	// Load configuration from file or environment variables using Viper
	viper.SetConfigFile("config.yaml")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return pb.Response{Status: 500, Message: fmt.Sprintf("error reading config file: %s", err)}
	}

	// Get InfluxDB configuration options from viper
	InfluxdbURL = viper.GetString("influxdb.url")
	InfluxdbOrg = viper.GetString("influxdb.org")
	InfluxdbBucket = viper.GetString("influxdb.bucket")
	InfluxdbToken = viper.GetString("influxdb.token")
	InfluxdbBatchSize = int(viper.GetUint("influxdb.batch_size"))
	InfluxdbFlushInterval = viper.GetDuration("influxdb.flush_interval")
	InfluxdbFlushIntervalMS := uint(InfluxdbFlushInterval.Milliseconds())

	// Connect to the InfluxDB database
	client := influxdb2.NewClientWithOptions(InfluxdbURL, InfluxdbToken,
		influxdb2.DefaultOptions().
			SetBatchSize(uint(InfluxdbBatchSize)).
			SetFlushInterval(InfluxdbFlushIntervalMS))
	defer client.Close()

	ready, err := client.Ready(context.Background())
	if err != nil {
		return pb.Response{Status: 500, Message: fmt.Sprintf("failed to establish connection to InfluxDB: %s", err)}
	}

	if !ready {
		return pb.Response{Status: 200, Message: "InfuxDB is not ready"}
	}

	writeAPI := client.WriteAPI(InfluxdbOrg, InfluxdbBucket)
	defer writeAPI.Close()

	// Connect to InfluxDB
	infdb := &dash.InfluxDB{}
	if err := infdb.InitConnection(InfluxdbURL, InfluxdbBucket, InfluxdbToken, InfluxdbOrg, uint(InfluxdbBatchSize), InfluxdbFlushInterval); err != nil {
		return pb.Response{Status: 500, Message: fmt.Sprintf("failed to initialize InfluxDB connection: %s", err)}
	}
	defer infdb.TerminateConnection()

	return pb.Response{Status: 200, Message: "InfluxDB connection initialized successfully"}
}
func buildSmartContractFunctionsMap() map[string]func(contractapi.TransactionContextInterface, []string) pb.Response {
	functions := make(map[string]func(contractapi.TransactionContextInterface, []string) pb.Response)
	pkg := reflect.ValueOf(cash)
	pkgType := pkg.Type()

	for i := 0; i < pkg.NumMethod(); i++ {
		method := pkg.Method(i)
		methodName := pkgType.Method(i).Name
		if strings.Contains(methodName, "Record") {
			fn := method.Interface().(func(contractapi.TransactionContextInterface, []string) ([]byte, error))
			functions[methodName] = convertToPBResponse(fn)
		}
	}

	return functions
}
func convertToPBResponse(fn func(contractapi.TransactionContextInterface, []string) ([]byte, error)) func(contractapi.TransactionContextInterface, []string) pb.Response {
	return func(ctx contractapi.TransactionContextInterface, args []string) pb.Response {
		resBytes, err := fn(ctx, args)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(resBytes)
	}
}
func (sc *SmartContract) Invoke(ctx contractapi.TransactionContextInterface) pb.Response {
	// Get the function name and arguments from the transaction proposal
	fn, args := ctx.GetStub().GetFunctionAndParameters()

	// Build the function map
	functions := buildSmartContractFunctionsMap()

	// Invoke the corresponding function from the functions map
	if f, ok := functions[fn]; ok {
		return f(ctx, args)
	}

	// Return an error if the function name is not found
	return shim.Error("Invalid Smart Contract function name.")
}

func main() {
	// See chaincode.env.example
	config := serverConfig{
		CCID:    viper.GetString("CHAINCODE_ID"),
		Address: viper.GetString("CHAINCODE_SERVER_ADDRESS"),
	}
	svRR, err := contractapi.NewChaincode(&SmartContract{})

	if err != nil {
		log.Panicf("error creating the influxdb chaincode: %s", err)
	}

	tlsConfig, err := getTLSProperties()
	if err != nil {
		log.Panicf("Error creating TLS configuration: %v", err)
	}
	// create a new SmartContract object
	smartContract := &SmartContract{}

	functions := buildSmartContractFunctionsMap()
	smartContract.SmartContractFunctions = functions


	server := &shim.ChaincodeServer{
		CCID:     config.CCID,
		Address:  config.Address,
		CC:       svRR,
		TLSProps: tlsConfig,
	}

	// start the chaincode server
	if err := server.Start(); err != nil {
		log.Panicf("error starting InfluxDB chaincode: %s", err)
	}
}
func getTLSProperties() (shim.TLSProperties, error) {
	tlsDisabled := viper.GetBool("CHAINCODE_TLS_DISABLED")

	if tlsDisabled {
		return shim.TLSProperties{}, nil
	}

	keyPath := viper.GetString("CHAINCODE_TLS_KEY")
	keyBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return shim.TLSProperties{}, fmt.Errorf("failed to read key file: %s", err)
	}

	certPath := viper.GetString("CHAINCODE_TLS_CERT")
	certBytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		return shim.TLSProperties{}, fmt.Errorf("failed to read cert file: %s", err)
	}

	clientCACertPath := viper.GetString("CHAINCODE_CLIENT_CA_CERT")
	clientCACertBytes, err := ioutil.ReadFile(clientCACertPath)
	if err != nil {
		return shim.TLSProperties{}, fmt.Errorf("failed to read client CA cert file: %s", err)
	}

	return shim.TLSProperties{
		Disabled:      tlsDisabled,
		Key:           keyBytes,
		Cert:          certBytes,
		ClientCACerts: clientCACertBytes,
	}, nil
}

func getEnvOrDefault(env, defaultVal string) string {
	viper.SetDefault(env, defaultVal)
	return viper.GetString(env)
}

// Note that the method returns default value if the string
// cannot be parsed!
func getBoolOrDefault(value string, defaultVal bool) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultVal
	}
	return parsed
}
