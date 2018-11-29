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
	"github.com/harrydb/go/img/grayscale"

	// "github.com/harrydb/go/img/grayscale"
	// "github.com/nfnt/resize"
	"github.com/disintegration/imaging"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
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

	// blah := inputs["content"].Value().(map[string]interface{})["Input"] //.(int32)
	// fmt.Println(blah)
	// trgobjs := inputs["content"].Value().(map[string]interface{})["Target"].([]interface{})

	filename := "/Users/avanderg@tibco.com/datasets/box_images/Box/boxes/google-image(0060).jpeg"
	src, err := imaging.Open(filename)

	imgsize := 256
	src = imaging.Resize(src, imgsize, imgsize, imaging.Lanczos)
	src = imaging.Grayscale(src)
	src = grayscale.Convert(src, grayscale.ToGrayLuminance)

	var flatimg []float32
	for x := 0; x < imgsize; x++ {
		for y := 0; y < imgsize; y++ {
			imageColor := src.At(x, y)
			rr, _, _, _ := imageColor.RGBA()
			gray := float32(rr) / 65535.
			flatimg = append(flatimg, gray)

		}
	}

	flatimgout, err := tf.NewTensor(flatimg)
	if err != nil {
		return nil, err
	}

	m := modelPath
	framework := "Tensorflow"
	var features []interface{}
	features = append(features, map[string]interface{}{
		"name": "xs",
		"data": flatimgout,
	})

	in := map[string]interface{}{"model": m, "framework": framework, "features": features}
	// fmt.Println(in["model"])

	out, err := flogo.EvalActivity(&inference.InferenceActivity{}, in)
	if err != nil {
		return nil, err
	}
	mapProb := out["result"].Value().(map[string]interface{})["prediction"].([][]float32)
	// fmt.Println(out, mapProb)

	// outfile, err := os.Create("Blah.jpeg")
	// if err != nil {
	// 	return nil, err
	// }
	// defer outfile.Close()
	// jpeg.Encode(outfile, src, &jpeg.Options{Quality: 100})

	// Creating output variable for catching inference results
	//   Shape of output is output["sourceName"]={"targetName", "matchProbability"}
	// output := make(map[string][]interface{})
	output := mapProb[0][0]

	// The return message is a map[string]*data.Attribute which we'll have to construct
	response := make(map[string]interface{})
	response["output"] = output

	ret := make(map[string]*data.Attribute)
	ret["code"], _ = data.NewAttribute("code", data.TypeInteger, 200)
	ret["data"], _ = data.NewAttribute("data", data.TypeAny, response)

	return ret, nil
}
