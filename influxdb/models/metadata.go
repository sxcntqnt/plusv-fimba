package models	

type Metadata struct { //must decide on the metadata : PENDING
	Digest  string `json:"digest"`  // the corresponding fromTimestamp_toTimestamp can be found in the related composite key
	IsEmpty bool   `json:"isEmpty"` // 'true' indicates an empty data batch; used to indicate a WriteTiInfluxError as well
}
type Auxiliary struct {
	MetaArray []Metadata
}
