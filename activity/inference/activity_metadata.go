package inference

import (
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
)

var jsonMetadata = `{
  "name": "flogo-inference",
  "type": "flogo:activity",
  "ref": "github.com/TIBCOSoftware/flogo-contrib/activity/inference",
  "version": "0.0.1",
  "title": "Invoke ML Model",
  "description": "Basic inferencing activity to invoke ML model using the flogo-ml framework.",
  "homepage": "https://github.com/TIBCOSoftware/flogo-contrib/tree/master/activity/inference",
  "input":[
    {
        "name": "model",
        "type": "string",
        "required": true
    },
    {
        "name": "inputName",
        "type": "string",
        "required": true
    },
    {
      "name": "features",
      "type": "object",
      "required": true
    },
    {
        "name": "framework",
        "type": "string",
        "required": true
    }
  ],
  "output": [
    {
      "name": "result",
      "type": "object"
    }
  ]
}
`

// init create & register activity
func init() {
	md := activity.NewMetadata(jsonMetadata)
	activity.Register(NewActivity(md))
}
