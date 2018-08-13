//go:generate go run ../../../../TIBCOSoftware/flogo-lib/flogo/gen/gen.go $GOPATH
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

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
	httpport  = os.Getenv("HTTPPORT")
	modelPath = os.Getenv("SMMODEL")
	//GLOVEFILE defined in preprocess.go
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
	trg.NewFuncHandler(map[string]interface{}{"method": "POST", "path": "/api"}, handler)

	return app
}

func handler(ctx context.Context, inputs map[string]*data.Attribute) (map[string]*data.Attribute, error) {

	//Getting source objects from json string - after cleaning/normalizing the strings
	s := strings.Map(normStr, inputs["queryParams"].Value().(map[string]string)["Source"])
	// fmt.Println(s)
	srcobjs, err := jsonStr2Obj(s)
	if err != nil {
		return nil, err
	}

	//Getting target objects from json string - after cleaning/normalizing the strings
	s = strings.Map(normStr, inputs["queryParams"].Value().(map[string]string)["Target"])
	// fmt.Println(s)
	trgobjs, err := jsonStr2Obj(s)
	if err != nil {
		return nil, err
	}

	// Get the ID from the path
	id := inputs["queryParams"].Value().(map[string]string)["id"]

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
	m := modelPath
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

			//YOU CAN USE THIS PORTION TO CREATE LABELED DATA, COPY PASTE THE PRINTED LINES AND ADD 1/0 FOR THE LABEL
			// if mapProb > 0.5 {
			// 	fmt.Printf("%s,%s,%s,%s,%d,%d,%s,%s,\n", obja.fieldType, objb.fieldType,
			// 		obja.label, objb.label, obja.fieldLength, objb.fieldLength, obja.name, objb.name)
			// }

		}
		output[obja.name] = listofcomps
	}

	// The return message is a map[string]*data.Attribute which we'll have to construct
	response := make(map[string]interface{})
	response["output"] = output2

	ret := make(map[string]*data.Attribute)
	ret["code"], _ = data.NewAttribute("code", data.TypeInteger, 200)
	ret["data"], _ = data.NewAttribute("data", data.TypeAny, response)

	return ret, nil
}
