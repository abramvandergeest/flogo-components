//go:generate go run ../../../../TIBCOSoftware/flogo-lib/flogo/gen/gen.go $GOPATH
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	// "github.com/TIBCOSoftware/flogo-contrib/activity/inference"
	"github.com/TIBCOSoftware/flogo-contrib/activity/log"
	rt "github.com/TIBCOSoftware/flogo-contrib/trigger/rest"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/engine"
	"github.com/TIBCOSoftware/flogo-lib/flogo"
	"github.com/TIBCOSoftware/flogo-lib/logger"

	// "github.com/abramvandergeest/flogo-components/activity/inference"

	"github.com/abramvandergeest/flogo-components/activity/inference"
	"github.com/retgits/flogo-components/activity/randomnumber"
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

func appBuilder() *flogo.App {
	app := flogo.NewApp()

	// Convert the HTTPPort to an integer
	port, err := strconv.Atoi(httpport)
	if err != nil {
		logger.Error(err)
	}

	// Register the HTTP trigger
	trg := app.NewTrigger(&rt.RestTrigger{}, map[string]interface{}{"port": port})
	trg.NewFuncHandler(map[string]interface{}{"method": "GET", "path": "/api/invoices/:id"}, handler)

	return app
}

func handler(ctx context.Context, inputs map[string]*data.Attribute) (map[string]*data.Attribute, error) {
	obj1 := objectField{name: "testingNameHere", label: "labeledDescription-Hi", fieldType: "STRING", fieldLength: 128}
	obj2 := objectField{name: "Phone", label: "mobilePhone", fieldType: "STRING", fieldLength: 128}
	obj1 = obj1.embedding()
	obj2 = obj2.embedding()
	features := objs2features(obj1, obj2)
	// features["map"] = 0
	// var features map[string]interface{}

	// Get the ID from the path
	id := inputs["pathParams"].Value().(map[string]string)["id"]

	// Execute the log activity
	in := map[string]interface{}{"message": id, "flowInfo": "true", "addToFlow": "true"}

	_, err := flogo.EvalActivity(&log.LogActivity{}, in)
	if err != nil {
		return nil, err
	}
	m := "/Users/avanders/working/working_python/Smart_mapper/Archive.zip"
	inputName := "inputs"
	framework := "Tensorflow"

	// Generate a random number for the amount
	// There are definitely better ways to do this with Go, but this keeps the flow consistent with the UI version
	in = map[string]interface{}{"model": m, "inputName": inputName, "framework": framework, "features": features}
	// in = make(map[string]interface{})
	// in = map[string]interface{}{"min": 0, "max": 2000}
	// fmt.Println(in["model"])
	fmt.Println(in)
	out, err := flogo.EvalActivity(&inference.InferenceActivity{}, in)
	fmt.Println(err)
	if err != nil {
		return nil, err
	}
	amount := strconv.Itoa(out["result"].Value().(int))
	// amount := "100"
	// // Instead of using the combine activity we'll concat the strings together
	ref := fmt.Sprintf("INV-%v", id)

	// Generate a random number for the balance
	// There are definitely better ways to do this with Go, but this keeps the flow consistent with the UI version
	in = map[string]interface{}{"min": 0, "max": 2000}
	out, err = flogo.EvalActivity(&randomnumber.MyActivity{}, in)
	if err != nil {
		return nil, err
	}
	balance := strconv.Itoa(out["result"].Value().(int))

	//expectedDate := out["data"].Value().(map[string]interface{})["expectedDate"].(string)

	// The return message is a map[string]*data.Attribute which we'll have to construct
	response := make(map[string]interface{})
	response["id"] = id
	response["ref"] = ref
	response["amount"] = amount
	response["balance"] = balance
	//response["expectedPaymentDate"] = expectedDate
	response["currency"] = "USD"

	ret := make(map[string]*data.Attribute)
	ret["code"], _ = data.NewAttribute("code", data.TypeInteger, 200)
	ret["data"], _ = data.NewAttribute("data", data.TypeAny, response)

	return ret, nil
}
