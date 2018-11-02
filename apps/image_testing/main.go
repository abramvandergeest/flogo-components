//go:generate go run ../../../../TIBCOSoftware/flogo-lib/flogo/gen/gen.go $GOPATH
package main

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"strconv"

	"github.com/TIBCOSoftware/flogo-contrib/activity/inference"
	rt "github.com/TIBCOSoftware/flogo-contrib/trigger/rest"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/engine"
	"github.com/TIBCOSoftware/flogo-lib/flogo"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

var (
	httpport  = os.Getenv("HTTPPORT")
	modelPath = os.Getenv("IMGMODEL")
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
	trg.NewFuncHandler(map[string]interface{}{"method": "POST", "path": "/api"}, handler)

	return app
}

func handler(ctx context.Context, inputs map[string]*data.Attribute) (map[string]*data.Attribute, error) {

	//Getting source objects as []interface{} from POST body
	fmt.Println("Gets Triggered.")

	blah := inputs["content"].Value().(map[string]interface{})["Input"] //.(int32)
	fmt.Println(blah)
	// trgobjs := inputs["content"].Value().(map[string]interface{})["Target"].([]interface{})

	filename := "/Users/avanderg@tibco.com/Desktop/Touch Bar Shot 2018-10-30 at 3.51.06 PM.png"
	infile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer infile.Close()

	src, _, err2 := image.Decode(infile)
	if err2 != nil {
		return nil, err2
	}
	m := modelPath
	framework := "Tensorflow"
	var features []interface{}
	features = append(features, map[string]interface{}{
		"name": "inputs",
		"data": 2,
	})

	in := map[string]interface{}{"model": m, "framework": framework, "features": features}
	fmt.Println(in["model"])

	out, err := flogo.EvalActivity(&inference.InferenceActivity{}, in)
	if err != nil {
		return nil, err
	}

	fmt.Println(out)

	outfile, err := os.Create("Blah.png")
	if err != nil {
		return nil, err
	}
	defer outfile.Close()
	png.Encode(outfile, src)

	// Creating output variable for catching inference results
	//   Shape of output is output["sourceName"]={"targetName", "matchProbability"}
	output := make(map[string][]interface{})

	// // Defining constant inference values
	// m := modelPath
	// inputName := "inputs"
	// framework := "Tensorflow"

	// //Looping over the source and target objects
	// var features map[string]interface{}
	// for _, ointa := range srcobjs {

	// 	//Converting the interface{} object into a objectField
	// 	oa := ointa.(map[string]interface{})
	// 	obja := objectField{
	// 		Name:        oa["name"].(string),
	// 		Label:       oa["label"].(string),
	// 		FieldType:   oa["type"].(string),
	// 		FieldLength: int32(oa["type_length"].(float64))}

	// 	for _, ointb := range trgobjs {
	// 		//Converting the interface{} object into a objectField
	// 		ob := ointb.(map[string]interface{})
	// 		objb := objectField{
	// 			Name:        ob["name"].(string),
	// 			Label:       ob["label"].(string),
	// 			FieldType:   ob["type"].(string),
	// 			FieldLength: int32(ob["type_length"].(float64))}

	// 		//Getting vector embeddings
	// 		obja = obja.embedding()
	// 		objb = objb.embedding()
	// 		features = objs2features(obja, objb)

	// 		// logger.Info(fmt.Sprintf("sourceName:%s  targetName:%s", obja.Name, objb.Name))

	// 		// Given inputs make inference with ML model
	// 		in := map[string]interface{}{"model": m, "inputName": inputName, "framework": framework, "features": features}
	// 		out, err := flogo.EvalActivity(&inference.InferenceActivity{}, in)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		mapProb := out["result"].Value().(map[string]interface{})["scores"].([][]float32)[0][1]

	// 		//Logging prediction from source and target name
	// 		s := fmt.Sprintf(`{"SourceName":"%s","TargetName":"%s","match":%f}`, obja.Name, objb.Name, float64(mapProb))
	// 		logger.Info(s)
	// 		output[obja.Name] = append(output[obja.Name],
	// 			outrow{
	// 				TargetName: objb.Name,
	// 				Match:      float64(mapProb),
	// 				Label:      objb.Label,
	// 				Type:       objb.FieldType,
	// 				TypeLength: objb.FieldLength,
	// 			},
	// 		)

	// 		//YOU CAN USE THIS PORTION TO CREATE LABELED DATA, COPY PASTE THE PRINTED LINES AND ADD 1/0 FOR THE LABEL
	// 		// if mapProb > 0.0 { //0.5
	// 		// 	fmt.Printf("%s,%s,%s,%s,%d,%d,%s,%s,\n", obja.FieldType, objb.FieldType,
	// 		// 		obja.Label, objb.Label, obja.FieldLength, objb.FieldLength, obja.Name, objb.Name)
	// 		// }

	// 	}
	// }

	// The return message is a map[string]*data.Attribute which we'll have to construct
	response := make(map[string]interface{})
	response["output"] = output

	ret := make(map[string]*data.Attribute)
	ret["code"], _ = data.NewAttribute("code", data.TypeInteger, 200)
	ret["data"], _ = data.NewAttribute("data", data.TypeAny, response)

	return ret, nil
}
