//go:generate go run ../../../../TIBCOSoftware/flogo-lib/flogo/gen/gen.go $GOPATH
package main

import (
	"context"
	// "log"
	"os"
	"strconv"

	// "github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/session"
	// "github.com/aws/aws-sdk-go/service/s3"
	// "github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/TIBCOSoftware/flogo-contrib/activity/inference"
	rt "github.com/TIBCOSoftware/flogo-contrib/trigger/rest"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/engine"
	"github.com/TIBCOSoftware/flogo-lib/flogo"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/harrydb/go/img/grayscale"

	// "github.com/harrydb/go/img/grayscale"
	// "github.com/nfnt/resize"
	"github.com/disintegration/imaging"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

var log = logger.GetLogger("activity-tibco-inference")

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

	// Creating file to hold file from s3
	path := "./tmp.jpg"
	file, err := os.Create(path)
	if err != nil {
		log.Error(err)
	}

	// Read in Bucket and filename/path for target image
	bucket := inputs["content"].Value().(map[string]interface{})["Bucket"].(string)
	item := inputs["content"].Value().(map[string]interface{})["Item"].(string)

	// Create AWS session and downloader
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	downloader := s3manager.NewDownloader(sess)

	// Download s3 image to "file" and checks error
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})
	if err != nil {
		log.Errorf("Unable to download item %q, %v", item, err)
	}

	log.Info("Downloaded", file.Name(), numBytes, "bytes")

	// Opening the image
	src, err := imaging.Open(file.Name())
	if err != nil {
		log.Errorf("Unable to download item %q, %v", file.Name(), err)
	}

	// Pre-Process image
	imgsize := 256
	src = imaging.Resize(src, imgsize, imgsize, imaging.Lanczos)
	src = imaging.Grayscale(src)
	src = grayscale.Convert(src, grayscale.ToGrayLuminance)

	//Converting Image to array
	var flatimg []float32
	for x := 0; x < imgsize; x++ {
		for y := 0; y < imgsize; y++ {
			imageColor := src.At(x, y)
			rr, _, _, _ := imageColor.RGBA()
			gray := float32(rr) / 65535.
			flatimg = append(flatimg, gray)

		}
	}

	// Array to TF tensor
	flatimgout, err := tf.NewTensor(flatimg)
	if err != nil {
		return nil, err
	}

	// Setting up inputs to Inference Activity
	m := modelPath
	framework := "Tensorflow"
	var features []interface{}
	features = append(features, map[string]interface{}{
		"name": "xs",
		"data": flatimgout,
	})
	in := map[string]interface{}{"model": m, "framework": framework, "features": features}

	// Making ML prediction
	out, err := flogo.EvalActivity(&inference.InferenceActivity{}, in)
	if err != nil {
		return nil, err
	}
	mapProb := out["result"].Value().(map[string]interface{})["prediction"].([][]float32)
	output := mapProb[0][0]

	// The return message is a map[string]*data.Attribute which we'll have to construct
	//    Including output
	response := make(map[string]interface{})
	response["output"] = output

	ret := make(map[string]*data.Attribute)
	ret["code"], _ = data.NewAttribute("code", data.TypeInteger, 200)
	ret["data"], _ = data.NewAttribute("data", data.TypeAny, response)

	return ret, nil
}
