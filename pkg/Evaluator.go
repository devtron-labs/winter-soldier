/*
Copyright 2021 Devtron Labs Pvt Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pkg

import (
	"fmt"
	"github.com/antonmedv/expr"
	"github.com/tidwall/gjson"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const variablePattern = `\{\{[^\}\}|\{\{]*\}\}`

var variableRegex *regexp.Regexp

//var memoryConversions = map[string]float64{
//	"E":  float64(1000000000000000000),
//	"P":  float64(1000000000000000),
//	"T":  float64(1000000000000),
//	"G":  float64(1000000000),
//	"M":  float64(1000000),
//	"K":  float64(1000),
//	"Ei": float64(1152921504606846976),
//	"Pi": float64(1125899906842624),
//	"Ti": float64(1099511627776),
//	"Gi": float64(1073741824),
//	"Mi": float64(1048576),
//	"Ki": float64(10241),
//}
//
//var cpuConversions = map[string]float64{
//	"m": .0001,
//}

type resourceParser struct {
	name        string
	pattern     string
	regex       *regexp.Regexp
	conversions map[string]float64
}

//const memoryPattern = "(\\d*e?\\d*)(Ei?|Pi?|Ti?|Gi?|Mi?|Ki?|$)"
//const cpuPattern = "(\\d*e?\\d*)(m?)"

var memoryParser *resourceParser
var cpuParser *resourceParser

//TODO: handle date and memory and cpu properly
func ExpressionEvaluator(expression, json string) bool {
	if variableRegex == nil {
		variableRegex, _ = regexp.Compile(variablePattern)
	}
	matches := variableRegex.FindAllString(expression, -1)
	variablesWithValue := make(map[string]interface{}, len(matches))
	finalExpression := expression
	var expressions, variables []string
	for index, match := range matches {
		originalMatch := match
		match = strings.Replace(match, "{{", "", 1)
		match = strings.Replace(match, "}}", "", 1)
		variableName := "var" + strconv.Itoa(index)
		variables = append(variables, variableName)
		expressions = append(expressions, match)
		//fmt.Println(match)
		finalExpression = strings.ReplaceAll(finalExpression, originalMatch, variableName)
	}
	res := gjson.GetMany(json, expressions...)
	for i, re := range res {
		variablesWithValue[variables[i]] = re.Value()
	}
	//fmt.Println(expression)
	//fmt.Println(finalExpression)
	//fmt.Println(variables)
	//fmt.Println(variablesWithValue)
	variablesWithValue["CpuToNumber"] = CpuToNumber
	variablesWithValue["MemoryToNumber"] = MemoryToNumber
	variablesWithValue["ParseTime"] = ParseTime
	variablesWithValue["Now"] = func() *time.Time {
		n := time.Now()
		return &n
	}
	variablesWithValue["AfterTime"] = func(t1, t2 *time.Time) (bool, error) {
		if t1 == nil || t2 == nil {
			return false, fmt.Errorf("t1: %v or t2: %v is nil", t1, t2)
		}
		return t1.After(*t2), nil
	}
	variablesWithValue["AddTime"] = AddTime
	program, err := expr.Compile(finalExpression, expr.Env(variablesWithValue))
	if err != nil {
		fmt.Println(err)
		return false
	}
	output, err := expr.Run(program, variablesWithValue)
	if err != nil {
		fmt.Println(err)
		return false
	}
	//fmt.Println(output)
	switch v := output.(type) {
	case bool:
		return v
	default:
		return false
	}
}

func ParseTime(dateTime, format string) (*time.Time, error) {
	t, err := time.Parse(format, dateTime)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(t)
	return &t, nil
}

func AddTime(t *time.Time, period string) (*time.Time, error) {
	chars := []rune(period)
	start := 0
	if chars[0] == '-' {
		start = 1
	}
	numeral := ""
	duration := 0
	for i := start; i < len(chars); i++ {
		if chars[i] == 'd' || chars[i] == 'h' || chars[i] == 'm' || chars[i] == 's' {
			d, err := strconv.Atoi(numeral)
			if err != nil {
				return nil, err
			}
			m, err := toSeconds(chars[i])
			if err != nil {
				return nil, err
			}
			duration += d * m
			numeral = ""
		} else {
			numeral += string(chars[i])
		}
	}
	if chars[0] == '-' {
		duration *= -1
	}
	ft := t.Add(time.Duration(duration) * time.Second)
	return &ft, nil
}

func toSeconds(symbol rune) (int, error) {
	switch symbol {
	case 'd':
		return 24 * 60 * 60, nil
	case 'h':
		return 60 * 60, nil
	case 'm':
		return 60, nil
	case 's':
		return 1, nil
	default:
		return 0, fmt.Errorf("unsupported format %c", symbol)
	}
}

func MemoryToNumber(memory string) (float64, error) {
	if memoryParser == nil {
		pattern := "(\\d*e?\\d*)(Ei?|Pi?|Ti?|Gi?|Mi?|Ki?|$)"
		re, _ := regexp.Compile(pattern)
		memoryParser = &resourceParser{
			name:    "memory",
			pattern: pattern,
			regex:   re,
			conversions: map[string]float64{
				"E":  float64(1000000000000000000),
				"P":  float64(1000000000000000),
				"T":  float64(1000000000000),
				"G":  float64(1000000000),
				"M":  float64(1000000),
				"K":  float64(1000),
				"Ei": float64(1152921504606846976),
				"Pi": float64(1125899906842624),
				"Ti": float64(1099511627776),
				"Gi": float64(1073741824),
				"Mi": float64(1048576),
				"Ki": float64(10241),
			},
		}
	}
	return convertResource(memoryParser, memory)
}

func CpuToNumber(cpu string) (float64, error) {
	if cpuParser == nil {
		pattern := "(\\d*e?\\d*)(m?)"
		re, _ := regexp.Compile(pattern)
		cpuParser = &resourceParser{
			name:    "cpu",
			pattern: pattern,
			regex:   re,
			conversions: map[string]float64{
				"m": .0001,
			},
		}
	}
	return convertResource(cpuParser, cpu)
}

func convertResource(rp *resourceParser, resource string) (float64, error) {
	matches := rp.regex.FindAllStringSubmatch(resource, -1)
	if len(matches[0]) < 2 {
		fmt.Printf("expected pattern for %s should match %s, found %s\n", rp.name, rp.pattern, resource)
		return float64(0), fmt.Errorf("expected pattern for %s should match %s, found %s", rp.name, rp.pattern, resource)
	}
	num, err := ParseFloat(matches[0][1])
	if err != nil {
		fmt.Println(err)
		return float64(0), err
	}
	if len(matches[0]) == 3 && matches[0][2] != "" {
		if suffix, ok := rp.conversions[matches[0][2]]; ok {
			return num * suffix, nil
		}
	} else {
		return num, nil
	}
	fmt.Printf("expected pattern for %s should match %s, found %s\n", rp.name, rp.pattern, resource)
	return float64(0), fmt.Errorf("expected pattern for %s should match %s, found %s", rp.name, rp.pattern, resource)
}

func ParseFloat(str string) (float64, error) {
	val, err := strconv.ParseFloat(str, 64)
	if err == nil {
		return val, nil
	}

	//Some number may be seperated by comma, for example, 23,120,123, so remove the comma firstly
	str = strings.Replace(str, ",", "", -1)

	//Some number is specifed in scientific notation
	pos := strings.IndexAny(str, "eE")
	if pos < 0 {
		return strconv.ParseFloat(str, 64)
	}

	var baseVal float64
	var expVal int64

	baseStr := str[0:pos]
	baseVal, err = strconv.ParseFloat(baseStr, 64)
	if err != nil {
		return 0, err
	}

	expStr := str[(pos + 1):]
	expVal, err = strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return baseVal * math.Pow10(int(expVal)), nil
}
