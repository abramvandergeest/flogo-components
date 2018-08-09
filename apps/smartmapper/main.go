//go:generate go run ../../../../TIBCOSoftware/flogo-lib/flogo/gen/gen.go $GOPATH
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/TIBCOSoftware/flogo-contrib/activity/inference"
	"github.com/TIBCOSoftware/flogo-contrib/activity/log"
	rt "github.com/TIBCOSoftware/flogo-contrib/trigger/rest"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/engine"
	"github.com/TIBCOSoftware/flogo-lib/flogo"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	// "github.com/abramvandergeest/flogo-components/activity/inference"
)

var (
	httpport       = os.Getenv("HTTPPORT")
	paymentservice = os.Getenv("PAYMENTSERVICE")
)

func main() {
	// Create a new Flogo app
	app := appBuilder()

	e, err := flogo.NewEngine(app)

	if err != nil {
		logger.Error(err)
		return
	}

	engine.RunEngine(e)
}

type outrow struct {
	TargetName string  `json:"targetName"`
	Match      float64 `json:"match"`
}

func appBuilder() *flogo.App {
	app := flogo.NewApp()

	// Convert the HTTPPort to an integer
	port, err := strconv.Atoi(httpport)
	if err != nil {
		logger.Error(err)
	}

	// Register the HTTP trigger
	trg := app.NewTrigger(&rt.RestTrigger{}, map[string]interface{}{"port": port})
	// trg.NewFuncHandler(map[string]interface{}{"method": "GET", "path": "/api/invoices/:id"}, handler)
	trg.NewFuncHandler(map[string]interface{}{"method": "POST", "path": "/api/invoices"}, handler)

	return app
}

func jsonStr2Obj(str string) (out []objectField, err error) {

	var data []map[string]interface{}
	err = json.Unmarshal([]byte(str), &data)
	if err != nil {
		return nil, err
	}
	for _, a := range data {
		o := objectField{
			name:        a["name"].(string),
			label:       a["label"].(string),
			fieldType:   a["type"].(string),
			fieldLength: int32(a["type_length"].(float64))}
		out = append(out, o)
	}
	return out, nil
}

func handler(ctx context.Context, inputs map[string]*data.Attribute) (map[string]*data.Attribute, error) {

	//Getting source objects from json string
	srcobjs, err := jsonStr2Obj(inputs["queryParams"].Value().(map[string]string)["Source"])
	if err != nil {
		return nil, err
	}
	fmt.Println("Wish", srcobjs)

	//Getting target objects from json string
	trgobjs, err := jsonStr2Obj(inputs["queryParams"].Value().(map[string]string)["Target"])
	if err != nil {
		return nil, err
	}
	fmt.Println("FEAR", trgobjs)

	// Get the ID from the path
	id := inputs["queryParams"].Value().(map[string]string)["id"]
	fmt.Println(id)

	// Execute the log activity
	// // // // I NEED TO LOG WHAT IS GOING ON!!!!!!
	in := map[string]interface{}{"message": id, "flowInfo": "true", "addToFlow": "true"}
	_, err = flogo.EvalActivity(&log.LogActivity{}, in)
	if err != nil {
		return nil, err
	}

	// Creating output variable for catching inference results
	//   Shape of output is output["sourceName"]={"targetName", "matchProbability"}
	output := make(map[string][]string)
	output2 := make(map[string][]interface{})

	// Defining constant inference values
	m := "/Users/avanders/working/working_python/Smart_mapper/Archive.zip"
	inputName := "inputs"
	framework := "Tensorflow"

	//Looping over the source and target objects
	var features map[string]interface{}
	for _, obja := range srcobjs {
		listofcomps := []string{}
		for _, objb := range trgobjs {
			obja = obja.embedding()
			objb = objb.embedding()
			features = objs2features(obja, objb)

			logger.Info(fmt.Sprintf("sourceName:%s  targetName:%s", obja.name, objb.name))

			// Given inputs make inference with ML model
			in = map[string]interface{}{"model": m, "inputName": inputName, "framework": framework, "features": features}
			out, err := flogo.EvalActivity(&inference.InferenceActivity{}, in)
			if err != nil {
				return nil, err
			}
			mapProb := out["result"].Value().(map[string]interface{})["scores"].([][]float32)[0][1]
			// orCurrent := outrow{TargetName: objb.name, match: float64(mapProb)}
			s := fmt.Sprintf(`{"TargetName":"%s","match":%f}`, objb.name, float64(mapProb))
			fmt.Println(s)
			listofcomps = append(listofcomps, s)
			output2[obja.name] = append(output2[obja.name], outrow{TargetName: objb.name, Match: float64(mapProb)})

		}
		output[obja.name] = listofcomps
	}

	fmt.Println("OUTPUT:", output)
	fmt.Println("OUTPUT2:", output2)

	// The return message is a map[string]*data.Attribute which we'll have to construct
	response := make(map[string]interface{})
	response["output"] = output2

	ret := make(map[string]*data.Attribute)
	ret["code"], _ = data.NewAttribute("code", data.TypeInteger, 200)
	ret["data"], _ = data.NewAttribute("data", data.TypeAny, response)

	return ret, nil
}
