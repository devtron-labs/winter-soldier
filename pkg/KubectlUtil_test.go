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
	"testing"
)

func TestArgsProcessor_ResourceTypeOrNameArgs(t *testing.T) {
	mapper := NewMapperFactory()
	type fields struct {
		mapper *Mapper
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "multi resource",
			fields: fields{
				mapper: mapper,
			},
			args: args{
				args: []string{"pod"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewFactory(tt.fields.mapper)
			resource, err := a.MappingFor(tt.args.args[0])
			if err != nil {
				fmt.Println("error occurred", err)
			}
			fmt.Println(resource)
			//a.ResourceTypeOrNameArgs(tt.args.args...)
			//fmt.Printf("size %d\n", len(a.resourceTuples))
			//fmt.Printf("err size %d\n", len(a.errs))
			//for _, rt := range a.ResourceTuples() {
			//	fmt.Printf("%s - %s\n", rt.Name, rt.Resource)
			//
			//}
		})
	}
}
