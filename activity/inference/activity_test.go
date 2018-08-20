package inference

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TIBCOSoftware/flogo-contrib/action/flow/test"
	"github.com/TIBCOSoftware/flogo-contrib/activity/inference/framework/tf"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
)

var _ tf.TensorflowModel

var activityMetadata *activity.Metadata

func getActivityMetadata() *activity.Metadata {

	if activityMetadata == nil {
		jsonMetadataBytes, err := ioutil.ReadFile("activity.json")
		if err != nil {
			panic("No Json Metadata found for activity.json path")
		}

		activityMetadata = activity.NewMetadata(string(jsonMetadataBytes))
	}

	return activityMetadata
}

func TestCreate(t *testing.T) {

	act := NewActivity(getActivityMetadata())

	if act == nil {
		t.Error("Activity Not Created")
		t.Fail()
		return
	}
}

func TestEval(t *testing.T) {

	defer func() {
		if r := recover(); r != nil {
			t.Failed()
			t.Errorf("panic during execution: %v", r)
		}
	}()

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())

	//setup attrs
	// tc.SetInput("model", "/Users/avanders/working/flogo_components/flogo/samples/tensorflow/helloworld/Archive.zip")
	tc.SetInput("model", "/Users/avanders/working/working_python/Smart_mapper/Archive.zip")
	tc.SetInput("inputName", "inputs")
	tc.SetInput("framework", "Tensorflow")

	var features = make(map[string]interface{})
	// features["z-axis-q75"] = 4.140586
	// features["corr-x-z"] = 0.1381063882214782
	// features["x-axis-mean"] = 1.7554575428900194
	// features["z-axis-sd"] = 4.6888631696380765
	// features["z-axis-skew"] = -0.3619011587545954
	// features["y-axis-sd"] = -7.959084724314854
	// features["y-axis-q75"] = 16.467001
	// features["corr-z-y"] = 0.3467060369518231
	// features["x-axis-sd"] = 6.450293741961166
	// features["x-axis-skew"] = 0.09756801680727022
	// features["y-axis-mean"] = 9.389463650669393
	// features["y-axis-skew"] = -0.49036224958471764
	// features["z-axis-mean"] = 1.1226106985139188
	// features["x-axis-q25"] = -3.1463003
	// features["x-axis-q75"] = 6.3198414
	// features["y-axis-q25"] = 3.0645783
	// features["z-axis-q25"] = -1.9477097
	// features["corr-x-y"] = 0.08100326860866637

	features["VEC_DIFF"] = 8.58175
	features["LVS_RATIO"] = 0.5
	features["X_IN_Y"] = 1.
	features["Y_IN_X"] = 0.
	features["0_x"] = -0.0234164
	features["0_y"] = 0.128719
	features["1_x"] = 0.0680591
	features["1_y"] = 0.134658
	features["2_x"] = -0.216889
	features["2_y"] = -0.0712427
	features["3_x"] = 0.0742827
	features["3_y"] = -0.164237
	features["4_x"] = 0.0249424
	features["4_y"] = -0.0398755
	features["5_x"] = -0.101374
	features["5_y"] = -0.0710295
	features["6_x"] = 0.122862
	features["6_y"] = -0.00998727
	features["7_x"] = -0.0667677
	features["7_y"] = -0.175369
	features["8_x"] = 0.234663
	features["8_y"] = 0.0445727
	features["9_x"] = 0.00989094
	features["9_y"] = 0.0497989
	features["10_x"] = 0.0132991
	features["10_y"] = -0.134546
	features["11_x"] = 0.109022
	features["11_y"] = 0.113493
	features["12_x"] = -1.0418
	features["12_y"] = -0.5509
	features["13_x"] = 0.0475975
	features["13_y"] = -0.0403612
	features["14_x"] = 0.12843
	features["14_y"] = 0.121347
	features["15_x"] = 0.18349
	features["15_y"] = -0.0135302
	features["16_x"] = 0.0440373
	features["16_y"] = 0.155696
	features["17_x"] = -0.0144409
	features["17_y"] = 0.0103455
	features["18_x"] = -0.176269
	features["18_y"] = -0.0884964
	features["19_x"] = 0.0423863
	features["19_y"] = 0.00913464
	features["20_x"] = -0.0854525
	features["20_y"] = -0.00878991
	features["21_x"] = -0.241307
	features["21_y"] = -0.0514073
	features["22_x"] = -0.0553236
	features["22_y"] = 0.0178455
	features["23_x"] = -0.08883
	features["23_y"] = 0.136497
	features["24_x"] = -0.23629
	features["24_y"] = -0.200955
	features["map"] = 0

	tc.SetInput("features", features)

	done, _ := act.Eval(tc)
	if done == false {
		assert.Fail(t, "Invalid framework specified")
	}

	//check result attr
	fmt.Println(tc.GetOutput("result"))
}
