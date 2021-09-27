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
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"regexp"
	"strconv"
	"strings"
)

const VariableRegex = `\{\{[^\}\}|\{\{]*\}\}`

func ExpressionEvaluator(expression, json string) bool {
	rp, _ := regexp.Compile(VariableRegex)
	matches := rp.FindAllString(expression, -1)
	variables := make(map[string]interface{}, len(matches))
	finalExpression := expression
	for index, match := range matches {
		originalMatch := match
		match = strings.Replace(match, "{{", "", 1)
		match = strings.Replace(match, "}}", "", 1)
		variableName := "var" + strconv.Itoa(index)
		//fmt.Println(match)
		res := gjson.Get(json, match)
		switch res.Type {
		case gjson.Null:
			variables[variableName] = nil
		case gjson.False:
			variables[variableName] = false
		case gjson.True:
			variables[variableName] = true
		case gjson.String:
			variables[variableName] = res.Str
		case gjson.Number:
			variables[variableName] = res.Float()
		case gjson.JSON:
			if strings.Index(res.Raw, "[") == 0 {
				var arr []interface{}
				err := jsoniter.Unmarshal([]byte(res.String()), &arr)
				if err != nil {
					variables[variableName] = res.Raw
				} else {
					variables[variableName] = arr
				}
			} else if strings.Index(res.Raw, "{") == 0 {
				var dict map[string]interface{}
				err := jsoniter.Unmarshal([]byte(res.String()), &dict)
				if err != nil {
					variables[variableName] = res.Raw
				} else {
					variables[variableName] = dict
				}
			} else {
				variables[variableName] = res.Raw
			}
		}
		finalExpression = strings.ReplaceAll(finalExpression, originalMatch, variableName)
	}
	//fmt.Println(expression)
	fmt.Println(finalExpression)
	fmt.Println(variables)
	program, err := expr.Compile(finalExpression, expr.Env(variables))
	if err != nil {
		fmt.Println(err)
		return false
	}
	output, err := expr.Run(program, variables)
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
