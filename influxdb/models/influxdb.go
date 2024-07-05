package models

import (
	"fmt"

influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/v2/api"
)
// The connection's struct
type InfluxDB struct {
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
	queryAPI api.QueryAPI
}

// for retrieving data from influx
type PointInflux struct {
	deviceID  string
	timestamp int64
	value     interface{}
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

func (p *PointInflux) String() string {
	return fmt.Sprintf(`{ 'deviceID' : '%s' , 'timestamp' : '%v' , 'value' : '%v' }`, p.deviceID, p.timestamp, p.value)
}
// Returns a string representation of a PointsJson struct
func (p *PointsJson) String() string {

        str := `[`

        for _, pnt := range p.Points {

                str = str + `(` + pnt.Measurement

                for _, t := range pnt.Tags {
                        str = str + `,` + t.Key + `:` + t.Value
                }

                for _, f := range pnt.Fields {
                        str = str + `,` + f.Key + `:` + f.Value
                }

                str = str + `)`
        }

        str = str + `]`

        return str
}

