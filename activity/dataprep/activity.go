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
	Type string `json:"type"`
	Dim  int    `json:"dimensions"`
}

// OperationObject is a struct of operations - to be developed further
type OperationObject struct {
	Input  string   `json:"input"`
	Inputs []string `json:"inputs"`
	Name   string   `json:"fnname"`
	List   []string `json:"list"`
	Label  string   `json:"label"`
}

// OutputObject is the sub-Json that handles input structure
type OutputObject struct {
	Type   string `json:"type"`
	Dim    int    `json:"dimensions"`
	OutVar string `json:"outvar"`
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

func readInJSON(filename string) (JSONObject, error) {
	// Reading json and putting it into variable
	file, err := os.Open(filename)
	if err != nil {
		log.Errorf("Can't open file %s", filename)
		return JSONObject{}, err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("Cannot read dataprep.json into bytes")
		return JSONObject{}, err
	}
	// Defining json varable
	var injson JSONObject
	json.Unmarshal(byteValue, &injson)
	return injson, err
}

// Eval implements activity.Activity.Eval
func (a *MyActivity) Eval(context activity.Context) (done bool, err error) {

	var filename string
	if context.GetInput("jsonfile") == "" {
		filename = "dataprep.json"
	} else {
		filename = context.GetInput("jsonfile").(string)
	}

	injson, err := readInJSON(filename)

	// creating maps to hold value and dimensions oftotal namespace
	nameSpace := make(map[string]interface{})
	dimSpace := make(map[string]interface{})

	if injson.InObj.Type == "object" || injson.InObj.Type == "array" {
		if injson.InObj.Dim == 0 {
			nameSpace["input"] = context.GetInput("input").(interface{})
		} else if injson.InObj.Dim == 1 {
			nameSpace["input"] = context.GetInput("input").([]interface{})
		} else if injson.InObj.Dim == 2 {
			nameSpace["input"] = context.GetInput("input").([][]interface{})
		} else if injson.InObj.Dim == 3 {
			nameSpace["input"] = context.GetInput("input").([][][]interface{})
		} else if injson.InObj.Dim == 4 {
			nameSpace["input"] = context.GetInput("input").([][][][]interface{})
		}

		dimSpace["input"] = injson.InObj.Dim
	}

	/////////I Think I have to create an object to be in nameSpace so I can track features on object
	fmt.Println("Printing inputs for each operation as it is reached")
	for _, op := range injson.OpObjs {
		if op.Input != "" {
			fmt.Println(op.Name, op.List, op.Label, op.Input)
		} else {
			fmt.Println(op.Name, op.List, op.Label, op.Inputs)
		}
		if op.Name == "rename" {
			nameSpace[op.Label], err = rename(nameSpace[op.Input])
			if err != nil {
				return false, err
			}
			dimSpace[op.Label] = dimSpace[op.Input]
		} else if op.Name == "magnitude" {
			nameSpace[op.Label], err = magnitude(nameSpace[op.Input].([][]interface{}))
			if err != nil {
				return false, err
			}
			dimSpace[op.Label] = 1
		} else if op.Name == "addCol2Tab" {
			nameSpace[op.Label], err = addCol2Tab(
				nameSpace[op.Inputs[0]].([][]interface{}),
				nameSpace[op.Inputs[1]].([]interface{}),
			)
			if err != nil {
				return false, err
			}
			dimSpace[op.Label] = 1
		} else if op.Name == "flatten" {
			nameSpace[op.Label], err = flatten(nameSpace[op.Input].([][]interface{}))
			if err != nil {
				return false, err
			}
			dimSpace[op.Label] = 1
		} else if op.Name == "toMap" {
			nameSpace[op.Label], err = toMap(nameSpace[op.Input].([]interface{}), op.List)
			if err != nil {
				return false, err
			}
			dimSpace[op.Label] = -1 // Use this to mean a dictionary?
		}
	}

	// Need to set up building an ouput variable from parts, I don't think I need it for the moment
	fmt.Println("OUTPUT:", nameSpace[injson.OutObj.OutVar], injson.OutObj.OutVar)
	context.SetOutput("output", nameSpace[injson.OutObj.OutVar])

	return true, nil
}
