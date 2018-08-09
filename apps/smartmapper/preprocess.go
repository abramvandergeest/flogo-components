package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var camel = regexp.MustCompile("(^[^A-Z0-9]*|[A-Z0-9]*)([A-Z0-9][^A-Z]+|$)")
var altspace = regexp.MustCompile("_|-|\\.|/")
var badchar = regexp.MustCompile("[\\[\\]()\\&\\?$]")
var regID = regexp.MustCompile("ID")

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

var gfile = "/Users/avanders/glove/glove_twitter/glove.twitter.27B.25d.txt"

const dim int = 25

var gdic = loadGloveModel(gfile)

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

func getembed(s string) []float64 {
	var a []float64
	for _, word := range string2List(s) {
		a = addVecs(a, gdic[word])
		//I might need to divide by length of string2list
	}
	return a
}

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

type objectField struct {
	name        string
	label       string
	fieldType   string
	fieldLength int32
	vec         []float64
	vecsNormed  bool
}

func (oF objectField) embedding() objectField {
	oF.vec = getembed(oF.label)
	return oF
}

func getNormedVec(vec []float64) []float64 {
	gnorm := 11.

	for i, ai := range vec {
		vec[i] = ai / gnorm
	}
	return vec
}

func (oF objectField) norming() objectField {
	// oF.vec = getembed(oF.label)
	oF.vec = getNormedVec(oF.vec)

	oF.vecsNormed = true
	return oF
}

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

// func objs2features(o1 objectField, o2 objectField) (features map[string]float64, err error) {

// 	return features, err
// }

type matchedResult struct {
	o1name string
	o2name string
	score  float64
}

func otherFunc() {

	var s string
	// var err error
	s = "helloLetsTalk$Today_and.tomorrow"

	fmt.Println(string2List(s))
	fmt.Println(string2List("Hi&How(_Are-You/_doing?"))
	fmt.Println(string2List("result.company.industry"))
	fmt.Println(string2List("result.company.industryIDBlah"))

	// fmt.Println(gdic[strings.ToLower("Hello")])

	// gnorm := 11. //This is the value to use to normalize the glove vectors
	// This is applied after taking vector diff

	t := "result.company.industryIDBlah"
	tvec := getembed(t)
	svec := getembed(s)
	mag, _ := diff(tvec, svec)
	fmt.Println(mag)

	obj1 := objectField{name: "testingNameHere", label: "labeledDescription-Hi", fieldType: "STRING", fieldLength: 128}
	obj1 = obj1.embedding()
	obj2 := objectField{name: "Phone", label: "mobilePhone", fieldType: "STRING", fieldLength: 128}
	obj2 = obj2.embedding()
	fmt.Println(obj1)
	obj1 = obj1.norming()
	fmt.Println(obj1)

	fmt.Println(objs2features(obj1, obj2))
}
