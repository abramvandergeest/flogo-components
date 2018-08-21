package dataprep

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	// "json"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

// log is the default package logger
var log = logger.GetLogger("activity-tibco-dataprep")

// JSONObject primary structure of dataprep json
type JSONObject struct {
	InObj  InputObject       `json:"input"`
	OpObjs []OperationObject `json:"operations"`
	OutObj OutputObject      `json:"output"`
}

// InputObject is the sub-Json that handles input structure
type InputObject struct {
	Type  string `json:"type"`
	Shape string `json:"shape"`
}

// OperationObject is a struct of operations - to be developed further
type OperationObject struct {
	Name string `json:"name"`
	List string `json:"list"`
}

// OutputObject is the sub-Json that handles input structure
type OutputObject struct {
	Type  string `json:"type"`
	Shape string `json:"shape"`
	Build string `json:"build"`
}

// MyActivity is a stub for your Activity implementation
type MyActivity struct {
	metadata *activity.Metadata
}

// NewActivity creates a new activity
func NewActivity(metadata *activity.Metadata) activity.Activity {
	return &MyActivity{metadata: metadata}
}

// Metadata implements activity.Activity.Metadata
func (a *MyActivity) Metadata() *activity.Metadata {
	return a.metadata
}

// Eval implements activity.Activity.Eval
func (a *MyActivity) Eval(context activity.Context) (done bool, err error) {

	filename := "dataprep.json"

	file, err := os.Open(filename)
	if err != nil {
		log.Errorf("Can't open file %s", filename)
		return false, err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("Cannot read dataprep.json into bytes")
	}

	var jsontype JSONObject
	json.Unmarshal(byteValue, &jsontype)
	fmt.Println(jsontype)

	return true, nil
}
