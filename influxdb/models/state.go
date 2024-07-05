package models

type Record struct {
	ObjectType  string  `json:"objectType"` // The type of the object, used as a key in the composite key
	Key         string  `json:"key"`        // The unique key for the record, used as part of the composite key
	Value       float64 `json:"value"`
	Timestamp   int64   `json:"timestamp"`   // The timestamp when the record was created
	Description string  `json:"description"` // A description of the record
}
type QuerySelector struct {
	DocType string `json:"docType"`
}

type Query struct {
	Selector QuerySelector `json:"selector"`
}
type queryResponse struct {
	Results        []Record `json:"results"`
	Metadata       struct {
		FetchedRecords int32 `json:"fetchedRecords"`
		Bookmark       string `json:"bookmark"`
	} `json:"metadata"`
}


