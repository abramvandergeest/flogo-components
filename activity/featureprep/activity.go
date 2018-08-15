package featureprep

import (
	"fmt"
	"math"
	"strconv"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
)

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

func magnitude(vec []float64) float64 {
	value := 0.
	for _, ai := range vec {
		value += ai * ai
	}
	return math.Sqrt(value)
}

// Eval implements activity.Activity.Eval
func (a *MyActivity) Eval(context activity.Context) (done bool, err error) {

	outobject := make(map[string]interface{})

	// do eval
	inputs := context.GetInput("input").([][]float64)

	li := len(inputs)
	fmt.Println(inputs)
	for i, ivec := range inputs {
		ind := li - i - 1
		for j := 0; j < len(ivec); j++ {
			outobject[strconv.Itoa(j)+"_"+strconv.Itoa(ind)] = ivec[j]
			fmt.Println(strconv.Itoa(j)+"_"+strconv.Itoa(ind), ivec[j])
		}
		outobject["amag_"+strconv.Itoa(ind)] = magnitude(ivec)
		fmt.Println("amag_"+strconv.Itoa(ind), magnitude(ivec))
	}

	// fmt.Println(outobject)
	context.SetOutput("result", outobject)

	return true, nil
}
