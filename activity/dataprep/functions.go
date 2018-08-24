package dataprep

import (
	"errors"
	"fmt"
	"math"
)

func rename(in interface{}) (interface{}, error) {
	return in, nil
}

func mag(vec []interface{}) interface{} {
	value := 0.
	for _, ai := range vec {
		aif := ai.(float64)
		value += aif * aif
	}
	return math.Sqrt(value)
}

func magnitude(matrix [][]interface{}) ([]interface{}, error) {
	var outvec []interface{}
	for _, vec := range matrix {
		outvec = append(outvec, mag(vec))
	}
	return outvec, nil
}

func addCol2Tab(matrix [][]interface{}, vec []interface{}) ([][]interface{}, error) {
	if len(matrix) != len(vec) {
		return nil, errors.New("vector being added to matrix, not matching in length")
	}
	var outmat [][]interface{}
	for i, row := range matrix {
		fmt.Println(row, vec[i], append(row, vec[i]))
		outmat = append(outmat, append(row, vec[i]))
	}

	return outmat, nil
}

func flatten(matrix [][]interface{}) (out []interface{}, err error) {
	for _, vec := range matrix {
		for _, item := range vec {
			out = append(out, item)
		}
	}
	return out, nil
}

func toMap(vec []interface{}, list []string) (out map[string]interface{}, err error) {
	if len(vec) != len(list) {
		return nil, errors.New("vector being added to matrix, not matching in length")
	}

	out = make(map[string]interface{})
	for i, v := range vec {
		out[list[i]] = v
	}
	return out, nil
}
