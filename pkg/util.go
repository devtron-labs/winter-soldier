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
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
)

type Operator interface {
	execute(lhs string, rhs string)
}

var operations = []string{"==", "!=", "!", ">=", "<=", "=>", "=<", ">", "<"}

func SplitByLogicalOperator(input string) ([]string, error) {
	var operands []string
	for _, operation := range operations {
		if strings.Index(input, operation) > -1 {
			operands = append(operands, operation)
			parts := strings.Split(input, operation)
			for _, part := range parts {
				if len(part) > 0 {
					operands = append(operands, part)
				}
			}
			return operands, nil
		}
	}
	operands = []string{"", input}
	return operands, nil
}

func ApplyLogicalOperator(result gjson.Result, ops []string) (bool, error) {
	switch ops[0] {
	case "==":
		return result.String() == ops[2], nil
	case "!=":
		return result.String() != ops[2], nil
	case "!":
		return !result.Exists(), nil
	case ">=", "=>", "<=", "=<", ">", "<":
		return Compare(result, ops)
	case "":
		return result.Exists(), nil
	default:
		return false, fmt.Errorf("unknown operator")
	}
	return false, fmt.Errorf("unknown operator")
}

func Compare(result gjson.Result, ops []string) (bool, error) {
	switch result.Type {
	default:
		return false, fmt.Errorf("unsupported type")
	case gjson.False, gjson.True:
		return false, fmt.Errorf("unsupported operation %s for boolean type", ops[0])
	case gjson.Number:
		// calculated result
		in, err := strconv.ParseFloat(ops[2], 64)
		if err != nil {
			return false, fmt.Errorf("error while typecasting to float %v", err)
		}
		switch ops[0] {
		case ">=", "=>":
			return result.Num >= in, nil
		case "<=", "=<":
			return result.Num <= in, nil
		case ">":
			return result.Num > in, nil
		case "<":
			return result.Num < in, nil
		}
		return false, fmt.Errorf("unsupported operation %s", ops[0])
	case gjson.String, gjson.JSON:
		switch ops[0] {
		case ">=", "=>":
			return result.Str >= ops[2], nil
		case "<=", "=<":
			return result.Str <= ops[2], nil
		case "<":
			return result.Str < ops[2], nil
		case ">":
			return result.Str > ops[2], nil
		}
		return false, fmt.Errorf("unsupported operator %s", ops[0])
	}
}
