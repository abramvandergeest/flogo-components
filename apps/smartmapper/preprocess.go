package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// defining my regex search parameters
var camel = regexp.MustCompile("(^[^A-Z0-9]*|[A-Z0-9]*)([A-Z0-9][^A-Z]+|$)")
var altspace = regexp.MustCompile("_|-|\\.|/")
var badchar = regexp.MustCompile("[\\[\\]()\\&\\?$]")
var regID = regexp.MustCompile("ID")

// Cleaning a Label into something usable
func camel2Space(s string) string {
	// https://gist.github.com/vermotr/dd9cfe74169234ef7380e8f32a8fbce9
	var a []string
	for _, sub := range camel.FindAllStringSubmatch(s, -1) {
		if sub[1] != "" {
			a = append(a, sub[1])
		}
		if sub[2] != "" {
			a = append(a, sub[2])
		}
	}
	return strings.ToLower(strings.Join(a, " "))
}

func cleanUnder(s string) string {
	return altspace.ReplaceAllString(s, " ")
}

func cleanString(s string) string {
	// var a []string
	return badchar.ReplaceAllString(cleanUnder(camel2Space(regID.ReplaceAllString(s, "Id"))), "")
}

func string2List(s string) []string {
	return strings.Fields(cleanString(s))
}

//Loading the glove vector from file to dictionary
func loadGloveModel(gfile string) map[string][]float64 {
	gdic := make(map[string][]float64)

	fmt.Println("Loading Glove Model")

	file, err := os.Open(gfile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var l []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		l = strings.Fields(scanner.Text())
		var num []float64
		for _, strfloat := range l {
			i, err := strconv.ParseFloat(strfloat, 64)
			if err == nil {
				num = append(num, i)
			}
		}
		gdic[l[0]] = num
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Done. ", len(gdic), " words loaded!")
	return gdic
}

// Defining basic values for glove file
var gfile = "/Users/avanders/glove/glove_twitter/glove.twitter.27B.25d.txt"

const dim int = 25

// Load the glove vector dictionary
var gdic = loadGloveModel(gfile)

// adding two vectors together
func addVecs(a []float64, b []float64) []float64 {
	la := len(a)
	lb := len(b)
	if la == 0 && lb > 0 {
		return b
	} else if lb == 0 && la > 0 {
		return a
	} else if lb > 0 && la > 0 && lb == la {
		for i, ai := range a {
			// fmt.Println(i)
			a[i] = ai + b[i]
		}
		return a
	}
	return nil
}

//  getting the embedded word value summed - not normed because that is done later
func getembed(s string) []float64 {
	var a []float64
	for _, word := range string2List(s) {
		a = addVecs(a, gdic[word])
		//I might need to divide by length of string2list
	}
	return a
}

//finding the difference between two vectors with edge cases for the word embeddnig values
func diff(svec []float64, tvec []float64) (mag float64, err error) {
	lt := len(tvec)
	ls := len(svec)

	if lt == 0 || ls == 0 {
		return 10., nil
	} else if lt > 0 && ls > 0 && lt != ls {
		return 10., fmt.Errorf("Cannot take difference of two vectors of different sizes")
	}

	// var mag float64 //=0
	for i, ai := range svec {
		mag = mag + (ai-tvec[i])*(ai-tvec[i])
	}
	mag = math.Sqrt(mag)

	if mag > 10. {
		mag = 10.
	}
	return mag, nil
}

// Above is functions not dependant on activity structure, below are
//Creating a structure for the source and target of a mapping
type objectField struct {
	name        string
	label       string
	fieldType   string
	fieldLength int32
	vec         []float64
	vecsNormed  bool
}

// assigned an embedding value to an objectField
func (oF objectField) embedding() objectField {
	oF.vec = getembed(oF.label)
	return oF
}

// Normalizing a vector - 11 is from the max value of the embedded vecs
func getNormedVec(vec []float64) []float64 {
	gnorm := 11.

	for i, ai := range vec {
		vec[i] = ai / gnorm
	}
	return vec
}

//  Applying gerNormedVec to an objectField's vec property
func (oF objectField) norming() objectField {
	// oF.vec = getembed(oF.label)
	oF.vec = getNormedVec(oF.vec)

	oF.vecsNormed = true
	return oF
}

// Given two objectField objects convert it into the features for the machine learning model
func objs2features(o1 objectField, o2 objectField) map[string]interface{} { // oF to feature vector
	out := make(map[string]interface{})

	// vecs should only be normed after diff is calculated
	out["VEC_DIFF"], _ = diff(o1.vec, o2.vec)
	out["map"] = 0

	if !o1.vecsNormed {
		o1 = o1.norming()
	}

	var s string
	for i, ai := range o1.vec {
		s = fmt.Sprintf("%d_x", i)
		out[s] = ai
	}

	if !o2.vecsNormed {
		o2 = o2.norming()
	}

	for i, ai := range o2.vec {
		s = fmt.Sprintf("%d_y", i)
		out[s] = ai
	}

	return out
}

type outrow struct {
	TargetName string  `json:"targetName"`
	Match      float64 `json:"match"`
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
