//go:generate go run ../../../../TIBCOSoftware/flogo-lib/flogo/gen/gen.go $GOPATH
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/TIBCOSoftware/flogo-contrib/activity/inference"
	rt "github.com/TIBCOSoftware/flogo-contrib/trigger/rest"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/engine"
	"github.com/TIBCOSoftware/flogo-lib/flogo"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	// "github.com/abramvandergeest/flogo-components/activity/inference"
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

	//Getting source objects as []interface{} from POST body

	srcobjs := inputs["content"].Value().(map[string]interface{})["Source"].([]interface{})
	trgobjs := inputs["content"].Value().(map[string]interface{})["Target"].([]interface{})

	// Creating output variable for catching inference results
	//   Shape of output is output["sourceName"]={"targetName", "matchProbability"}
	output := make(map[string][]interface{})

	// Defining constant inference values
	m := modelPath
	fmt.Println(m)
	inputName := "inputs"
	framework := "Tensorflow"

	//Looping over the source and target objects
	var features []interface{}

	for _, ointa := range srcobjs {

		//Converting the interface{} object into a objectField
		oa := ointa.(map[string]interface{})
		obja := objectField{
			Name:        oa["name"].(string),
			Label:       oa["label"].(string),
			FieldType:   oa["type"].(string),
			FieldLength: int32(oa["type_length"].(float64))}

		for _, ointb := range trgobjs {
			//Converting the interface{} object into a objectField
			ob := ointb.(map[string]interface{})
			objb := objectField{
				Name:        ob["name"].(string),
				Label:       ob["label"].(string),
				FieldType:   ob["type"].(string),
				FieldLength: int32(ob["type_length"].(float64))}

			//Getting vector embeddings
			obja = obja.embedding()
			objb = objb.embedding()
			features = append(features, map[string]interface{}{
				"name": "inputs",
				"data": objs2features(obja, objb),
			})

			// logger.Info(fmt.Sprintf("sourceName:%s  targetName:%s", obja.Name, objb.Name))

			// Given inputs make inference with ML model
			in := map[string]interface{}{"model": m, "inputName": inputName, "framework": framework, "features": features}
			out, err := flogo.EvalActivity(&inference.InferenceActivity{}, in)
			if err != nil {
				return nil, err
			}
			mapProb := out["result"].Value().(map[string]interface{})["Mapping"].(float32)

			//Logging prediction from source and target name
			s := fmt.Sprintf(`{"SourceName":"%s","TargetName":"%s","match":%f}`, obja.Name, objb.Name, float64(mapProb))
			logger.Info(s)
			output[obja.Name] = append(output[obja.Name],
				outrow{
					TargetName: objb.Name,
					Match:      float64(mapProb),
					Label:      objb.Label,
					Type:       objb.FieldType,
					TypeLength: objb.FieldLength,
				},
			)

			//YOU CAN USE THIS PORTION TO CREATE LABELED DATA, COPY PASTE THE PRINTED LINES AND ADD 1/0 FOR THE LABEL
			// if mapProb > 0.0 { //0.5
			// 	fmt.Printf("%s,%s,%s,%s,%d,%d,%s,%s,\n", obja.FieldType, objb.FieldType,
			// 		obja.Label, objb.Label, obja.FieldLength, objb.FieldLength, obja.Name, objb.Name)
			// }

		}
	}

	// The return message is a map[string]*data.Attribute which we'll have to construct
	response := make(map[string]interface{})
	response["output"] = output

	ret := make(map[string]*data.Attribute)
	ret["code"], _ = data.NewAttribute("code", data.TypeInteger, 200)
	ret["data"], _ = data.NewAttribute("data", data.TypeAny, response)

	return ret, nil
}
